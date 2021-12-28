package ucscclasses

import (
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type ClassStatus string

const (
	ClassStatusOpen           ClassStatus = "Open"
	ClassStatusClosedWaitlist ClassStatus = "Waitlist"
	ClassStatusClosed         ClassStatus = "Closed"
)

type RegistrationStatus string

const (
	// RegistrationStatusAll includes open and closed courses
	RegistrationStatusAll RegistrationStatus = "all"

	// RegistrationStatusOpen only includes open courses
	RegistrationStatusOpen RegistrationStatus = "O"
)

var statusStrStatusMap = map[string]ClassStatus{
	"Open":      ClassStatusOpen,
	"Wait List": ClassStatusClosedWaitlist,
	"Closed":    ClassStatusClosed,
}

type ClassBriefInfo struct {
	// AKA ClassNumber (not to be confused with number for search)
	ID string

	DetailsURL string

	// Title, eg. "CSE 13S - 01 Comp Sys and C Prog"
	FullTitle string

	// Full number, eg. "CSE 13S - 01".
	// Note that this is not to be confused with Number or ID. It just differentiates the specific class from other
	// classes of the same subject and number.
	FullNumber string

	// Name, eg. "Comp Sys and C Prog"
	Name string

	Subject    string
	Number     string
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
	Term string

	// Subject is an optional selector. Use "" to ignore.
	Subject string

	// Number is an optional selector. Use "" to ignore.
	Number             string
	NumberSearchMethod SearchMethod

	// RegistrationStatus is the registrability of a course
	// Leave as nil to use RegistrationStatusAll
	RegistrationStatus *RegistrationStatus

	// GE is an optional selector. Use "" to ignore.
	GE string

	// Title (title keyword) is an optional selector. Use "" to ignore.
	Title string
}

func SearchClasses(c *http.Client, opt *SearchOptions) ([]*ClassBriefInfo, error) {
	useRegStatus := RegistrationStatusAll

	if opt.RegistrationStatus != nil {
		useRegStatus = *opt.RegistrationStatus
	}

	fData := map[string]string{
		"action":                   "results",
		"binds[:term]":             opt.Term,
		"binds[:reg_status]":       string(useRegStatus),
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

		classDetailInputs := s.Find(`form input`)

		if classDetailInputs.Length() < 22 {
			return
		}

		cbi.ID = classDetailInputs.Eq(2).AttrOr("value", "")

		statusStr := s.Find("span.sr-only").Text()

		var ok bool
		cbi.Status, ok = statusStrStatusMap[statusStr]

		if !ok {
			return
		}

		title := s.Find("h2 > a")

		cbi.FullTitle = title.Text()
		titleSplit := strings.Split(cbi.FullTitle, "\u00A0\u00A0\u00A0")

		if len(titleSplit) > 1 {
			cbi.FullNumber = titleSplit[0]
			cbi.Name = titleSplit[1]
		}

		cbi.FullTitle = cleanString(cbi.FullTitle)

		cbi.Subject, _ = classDetailInputs.Eq(15).Attr("value")
		cbi.Number, _ = classDetailInputs.Eq(4).Attr("value")

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

		cnbrLink := s.Find("div a").Eq(1)

		cbi.DetailsURL, ok = cnbrLink.Attr("href")

		if !ok {
			return
		}

		cbi.Instructor = stringRemovePrefix(s.Find("div.col-xs-6").Eq(1).Text())
		cbi.Location = stringRemovePrefix(s.Find(".col-xs-12 > .col-xs-6").Eq(0).Text())
		cbi.TimeDay = cleanString(stringRemovePrefix(s.Find(".col-xs-12 > .col-xs-6").Eq(1).Text()))

		cbis = append(cbis, cbi)
	})

	return cbis, nil
}
