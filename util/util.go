/*
package util provides utilities used by stress
*/
package util

import (
	"net/url"
	"strings"
)

const (
	ConnNumberKey    = "stress/ConnectionNumber"
	ReceivedBytesKey = "stress/ReceivedBytes"
	RequestCountKey  = "stress/RequestCount"
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

// ParseUrlsToStrings parse []url.URL to []string, useful for get endpoints
func ParseUrlsToStrings(urls []url.URL) []string {
	ret := make([]string, 0)
	for _, u := range urls {
		s := convertLocalhost(u.String())
		ret = append(ret, s)
	}
	return ret
}

// to ensure localhost is translated to 127.0.0.1
func convertLocalhost(raw string) string {
	return strings.Replace(raw, "localhost", "127.0.0.1", 1)
}
