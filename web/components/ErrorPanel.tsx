export default function ErrorPanel({ message }: { message: string }) {
  return (
    <div className="banner error">
      <strong>Something went wrong.</strong> {message}
    </div>
  );
}
