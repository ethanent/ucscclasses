package ucscclasses

import (
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"strconv"
)

type ClassStatus string

const (
	ClassStatusOpen           ClassStatus = "Open"
	ClassStatusClosedWaitlist ClassStatus = "Waitlist"
	ClassStatusClosed         ClassStatus = "Closed"
)

var statusStrStatusMap = map[string]ClassStatus{
	"Open":      ClassStatusOpen,
	"Wait List": ClassStatusClosedWaitlist,
	"Closed":    ClassStatusClosed,
}

type ClassBriefInfo struct {
	// AKA ClassNumber (not to be confused with number for search)
	ID         string
	DetailsURL string
	Name       string
	Location   string
	TimeDay    string
	Instructor string
	Status     ClassStatus
	Enrolled   int
	Capacity   int
}

type SearchMethod string

const (
	SearchMethodEqual    SearchMethod = "="
	SearchMethodContains SearchMethod = "contains"
)

type SearchOptions struct {
	// Term is required
	Term               string

	// Subject is an optional selector. Use "" to ignore.
	Subject            string

	// Number is an optional selector. Use "" to ignore.
	Number             string
	NumberSearchMethod SearchMethod

	// GE is an optional selector. Use "" to ignore.
	GE                 string

	// Title is an optional selector. Use "" to ignore.
	Title              string
}

func SearchClasses(c *http.Client, opt *SearchOptions) ([]*ClassBriefInfo, error) {
	fData := map[string]string{
		"action":                   "results",
		"binds[:term]":             opt.Term,
		"binds[:reg_status]":       "all",
		"binds[:subject]":          opt.Subject,
		"binds[:catalog_nbr_op]":   string(opt.NumberSearchMethod),
		"binds[:catalog_nbr]":      opt.Number,
		"binds[:title]":            opt.Title,
		"binds[:instr_name_op]":    "=",
		"binds[:instructor]":       "",
		"binds[:ge]":               opt.GE,
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

		var ok bool
		cbi.Status, ok = statusStrStatusMap[statusStr]

		if !ok {
			return
		}

		title := s.Find("h2 > a")

		cbi.Name = cleanString(title.Text())

		enrolledText := s.Find("div.row").Children().Eq(3).Text()

		enrolledNums := numberRegex.FindAllString(enrolledText, 2)

		if enrolledNums == nil || len(enrolledNums) < 2 {
			return
		}

		cbi.Enrolled, err = strconv.Atoi(enrolledNums[0])

		if err != nil {
			return
		}

		cbi.Capacity, err = strconv.Atoi(enrolledNums[1])

		if err != nil {
			return
		}

		cnbrLink := s.Find("div > a")

		cbi.DetailsURL, ok = cnbrLink.Attr("href")

		if !ok {
			return
		}

		cbi.Instructor = stringRemovePrefix(s.Find("div.col-xs-6").Eq(1).Text())
		cbi.Location = stringRemovePrefix(s.Find(".col-xs-12 > .col-xs-6").Eq(0).Text())
		cbi.TimeDay = stringRemovePrefix(s.Find(".col-xs-12 > .col-xs-6").Eq(1).Text())

		cbis = append(cbis, cbi)
	})

	return cbis, nil
}
