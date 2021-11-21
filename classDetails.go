package main

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"regexp"
	"strconv"
)

type DiscussionSection struct {
	ID               string
	Name             string
	Status           ClassStatus
	Location         string
	Instructor       string
	Capacity         int
	Enrolled         int
	TimeDay          string
	WaitlistTotal    int
	WaitlistCapacity int
}

type ClassDetails struct {
	// AKA ClassNumber (not to be confused with number for search.
	ID               string
	Name             string
	Status           ClassStatus
	Capacity         int
	Enrolled         int
	WaitlistTotal    int
	WaitlistCapacity int
	Career           string
	Description      string
	ClassNotes       string
	Units            int

	// Observed types have been "Lecture" and "Seminar"
	Type string

	GE                 string
	Location           string
	TimeDay            string
	Instructor         string
	MeetingDates       string
	DiscussionSections []*DiscussionSection
}

var unitsRegex = regexp.MustCompile(`([0-9]+) units`)
var discussionSectionIDNameRegex = regexp.MustCompile(`#([0-9]+) (.+)`)

func GetClassDetails(c *http.Client, detailsURL string) (*ClassDetails, error) {
	resp, err := c.Get(detailsURL)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		return nil, err
	}

	details := &ClassDetails{
		DiscussionSections: []*DiscussionSection{},
	}

	dds := doc.Find("dd")

	details.ID = dds.Eq(2).Text()
	details.Name = cleanString(doc.Find("h2").Eq(0).Text())

	var ok bool
	details.Status, ok = statusStrStatusMap[cleanString(dds.Eq(6).Text())]

	if !ok {
		return nil, errors.New("unexpected class status")
	}

	details.Capacity, err = strconv.Atoi(dds.Eq(8).Text())

	if err != nil {
		return nil, err
	}

	details.Enrolled, err = strconv.Atoi(dds.Eq(9).Text())

	if err != nil {
		return nil, err
	}

	details.WaitlistTotal, err = strconv.Atoi(dds.Eq(11).Text())

	if err != nil {
		return nil, err
	}

	details.WaitlistCapacity, err = strconv.Atoi(dds.Eq(10).Text())

	if err != nil {
		return nil, err
	}

	pbs := doc.Find(".panel > .panel-body")

	details.Career = dds.Eq(0).Text()
	details.Description = cleanString(pbs.Eq(2).Text())
	details.ClassNotes = cleanString(pbs.Eq(3).Text())

	unitsStr := dds.Eq(4).Text()
	unitsRegexRes := unitsRegex.FindStringSubmatch(unitsStr)

	if unitsRegexRes == nil || len(unitsRegexRes) < 2 {
		return nil, errors.New("failed to parse class units")
	}

	details.Units, err = strconv.Atoi(unitsRegexRes[1])

	if err != nil {
		return nil, err
	}

	tds := doc.Find("td")

	details.Type = dds.Eq(3).Text()
	details.GE = cleanString(dds.Eq(5).Text())
	details.Location = cleanString(tds.Eq(1).Text())
	details.TimeDay = cleanString(tds.Eq(0).Text())
	details.Instructor = cleanString(tds.Eq(2).Text())
	details.MeetingDates = cleanString(tds.Eq(3).Text())

	rss := doc.Find(".row-striped")

	rss.Each(func(i int, sel *goquery.Selection) {
		curDS := &DiscussionSection{}

		divs := sel.Find("div")

		idnrResult := discussionSectionIDNameRegex.FindStringSubmatch(divs.Eq(0).Text())

		if idnrResult == nil || len(idnrResult) < 3 {
			return
		}

		curDS.ID = idnrResult[1]
		curDS.Name = idnrResult[2]

		curDS.Status, ok = statusStrStatusMap[cleanString(divs.Eq(6).Text())]

		if !ok {
			return
		}

		curDS.Location = stringSubmatchNoError(divs.Eq(3).Text(), prefixRegex)
		curDS.Instructor = cleanString(divs.Eq(2).Text())

		enrollmentNums := extractStringNumbers(divs.Eq(4).Text(), 2)
		waitlistNums := extractStringNumbers(divs.Eq(5).Text(), 2)

		curDS.Enrolled = enrollmentNums[0]
		curDS.Capacity = enrollmentNums[1]

		curDS.WaitlistTotal = waitlistNums[0]
		curDS.WaitlistCapacity = waitlistNums[1]

		curDS.TimeDay = cleanString(divs.Eq(1).Text())

		details.DiscussionSections = append(details.DiscussionSections, curDS)
	})

	return details, nil
}
