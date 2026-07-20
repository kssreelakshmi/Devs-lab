# Lab II.7 — ZTNA Gateway System Design

System design for a Zero Trust Network Access (ZTNA) gateway capable of handling 1,000,000 concurrent authenticated sessions, plus a small working Go service that implements and proves out the core session/auth mechanism the design depends on.

## What's in this submission

| File | Description |
|---|---|
| `system-design.pdf` | The full design document (6 pages) — requirements, scale estimation, architecture, data layer, scaling strategy, failure modes, and NIST 800-207 alignment |
| `system-design.docx` | Editable Word version of the same document |
| `architecture-diagram.png` | Component architecture diagram (built in draw.io), also embedded in the PDF |
| `ztna-auth-service/` | Working Go implementation of the Auth Service + Session Store slice of the design |

## The design document

Covers, in order:

1. **Requirements & scale estimation** — functional/non-functional requirements, and the scale math (10GB session data, ~17K auth checks/sec) that drives every later architecture decision
2. **Architecture** — component diagram tracing a request end-to-end: DNS/Anycast → Global Load Balancer → Auth Service → Policy Engine → App Connector → Application, with Identity Provider, Session Store (Redis), and Audit Logging as supporting services
3. **Data layer** — justifies Redis Cluster for session state, PostgreSQL for the device registry, and etcd for the policy store; defines the session schema; explains the caching/invalidation strategy
4. **Scaling & failure modes** — how each component scales horizontally, three failure mode analyses (Auth Service down, Redis Cluster partition, Identity Provider unavailable), and an availability calculation showing why component-level redundancy is necessary to hit a 99.99% SLA
5. **Zero Trust alignment** — maps the design to NIST 800-207 tenets, with the session-revocation test (below) as concrete evidence of "never trust, always verify"

## Why there's working code alongside a design doc

Rather than leave the Auth Service and Session Store as boxes on a diagram, I built a minimal but real version of that slice in Go, so the design's core claim — that a valid JWT is not enough, the session must still be live in the session store — could actually be demonstrated instead of just asserted.

### What it does

- `POST /login` — issues a JWT containing the session schema fields from the design doc (`user_id`, `device_id`, `policy_hash`, `app_id`, `region`, `exp`, `iat`), and writes the session to Redis with a TTL matching the token's expiry
- `GET /verify` — validates the JWT's signature and expiry, then cross-checks the claims against the live session in Redis. If the session has been deleted (revoked) or the `policy_hash` no longer matches, access is denied — even if the JWT itself is still technically valid

### Running it

Requires Docker and Go 1.21+.

```bash
# start Redis
docker run -d --name ztna-redis -p 6379:6379 redis:7-alpine

# from the ztna-auth-service directory
go mod tidy
go run main.go
```

Then, in another terminal:

```bash
# log in and get a token
curl -X POST http://localhost:8080/login -H "Content-Type: application/json" \
  -d '{"user_id":"u123","device_id":"d456","app_id":"crm-app","region":"ap-south-1"}'

# verify it (use the token from the response above)
curl http://localhost:8080/verify -H "Authorization: Bearer <token>"
```

### The revocation test

This is the proof point referenced in Section 5 of the design doc:

1. Log in, get a token
2. Verify it — returns `{"decision":"allow", ...}`
3. Delete the session directly from Redis: `docker exec -it ztna-redis redis-cli DEL "session:u123:d456"`
4. Verify again with the **same token** — returns `{"decision":"deny","reason":"session not found or revoked"}`

The token itself never changed and hadn't expired; only the live session state did. That's the practical difference between a system that validates tokens and one that actually enforces Zero Trust.

## Still outstanding

- **Peer review note** — the design doc has a placeholder in Section 6 for a colleague's review paragraph, per the submission checklist. Needs to be filled in before final submission.
