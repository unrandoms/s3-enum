package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/unrandoms/s3-enum/internal/prober"
)

func sampleRows() []Row {
	return []Row{
		{Bucket: "acme-dev", Status: "PUBLIC", Region: "us-east-1", DNSHit: true, ObjCount: 3,
			Objects: []string{"a.txt", "b.txt", "c.txt"}, RawStatus: prober.StatusPublic},
		{Bucket: "acme-staging", Status: "PRIVATE", Region: "eu-west-1", DNSHit: true, RawStatus: prober.StatusPrivate},
		{Bucket: "acme-backup", Status: "NOT_FOUND", RawStatus: prober.StatusNotFound},
	}
}

func TestWriteTableContainsBucketNames(t *testing.T) {
	var buf bytes.Buffer
	WriteTable(&buf, sampleRows(), true)
	out := buf.String()
	for _, name := range []string{"acme-dev", "acme-staging", "acme-backup"} {
		if !strings.Contains(out, name) {
			t.Errorf("table output missing bucket name %q", name)
		}
	}
}

func TestWriteJSONRoundTrip(t *testing.T) {
	rows := sampleRows()
	var buf bytes.Buffer
	if err := WriteJSON(&buf, rows); err != nil {
		t.Fatalf("WriteJSON error: %v", err)
	}
	var decoded []Row
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}
	if len(decoded) != len(rows) {
		t.Errorf("got %d rows, want %d", len(decoded), len(rows))
	}
	if decoded[0].Bucket != "acme-dev" {
		t.Errorf("first bucket = %q, want acme-dev", decoded[0].Bucket)
	}
}

func TestWriteTableNoColor(t *testing.T) {
	var buf bytes.Buffer
	WriteTable(&buf, sampleRows(), true)
	if strings.Contains(buf.String(), "\033[") {
		t.Error("noColor=true but ANSI codes found in output")
	}
}
