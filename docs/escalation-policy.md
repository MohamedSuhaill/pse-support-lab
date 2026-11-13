# Escalation policy

## Levels

| Level | Who | Owns |
|---|---|---|
| **L1** | Product Support Engineer (you) | The customer conversation, end to end. Triage, priority, investigation, workarounds, customer updates. |
| **L2** | Senior Support / Implementation | Complex configuration, data fixes, onboarding questions, second opinion on triage. |
| **L3** | Engineering (on-call for P1) | Code changes, releases, infrastructure, data repairs that need engineering access. |

## Ownership stays with L1

Escalating hands over a **task**, not the **customer**. After you escalate you still:

- keep the customer updated at the cadence in [sla-matrix.md](sla-matrix.md);
- chase the engineering team for progress and keep the escalation note up to date;
- confirm the fix with the customer and close the case.

"Engineering is looking at it" is not an update. Say what is known, what is being done, and when you will update next.

## When to escalate

Escalate to **L3** when:

- the evidence points to a defect in the product (code, configuration or infrastructure) that support cannot fix or work around safely;
- data needs to be repaired in a way that support cannot do through the product;
- it is a P1 — page engineering on-call as soon as you have confirmed the impact. Do not wait for a full root cause.

Do **not** escalate when the product is working as documented. Explain it to the customer, help them fix their side, and consider whether the documentation or the error message could be clearer (raise that as a suggestion, not an incident).

When unsure, ask L2.

## What an escalation must contain

An escalation that engineering cannot act on comes back to you. Use this template:

```markdown
### Escalation: <one-line summary>

**Priority:** P? — <why; business impact>
**Customer / seller:** <company, seller id>
**Reported by / at:** <name, channel, time received>
**Affected:** <import ids, invoice numbers / invoice ids, how many, since when>

**Symptoms (customer view):**
<what the customer sees>

**Evidence:**
- Payload excerpt: <relevant JSON>
- ASP response / error code: <status code, rule id, message>
- Logs: <query used and relevant lines with timestamps>
- Metrics / dashboard: <panel name, query, time range, what it shows>
- Database: <query and result>

**Steps to reproduce:**
1. …

**What I have already tried / ruled out:**
- …

**Suspected cause (with confidence):**
<file:line or component, why you think so — or "unknown">

**Workaround in place:** <yes/no — what>

**Timeline:**
- hh:mm — …

**Ask:** <exactly what you need from engineering, and by when>
**Next customer update due:** <time>
```
