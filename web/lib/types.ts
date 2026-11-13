export type InvoiceStatus = 'Parsed' | 'Validated' | 'Submitted' | 'Accepted' | 'Rejected' | 'Failed';
export type ImportStatus = 'Uploaded' | 'Processing' | 'Completed' | 'Failed';

export interface Seller {
  id: number;
  name: string;
  trn: string;
  peppolId: string;
  emirate: string;
}

export interface InvoiceSummary {
  id: string;
  invoiceNumber: string;
  buyerName: string;
  issueDate: string;
  totalInclVat: number;
  currency: string;
  status: InvoiceStatus;
  rejectedBy: 'portal' | 'asp' | '';
  errorCount: number;
  aspMessageId: string;
  updatedAt: string;
}

export interface Import {
  id: string;
  sellerId: number;
  sellerName: string;
  filename: string;
  status: ImportStatus;
  error: string;
  rowCount: number;
  invoiceCount: number;
  durationMs: number | null;
  createdAt: string;
  completedAt: string | null;
  statusCounts: Record<InvoiceStatus, number>;
  invoices?: InvoiceSummary[];
}

export interface ValidationError {
  rule: string;
  field: string;
  message: string;
  row: number;
}

export interface InvoiceEvent {
  status: string;
  detail: string;
  at: string;
}

export interface InvoiceDetail extends InvoiceSummary {
  importId: string;
  sellerName: string;
  payload: unknown;
  errors: ValidationError[];
  aspResponse: unknown;
  events: InvoiceEvent[];
}

export interface SourceRows {
  available: boolean;
  headers?: string[];
  rows?: string[][];
}

export const INVOICE_STATUSES: InvoiceStatus[] = ['Parsed', 'Validated', 'Submitted', 'Accepted', 'Rejected', 'Failed'];
