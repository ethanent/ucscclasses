package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	terms, subjects, ges, err := GetFixedData(http.DefaultClient)

	if err != nil {
		panic(err)
	}

	fmt.Println(terms[0])
	fmt.Println(subjects[0])
	fmt.Println(ges[0])

	sres, err := SearchClasses(http.DefaultClient, terms[0].Value, "CSE", "20")

	if err != nil {
		panic(err)
	}

	c, err := GetClassDetails(http.DefaultClient, sres[0].DetailsURL)

	if err != nil {
		panic(err)
	}

	v, _ := json.Marshal(c)

	fmt.Println(string(v))
}
