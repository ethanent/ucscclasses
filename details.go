package ucscclasses

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
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
	ID string

	// Title, eg. "CSE 13S - 01 Comp Sys and C Prog"
	FullTitle string

	// Full number, eg. "CSE 13S - 01".
	// Note that this is not to be confused with Number or ID. It just differentiates the specific class from other
	// classes of the same subject and number.
	FullNumber string

	// Name, eg. "Computer Systems and C Programming"
	Name string

	Subject string
	Number  string

	Status                 ClassStatus
	Capacity               int
	Enrolled               int
	WaitlistTotal          int
	WaitlistCapacity       int
	Career                 string
	Description            string
	EnrollmentRequirements string
	ClassNotes             string
	Units                  int

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

var classFullTitleRegex = regexp.MustCompile(`^([A-Za-z]+) ([0-9][A-Za-z0-9]*) - ([0-9]+) (.+)$`)

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
	details.FullTitle = cleanString(doc.Find("h2").Eq(0).Text())

	sm := classFullTitleRegex.FindStringSubmatch(details.FullTitle)

	if len(sm) > 4 {
		details.Subject = sm[1]
		details.Number = sm[2]
		details.FullNumber = details.Subject + " " + details.Number + " - " + sm[3]
		details.Name = sm[4]
	}

	var ok bool
	details.Status, ok = statusStrStatusMap[cleanString(dds.Eq(7).Text())]

	if !ok {
		return nil, fmt.Errorf("while parsing detail status: unexpected status")
	}

	details.Capacity, err = strconv.Atoi(dds.Eq(9).Text())

	if err != nil {
		return nil, fmt.Errorf("while parsing detail capacity: %w", err)
	}

	details.Enrolled, err = strconv.Atoi(dds.Eq(10).Text())

	if err != nil {
		return nil, fmt.Errorf("while parsing detail enrolled: %w", err)
	}

	details.WaitlistTotal, err = strconv.Atoi(dds.Eq(12).Text())

	if err != nil {
		return nil, fmt.Errorf("while parsing detail waitlist total: %w", err)
	}

	details.WaitlistCapacity, err = strconv.Atoi(dds.Eq(11).Text())

	if err != nil {
		return nil, fmt.Errorf("while parsing detail waitlist capacity: %w", err)
	}

	pbs := doc.Find(".panel > .panel-body")

	details.Career = dds.Eq(0).Text()
	details.Description = cleanString(pbs.Eq(2).Text())
	details.EnrollmentRequirements = cleanString(pbs.Eq(3).Text())
	details.ClassNotes = cleanString(pbs.Eq(4).Text())

	if strings.Index(details.EnrollmentRequirements, "Days Times Room Instructor") == 0 {
		details.EnrollmentRequirements = ""
	}

	if strings.Index(details.ClassNotes, "Days Times Room Instructor") == 0 {
		details.ClassNotes = ""
	}

	unitsStr := dds.Eq(5).Text()
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
