package rundeck

import (
	"regexp"
	"strings"
)

var (
	re1 = regexp.MustCompile(`(?i)[^-_a-z0-9 ]`)
	re2 = regexp.MustCompile(`[ _-]+`)
	re3 = regexp.MustCompile(`^-`)
)

func normalize(s string) string {
	s = re1.ReplaceAllString(s, "")
	s = re2.ReplaceAllString(s, "-")
	s = re3.ReplaceAllString(s, "")

	return strings.ToLower(s)
}
