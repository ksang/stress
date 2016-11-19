package util

import (
	"testing"
)

func TestParseStringToUrl(t *testing.T) {
	var tests = []struct {
		s   string
		err bool
	}{
		{
			"http://127.0.0.1,",
			false,
		},
		{
			"http://127.0.0.1,http://0.0.0.0:80/",
			false,
		},
		{
			"",
			false,
		},
	}

	for caseid, c := range tests {
		res, err := ParseStringToUrl(c.s)
		if err != nil {
			if c.err {
				t.Logf("case #%d, err: %v", caseid+1, err)
			} else {
				t.Errorf("case #%d, err: %v", caseid+1, err)
			}
		}
		if c.err {
			t.Errorf("case #%d, no error returned, expecting error", caseid+1)
		}
		t.Logf("Result: %v", res)
	}
}
