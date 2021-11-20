package main

import (
	"golang.org/x/net/html"
	"io"
	"net/http"
)

type Term struct {
	Name     string `json:"name"`
	TermCode string `json:"termCode"`
	selected bool
}

// GetTerms returns the currently available terms, ensuring the default is first
func GetTerms(c *http.Client) ([]*Term, error) {
	resp, err := c.Get("https://pisa.ucsc.edu/class_search/index.php")

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)
	nextText := false
	var terms []*Term
	var selectedTerm *Term
	curTerm := &Term{}

	for {
		t := z.Next()
		exitTokenLoop := false

		switch t {
		case html.ErrorToken:
			e := z.Err()
			if e == io.EOF {
				exitTokenLoop = true
				break
			}

			return nil, e
		case html.StartTagToken:
			tn, hasAttr := z.TagName()

			if string(tn) == "option" {
				if hasAttr {
					attrs := map[string]string{}

					for {
						tk, tv, more := z.TagAttr()

						attrs[string(tk)] = string(tv)

						if !more {
							break
						}
					}

					curTerm.TermCode, _ = attrs["value"]
					_, curTerm.selected = attrs["selected"]
					nextText = true
				}
			}
		case html.TextToken:
			if nextText {
				curTerm.Name = string(z.Text())

				if curTerm.selected {
					selectedTerm = curTerm
				} else {
					terms = append(terms, curTerm)
				}

				curTerm = &Term{}
				nextText = false
			}
		case html.EndTagToken:
			tn, _ := z.TagName()

			if len(terms) > 0 && string(tn) == "div" {
				exitTokenLoop = true
				break
			}
		}

		if exitTokenLoop {
			break
		}
	}

	// Place selected term first in terms
	terms = append([]*Term{selectedTerm}, terms...)

	return terms, nil
}
