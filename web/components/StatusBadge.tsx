export default function StatusBadge({ status, count }: { status: string; count?: number }) {
  return (
    <span className={`badge badge-${status.toLowerCase()}`}>
      {status}
      {count !== undefined && <strong> {count}</strong>}
    </span>
  );
}
