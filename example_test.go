package ucscclasses_test

import (
	"fmt"
	"github.com/ethanent/ucscclasses"
	"net/http"
	"time"
)

var c = &http.Client{Timeout: time.Second * 5}

func ExampleSearchClasses() {
	// First, get current term

	terms, _, _, err := ucscclasses.GetFixedData(c)

	if err != nil {
		panic(err)
	}

	currentTerm := terms[0].Value

	// Now, search for the class CSE 13S

	cbis, err := ucscclasses.SearchClasses(c, &ucscclasses.SearchOptions{
		Term:               currentTerm,
		Subject:            "CSE",
		Number:             "13S",
		NumberSearchMethod: ucscclasses.SearchMethodEqual,
	})

	if err != nil {
		panic(err)
	}

	// Let's iterate the CBIs to print results!

	fmt.Println(len(cbis), "results:")

	for _, cc := range cbis {
		fmt.Println("===== (" + cc.ID + ") " + cc.FullNumber + ": " + cc.Name + " =====")
		fmt.Println("This is a", cc.Subject, cc.Number, "class!")
		fmt.Println("(", cc.Enrolled, "/", cc.Capacity, ")", cc.Instructor)
		fmt.Println(cc.Location, "|", cc.TimeDay, "OPEN?", cc.Status == ucscclasses.ClassStatusOpen, cc.DetailsURL)
	}

	// You can use the DetailsURL of a CBI to retrieve class details using the Details function
}

func ExampleGetClassDetails() {
	// You can get a DetailsURL from a search result or elsewhere
	exampleDetailsURL := "https://pisa.ucsc.edu/class_search/index.php?action=detail&class_data=YToyOntzOjU6IjpTVFJNIjtzOjQ6IjIyMjAiO3M6MTA6IjpDTEFTU19OQlIiO3M6NToiNDQ1NzkiO30%253D"

	details, err := ucscclasses.GetClassDetails(c, exampleDetailsURL)

	if err != nil {
		panic(err)
	}

	fmt.Println(details.FullTitle)
	fmt.Println(details.Instructor)
	fmt.Println("Desc:", details.Description)
	fmt.Println(details.Enrolled, "/", details.Capacity)

	for _, ds := range details.DiscussionSections {
		fmt.Println(ds.Name, ds.Location, "(", ds.Enrolled, "/", ds.Capacity, ")")
	}
}
