package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/unrandoms/s3-enum/internal/prober"
)

// Row is one result entry combining HTTP and DNS findings for a bucket.
type Row struct {
	Bucket    string         `json:"bucket"`
	Status    string         `json:"status"`
	Region    string         `json:"region,omitempty"`
	DNSHit    bool           `json:"dns_resolves"`
	ObjCount  int            `json:"object_count,omitempty"`
	Objects   []string       `json:"objects,omitempty"`
	ProbeURL  string         `json:"probe_url,omitempty"`
	RawStatus prober.Status  `json:"-"`
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorDim    = "\033[2m"
)

// WriteTable renders results as a colored tab-separated table to w.
// If noColor is true, ANSI codes are omitted.
func WriteTable(w io.Writer, rows []Row, noColor bool) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "BUCKET\tSTATUS\tREGION\tDNS\tOBJECTS\tNOTES")
	fmt.Fprintln(tw, strings.Repeat("-", 80))
	for _, r := range rows {
		color, reset := pickColor(r.RawStatus, noColor)
		dns := "-"
		if r.DNSHit {
			dns = "yes"
		}
		region := r.Region
		if region == "" {
			region = "-"
		}
		objCount := "-"
		if r.RawStatus == prober.StatusPublic {
			objCount = fmt.Sprintf("%d", r.ObjCount)
		}
		notes := ""
		if r.RawStatus == prober.StatusPublic && len(r.Objects) > 0 {
			notes = "keys: " + strings.Join(r.Objects[:min(3, len(r.Objects))], ", ")
			if len(r.Objects) > 3 {
				notes += "…"
			}
		}
		fmt.Fprintf(tw, "%s%s\t%s\t%s\t%s\t%s\t%s%s\n",
			color, r.Bucket, r.Status, region, dns, objCount, notes, reset)
	}
	tw.Flush()
}

// WriteJSON encodes rows as a JSON array to w.
func WriteJSON(w io.Writer, rows []Row) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}

func pickColor(s prober.Status, noColor bool) (string, string) {
	if noColor {
		return "", ""
	}
	switch s {
	case prober.StatusPublic:
		return colorGreen, colorReset
	case prober.StatusPrivate:
		return colorYellow, colorReset
	default:
		return colorDim, colorReset
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
