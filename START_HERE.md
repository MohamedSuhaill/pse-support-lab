# Start here — Invoicely Lite support lab

Welcome, and thank you for taking the time to do this exercise.

## What this is

Invoicely Lite is a small, **fictional** UAE e-invoicing portal. Businesses upload an Excel file, the portal turns each invoice into a PINT AE e-invoice and sends it to an Accredited Service Provider (ASP) on the Peppol network.

This repository contains a working copy of the whole system. It runs on your laptop with Docker: the portal, its API, a simulated ASP, a database, and a monitoring stack (Grafana, Prometheus, Loki). Every company, person, TRN and email in this lab is made up.

## Your role

You are the on-shift **L1 Product Support Engineer** at Invoicely Lite.

It is **Monday 16 November 2026**. Your shift starts at **09:00 Asia/Dubai** (UTC+4). During the morning, five items arrive in [`inbox/`](inbox/): customer emails, a partner email and an internal alert. Treat them as they arrive, in the order of their timestamps.

For each item, do what you would do on a real shift: acknowledge, prioritise, investigate, find the root cause with evidence, reply to the customer, and escalate where that is the right call.

## What we assess

- **Acknowledgement within SLA** — a meaningful first reply, sent within the first-response target for the priority you chose.
- **Prioritisation** — which item you work first, and why.
- **Investigation and root cause, with evidence** — dashboards, log lines, database queries and code references that support your conclusion.
- **Clear customer communication** — plain language, honest, no blame, a useful next step.
- **Escalation quality and ownership** — when you escalate, the note is complete and actionable, and you still own the customer conversation.
- **Learning mindset** — what you did not know, how you found out, and what you would do differently.

## Time

- You have about **2 days** of elapsed time. We expect roughly **6–8 hours** of actual work.
- **Depth over polish.** A well-evidenced root cause for a few items is worth more than shallow notes on all five.
- It is fine not to finish everything. If you run out of time, write down what you would do next.
- Keep two clocks apart: the **exercise clock** (Monday 16 Nov 2026, Asia/Dubai) for your replies and SLA decisions, and your **real time** for the time log.

## How to run it

You need Docker (with at least 4 GB of memory), `make` and `curl`. Full details are in [`README.md`](README.md).

```bash
make up      # build and start everything (first build takes 5–10 minutes)
```

Then open:

- Portal: http://localhost:3000
- Grafana: http://localhost:3001 (dashboard "Invoicely Lite – Overview")

Other useful commands:

```bash
make load    # replay this morning's customer traffic (~3–4 minutes)
make reset   # wipe the database and start clean
```

## Suggested approach

1. Read the docs in [`docs/`](docs/), starting with [`docs/architecture.md`](docs/architecture.md).
2. Upload the clean sample `samples/desert-rose-nov-2026.xlsx` (seller: Desert Rose Trading LLC). Watch what normal looks like in the portal, in Grafana and in the logs.
3. Read the whole inbox, decide your priority order, and send your acknowledgements.
4. Work the items in your priority order. Use `make load` or `make replay CUSTOMER=<name>` to reproduce what customers did.

## What to submit

Send us a **link to a git repository** containing your copy of this repo with:

- [`SUBMISSION.md`](SUBMISSION.md) completed.
- Your evidence in an `evidence/` folder: screenshots, log excerpts, queries and their output. Link to them from `SUBMISSION.md`.

**Optional bonus:** code fix(es) on a separate git branch, with an explanation in `SUBMISSION.md` of what you changed and why.

## Rules

- You may read and change any code in this repository.
- AI tools are allowed. Disclose which tools you used and how, and verify everything they tell you against the system itself.
- Do not contact any real company, person or service. Everything here is fictional.
- Replies you write are drafts in `SUBMISSION.md`; do not send real emails.

Good luck, and enjoy it.
