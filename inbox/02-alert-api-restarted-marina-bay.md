```
From:        Grafana Alerting <alerts@invoicely-lite.example>
To:          support-oncall@invoicely-lite.example
Date:        Mon, 16 Nov 2026 09:42 +0400
Subject:     [FIRING:1] Invoicely API restarted
```

**[FIRING:1] Invoicely API restarted**

**Firing**

```
Value: restarts=3
Labels:
 - alertname = InvoicelyApiRestarted
 - service = api
 - severity = critical
 - env = production
 - grafana_folder = Invoicely Lite
Annotations:
 - summary = API process restarted 3 times in the last 15 minutes
 - description = The api container has restarted more than twice within 15 minutes.
Source: http://localhost:3001
Dashboard: http://localhost:3001  (Invoicely Lite – Overview)
Silence: http://localhost:3001/alerting/silences
```

Sent by Grafana Alerting

---

```
From:        Priya Nair <priya.nair@invoicely-lite.example>
To:          support-oncall@invoicely-lite.example
Date:        Mon, 16 Nov 2026 09:50 +0400
Subject:     Fwd: [FIRING:1] Invoicely API restarted
```

Forwarding from #cs-escalations so it's in the support queue:

> **Priya Nair** 09:48
> Hey support, a few of my customers are pinging me. They're getting "502 Bad Gateway" on and off in the portal this morning, mostly when they upload. A couple of them say invoices have been sitting in "Validated" since about 9:30 and not moving.
> Is this the same thing as the alert above?
> Also flagging: **Marina Bay Hospitality goes live on Wednesday (18 Nov)**. Their GM is already nervous after the pilot. I need to know whether we're OK for Wednesday, ideally before my call with them this afternoon.
> Can someone own this and keep me posted? 🙏

Priya Nair
Customer Success Manager, Invoicely Lite
