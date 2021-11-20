package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	t, err := GetTerms(http.DefaultClient)

	if err != nil {
		panic(err)
	}

	v, _ := json.Marshal(t)

	fmt.Println(string(v))
}
