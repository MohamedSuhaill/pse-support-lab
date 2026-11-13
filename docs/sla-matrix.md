# SLA matrix

| Priority | Definition | First response | Resolution target | Update cadence |
|---|---|---|---|---|
| **P1** | Production down / no invoices reaching the ASP for one or more customers | 15–30 min | 4 h | every 30–60 min |
| **P2** | Subset of invoices rejected or major degradation | 1 h | 8 h | every 2–4 h |
| **P3** | Single user or single invoice type affected | 4 h | 3 days | daily |
| **P4** | How-to / question | 1 business day | 5–10 days | as agreed |

## Hours

- **P1:** 24×7.
- **P2–P4:** business hours, Monday–Friday 08:00–18:00 Asia/Dubai (UTC+4). The clock pauses outside business hours.

## How the clock works

- The clock starts when the message is **received**, not when you open it.
- **First response** means a meaningful reply from a person: it shows you understood the problem, says what happens next and when the customer will hear from you again. An auto-reply does not count.
- **Resolution** means the customer's problem is solved or a workaround is in place and agreed. A response deadline is not a resolution promise.
- Escalating to another team does **not** reset or pause the clock. You still own the customer update cadence.

## Choosing a priority

Ask, in this order:

1. **Can customers get invoices to the ASP at all?** If not, for one or more customers, it is P1.
2. **How many invoices or customers are affected, and how badly?** A subset of invoices rejected, or a major slowdown, is P2.
3. **Is it one user, one invoice type, or a cosmetic issue with a workaround?** P3.
4. **Is it a question, or the product working as documented?** P4.

Consider business impact as well: a go-live date, month-end close, an auditor or a payment run can make the same technical problem more urgent. Write down why you chose the priority.

## Re-prioritise when impact changes

Priority is not fixed. Raise it as soon as new evidence shows wider impact (more customers, production down, compliance risk); lower it when a workaround is in place. Tell the customer when the priority changes and why.
