
# Lab 0.4 — OWASP SAMM Fundamentals

**Course:** OWASP SAMM Fundamentals | **Role:** Developer

## Objective

Evaluate InstaSafe's software development process against the OWASP Software Assurance Maturity Model (SAMM) across all 5 business functions, identify maturity gaps, and propose 2 actionable improvements.

## Methodology

This assessment is based on:
- **docs.instasafe.com** — product documentation across all 4 InstaSafe products (ZTAA, ZTAA Gov Edition, IdSec Enterprise, ISA), including architecture pages, release notes, product known issues, and end-of-support pages
- **instasafe.com** — the marketing site, including the Responsible Disclosure Policy and Business Continuity Planning page
- **COGS onboarding lab practices** — hands-on labs covering networking, identity, monitoring, incident simulation, and change management

These sources describe the product itself and how customers configure and use it. They do not contain internal information about how InstaSafe's engineering team plans, builds, or tests its software. Many SAMM practices may well be followed internally, but this could not be confirmed or found from the available sources — so those practices are scored conservatively rather than assumed.

## Step 1 — The 15 SAMM Practices

| Business Function | Security Practices |
|---|---|
| Governance | Strategy & Metrics, Policy & Compliance, Education & Guidance |
| Design | Threat Assessment, Security Requirements, Secure Architecture |
| Implementation | Secure Build, Secure Deployment, Defect Management |
| Verification | Architecture Assessment, Requirements-driven Testing, Security Testing |
| Operations | Incident Management, Environment Management, Operational Management |

## Step 2 — Scorecard

All 90 underlying questions (15 practices × 2 streams × 3 maturity levels) were answered with an evidence note. Full detail: [SAMM Scorecard — Google Sheet](https://docs.google.com/spreadsheets/d/19q9-bSIQ1v6ZBm1-_YPQ0X5i6VsuXMjV_YcrT6aq_lU/edit?usp=sharing)

| Business Function | Security Practice | Score |
|---|---|---|
| Governance | Strategy & Metrics | 0 |
| Governance | Policy & Compliance | 0 |
| Governance | Education & Guidance | 0.125 |
| Design | Threat Assessment | 0 |
| Design | Security Requirements | 0 |
| Design | Secure Architecture | 0.25 |
| Implementation | Secure Build | 0 |
| Implementation | Secure Deployment | 0.125 |
| Implementation | Defect Management | 0.125 |
| Verification | Architecture Assessment | 0 |
| Verification | Requirements-driven Testing | 0 |
| Verification | Security Testing | 0 |
| Operations | Incident Management | 0 |
| Operations | Environment Management | 0 |
| Operations | Operational Management | 0.125 |

**Summary:** 10 of 15 practices score 0 (no visibility from available sources). 5 practices have small, evidence-backed partial scores — Security Architecture, Secure Deployment, Defect Management, Education & Guidance, and Operational Management — each backed by a specific citation (architecture pages, release notes, product known issues, responsible disclosure policy, business continuity page, and the COGS onboarding curriculum).

## Step 3 — Maturity Radar Chart

![SAMM Radar Chart](samm-radar-chart.png)

Current maturity across all 5 domains sits below the Level 1 "starter" target — a common baseline for organizations early in a SAMM journey. Design and Implementation are slightly ahead of the other domains due to the Secure Architecture, Secure Deployment, and Defect Management evidence found. Verification sits at a flat 0, since testing and code-review practices are never publicly disclosed by any vendor.

## Step 4 — Gap Analysis & Improvement Proposals

Full detail: [Gap Analysis — Google Doc](https://docs.google.com/document/d/16nPaEvMuuIGQGh9QOpYMnnazl4T3_H5XBaVEQTgHX_w/edit?usp=sharing)

Of the 10 practices scoring 0, several are tied at the same gap against the Level 1 target. The two below were selected using a simple rule: which gaps, if closed, would help the most other practices improve too.

### Gap 1 — Governance: Strategy & Metrics (Current: 0 → Target: 1)
- **Evidence/rationale:** No enterprise-wide risk appetite, application security strategic plan, or security KPIs are described anywhere in the available sources. This is the most foundational gap — without a stated strategy or basic metrics, it's hard to prioritize or measure improvement in any other practice.
- **Proposed change:** Document a simple, one-page Application Security Strategy — state the risk appetite in plain terms, list 2–3 measurable near-term goals (e.g., "SAST scanning on all builds by Q3"), and assign an owner. Pair it with 2–3 easy-to-track basic metrics (e.g., open security defects by severity, time-to-fix for critical issues).
- **Estimated effort:** Low — a documentation and alignment exercise, not new tooling or headcount.
- **Expected benefit:** Gives every other Governance, Design, and Implementation gap a reference point to be measured and prioritized against.

### Gap 2 — Design: Threat Assessment (Current: 0 → Target: 1)
- **Evidence/rationale:** No application risk classification method or threat modeling practice is described anywhere, despite Secure Architecture being the highest-scoring Design practice (0.25). The product shows good secure-by-design patterns, but there's no visible sign that new features are threat-modeled to make sure those patterns hold up as the product evolves.
- **Proposed change:** Introduce a simple, mandatory threat-modeling checklist (e.g., a lightweight STRIDE pass) for new features touching authentication, gateway routing, or tenant data isolation — the highest-risk areas for this product line.
- **Estimated effort:** Medium — needs a template, brief team training, and a checkpoint added to the existing release process.
- **Expected benefit:** Protects the architectural strengths already in place by making sure they get re-checked as the product changes.

Both proposals are scoped as "move from 0 to 1" — achievable near-term improvements rather than aspirational leaps to Level 3.

## Submission Checklist

- [x] `samm-scorecard.xlsx` — completed scoring toolbox with evidence column
- [x] `samm-radar-chart.png` — visualisation
- [x] Gap analysis — 1-page per improvement proposal

## Key Takeaway

Assessing SAMM maturity purely from public product documentation and marketing content has a hard ceiling: it can surface signals about practices that happen to leave a visible trace (architecture, release cadence, defect tracking), but it cannot speak to internal-only practices (strategy, threat modeling methodology, build pipelines, test coverage, incident response). The honest result — 10 of 15 practices scoring 0 — reflects that constraint, not a judgment that InstaSafe lacks internal maturity.
