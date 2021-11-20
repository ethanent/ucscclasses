package main

import (
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

type ClassStatus string

const (
	ClassStatusOpen           ClassStatus = "Open"
	ClassStatusClosedWaitlist ClassStatus = "Waitlist"
	ClassStatusClosed         ClassStatus = "Closed"
)

type ClassBriefInfo struct {
	ID         string
	DetailsURL string
	Name       string
	Location   string
	TimeDay    string
	Instructor string
	Status     ClassStatus
	Enrolled   int
	Spaces     int
}

var numberRegex = regexp.MustCompile(`[0-9]+`)
var multispaceRegex = regexp.MustCompile(`(?:\s){2,}`)
var prefixRegex = regexp.MustCompile(`[A-Za-z0-9]: (.+)`)

func stringSubmatchNoError(s string, r *regexp.Regexp) string {
	sm := r.FindStringSubmatch(s)

	if sm == nil || len(sm) < 2 {
		return ""
	} else {
		return sm[1]
	}
}

func SearchClasses(c *http.Client, term string, subject string, number string) ([]*ClassBriefInfo, error) {
	fData := map[string]string{
		"action":                   "results",
		"binds[:term]":             term,
		"binds[:reg_status]":       "all",
		"binds[:subject]":          subject,
		"binds[:catalog_nbr_op]":   "=",
		"binds[:catalog_nbr]":      number,
		"binds[:title]":            "",
		"binds[:instr_name_op]":    "=",
		"binds[:instructor]":       "",
		"binds[:ge]":               "",
		"binds[:crse_units_op]":    "=",
		"binds[:crse_units_from]":  "",
		"binds[:crse_units_to]":    "",
		"binds[:crse_units_exact]": "",
		"binds[:days]":             "",
		"binds[:times]":            "",
		"binds[:acad_career]":      "",
	}

	fDataValues := url.Values{}

	for k, v := range fData {
		fDataValues[k] = []string{v}
	}

	resp, err := c.PostForm("https://pisa.ucsc.edu/class_search/index.php", fDataValues)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		return nil, err
	}

	var cbis []*ClassBriefInfo

	doc.Find("div.panel-default").Each(func(i int, s *goquery.Selection) {
		cbi := &ClassBriefInfo{}

		classNumberFR := s.Find(`form > input`)

		if classNumberFR.Length() < 22 {
			return
		}

		classNumberAttrs := classNumberFR.Get(2).Attr

		if len(classNumberAttrs) < 3 {
			return
		}

		cbi.ID = classNumberAttrs[2].Val

		statusStr := s.Find("span.sr-only").Text()

		cbi.Status, _ = map[string]ClassStatus{
			"Open":      ClassStatusOpen,
			"Wait List": ClassStatusClosedWaitlist,
			"Closed":    ClassStatusClosed,
		}[statusStr]

		title := s.Find("h2 > a")

		cbi.Name = multispaceRegex.ReplaceAllString(title.Text(), " ")

		enrolledText := s.Find("div.row").Children().Eq(3).Text()

		enrolledNums := numberRegex.FindAllString(enrolledText, 2)

		if enrolledNums == nil || len(enrolledNums) < 2 {
			return
		}

		cbi.Enrolled, err = strconv.Atoi(enrolledNums[0])

		if err != nil {
			return
		}

		cbi.Spaces, err = strconv.Atoi(enrolledNums[1])

		if err != nil {
			return
		}

		cnbrLink := s.Find("div > a")

		var ok bool
		cbi.DetailsURL, ok = cnbrLink.Attr("href")

		if !ok {
			return
		}

		cbi.Instructor = stringSubmatchNoError(s.Find("div.col-xs-6").Eq(1).Text(), prefixRegex)
		cbi.Location = stringSubmatchNoError(s.Find(".col-xs-12 > .col-xs-6").Eq(0).Text(), prefixRegex)
		cbi.TimeDay = stringSubmatchNoError(s.Find(".col-xs-12 > .col-xs-6").Eq(1).Text(), prefixRegex)

		cbis = append(cbis, cbi)
	})

	return cbis, nil
}
