import ErrorPanel from '@/components/ErrorPanel';
import UploadForm from '@/components/UploadForm';
import { apiGet, load } from '@/lib/api';
import type { Seller } from '@/lib/types';

export const dynamic = 'force-dynamic';

export default async function UploadPage() {
  const { data: sellers, error } = await load(() => apiGet<Seller[]>('/sellers'));
  return (
    <>
      <h1>Upload invoices</h1>
      <p className="muted">Files must follow the Invoicely UAE template (sheet &apos;Invoices&apos;).</p>
      {error ? <ErrorPanel message={error} /> : <UploadForm sellers={sellers ?? []} />}
    </>
  );
}
