/*
package util provides utilities used by stress
*/
package util

import (
	"net/url"
	"strings"
)

// ParseStringToUrl parse comma saperated url string to []url.URL
func ParseStringToUrl(raw string) ([]url.URL, error) {
	raws := strings.Split(raw, ",")
	res := make([]url.URL, 0)
	for _, s := range raws {
		if len(s) == 0 {
			continue
		}
		u, err := url.Parse(s)
		if err != nil {
			return nil, err
		}
		res = append(res, *u)
	}
	return res, nil
}
