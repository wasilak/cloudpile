package libs

import (
	"regexp"
	"strings"

	"github.com/newrelic/go-agent/v3/newrelic"
)

var NewRelicApp *newrelic.Application

// Deduplicate returns a new slice with duplicates values removed.
func Deduplicate(s []string) []string {
	if len(s) == 0 {
		return s
	}

	result := []string{}
	seen := make(map[string]struct{})
	for _, val := range s {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = struct{}{}
		}
	}
	return result
}

func RemoveEmptyStrings(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func getTagsFromString(tags string) map[string]string {
	result := make(map[string]string)
	regexpExp := regexp.MustCompile(`^(?P<name>.+?)\=(?P<value>.+)$`)
	match := regexpExp.FindStringSubmatch(tags)

	for i, name := range regexpExp.SubexpNames() {
		if i > 0 && i <= len(match) {
			result[strings.TrimSpace(name)] = strings.TrimSpace(match[i])
		}
	}

	return result
}
