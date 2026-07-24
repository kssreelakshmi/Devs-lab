# vulnerable-go-app

A small, dependency-free Go app built for Lab II.9 (Secure Coding Practices).
It contains three intentional vulnerabilities, each exploitable at runtime,
not just flaggable statically:

| # | Vulnerability | CWE | File | Status | Notes |
|---|---|---|---|---|---|
| 1 | SQL Injection | CWE-89 | `store.go` — `FindUserByName` | **Fixed** | Was `fmt.Sprintf` + string interpolation; now a bound `?` parameter |
| 2 | Hardcoded secret | CWE-798 | `auth.go` — `jwtSecret`, `stripeAPIKey` | **Fixed** | Was string literals; now loaded from env vars with a fail-fast startup check |
| 3 | Insecure JWT validation | — | `auth.go` — `VerifyToken` | **Fixed** | Was trusting the token's own `alg`; now checked against a server-owned allowlist (`HS256` only) |

## Running it

Requires two environment variables (the app refuses to start without them):

- `JWT_SIGNING_SECRET` — HMAC key used to sign/verify tokens
- `STRIPE_API_KEY` — demo third-party API key

```
go build -o vulnapp .
JWT_SIGNING_SECRET="local-dev-secret-do-not-use-in-prod" STRIPE_API_KEY="sk_test_fake" ./vulnapp
```

Server listens on `:8080`.

## Exploiting it

**SQL injection (fixed)** — normal vs. injected search:
```
curl -s "localhost:8080/search?name=bob"
curl -s -G "localhost:8080/search" --data-urlencode "name=x' OR '1'='1"
```
The second call now returns `null` (no match) instead of every row — the
payload is just a literal string being compared, not query syntax.

**JWT alg:none forgery (fixed)** — the same forged token that used to work:
```
HEADER=$(echo -n '{"alg":"none","typ":"JWT"}' | base64 | tr -d '=' | tr '/+' '_-')
PAYLOAD=$(echo -n '{"sub":"eve","admin":true,"exp":9999999999}' | base64 | tr -d '=' | tr '/+' '_-')
TOKEN="$HEADER.$PAYLOAD."
curl -s -H "Authorization: Bearer $TOKEN" localhost:8080/admin
```
Now returns `401 unauthorized` — `alg:none` isn't on the server's allowlist,
so the request is rejected before the payload is even trusted.

Run `go test ./... -v` to see the full test suite, including the
lab-required test confirming `alg:none` tokens are rejected.

Compare against a legitimately issued, non-admin token:
```
TOKEN=$(curl -s "localhost:8080/login?user=bob" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
curl -s -H "Authorization: Bearer $TOKEN" localhost:8080/admin
```
Returns `forbidden`, as expected — Bob isn't an admin and his token is properly signed.

## Running Semgrep against it (Lab II.9, Step 1)

```
pip install semgrep
semgrep --config=p/golang . --output=semgrep-baseline.html --html
```
(Needs network access to `semgrep.dev` to pull the ruleset — not available in
every sandboxed environment.)

## `smoketest/`

A standalone script (not part of the app) that prints example curl commands
and a pre-forged token, for quick manual verification. Safe to delete before
you start fixing things — it's scaffolding, not part of the deliverable.

## Why these three, and not "safer" versions

Each vulnerability is written to be the textbook version so it's unambiguous
what "fixed" looks like:

- The injection isn't hidden behind an ORM or query builder — it's a bare
  `fmt.Sprintf` straight into a query call.
- The secret isn't in a config file that's merely `.gitignore`'d and easy to
  miss — it's a literal `const` in source.
- The JWT bug isn't a subtle timing issue — it's the canonical "alg:none"
  mistake: trusting the token to say how it should be checked.
