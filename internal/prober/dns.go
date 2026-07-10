package prober

import (
	"context"
	"net"
	"time"
)

// DNSResult holds the outcome of a DNS probe for one bucket name.
type DNSResult struct {
	Bucket  string
	Resolves bool
	Addrs   []string
}

// ProbeDNS checks whether BUCKET.s3.amazonaws.com resolves to at least one
// address. A successful resolution is strong evidence the bucket exists (even
// if access is denied over HTTP).
func ProbeDNS(ctx context.Context, bucket string) DNSResult {
	host := bucket + ".s3.amazonaws.com"
	r := &net.Resolver{}
	ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	addrs, err := r.LookupHost(ctx2, host)
	if err != nil {
		return DNSResult{Bucket: bucket, Resolves: false}
	}
	return DNSResult{Bucket: bucket, Resolves: len(addrs) > 0, Addrs: addrs}
}
