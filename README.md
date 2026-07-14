# s3-enum

S3 bucket enumeration from a domain name or organization name using permutation wordlists, HTTP probing, and DNS resolution.

## How it works

1. **Permutation generation** — strip the TLD and expand into hundreds of candidates using built-in suffix/prefix lists (`-dev`, `-staging`, `-prod`, `-backup`, …). Supply a custom wordlist with `-w` and enable full cross-product permutations with `--permutations`.
2. **HTTP probing** — each candidate is checked via virtual-hosted-style and path-style URLs across multiple regions. Responses are classified:
   - `PUBLIC` (200) — exists and allows anonymous listing; object keys are extracted.
   - `PRIVATE` (403) — exists but access is denied.
   - `NOT_FOUND` (404 + `NoSuchBucket`) — bucket does not exist.
   - `AMBIGUOUS` — unexpected response.
3. **Region detection** — extracted from the `x-amz-bucket-region` response header.
4. **DNS probing** — `BUCKET.s3.amazonaws.com` is resolved; a successful lookup is independent evidence the bucket exists.

## Install

```sh
go install github.com/unrandoms/s3-enum@latest
```

Or build from source:

```sh
git clone https://github.com/unrandoms/s3-enum.git
cd s3-enum
go build -o s3-enum .
```

## Usage

```
s3-enum -t TARGET [flags]

Flags:
  -t, --target TARGET       Domain or org name to enumerate (required)
  -w, --wordlist FILE        Append words from FILE to the built-in list
      --threads N            Concurrency level (default 20, max 200)
  -o, --output FILE          Write results to FILE (.json or plain text)
      --public-only          Only report PUBLIC buckets
      --permutations         Cross-product: target-WORD and WORD-target for all words
      --no-color             Disable ANSI colors
      --no-dns               Skip DNS resolution probes
  -h, --help                 Show this help
```

## Examples

```sh
# Basic enumeration
s3-enum -t acme.com

# Custom wordlist + more threads
s3-enum -t acme -w my-words.txt --threads 50

# Save JSON report, public buckets only
s3-enum -t acme -o results.json --public-only

# Full permutation mode
s3-enum -t acme --permutations --threads 100
```

## Output

```
BUCKET          STATUS     REGION      DNS  OBJECTS  NOTES
--------------------------------------------------------------------------------
acme-dev        PUBLIC     us-east-1   yes  42       keys: index.html, app.js…
acme-staging    PRIVATE    eu-west-1   yes  -
acme-backup     NOT_FOUND  -           -    -
```

Green = PUBLIC, Yellow = PRIVATE, dim = NOT_FOUND.

## License

MIT
