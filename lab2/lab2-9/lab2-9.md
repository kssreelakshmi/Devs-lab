# Security Audit Report

**vulnerable-go-app — Lab II.9: Secure Coding Practices**

---

## Executive Summary

A Semgrep-assisted static analysis and manual review of the vulnerable-go-app codebase identified three critical findings: a SQL injection vulnerability, a hardcoded credential, and an insecure JWT signature-validation pattern that trusted a token's self-declared algorithm. All three were remediated with parameterized query binding, environment-based secret loading with a fail-fast startup check, and a server-owned algorithm allowlist, respectively. Each fix was verified both by automated tests and by re-running the original exploit against the live application, confirming the vulnerable behavior no longer reproduces.

## Findings Table

| Vulnerability | CWE | Severity | File / Line | Fix Applied |
|---|---|---|---|---|
| SQL Injection | CWE-89 | High | `store.go:48` | Replaced `fmt.Sprintf` string interpolation with a bound `?` parameter passed separately from the query string, so untrusted input is never parsed as query syntax. |
| Hardcoded Credentials | CWE-798 | High | `auth.go:27–46` | Removed string-literal secrets; both the JWT signing key and API key now load from environment variables via a startup routine that calls `log.Fatal` if either is missing, preventing the app from running with an empty or default secret. |
| Insecure JWT Validation (Algorithm Confusion) | — | Critical | `auth.go:88, 106` | Replaced logic that trusted the token's own `alg` header (including an `alg:none` branch that skipped signature verification entirely) with a server-owned allowlist checked before any claims are trusted. Unlisted or `none` algorithms are rejected outright. |

> **Note:** An automated Semgrep scan (`p/golang` ruleset) was run before and after remediation. It did not flag any of the three findings above in either scan — see Lessons Learned.

## Lessons Learned

All three vulnerabilities shared a common root cause: trusting a value's shape without constraining what it could actually do. The SQL injection let a string's content change the meaning of a query; the hardcoded secret let source access substitute for infrastructure access; and the JWT flaw let the token's own metadata dictate how much verification it received. In each case the fix was to move the decision from attacker-influenced input to a fixed point the server controls — a bound parameter, an environment boundary, and an explicit allowlist — rather than trying to sanitize or special-case the input after the fact.

A notable gap surfaced during this exercise: Semgrep's `p/golang` ruleset, run both before and after remediation, returned a single unrelated finding (a missing-TLS warning) and never flagged the SQL injection, the hardcoded secret, or the JWT algorithm-confusion bug in either scan. This is a useful reminder that static analysis is strongest against syntactic patterns it has rules for, and weaker against logic-level and semantic vulnerabilities — like a validator trusting a field it should be constraining — that require understanding what the code is supposed to prevent, not just what it contains. Automated scanning should be treated as one layer of a review process, not a substitute for manual security review against a threat model.
