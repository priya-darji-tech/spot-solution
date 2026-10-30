# AGENTS.md — Spot Solutions

Guidance for AI coding agents working in this repository.

## What this project is

Corporate marketing site + CRM lead pipeline for [Spot Solutions](https://spotsolutions.com/), rebuilt in **Go / Gin** with **PostgreSQL** and **server-rendered Go HTML templates**. Content and IA should stay aligned with the live site whenever possible.

## Stack

| Layer | Choice |
|---|---|
| Language | Go (`go.mod` → module `spot-solution`) |
| HTTP | Gin |
| DB | PostgreSQL (`pgx/v5`), migrations + admin seed on startup |
| UI | Go `html/template` under `web/templates/` |
| Assets | `web/static/` (CSS, JS, images) |
| Email | SMTP + transactional `email_outbox` worker |

## How to run

```bash
cp .env.example .env   # if needed; configure DB_* or DATABASE_URL
psql -U postgres -c "CREATE DATABASE spot_solution_db;"
go run ./cmd/server
```

- Site: http://localhost:8080  
- Admin: http://localhost:8080/admin/login (seeded via `DEFAULT_ADMIN_*` in `.env`)  
- Tests: `go test -v ./...`

Do not commit secrets from `.env`. Prefer `.env.example` for documenting vars.

## Layout

```
cmd/server/main.go     # entry: config, DB, middleware, routes, outbox worker
internal/
  config/              # env loading
  database/            # connection, migrations, seed
  handlers/            # HTTP handlers (pages, contact, leads, auth, SEO)
  middleware/          # security headers, CSRF, rate limit, auth
  models/
  repository/
  services/            # contact, CRM leads, auth
  email/               # mailer + outbox worker
web/
  templates/
    layouts/           # header, footer
    pages/             # public SSR pages (incl. service_detail.html)
    admin/
    emails/
  static/css|js|images
```

Templates load via `r.LoadHTMLGlob("web/templates/**/*")`. Static files: `/static` → `./web/static`.

## Service pages (important)

- Catalog + metadata: `internal/handlers/page_handler.go` → `ServicesCatalog`
- Rich HTML: `web/templates/pages/service_detail.html` (slug-specific `{{if eq .Service.Slug "..."}}` branches)
- Nav submenu: `web/templates/layouts/header.html`

SharePoint submenu URLs (match live spotsolutions.com):

| Menu label | Local path |
|---|---|
| SharePoint Migration | `/services/sharepoint-migration-2` |
| Microsoft 365 Experts | `/services/miscrosoft-365-experts` (typo slug is intentional — matches live) |
| Power Platform | `/services/powerplatform` |
| SharePoint Intranet | `/services/intranet` |
| SharePoint Custom Development | `/services/sharepoint-custom-development` |
| SharePoint Support | `/services/sharepoint-support` |

When adding or updating a service page:

1. Add/update the catalog entry in `ServicesCatalog` (slug must match the URL).
2. Add a dedicated content branch in `service_detail.html` — do not leave SharePoint sub-pages on the generic fallback.
3. Wire the header submenu link to that slug.
4. Prefer copying structure/copy from the corresponding live URL when the user asks for content parity.

## Design / UI

Follow `DESIGN.md` and CSS variables in `web/static/css/main.css`:

- Navy (`#07172c`) + accent blue (`#0084ff`) / cyan (`#00b4d8`)
- Headings: Montserrat; body fonts via existing CSS vars
- Prefer existing layout patterns (hero bar, sidebar + content for generic services, full custom sections for flagship pages like `custom-software` / `sharepoint`)

Do not invent a new purple/cream AI-default look. Stay on-brand with Spot.

## Architecture conventions

- Handlers stay thin; business logic in `internal/services/`.
- Contact → lead + email queue happens in one DB transaction (outbox pattern). Do not send SMTP inline in the request path.
- Auth: session cookies + RBAC (`admin` / `manager` / `viewer`). Protect admin/CRM routes.
- CSRF is middleware-wide; forms need `csrf_token` / `{{.CSRFToken}}`.
- Prefer aliases for live URL typos/legacy slugs (e.g. both `sharepoint-migration` and `sharepoint-migration-2`) rather than breaking existing links.

## What agents should / should not do

**Do**

- Keep public marketing content faithful to spotsolutions.com when asked
- Update both Go catalog data and HTML templates for service pages
- Run `go build ./cmd/server` or `go test ./...` after non-trivial Go changes
- Reuse existing CSS classes and template partials (`header` / `footer`)

**Don't**

- Add a Node/React frontend unless explicitly requested — this app is SSR Go templates
- Put secrets, production passwords, or real SMTP credentials in docs or commits
- “Fix” the live slug typo `miscrosoft-365-experts` without an explicit redirect plan
- Expand scope into unrelated redesigns when the task is content or routing

## Key docs

- `README.md` — setup, architecture, deployment
- `DESIGN.md` — brand, tokens, UI patterns
- `.env.example` — required configuration keys
