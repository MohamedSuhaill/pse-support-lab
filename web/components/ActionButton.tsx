'use client';

import { useRouter } from 'next/navigation';
import { useState } from 'react';

export default function ActionButton({ url, label }: { url: string; label: string }) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<{ ok: boolean; text: string } | null>(null);

  async function run() {
    setBusy(true);
    setMessage(null);
    try {
      const res = await fetch(url, { method: 'POST' });
      const body = await res.json().catch(() => ({}));
      if (res.ok) {
        setMessage({ ok: true, text: `Requeued ${body.requeued ?? 0} invoice(s)` });
        router.refresh();
      } else {
        setMessage({ ok: false, text: body.error ?? `Request failed (${res.status})` });
      }
    } catch {
      setMessage({ ok: false, text: 'Request failed. Please try again.' });
    } finally {
      setBusy(false);
    }
  }

  return (
    <span className="action">
      <button className="button secondary" onClick={run} disabled={busy}>
        {busy ? 'Working…' : label}
      </button>
      {message && <span className={message.ok ? 'note ok' : 'note err'}>{message.text}</span>}
    </span>
  );
}
