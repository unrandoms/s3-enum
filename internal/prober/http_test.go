package prober

import (
	"testing"
)

func TestClassifyStatus(t *testing.T) {
	cases := []struct {
		code int
		body string
		want Status
	}{
		{200, "", StatusPublic},
		{403, "", StatusPrivate},
		{404, "<Error><Code>NoSuchBucket</Code></Error>", StatusNotFound},
		{404, "<html>Not Found</html>", StatusAmbiguous},
		{301, "", StatusAmbiguous},
		{500, "", StatusAmbiguous},
	}
	for _, c := range cases {
		got := ClassifyStatus(c.code, c.body)
		if got != c.want {
			t.Errorf("ClassifyStatus(%d, %q) = %v, want %v", c.code, c.body, got, c.want)
		}
	}
}

func TestStatusString(t *testing.T) {
	cases := []struct {
		s    Status
		want string
	}{
		{StatusPublic, "PUBLIC"},
		{StatusPrivate, "PRIVATE"},
		{StatusNotFound, "NOT_FOUND"},
		{StatusAmbiguous, "AMBIGUOUS"},
		{StatusUnknown, "UNKNOWN"},
	}
	for _, c := range cases {
		if got := c.s.String(); got != c.want {
			t.Errorf("Status(%d).String() = %q, want %q", c.s, got, c.want)
		}
	}
}

func TestBuildURLs(t *testing.T) {
	urls := buildURLs("mybucket")
	if len(urls) == 0 {
		t.Fatal("buildURLs returned empty list")
	}
	for _, u := range urls {
		if len(u) == 0 {
			t.Error("empty URL in buildURLs result")
		}
	}
	// virtual-hosted style must be first
	if got := urls[0]; got != "https://mybucket.s3.amazonaws.com/" {
		t.Errorf("expected virtual-hosted URL first, got %q", got)
	}
}
