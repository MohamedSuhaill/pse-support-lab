-- Fictional customers and a few earlier batches (October 2026).

INSERT INTO sellers (id, name, trn, peppol_id, emirate) VALUES
    (1, 'Desert Rose Trading LLC',        '100234567800003', '0235:1002345678', 'DXB'),
    (2, 'Oasis Builders Supply LLC',      '100318274600003', '0235:1003182746', 'SHJ'),
    (3, 'Marina Bay Hospitality LLC',     '100427193500003', '0235:1004271935', 'DXB'),
    (4, 'Al Noor Fuel & Lubricants LLC',  '100536284100003', '0235:1005362841', 'AUH'),
    (5, 'Sahara Medical Supplies FZE',    '100645371900003', '0235:1006453719', 'AJM'),
    (6, 'Ghaf Tree Café LLC',             '100754862300003', '0235:1007548623', 'FUJ'),
    (7, 'Palm Coast Electronics LLC',     '100863917200003', '0235:1008639172', 'DXB'),
    (8, 'Zayed Port Logistics LLC',       '100972845600003', '0235:1009728456', 'AUH');
SELECT setval('sellers_id_seq', 8);

-- Earlier batches: (import id, seller, file, invoice prefix, invoice count, created_at)
CREATE TEMP TABLE seed_batches (import_id UUID, seller_id BIGINT, filename TEXT, prefix TEXT, n INT, at TIMESTAMPTZ);
INSERT INTO seed_batches VALUES
    ('0c6f3a52-1f0e-4f43-9d7a-6a1f1b2f0a01', 1, 'desert-rose-oct-2026.xlsx',      'DRT-2610-', 6,  '2026-10-31 06:12:00+00'),
    ('0c6f3a52-1f0e-4f43-9d7a-6a1f1b2f0a02', 2, 'oasis-builders-oct-2026.xlsx',   'OBS-2610-', 8,  '2026-10-31 07:40:00+00'),
    ('0c6f3a52-1f0e-4f43-9d7a-6a1f1b2f0a03', 4, 'al-noor-fuel-oct-2026.xlsx',     'ANF-2610-', 4,  '2026-11-02 05:05:00+00'),
    ('0c6f3a52-1f0e-4f43-9d7a-6a1f1b2f0a04', 5, 'sahara-medical-batch-1029.xlsx', 'SMS-1029-', 12, '2026-10-29 10:20:00+00'),
    ('0c6f3a52-1f0e-4f43-9d7a-6a1f1b2f0a05', 7, 'palm-coast-oct-wk4.xlsx',        'PCE-2610-', 5,  '2026-10-30 08:00:00+00');

INSERT INTO imports (id, seller_id, filename, status, row_count, invoice_count, duration_ms, created_at, completed_at)
SELECT import_id, seller_id, filename, 'Completed', n * 2, n, 1800 + n * 450, at, at + interval '3 seconds'
FROM seed_batches;

INSERT INTO invoices (id, import_id, seller_id, invoice_number, buyer_name, issue_date, currency,
                      total_excl_vat, total_vat, total_incl_vat, status, payload, asp_message_id, asp_response,
                      created_at, updated_at)
SELECT gen_random_uuid(), b.import_id, b.seller_id,
       b.prefix || lpad(g::text, 3, '0'),
       (ARRAY['Emirates Palm Hospital LLC', 'Bluewater Facilities Management LLC', 'Al Waha Hypermarkets LLC'])[1 + g % 3],
       to_char(b.at, 'YYYY-MM-DD'), 'AED',
       1000 + g * 100, (1000 + g * 100) * 0.05, (1000 + g * 100) * 1.05,
       'Accepted',
       jsonb_build_object('invoiceNumber', b.prefix || lpad(g::text, 3, '0'), 'currency', 'AED',
                          'customizationId', 'urn:peppol:pint:billing-1@ae-1',
                          'totals', jsonb_build_object('taxExclusiveAmount', 1000 + g * 100,
                                                       'taxAmount', (1000 + g * 100) * 0.05,
                                                       'taxInclusiveAmount', (1000 + g * 100) * 1.05)),
       gen_random_uuid()::text,
       jsonb_build_object('status', 'accepted'),
       b.at, b.at + interval '20 seconds'
FROM seed_batches b, generate_series(1, b.n) AS g;

INSERT INTO invoice_events (invoice_id, status, detail, created_at)
SELECT i.id, s.status, s.detail, i.created_at + s.offset_s * interval '1 second'
FROM invoices i
CROSS JOIN (VALUES ('Parsed', 'Created from uploaded file', 1),
                   ('Validated', 'Passed portal validation', 2),
                   ('Submitted', 'Sent to ASP', 10),
                   ('Accepted', 'Accepted by ASP', 11)) AS s(status, detail, offset_s);
