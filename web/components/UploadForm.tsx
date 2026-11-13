'use client';

import { useRouter } from 'next/navigation';
import { useEffect, useRef, useState } from 'react';
import type { Seller } from '@/lib/types';

function errorMessage(status: number, body: { error?: string }): string {
  if (status === 504) return 'The upload did not complete in time (504 Gateway Timeout). Please try again.';
  if (status === 502) return 'The service is temporarily unavailable (502 Bad Gateway). Please try again.';
  return body.error ?? `Upload failed (HTTP ${status}).`;
}

export default function UploadForm({ sellers }: { sellers: Seller[] }) {
  const router = useRouter();
  const input = useRef<HTMLInputElement>(null);
  const [sellerId, setSellerId] = useState(sellers[0]?.id.toString() ?? '');
  const [file, setFile] = useState<File | null>(null);
  const [dragging, setDragging] = useState(false);
  const [busy, setBusy] = useState(false);
  const [elapsed, setElapsed] = useState(0);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!busy) return;
    const started = Date.now();
    const timer = setInterval(() => setElapsed(Math.floor((Date.now() - started) / 1000)), 1000);
    return () => clearInterval(timer);
  }, [busy]);

  function pick(files: FileList | null) {
    const f = files?.[0];
    if (!f) return;
    if (!f.name.toLowerCase().endsWith('.xlsx')) {
      setError('Please choose an .xlsx file.');
      return;
    }
    setError(null);
    setFile(f);
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!file || !sellerId) return;
    setBusy(true);
    setElapsed(0);
    setError(null);
    const form = new FormData();
    form.append('seller_id', sellerId);
    form.append('file', file);
    try {
      const res = await fetch('/api/imports', { method: 'POST', body: form });
      const body = await res.json().catch(() => ({}));
      if (res.status === 201 && body.id) {
        router.push(`/imports/${body.id}`);
        return;
      }
      setError(errorMessage(res.status, body));
    } catch {
      setError('The upload could not be sent. Check your connection and try again.');
    }
    setBusy(false);
  }

  return (
    <form className="card upload" onSubmit={submit}>
      <label className="field">
        <span>Company</span>
        <select value={sellerId} onChange={(e) => setSellerId(e.target.value)} disabled={busy}>
          {sellers.map((s) => (
            <option key={s.id} value={s.id}>
              {s.name} — TRN {s.trn}
            </option>
          ))}
        </select>
      </label>

      <div
        className={`dropzone${dragging ? ' dragging' : ''}`}
        onClick={() => !busy && input.current?.click()}
        onDragOver={(e) => {
          e.preventDefault();
          setDragging(true);
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={(e) => {
          e.preventDefault();
          setDragging(false);
          if (!busy) pick(e.dataTransfer.files);
        }}
      >
        <input ref={input} type="file" accept=".xlsx" hidden onChange={(e) => pick(e.target.files)} />
        {file ? (
          <>
            <strong>{file.name}</strong>
            <span className="muted">{(file.size / 1024).toFixed(0)} KB — click to choose another file</span>
          </>
        ) : (
          <>
            <strong>Drop your .xlsx file here</strong>
            <span className="muted">or click to browse</span>
          </>
        )}
      </div>

      {error && <div className="banner error">{error}</div>}

      <div className="row">
        <button className="button" type="submit" disabled={!file || busy}>
          Upload
        </button>
        {busy && (
          <span className="processing">
            <span className="spinner" /> Processing your file… {elapsed}s
          </span>
        )}
      </div>
    </form>
  );
}
