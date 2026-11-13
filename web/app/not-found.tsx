import Link from 'next/link';

export default function NotFound() {
  return (
    <div className="card">
      <h1>Not found</h1>
      <p>We could not find that batch or invoice.</p>
      <Link href="/imports">Back to batches</Link>
    </div>
  );
}
