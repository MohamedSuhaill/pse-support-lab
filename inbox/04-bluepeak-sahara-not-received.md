```
From:        Daniel Okafor <daniel.okafor@bluepeak-erp.example>
To:          support@invoicely-lite.example
Cc:          it@saharamedical.example
Date:        Mon, 16 Nov 2026 11:30 +0400
Subject:     Sahara Medical - invoices "Submitted" but buyer not receiving them
Attachments: samples/sahara-medical-batch-1116.xlsx
```

Hi team,

Daniel from Bluepeak ERP Integrations here. We look after the ERP connector for Sahara Medical Supplies FZE.

Our connector pushed this morning's batch for Sahara Medical (40 invoices, file attached) to your import endpoint. The connector got an HTTP 201 back, so from our side the hand-off was successful. In the portal the invoices show as "Submitted".

However, Sahara's customer, Emirates Palm Hospital, says most of these invoices have not arrived on their side. They received a few, not the rest.

A few things I noticed:

- The "ASP reference" field in the portal is empty for a lot of the invoices.
- Nothing in our connector logs shows an error.

Two questions:

1. Is Invoicely losing invoices somewhere between the portal and the network? Where are the rest of them?
2. Should we resend the batch from the connector? I'd rather not create duplicates at the hospital, so I'll hold off until I hear from you.

Happy to share connector logs or jump on a call if useful. Sahara's AR team would like an answer today if possible, as the hospital's payment run is this week.

Thanks,

Daniel Okafor
Integration Consultant
Bluepeak ERP Integrations
