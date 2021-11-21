package ucscclasses

import (
	"golang.org/x/net/html"
	"io"
	"net/http"
	"strings"
)

var fixedDataTopics = map[string]string{
	"Term":    "terms",
	"Subject": "subjects",
	"Geneds":  "ges",
}

const END_AT_TOPIC = "Course Units"

type Option struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	// Default bool   `json:"default"`
}

// GetFixedData returns the current terms, subjects, and GE categories.
func GetFixedData(c *http.Client) (terms []*Option, subjects []*Option, ges []*Option, err error) {
	resp, err := c.Get("https://pisa.ucsc.edu/class_search/index.php")

	if err != nil {
		return nil, nil, nil, err
	}

	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)
	nextText := false

	fixedData := map[string][]*Option{}

	for _, tc := range fixedDataTopics {
		fixedData[tc] = []*Option{}
	}

	fixedData["subjects"] = append(fixedData["subjects"], &Option{
		Name:  "All",
		Value: "",
	})

	curTopic := ""
	curOption := &Option{}

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

			return nil, nil, nil, e
		case html.CommentToken:
			ctxt := strings.Trim(string(z.Text()), " ")

			fdtName, ok := fixedDataTopics[ctxt]

			if !ok {
				if ctxt == END_AT_TOPIC {
					exitTokenLoop = true
					break
				}

				// Set topic to none if it's not one we're looking for. (Ignore the comment containing "align")
				if strings.Index(ctxt, "align") == -1 {
					curTopic = ""
				}
			} else {
				curTopic = fdtName
			}
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

					curOption.Value = attrs["value"]

					// Probably not necessary to keep track / order selected first because it seems that is already done by server for terms.
					// curOptionSelected = attrs["selected"]
					nextText = true
				}
			}
		case html.TextToken:
			if nextText && curTopic != "" {
				curOption.Name = string(z.Text())

				if curOption.Value != "begins" && curOption.Value != "" && curOption.Value != "\n" {
					fixedData[curTopic] = append(fixedData[curTopic], curOption)
				}

				curOption = &Option{}

				nextText = false
			}
		}

		if exitTokenLoop {
			break
		}
	}

	return fixedData["terms"], fixedData["subjects"], fixedData["ges"], nil
}
