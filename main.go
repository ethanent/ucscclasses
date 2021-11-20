package main

import (
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

}
