import Link from 'next/link';
import ActionButton from '@/components/ActionButton';
import AutoRefresh from '@/components/AutoRefresh';
import ErrorPanel from '@/components/ErrorPanel';
import StatusBadge from '@/components/StatusBadge';
import { apiGet, load } from '@/lib/api';
import { formatDateTime, formatMoney } from '@/lib/format';
import type { InvoiceDetail } from '@/lib/types';

export const dynamic = 'force-dynamic';

interface Payload {
  dueDate?: string;
  buyer?: { trn?: string; peppolId?: string };
  totals?: { taxExclusiveAmount?: number; taxAmount?: number; taxInclusiveAmount?: number };
}

export default async function InvoicePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const { data: inv, error } = await load(() => apiGet<InvoiceDetail>(`/invoices/${id}`));
  if (error || !inv) return <ErrorPanel message={error ?? 'Unknown error'} />;
  const payload = (inv.payload ?? {}) as Payload;
  const totals = payload.totals ?? {};
  const canRetry = inv.status === 'Validated' || inv.status === 'Failed';

  return (
    <>
      <AutoRefresh />
      <p>
        <Link href={`/imports/${inv.importId}`}>← Back to batch</Link>
      </p>
      <div className="page-head">
        <h1>Invoice {inv.invoiceNumber || '(no number)'}</h1>
        {canRetry && <ActionButton url={`/api/invoices/${inv.id}/retry`} label="Retry submission" />}
      </div>

      <div className="card facts">
        <div>
          <span>Company</span>
          {inv.sellerName}
        </div>
        <div>
          <span>Buyer</span>
          {inv.buyerName}
          {payload.buyer?.trn && <small className="muted"> TRN {payload.buyer.trn}</small>}
        </div>
        <div>
          <span>Buyer Peppol id</span>
          <code>{payload.buyer?.peppolId ?? '—'}</code>
        </div>
        <div>
          <span>Issue / due date</span>
          {inv.issueDate} / {payload.dueDate ?? '—'}
        </div>
        <div>
          <span>Total without VAT</span>
          {formatMoney(totals.taxExclusiveAmount ?? 0, inv.currency)}
        </div>
        <div>
          <span>VAT</span>
          {formatMoney(totals.taxAmount ?? 0, inv.currency)}
        </div>
        <div>
          <span>Total with VAT</span>
          {formatMoney(totals.taxInclusiveAmount ?? inv.totalInclVat, inv.currency)}
        </div>
        <div>
          <span>Status</span>
          <StatusBadge status={inv.status} />
          {inv.rejectedBy && <small className="muted"> by {inv.rejectedBy === 'asp' ? 'ASP' : 'portal validation'}</small>}
        </div>
        <div>
          <span>ASP reference</span>
          <code>{inv.aspMessageId || '—'}</code>
        </div>
        <div>
          <span>Last updated</span>
          {formatDateTime(inv.updatedAt)}
        </div>
      </div>

      <h2>Errors</h2>
      {inv.errors.length === 0 ? (
        <div className="card muted">No errors.</div>
      ) : (
        <div className="card table-wrap">
          <table>
            <thead>
              <tr>
                <th>Rule</th>
                <th>Field</th>
                <th className="num">Row</th>
                <th>Message</th>
              </tr>
            </thead>
            <tbody>
              {inv.errors.map((e, i) => (
                <tr key={i}>
                  <td className="nowrap">
                    <code>{e.rule}</code>
                  </td>
                  <td>{e.field}</td>
                  <td className="num">{e.row || ''}</td>
                  <td>{e.message}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <h2>ASP response</h2>
      {inv.aspResponse ? (
        <pre className="card json">{JSON.stringify(inv.aspResponse, null, 2)}</pre>
      ) : (
        <div className="card muted">No response recorded.</div>
      )}

      <h2>Status history</h2>
      <div className="card">
        <ol className="timeline">
          {inv.events.map((e, i) => (
            <li key={i}>
              <span className="muted nowrap">{formatDateTime(e.at)}</span> <StatusBadge status={e.status} /> {e.detail}
            </li>
          ))}
        </ol>
      </div>

      <h2>Payload</h2>
      <details className="card">
        <summary>PINT AE payload (JSON)</summary>
        <pre className="json">{JSON.stringify(inv.payload, null, 2)}</pre>
      </details>
    </>
  );
}
