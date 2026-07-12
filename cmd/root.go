package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/unrandoms/s3-enum/internal/output"
	"github.com/unrandoms/s3-enum/internal/prober"
	"github.com/unrandoms/s3-enum/internal/wordlist"
)

// Config holds all parsed CLI options.
type Config struct {
	Target       string
	WordlistFile string
	Threads      int
	OutFile      string
	PublicOnly   bool
	Permutations bool
	NoColor      bool
	NoDNS        bool
}

// Execute parses os.Args and runs the enumeration.
func Execute() {
	cfg := parseFlags()
	if cfg.Target == "" {
		fmt.Fprintln(os.Stderr, "usage: s3-enum -t TARGET [-w FILE] [--threads N] [-o results.json] [--public-only] [--permutations]")
		os.Exit(1)
	}

	var customWords []string
	if cfg.WordlistFile != "" {
		var err error
		customWords, err = wordlist.LoadFile(cfg.WordlistFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading wordlist %q: %v\n", cfg.WordlistFile, err)
			os.Exit(1)
		}
	}

	candidates := wordlist.Generate(cfg.Target, customWords, cfg.Permutations)
	fmt.Fprintf(os.Stderr, "[*] target: %s  candidates: %d  threads: %d\n",
		cfg.Target, len(candidates), cfg.Threads)

	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	ctx := context.Background()
	rows := enumerate(ctx, candidates, cfg, client)

	// filter if requested
	if cfg.PublicOnly {
		filtered := rows[:0]
		for _, r := range rows {
			if r.RawStatus == prober.StatusPublic {
				filtered = append(filtered, r)
			}
		}
		rows = filtered
	}

	if cfg.OutFile != "" {
		f, err := os.Create(cfg.OutFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		if strings.HasSuffix(cfg.OutFile, ".json") {
			if err := output.WriteJSON(f, rows); err != nil {
				fmt.Fprintf(os.Stderr, "error writing JSON: %v\n", err)
				os.Exit(1)
			}
		} else {
			output.WriteTable(f, rows, true)
		}
		fmt.Fprintf(os.Stderr, "[+] results written to %s\n", cfg.OutFile)
	}

	// Always print table to stdout
	output.WriteTable(os.Stdout, rows, cfg.NoColor)
}

func enumerate(ctx context.Context, candidates []string, cfg Config, client *http.Client) []output.Row {
	type job struct{ bucket string }

	jobs := make(chan job, len(candidates))
	for _, c := range candidates {
		jobs <- job{c}
	}
	close(jobs)

	resultsCh := make(chan output.Row, len(candidates))
	var wg sync.WaitGroup

	for i := 0; i < cfg.Threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				hr := prober.ProbeHTTP(ctx, j.bucket, client)
				row := output.Row{
					Bucket:    j.bucket,
					Status:    hr.Status.String(),
					Region:    hr.Region,
					ObjCount:  hr.ObjCount,
					Objects:   hr.Objects,
					ProbeURL:  hr.URL,
					RawStatus: hr.Status,
				}
				if !cfg.NoDNS {
					dr := prober.ProbeDNS(ctx, j.bucket)
					row.DNSHit = dr.Resolves
				}
				resultsCh <- row
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var rows []output.Row
	for r := range resultsCh {
		rows = append(rows, r)
	}
	return rows
}

func parseFlags() Config {
	cfg := Config{Threads: 20}
	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-t", "--target":
			i++
			if i < len(args) {
				cfg.Target = args[i]
			}
		case "-w", "--wordlist":
			i++
			if i < len(args) {
				cfg.WordlistFile = args[i]
			}
		case "--threads":
			i++
			if i < len(args) {
				fmt.Sscanf(args[i], "%d", &cfg.Threads)
			}
		case "-o", "--output":
			i++
			if i < len(args) {
				cfg.OutFile = args[i]
			}
		case "--public-only":
			cfg.PublicOnly = true
		case "--permutations":
			cfg.Permutations = true
		case "--no-color":
			cfg.NoColor = true
		case "--no-dns":
			cfg.NoDNS = true
		case "-h", "--help":
			printHelp()
			os.Exit(0)
		}
	}

	if cfg.Threads < 1 {
		cfg.Threads = 1
	}
	if cfg.Threads > 200 {
		cfg.Threads = 200
	}
	return cfg
}

func printHelp() {
	fmt.Print(`s3-enum — S3 bucket enumeration tool

Usage:
  s3-enum -t TARGET [flags]

Flags:
  -t, --target TARGET       Domain or organization name to enumerate (required)
  -w, --wordlist FILE        Append words from FILE to the built-in list
      --threads N            Concurrency level (default 20, max 200)
  -o, --output FILE          Write results to FILE (.json or plain text)
      --public-only          Only report PUBLIC buckets
      --permutations         Also generate WORD-target and target-WORD for all words
      --no-color             Disable ANSI color output
      --no-dns               Skip DNS resolution probes
  -h, --help                 Show this help

Examples:
  s3-enum -t acme.com
  s3-enum -t acme -w words.txt --threads 50
  s3-enum -t acme -o results.json --public-only
`)
}
