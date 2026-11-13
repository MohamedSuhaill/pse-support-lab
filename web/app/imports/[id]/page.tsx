import Link from 'next/link';
import ActionButton from '@/components/ActionButton';
import AutoRefresh from '@/components/AutoRefresh';
import ErrorPanel from '@/components/ErrorPanel';
import StatusBadge from '@/components/StatusBadge';
import { apiGet, load } from '@/lib/api';
import { formatDateTime, formatDuration, formatMoney } from '@/lib/format';
import { INVOICE_STATUSES, type Import, type SourceRows } from '@/lib/types';

export const dynamic = 'force-dynamic';

const MAX_SOURCE_ROWS = 200;

export default async function ImportPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const { data: imp, error } = await load(() => apiGet<Import>(`/imports/${id}`));
  if (error || !imp) return <ErrorPanel message={error ?? 'Unknown error'} />;
  const { data: source } = await load(() => apiGet<SourceRows>(`/imports/${id}/rows`));

  return (
    <>
      <AutoRefresh />
      <p>
        <Link href="/imports">← Batches</Link>
      </p>
      <div className="page-head">
        <h1>{imp.filename}</h1>
        <div className="row">
          <a className="button secondary" href={`/api/imports/${imp.id}/source`}>
            Download original file
          </a>
          <ActionButton url={`/api/imports/${imp.id}/retry`} label="Retry stuck invoices" />
        </div>
      </div>

      <div className="card facts">
        <div>
          <span>Company</span>
          {imp.sellerName}
        </div>
        <div>
          <span>Batch id</span>
          <code>{imp.id}</code>
        </div>
        <div>
          <span>Status</span>
          <StatusBadge status={imp.status} />
        </div>
        <div>
          <span>Uploaded</span>
          {formatDateTime(imp.createdAt)}
        </div>
        <div>
          <span>Completed</span>
          {formatDateTime(imp.completedAt)}
        </div>
        <div>
          <span>Processing time</span>
          {formatDuration(imp.durationMs)}
        </div>
        <div>
          <span>Rows / invoices</span>
          {imp.rowCount} / {imp.invoiceCount}
        </div>
        <div className="chips">
          <span>Invoice statuses</span>
          {INVOICE_STATUSES.filter((s) => imp.statusCounts[s] > 0).map((s) => (
            <StatusBadge key={s} status={s} count={imp.statusCounts[s]} />
          ))}
        </div>
      </div>
      {imp.error && <div className="banner error">{imp.error}</div>}

      <h2>Invoices</h2>
      <div className="card table-wrap">
        <table>
          <thead>
            <tr>
              <th>Invoice number</th>
              <th>Buyer</th>
              <th>Issue date</th>
              <th className="num">Total incl. VAT</th>
              <th>Status</th>
              <th>Rejected by</th>
              <th className="num">Errors</th>
              <th>ASP reference</th>
              <th>Last updated (GST)</th>
            </tr>
          </thead>
          <tbody>
            {(imp.invoices ?? []).map((inv) => (
              <tr key={inv.id}>
                <td>
                  <Link href={`/invoices/${inv.id}`}>{inv.invoiceNumber || '(no number)'}</Link>
                </td>
                <td>{inv.buyerName}</td>
                <td className="nowrap">{inv.issueDate}</td>
                <td className="num nowrap">{formatMoney(inv.totalInclVat, inv.currency)}</td>
                <td>
                  <StatusBadge status={inv.status} />
                </td>
                <td>{inv.rejectedBy === 'asp' ? 'ASP' : inv.rejectedBy === 'portal' ? 'Portal' : ''}</td>
                <td className="num">{inv.errorCount || ''}</td>
                <td>
                  <code>{inv.aspMessageId || '—'}</code>
                </td>
                <td className="nowrap">{formatDateTime(inv.updatedAt)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <h2>Source rows</h2>
      {!source?.available || !source.headers ? (
        <div className="card muted">Source preview is no longer available for this upload.</div>
      ) : (
        <div className="card table-wrap source">
          <table>
            <thead>
              <tr>
                <th>Row</th>
                {source.headers.map((h, i) => (
                  <th key={i} className="nowrap">
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {(source.rows ?? []).slice(0, MAX_SOURCE_ROWS).map((row, r) => (
                <tr key={r}>
                  <td className="muted">{r + 2}</td>
                  {source.headers!.map((_, c) => (
                    <td key={c} className="nowrap">
                      {row[c] ?? ''}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
          {(source.rows?.length ?? 0) > MAX_SOURCE_ROWS && (
            <p className="muted">
              Showing the first {MAX_SOURCE_ROWS} of {source.rows!.length} rows.
            </p>
          )}
        </div>
      )}
    </>
  );
}
