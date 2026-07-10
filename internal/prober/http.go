package prober

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Status classifies the HTTP probe result for a bucket.
type Status int

const (
	StatusUnknown  Status = iota
	StatusPublic          // 200 — exists and allows anonymous listing
	StatusPrivate         // 403 — exists but access denied
	StatusNotFound        // 404 with NoSuchBucket body
	StatusAmbiguous       // 404 with unexpected body
)

func (s Status) String() string {
	switch s {
	case StatusPublic:
		return "PUBLIC"
	case StatusPrivate:
		return "PRIVATE"
	case StatusNotFound:
		return "NOT_FOUND"
	case StatusAmbiguous:
		return "AMBIGUOUS"
	default:
		return "UNKNOWN"
	}
}

// HTTPResult is the outcome of probing a single bucket over HTTP.
type HTTPResult struct {
	Bucket   string
	URL      string
	Status   Status
	Region   string
	Objects  []string // first 10 keys when PUBLIC
	ObjCount int      // total key count from XML
	Error    error
}

// listBucketResult is the top-level XML element returned for a public listing.
type listBucketResult struct {
	XMLName     xml.Name  `xml:"ListBucketResult"`
	KeyCount    int       `xml:"KeyCount"`
	Contents    []content `xml:"Contents"`
	IsTruncated bool      `xml:"IsTruncated"`
}

type content struct {
	Key  string `xml:"Key"`
	Size int64  `xml:"Size"`
}

var probeRegions = []string{"", "us-east-1", "us-west-2", "eu-west-1"}

// ProbeHTTP tries virtual-hosted and path-style URLs for bucket, stopping at
// the first definitive answer (PUBLIC or PRIVATE). NOT_FOUND / AMBIGUOUS
// results are collected and the least-ambiguous one is returned.
func ProbeHTTP(ctx context.Context, bucket string, client *http.Client) HTTPResult {
	best := HTTPResult{Bucket: bucket, Status: StatusUnknown}

	candidates := buildURLs(bucket)
	for _, u := range candidates {
		r := probe(ctx, bucket, u, client)
		switch r.Status {
		case StatusPublic, StatusPrivate:
			return r
		case StatusNotFound:
			if best.Status == StatusUnknown || best.Status == StatusAmbiguous {
				best = r
			}
		case StatusAmbiguous:
			if best.Status == StatusUnknown {
				best = r
			}
		}
	}
	return best
}

func buildURLs(bucket string) []string {
	urls := []string{
		fmt.Sprintf("https://%s.s3.amazonaws.com/", bucket),
	}
	for _, region := range probeRegions[1:] {
		urls = append(urls, fmt.Sprintf("https://%s.s3.%s.amazonaws.com/", bucket, region))
	}
	urls = append(urls, fmt.Sprintf("https://s3.amazonaws.com/%s/", bucket))
	return urls
}

func probe(ctx context.Context, bucket, u string, client *http.Client) HTTPResult {
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx2, http.MethodGet, u, nil)
	if err != nil {
		return HTTPResult{Bucket: bucket, URL: u, Status: StatusUnknown, Error: err}
	}

	resp, err := client.Do(req)
	if err != nil {
		return HTTPResult{Bucket: bucket, URL: u, Status: StatusUnknown, Error: err}
	}
	defer resp.Body.Close()

	region := resp.Header.Get("x-amz-bucket-region")
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MiB max

	res := HTTPResult{Bucket: bucket, URL: u, Region: region}

	switch resp.StatusCode {
	case http.StatusOK:
		res.Status = StatusPublic
		parseListResult(body, &res)
	case http.StatusForbidden:
		res.Status = StatusPrivate
	case http.StatusNotFound:
		if strings.Contains(string(body), "NoSuchBucket") {
			res.Status = StatusNotFound
		} else {
			res.Status = StatusAmbiguous
		}
	default:
		res.Status = StatusAmbiguous
	}
	return res
}

func parseListResult(body []byte, res *HTTPResult) {
	var lr listBucketResult
	if err := xml.Unmarshal(body, &lr); err != nil {
		return
	}
	res.ObjCount = len(lr.Contents)
	for i, c := range lr.Contents {
		if i >= 10 {
			break
		}
		res.Objects = append(res.Objects, c.Key)
	}
}

// ClassifyStatus maps an HTTP status code + body to a Status.
// Exported for use in tests without requiring a live HTTP call.
func ClassifyStatus(code int, body string) Status {
	switch code {
	case http.StatusOK:
		return StatusPublic
	case http.StatusForbidden:
		return StatusPrivate
	case http.StatusNotFound:
		if strings.Contains(body, "NoSuchBucket") {
			return StatusNotFound
		}
		return StatusAmbiguous
	default:
		return StatusAmbiguous
	}
}
