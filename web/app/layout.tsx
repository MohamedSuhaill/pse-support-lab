import type { Metadata } from 'next';
import Link from 'next/link';
import { GRAFANA_URL } from '@/lib/api';
import './globals.css';

export const metadata: Metadata = {
  title: 'Invoicely Lite',
  description: 'UAE e-invoicing portal',
};

export const dynamic = 'force-dynamic';

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <header className="topbar">
          <Link href="/imports" className="brand">
            Invoicely <span>Lite</span>
          </Link>
          <nav>
            <Link href="/upload">Upload</Link>
            <Link href="/imports">Batches</Link>
            <a href={GRAFANA_URL} target="_blank" rel="noreferrer">
              Monitoring ↗
            </a>
          </nav>
        </header>
        <main className="container">{children}</main>
      </body>
    </html>
  );
}
