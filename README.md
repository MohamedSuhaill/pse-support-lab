# Invoicely Lite — support lab

A small, fictional UAE e-invoicing portal that runs locally with Docker Compose. Read [`START_HERE.md`](START_HERE.md) first for the exercise itself. This file covers the technical side: running the system, where things are, and what to do when something will not start.

## Prerequisites

- Docker Desktop, or Docker Engine with Compose v2.
- At least **4 GB** of memory allotted to Docker.
- `make` and `curl`.
- Apple Silicon and x86 machines both work.
- The first build downloads about 2–3 GB of images and takes 5–10 minutes.

## Quick start

```bash
make up                         # build and start all services
open http://localhost:3000      # portal
open http://localhost:3001      # Grafana
```

Upload `samples/desert-rose-nov-2026.xlsx` for **Desert Rose Trading LLC** to see a normal, clean import.

## Services and ports

| Service | URL / host port | What it is |
|---|---|---|
| web | http://localhost:3000 | Portal UI (Next.js). Proxies `/api/*` to the api. |
| api | http://localhost:8080 | Invoicely Lite REST API (Go). Also `/metrics` and `/healthz`. |
| asp-mock | http://localhost:8090 | Simulated ASP / Peppol access point (Go). |
| postgres | localhost:5433 (container port 5432) | PostgreSQL 16. Database `invoicely`, user `invoicely`, password `invoicely`. |
| grafana | http://localhost:3001 | Dashboards and log search. Anonymous admin access. Dashboard: "Invoicely Lite – Overview". |
| prometheus | http://localhost:9090 | Metrics. |
| loki | internal (3100) | Logs from all containers, collected by Grafana Alloy. Label `service` = compose service name. |
| cadvisor | internal | Container CPU and memory metrics. |

## Make targets

| Command | What it does |
|---|---|
| `make up` | Build and start everything. |
| `make down` | Stop everything. |
| `make reset` | Wipe the database and volumes, then restart. |
| `make load` | Replay recent customer traffic: the four customer files plus a burst of peak-hour uploads from other customers. Takes about 3–4 minutes. |
| `make replay CUSTOMER=<name>` | Replay one customer's upload. `<name>` is one of `oasis`, `al-noor`, `sahara`, `ghaf`, `desert-rose`. |
| `make logs` | Tail the api and asp-mock logs. |
| `make ps` | Show container status. |
| `make psql` | Open a `psql` session on the `invoicely` database. |
| `make samples` | Regenerate the sample files in `samples/`. |

## Sample files

| File | Description |
|---|---|
| `samples/desert-rose-nov-2026.xlsx` | 5 invoices / 12 lines. Clean reference file; should reach Accepted. |
| `samples/oasis-builders-nov-2026.xlsx` | Monthly consolidated file: 8 invoices / 600 lines. |
| `samples/al-noor-fuel-nov-2026.xlsx` | 6 fuel invoices; unit prices per litre with 3 decimals. |
| `samples/sahara-medical-batch-1116.xlsx` | 40 invoices pushed by an ERP connector. |
| `samples/ghaf-cafe-nov-2026.xlsx` | 4 invoices, typed by hand. |

The file layout is described in [`docs/template-field-guide.md`](docs/template-field-guide.md).

## Uploading a file

**In the portal:** open http://localhost:3000, choose the company (seller), pick the `.xlsx` file and upload. The import page then shows the import status and each invoice with its own status.

**With curl:**

```bash
curl -F file=@samples/desert-rose-nov-2026.xlsx -F seller_id=1 http://localhost:8080/v1/imports
```

Seller ids are listed by `curl http://localhost:8080/v1/sellers`.

## API quick reference

All responses are JSON. Times are RFC3339 in UTC.

| Method and path | Purpose |
|---|---|
| `GET /healthz` | Health check. Returns `{"status":"ok"}`. |
| `GET /v1/sellers` | List of companies (sellers). |
| `POST /v1/imports` | Upload a file. Multipart form fields `file` (.xlsx) and `seller_id` (int). Returns `201` with the import once the file is parsed and validated; `400` with `{"error":"..."}` for a bad form or file. |
| `GET /v1/imports?limit=50` | Recent imports, newest first. |
| `GET /v1/imports/{id}` | One import, with its invoices. |
| `GET /v1/imports/{id}/rows` | Source rows of the uploaded sheet, as shown on the batch page (`available:false` when the preview is no longer available). |
| `GET /v1/imports/{id}/source` | Download the original uploaded file. |
| `POST /v1/imports/{id}/retry` | Retry submission for invoices of that import in status Validated or Failed. |
| `GET /v1/invoices/{id}` | One invoice: payload, errors, ASP response and status history (events). |
| `POST /v1/invoices/{id}/retry` | Retry submission for one invoice (`409` if its status is not Validated or Failed). |

From the browser, the same API is reachable through the portal at `http://localhost:3000/api/...`.

## Logs and metrics

- **Grafana dashboard:** http://localhost:3001 → Dashboards → "Invoicely Lite – Overview".
- **Logs:** Grafana → Explore → choose the Loki data source. Example query: `{service="api"} | json`. Use `{service="asp-mock"}`, `{service="web"}` and so on for other services.
- **Metrics:** Prometheus at http://localhost:9090, or Grafana → Explore → Prometheus. Scrape targets: http://localhost:9090/targets.
- **Terminal:** `make logs`, or `docker compose logs <service>` (for example `docker compose logs api`).
- **Database:** `make psql`.

More detail is in [`docs/architecture.md`](docs/architecture.md).

## Repository layout

| Path | Contents |
|---|---|
| `web/` | Portal UI (Next.js). |
| `services/api` | Invoicely Lite API (Go). |
| `services/asp-mock` | Simulated ASP (Go). |
| `tools/` | Sample file generator and traffic replay. |
| `db/` | Database schema and seed data. |
| `observability/` | Prometheus, Grafana, Loki and Alloy configuration. |
| `samples/` | Sample Excel files. |
| `inbox/` | Customer emails and alerts for the exercise. |
| `docs/` | Product and support documentation. |

## Troubleshooting

| Problem | What to try |
|---|---|
| "Port is already allocated" / port already in use | Stop the other process using that port, or change the host port mapping in `docker-compose.yml`. |
| The first build is killed, or containers never become healthy | Check Docker has at least 4 GB of memory (Docker Desktop → Settings → Resources). |
| First `make up` is very slow | Expected: images are downloaded and built once (5–10 minutes). Later starts are quick. |
| Apple Silicon | Supported; no extra steps. |
| Things are in a strange state | `make reset` wipes the database and volumes and starts clean. |
| Grafana shows no data | Wait about 30 seconds after start-up, then refresh. Check that targets are UP at http://localhost:9090/targets. Make sure the time range in Grafana covers the period you are looking at. |
| A service is not responding | `make ps` to see its state, then `docker compose logs <service>`. |
