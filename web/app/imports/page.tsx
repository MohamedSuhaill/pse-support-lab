import Link from 'next/link';
import AutoRefresh from '@/components/AutoRefresh';
import ErrorPanel from '@/components/ErrorPanel';
import StatusBadge from '@/components/StatusBadge';
import { apiGet, load } from '@/lib/api';
import { formatDateTime, formatDuration } from '@/lib/format';
import { INVOICE_STATUSES, type Import } from '@/lib/types';

export const dynamic = 'force-dynamic';

export default async function ImportsPage() {
  const { data: imports, error } = await load(() => apiGet<Import[]>('/imports?limit=50'));
  return (
    <>
      <AutoRefresh />
      <div className="page-head">
        <h1>Batches</h1>
        <Link className="button" href="/upload">
          Upload file
        </Link>
      </div>
      {error ? (
        <ErrorPanel message={error} />
      ) : imports && imports.length === 0 ? (
        <div className="card muted">No uploads yet.</div>
      ) : (
        <div className="card table-wrap">
          <table>
            <thead>
              <tr>
                <th>Uploaded at (GST)</th>
                <th>Company</th>
                <th>File</th>
                <th>Batch status</th>
                <th className="num">Rows</th>
                <th className="num">Invoices</th>
                <th>Invoice statuses</th>
                <th className="num">Processing time</th>
              </tr>
            </thead>
            <tbody>
              {imports?.map((imp) => (
                <tr key={imp.id}>
                  <td className="nowrap">
                    <Link href={`/imports/${imp.id}`}>{formatDateTime(imp.createdAt)}</Link>
                  </td>
                  <td>{imp.sellerName}</td>
                  <td>
                    <Link href={`/imports/${imp.id}`}>{imp.filename}</Link>
                  </td>
                  <td>
                    <StatusBadge status={imp.status} />
                  </td>
                  <td className="num">{imp.rowCount}</td>
                  <td className="num">{imp.invoiceCount}</td>
                  <td className="chips">
                    {INVOICE_STATUSES.filter((s) => imp.statusCounts[s] > 0).map((s) => (
                      <StatusBadge key={s} status={s} count={imp.statusCounts[s]} />
                    ))}
                  </td>
                  <td className="num">{formatDuration(imp.durationMs)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </>
  );
}
