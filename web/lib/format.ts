const dateTime = new Intl.DateTimeFormat('sv-SE', {
  timeZone: 'Asia/Dubai',
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit',
});

export function formatDateTime(value: string | null | undefined): string {
  if (!value) return '—';
  return dateTime.format(new Date(value));
}

export function formatDuration(ms: number | null | undefined): string {
  if (ms === null || ms === undefined) return '—';
  return `${(ms / 1000).toFixed(1)} s`;
}

export function formatMoney(amount: number, currency = 'AED'): string {
  return `${currency || 'AED'} ${amount.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}
