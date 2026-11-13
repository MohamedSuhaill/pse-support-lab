import { notFound } from 'next/navigation';

export const API_URL = process.env.API_URL ?? 'http://api:8080';
export const GRAFANA_URL = process.env.GRAFANA_URL ?? 'http://localhost:3001';

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
  }
}

/** Server-side GET against the Invoicely API. Calls notFound() on 404. */
export async function apiGet<T>(path: string): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${API_URL}/v1${path}`, { cache: 'no-store', signal: AbortSignal.timeout(15_000) });
  } catch {
    throw new ApiError(502, 'Could not reach the Invoicely API (502 Bad Gateway).');
  }
  if (res.status === 404) notFound();
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError(res.status, `The Invoicely API returned ${res.status}: ${body.error ?? res.statusText}`);
  }
  return (await res.json()) as T;
}

/** Runs a loader and converts API failures into a renderable error message. */
export async function load<T>(fn: () => Promise<T>): Promise<{ data: T; error: null } | { data: null; error: string }> {
  try {
    return { data: await fn(), error: null };
  } catch (err) {
    if (err instanceof ApiError) return { data: null, error: err.message };
    throw err;
  }
}
