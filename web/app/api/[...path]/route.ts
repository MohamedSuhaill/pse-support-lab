import type { NextRequest } from 'next/server';
import { API_URL } from '@/lib/api';

export const dynamic = 'force-dynamic';

const UPSTREAM_TIMEOUT_MS = 60_000;

function log(entry: Record<string, unknown>) {
  console.log(JSON.stringify({ ts: new Date().toISOString(), msg: 'proxy', ...entry }));
}

async function proxy(req: NextRequest, ctx: { params: Promise<{ path: string[] }> }) {
  const { path } = await ctx.params;
  const target = `${API_URL}/v1/${path.map(encodeURIComponent).join('/')}${req.nextUrl.search}`;
  const started = Date.now();

  const headers = new Headers();
  const contentType = req.headers.get('content-type');
  if (contentType) headers.set('content-type', contentType);

  const init: RequestInit & { duplex?: 'half' } = {
    method: req.method,
    headers,
    cache: 'no-store',
    signal: AbortSignal.timeout(UPSTREAM_TIMEOUT_MS),
  };
  if (req.method !== 'GET' && req.method !== 'HEAD') {
    init.body = req.body;
    init.duplex = 'half';
  }

  try {
    const res = await fetch(target, init);
    log({ level: 'info', method: req.method, path: `/${path.join('/')}`, status: res.status, durationMs: Date.now() - started });
    const out = new Headers();
    for (const name of ['content-type', 'content-disposition']) {
      const value = res.headers.get(name);
      if (value) out.set(name, value);
    }
    return new Response(res.body, { status: res.status, headers: out });
  } catch (err) {
    const timedOut = err instanceof Error && (err.name === 'TimeoutError' || err.name === 'AbortError');
    const status = timedOut ? 504 : 502;
    log({
      level: 'error',
      method: req.method,
      path: `/${path.join('/')}`,
      status,
      durationMs: Date.now() - started,
      error: err instanceof Error ? `${err.name}: ${err.message}${err.cause ? ` (${String(err.cause)})` : ''}` : String(err),
    });
    return Response.json({ error: timedOut ? 'Gateway Timeout' : 'Bad Gateway' }, { status });
  }
}

export { proxy as GET, proxy as POST };
