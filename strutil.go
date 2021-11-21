package main

import (
	"regexp"
	"strconv"
	"strings"
)

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

func stringRemovePrefix(s string) string {
	return stringSubmatchNoError(s, prefixRegex)
}

func cleanString(s string) string {
	return strings.Trim(s, "\n \t")
}

func extractStringNumbers(s string, expected int) []int {
	sm := numberRegex.FindAllString(s, expected)

	if sm == nil || len(sm) < 1 {
		return make([]int, expected)
	}

	var values []int

	for _, sv := range sm {
		v, err := strconv.Atoi(sv)

		if err != nil {
			v = 0
		}

		values = append(values, v)
	}

	if len(values) < expected {
		values = append(values, make([]int, expected-len(values))...)
	}

	return values
}
