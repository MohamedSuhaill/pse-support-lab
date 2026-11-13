# Product Support Engineer — Assignment Submission

- **Name:** Mohamed Athip Sukaill A M
- **Date submitted:** 14 November 2026
- **Repository link:** 
- **Total time spent (real time):** 4.5 hours

> All customer replies below are drafts written against the exercise clock. They have not been sent.

## Executive Summary

I reviewed all five support cases, assigned priorities using the SLA matrix, and identified the next actions for each case.

I fully investigated:
1. Oasis Builders Supply — duplicate upload concern.
2. API restart and intermittent 502 incident.
3. Al Noor Fuel — BR-CO-15 invoice rejection.

For Sahara Medical and Ghaf Tree Café, I documented the next investigation steps and safe customer guidance.

---

# Incident 1 — Oasis Builders Supply LLC

File: `inbox/01-oasis-builders-upload-stuck.md`

### Acknowledgement

- **Sent at (exercise clock):** 09:30, 16 November 2026
- **Reply:**

Hello Rashid,

Thank you for reporting this issue. I understand that you uploaded the November invoice file multiple times and are concerned about duplicate submissions.

I am checking the upload history and invoice processing records to identify which invoices have already been accepted by the ASP. Please do not upload the same file again until we confirm the status.

I will share an update by 11:30 Dubai time with the invoice-level findings and recommended next steps.

Regards,  
Product Support Team

### Priority and justification

- **Priority:** P2
- **Why:** Multiple invoices are affected, and the customer cannot confidently determine which submissions have been processed. This creates a risk of duplicate business submissions. However, the available evidence does not establish a complete production outage.
- **First response target:** Within 1 hour of receipt.
- **Next update:** 11:30 Dubai time.
- **Resolution target:** Within 8 business hours, subject to investigation and escalation.

### Investigation log

| Time (exercise clock) | What I checked / did | Result |
|---|---|---|
| 09:35 | Reviewed the customer report and sample file reference | Eight invoices and 600 lines reported |
| 09:45 | Replayed the supplied Oasis sample in the lab | Import completed successfully |
| 10:00 | Checked import details through the API | Eight invoices, all Accepted |
| 10:10 | Queried the database for previous Oasis imports | Found two November imports |
| 10:20 | Compared invoice numbers and statuses | Both imports contained eight Accepted invoices |
| 10:30 | Compared buyers and invoice totals | Matching buyers and amounts, but different invoice numbers |
| 10:40 | Checked ASP message IDs | Both sets had ASP message IDs |

### Findings

The earlier November import was already marked Completed, with all eight invoices Accepted. A subsequent controlled replay also produced eight Accepted invoices.

The two sets have different invoice numbers and different ASP message IDs, although the corresponding buyers and invoice amounts match.

This indicates a potential duplicate business-submission risk. However, the available records do not confirm whether both sets were delivered to or processed by the buyers.

### Recommended action

- Do not ask the customer to upload the same file again.
- Escalate to L2 to verify ASP delivery and determine which invoice set should be treated as authoritative.
- Confirm the buyer-side status before recommending any further submission.
- Provide the customer with an invoice-level status report after verification.

---

# Incident 2 — API Restarts and Intermittent 502 Errors

File: `inbox/02-api-restarts-and-502s.md`

### Acknowledgement

- **Sent at (exercise clock):** 10:00, 16 November 2026
- **Reply:**

Hello Team,

We have received the report about intermittent upload failures and API availability. We are investigating the service health, restart history, and affected invoice processing records.

We will verify whether invoices are still reaching the ASP and whether any submissions require recovery. Please avoid repeatedly submitting the same invoices while we check their processing status.

We will provide the next update by 10:30 Dubai time.

Regards,  
Product Support Team

### Priority and justification

- **Initial priority:** P1 provisional, pending confirmation of impact.
- **Why:** The reported multiple API restarts, intermittent 502 errors, and impact across customers indicate a potentially serious service incident.
- **Important qualification:** The SLA policy defines P1 around production being down or invoices not reaching the ASP. The evidence collected so far does not confirm that threshold. Reassess as P2 if invoices continue reaching the ASP and the incident remains an intermittent degradation.
- **Next update:** 10:30 Dubai time.

### Investigation log

| Time (exercise clock) | What I checked / did | Result |
|---|---|---|
| 09:50 | Reviewed the Grafana alert and Customer Success report | Reported three restarts in 15 minutes and intermittent 502 errors |
| 10:00 | Checked API health endpoint | Returned `{"status":"ok"}` |
| 10:05 | Checked Docker container status | API container was running and healthy |
| 10:10 | Inspected container restart information | RestartCount was 2; OOMKilled was false |
| 10:15 | Checked container events | Observed health-check events; no container die or restart event in the captured event window |
| 10:20 | Reviewed API logs for errors and shutdown messages | No matching error or shutdown message found in the checked log output |

### Findings

The API was running and healthy at the time of investigation.

Docker inspection reported:
- RestartCount: 2
- OOMKilled: false
- ExitCode: 0
- Status: running

The customer-facing report and Grafana alert described three restarts in a 15-minute window. The Docker inspection result does not independently confirm that exact count or timeframe.

No root cause for the reported 502 errors has been established. The current health check also does not prove that all customer uploads are working correctly.

### Recommended action

- Escalate to L2/Engineering to correlate Grafana alert timestamps with API and infrastructure logs.
- Check reverse-proxy and application logs for 502 responses.
- Identify affected customer requests and import IDs.
- Verify whether affected invoices reached the ASP.
- Keep the incident open until stability and successful processing are confirmed.

---

# Incident 3 — Al Noor Fuel: BR-CO-15 Rejections

File: `inbox/03-al-noor-br-co-15-rejections.md`

### Acknowledgement

- **Sent at (exercise clock):** 10:30, 16 November 2026
- **Reply:**

Hello Fatima,

Thank you for sharing the rejection details. We understand the importance of resolving this issue, particularly while your team is reviewing its VAT position.

We have started checking the rejected invoices and their validation history. The rejection indicates that the invoice total does not satisfy the BR-CO-15 calculation requirement.

Please do not resubmit the unchanged rejected invoices. We will share the invoice-level findings and recommended correction steps by 12:30 Dubai time.

Regards,  
Product Support Team

### Priority and justification

- **Priority:** P2
- **Why:** Three invoices in the supplied batch were rejected by the ASP, while other invoices in the same batch were accepted. This affects a subset of invoice submissions and requires investigation.
- **First response target:** Within 1 hour.
- **Next update:** 12:30 Dubai time.
- **Resolution target:** Within 8 business hours.

### Investigation log

| Time (exercise clock) | What I checked / did | Result |
|---|---|---|
| 10:35 | Reviewed the supplied Al Noor sample | Six invoices in the sample |
| 10:45 | Replayed the sample in the lab | Import completed |
| 10:50 | Checked import details through the API | Three Accepted and three Rejected |
| 11:00 | Reviewed invoice-level rejection details | All three rejected invoices had BR-CO-15 |
| 11:10 | Checked invoice event history | Portal validation passed before ASP rejection |
| 11:20 | Compared actual and expected invoice totals | Differences of AED 0.01, AED 0.02 and AED 0.01 |

### Findings

Three invoices were rejected by the ASP with BR-CO-15.

| Invoice number | Actual total incl. VAT | Expected total | Difference |
|---|---:|---:|---:|
| ANF-2611-0042-GJ9A | 15,022.42 | 15,022.43 | 0.01 |
| ANF-2611-0044-GJ9A | 11,817.57 | 11,817.59 | 0.02 |
| ANF-2611-0046-GJ9A | 5,610.99 | 5,611.00 | 0.01 |

The event history shows that these invoices passed portal validation, were submitted, and were subsequently rejected by the ASP.

The differences indicate an invoice arithmetic inconsistency. A rounding or calculation issue is possible, but the exact source of the mismatch has not yet been established.

### Recommended action

- Escalate the three rejected invoice records to L2 for calculation and transformation review.
- Compare source spreadsheet values with the generated invoice payload.
- Verify the VAT and total calculation before retrying.
- Ask the customer not to resubmit unchanged invoices.
- Provide a clear explanation of the three affected invoice numbers.

---

# Incident 4 — Sahara Medical: ASP References Missing

File: `inbox/04-sahara-medical-asp-reference-missing.md`

### Acknowledgement

- **Sent at (exercise clock):** 11:45, 16 November 2026
- **Reply:**

Hello Daniel,

Thank you for reporting the issue. We understand that your connector received HTTP 201 responses, but the buyer has not received all the invoices and some ASP references are missing.

HTTP 201 confirms that the portal accepted the request, but it does not by itself confirm successful ASP delivery.

Please do not resend the affected invoices until we verify their processing status. We will check the import records, invoice events, and ASP responses and provide an update by 13:45 Dubai time.

Regards,  
Product Support Team

### Priority and justification

- **Priority:** P2
- **Why:** A subset of submitted invoices may not have completed ASP processing or delivery. Resending without checking could create duplicate submissions.
- **First response target:** Within 1 hour.
- **Next update:** 13:45 Dubai time.
- **Resolution target:** Within 8 business hours.

### Investigation status

A full invoice-level investigation has not yet been completed.

### Next checks

1. Identify the import ID associated with the 40 invoices.
2. Retrieve the import details using the API.
3. Check each invoice's current status.
4. Review invoice event history to determine whether each invoice reached the Submitted, Accepted, or Rejected state.
5. Check whether ASP message IDs are present.
6. Correlate missing ASP references with ASP mock logs.
7. Determine whether the invoices are pending, rejected, or accepted.
8. Escalate missing or inconsistent ASP responses to L2.

### Recommended action

Do not resend the affected invoices until their existing processing state is confirmed. If the portal accepted the request but ASP delivery is unconfirmed, investigate the original submission before attempting a retry.

---

# Incident 5 — Ghaf Tree Café: Date Format Rejection

File: `inbox/05-ghaf-tree-cafe-date-format.md`

### Acknowledgement

- **Sent at (exercise clock):** 13:00, 16 November 2026
- **Reply:**

Hello Hessa,

Thank you for sharing the file and rejection details.

The rejection indicates that the issue date format does not match the required format. Your file contains dates in `DD/MM/YYYY` format, while the portal expects `YYYY-MM-DD`.

Please update the issue date values to the required format and validate the corrected file before uploading it again.

We will provide any additional guidance needed by the next business day.

Regards,  
Product Support Team

### Priority and justification

- **Priority:** P3
- **Why:** The issue affects one customer's four invoices and appears to be related to the input file format rather than a platform-wide incident.
- **First response target:** Within 4 hours.
- **Next update:** Next business day.
- **Resolution target:** Within 3 days.

### Investigation status

The reported rejection message and sample description indicate a date-format mismatch. A complete replay and invoice-level verification have not yet been performed.

### Next checks

1. Review the original date values in the supplied spreadsheet.
2. Confirm that the affected values use `DD/MM/YYYY`.
3. Verify that the portal expects `YYYY-MM-DD`.
4. Correct the date format while preserving the actual date.
5. Validate the corrected file.
6. Confirm whether all four invoices are accepted after resubmission.

### Recommended action

Explain the required date format clearly and guide the customer to correct the spreadsheet. This is an input-format issue based on the available report, not evidence of a platform outage.

---

# Escalation and Ownership

As L1 Product Support, I remain responsible for the customer conversation and follow-up.

| Team | When to involve |
|---|---|
| L1 Support | Initial triage, customer communication, evidence gathering and status updates |
| L2 Support | Complex invoice processing, configuration, data investigation and second-level review |
| L3 Engineering | Application defects, infrastructure failures, code changes and engineering-level data repairs |

### Escalation priorities

- **API incident:** Escalate to L2/Engineering to investigate restarts and intermittent 502 responses. Page on-call if P1 impact is confirmed.
- **Oasis Builders:** Escalate to L2 to verify ASP delivery and determine the authoritative invoice set.
- **Al Noor Fuel:** Escalate to L2 for invoice calculation and payload comparison.
- **Sahara Medical:** Escalate if ASP delivery or message-reference records are inconsistent.
- **Ghaf Tree Café:** L1 guidance should be sufficient initially; escalate if corrected files continue to fail.

Every escalation should include the customer, priority, timestamps, import and invoice IDs, observed symptoms, supporting evidence, attempted checks, suspected cause, workaround, and the exact help required.

---

# SLA and Customer Communication Principles

I used the supplied SLA matrix to classify the incidents and set response expectations.

- P1: 15–30 minute response target, 4-hour resolution target, and 30–60 minute updates.
- P2: 1-hour response target, 8 business-hour resolution target, and 2–4 hour updates.
- P3: 4-hour response target, 3-day resolution target, and daily updates.
- P4: 1-business-day response target, 5–10-day resolution target.

P2–P4 business hours are Monday to Friday, 08:00–18:00 Dubai time. Escalation does not pause or reset the SLA.

I followed these communication principles:
- Acknowledge the issue clearly.
- Explain what is known and what is still being investigated.
- Avoid claiming a root cause without evidence.
- Never advise a customer to resend an invoice before checking its current state.
- Provide a realistic next update time.
- Maintain ownership until resolution.

---

# Evidence Files

The following investigation notes are included in the `evidence/` directory:

- `evidence/oasis-investigation.txt`
- `evidence/api-restart-investigation.txt`
- `evidence/al-noor-investigation.txt`
- `evidence/sahara-next-checks.txt`
- `evidence/ghaf-cafe-investigation.txt`

These notes document the checks and findings gathered during the lab investigation.

---

# AI Usage Disclosure

AI assistance was used to help structure the submission, improve customer communication drafts, and organize investigation findings.

Technical conclusions were checked against the lab environment, API responses, Docker inspection, logs, and database query results. Unverified causes and assumptions have been identified as such rather than presented as confirmed facts.

---

# Final Status

All five inbox items have been reviewed and assigned priorities. Three incidents have been investigated using the available lab evidence. The remaining two have documented next checks and customer guidance.

No customer replies have been sent, no production customer data has been changed, and no unverified root cause has been presented as confirmed.
