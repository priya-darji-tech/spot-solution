# Spot Solutions - Enterprise Software Development & SharePoint Platform

A high-performance corporate website, CRM lead pipeline, and transactional notification system built in **Go (Golang)** with **Gin**, **PostgreSQL**, and server-rendered **Go HTML templates**, replicating the exact structure and content of [Spot Solutions](https://spotsolutions.com/).

Spot Solutions showcases custom software development, SharePoint/M365 consulting, enterprise AI & machine learning, mobile solutions, database architecture, and legacy system modernization.

---

## 1. System Architecture

```mermaid
graph TD
    Visitor[Visitor / Client Browser] -->|HTTPS| Nginx[Nginx Reverse Proxy / SSL]
    Nginx -->|Port 8080| GoApp[Go / Gin Web Application]
    
    subgraph Web & Middleware
        GoApp --> SecHeaders[Security Headers & CSP]
        SecHeaders --> RateLimit[Token Bucket Rate Limiter]
        RateLimit --> CSRF[CSRF Protection]
        CSRF --> SessionAuth[Session / RBAC Auth]
    end

    subgraph Service Layer
        SessionAuth --> PageHandler[Page Handler (SSR Go Templates)]
        SessionAuth --> ContactHandler[Contact Handler]
        SessionAuth --> LeadHandler[CRM Lead Handler]
        
        ContactHandler --> ContactService[Contact Service]
        LeadHandler --> LeadService[CRM Lead Service]
    end

    subgraph Transactional Persistence
        ContactService -->|1. Atomic BEGIN Tx| DB[(PostgreSQL 16)]
        ContactService -->|2. Insert Submission| DB
        ContactService -->|3. Record New Lead| DB
        ContactService -->|4. Queue Outbox Emails| DB
        ContactService -->|5. COMMIT Tx| DB
    end

    subgraph Email Delivery Outbox Pattern
        OutboxWorker[Background Outbox Worker] -->|Polls Pending / Retries| DB
        OutboxWorker -->|Exponential Backoff SMTP| MailProvider[SMTP Mail Server]
        MailProvider --> AdminNotification[Admin Team Alert]
        MailProvider --> CustomerReceipt[Customer Acknowledgment]
    end
```

---

## 2. Technology Stack

- **Backend**: Go 1.27+
- **HTTP Routing & Engine**: Gin Web Framework (`github.com/gin-gonic/gin`)
- **Database**: PostgreSQL 16 with connection pooling (`github.com/jackc/pgx/v5`)
- **Template Engine**: Go `html/template` (Server-Side Rendered for maximum SEO)
- **Security**:
  - Passwords hashed with `golang.org/x/crypto/bcrypt`
  - Encrypted, HTTP-only, SameSite session cookies
  - Double-submit CSRF tokens
  - IP-based token-bucket rate limiting (`golang.org/x/time/rate`)
  - Strict Content-Security-Policy & HSTS headers
- **Email Delivery**: SMTP via `gopkg.in/mail.v2` backed by a transactional `email_outbox` table with automatic retries

---

## 3. Key Pages & Features

1. **Public Marketing Pages**:
   - **Home (`/`)**: Hero banner, client logos, 8 core service cards, why choose us highlights, delivery process, testimonials, and footer.
   - **About Us (`/about`, `/about-us`)**: Company history, mission, core values, Canadian enterprise delivery track record.
   - **Services Catalog (`/services`)**: Full service listing with sidebar navigation.
   - **10 Core Service Detail Pages (`/services/:slug`)**:
     - [Custom Software Development](http://localhost:8080/services/custom-software) (`/services/custom-software`)
     - [SharePoint & Microsoft 365](http://localhost:8080/services/sharepoint) (`/services/sharepoint`)
     - [AI & Machine Learning](http://localhost:8080/services/ai) (`/services/ai`)
     - [Mobile Solutions](http://localhost:8080/services/mobile) (`/services/mobile`)
     - [Web Applications](http://localhost:8080/services/web-applications) (`/services/web-applications`)
     - [Requirements Analysis](http://localhost:8080/services/requirements) (`/services/requirements`)
     - [Change Management](http://localhost:8080/services/change-managment) (`/services/change-managment`)
     - [Databases](http://localhost:8080/services/database) (`/services/database`)
     - [Enterprise Hosting](http://localhost:8080/services/hosting) (`/services/hosting`)
     - [Legacy System Upgrades](http://localhost:8080/services/legacy-system-upgrades) (`/services/legacy-system-upgrades`)
   - **Industries (`/industries`)**: Specialized industry verticals (Healthcare, Public Sector, Energy).
   - **Why Choose Us (`/why-us`)**: 6 key competitive differentiators.
   - **Contact Us (`/contact`, `/contact-us`)**: Canadian office details (Vancouver & Calgary), phone `604-551-4640`, email `info@spotsolutions.com`, and inquiry form.
   - **Legal (`/privacy`, `/terms`)**: Privacy Policy and Terms & Conditions.
   - **404 Page**: Custom branded not-found view with navigation recovery links.
2. **SEO & Indexing**:
   - `/robots.txt` dynamic generation
   - `/sitemap.xml` with canonical links for all pages & services
   - JSON-LD Schema.org `ProfessionalService` structured data
   - Open Graph & Twitter card metadata
3. **CRM Lead Management**:
   - Role-Based Access Control (`admin`, `manager`, `viewer`)
   - Protected Admin Portal (`/admin/leads`)
   - RESTful API (`GET /api/leads`, `GET /api/leads/:id`, `PATCH /api/leads/:id/status`, `DELETE /api/leads/:id`)
   - Instant lead status transitions: `New` $\to$ `Contacted` $\to$ `Qualified` $\to$ `Converted` $\to$ `Lost`
4. **Reliable Outbox Email Pattern**:
   - Contact inquiries queue emails within the same atomic database transaction.
   - Zero lost emails if SMTP is temporarily unreachable.
   - Background worker processes pending emails and retries with exponential backoff up to 5 times.

---

## 4. Local Development Setup

### Prerequisites
- Go 1.22+ installed
- PostgreSQL 14+ running locally

### Step 1: Clone & Configure Environment
```bash
git clone <repository_url>
cd "spot solution"

# Copy example environment configuration
cp .env.example .env
```

Ensure `.env` matches your local PostgreSQL credentials:
```ini
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=spot_solution_db
```

### Step 2: Create Local Database
```bash
psql -U postgres -c "CREATE DATABASE spot_solution_db;"
```

### Step 3: Run the Application
The server automatically applies embedded schema migrations and seeds the default administrator on startup:
```bash
go run ./cmd/server
```

Open your browser to:
- Public Website: [http://localhost:8080](http://localhost:8080)
- Admin CRM Dashboard: [http://localhost:8080/admin/login](http://localhost:8080/admin/login)
  - **Email**: `admin@nexoradynamics.com`
  - **Password**: `AdminPass123!`

---

## 5. Running Automated Tests

Run the complete test suite covering authentication, RBAC, input validation, duplicate prevention, and CRM endpoints:
```bash
go test -v ./...
```

---

## 6. Hostinger VPS Production Deployment Guide

Deploying a Go application with PostgreSQL on Hostinger requires a **Hostinger KVM VPS** (running Ubuntu 22.04 LTS or 24.04 LTS). Standard shared web hosting accounts do not support long-running Go binary processes.

### Deployment Architecture

```mermaid
graph LR
    User[Internet Traffic] -->|Port 80/443| Nginx[Hostinger VPS: Nginx Reverse Proxy]
    Nginx -->|Let's Encrypt SSL| Certbot[Certbot TLS 1.3]
    Nginx -->|Proxy Pass 127.0.0.1:8080| Systemd[systemd: spot-solution.service]
    Systemd --> GoBinary[/opt/spot-solution/server]
    GoBinary --> Postgres[(PostgreSQL 16 Database)]
```

### Step-by-Step Server Setup on Hostinger VPS

#### 1. Prepare VPS & Install Dependencies
SSH into your Hostinger VPS:
```bash
ssh root@your_vps_ip

# Update packages and install PostgreSQL, Nginx, and Certbot
sudo apt update && sudo apt upgrade -y
sudo apt install -y postgresql postgresql-contrib nginx certbot python3-certbot-nginx
```

#### 2. Configure Production PostgreSQL
```bash
sudo -u postgres psql
```
In the PostgreSQL prompt:
```sql
CREATE DATABASE nexora_production_db;
CREATE USER nexora_user WITH ENCRYPTED PASSWORD 'StrongProductionPasswordHere!';
GRANT ALL PRIVILEGES ON DATABASE nexora_production_db TO nexora_user;
\q
```

#### 3. Build & Transfer Go Binary
On your local machine, cross-compile the binary for Linux AMD64:
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o server_linux ./cmd/server
```

Transfer files to your VPS:
```bash
ssh root@your_vps_ip "mkdir -p /opt/spot-solution/web"
scp server_linux root@your_vps_ip:/opt/spot-solution/server
scp -r web/static web/templates root@your_vps_ip:/opt/spot-solution/web/
scp .env.example root@your_vps_ip:/opt/spot-solution/.env
```

#### 4. Configure Production `.env` on VPS
Edit `/opt/spot-solution/.env`:
```ini
PORT=8080
APP_ENV=production
BASE_URL=https://yourdomain.com
APP_NAME="Nexora Dynamics"

DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=nexora_user
DB_PASSWORD=StrongProductionPasswordHere!
DB_NAME=nexora_production_db
DB_SSLMODE=disable

COOKIE_SECURE=true
SESSION_SECRET=GENERATE_RANDOM_32_BYTE_HEX_KEY
CSRF_SECRET=GENERATE_ANOTHER_RANDOM_32_BYTE_HEX_KEY

DEFAULT_ADMIN_EMAIL=admin@yourdomain.com
DEFAULT_ADMIN_PASSWORD=ComplexAdminPassword99!

SMTP_HOST=smtp.mailgun.org
SMTP_PORT=587
SMTP_USER=postmaster@yourdomain.com
SMTP_PASSWORD=your_production_smtp_password
SMTP_FROM_EMAIL=no-reply@yourdomain.com
NOTIFICATION_RECIPIENT_EMAIL=leads@yourdomain.com
```

#### 5. Configure `systemd` Managed Service
Create `/etc/systemd/system/spot-solution.service`:
```ini
[Unit]
Description=Nexora Dynamics Corporate Web Server & CRM
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/spot-solution
ExecStart=/opt/spot-solution/server
Restart=always
RestartSec=5s
EnvironmentFile=/opt/spot-solution/.env

# Security restrictions
LimitNOFILE=65535
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Enable and start the service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable spot-solution
sudo systemctl start spot-solution
sudo systemctl status spot-solution
```

#### 6. Configure Nginx Reverse Proxy
Create `/etc/nginx/sites-available/nexora`:
```nginx
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;

    client_max_body_size 2M;

    location /static/ {
        alias /opt/spot-solution/web/static/;
        expires 30d;
        add_header Cache-Control "public, no-transform";
    }

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Enable the site:
```bash
sudo ln -s /etc/nginx/sites-available/nexora /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

#### 7. Install SSL with Let's Encrypt
```bash
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com
```

Certbot will automatically install TLS certificates and configure HTTPS redirects.

---

## 7. Verifying Production Operations

1. **Verify Web Pages**: Navigate to `https://yourdomain.com` and test navigation across all 8 service detail views.
2. **Submit Contact Form**: Submit a test inquiry via the homepage or contact page.
3. **Verify Lead Creation**: Log into `https://yourdomain.com/admin/login` and check that the inquiry appears in the CRM dashboard with status `New`.
4. **Test Status Update**: Change status to `Contacted` or `Qualified` directly in the dropdown.
5. **Verify Email Outbox**: Check server journal logs:
   ```bash
   sudo journalctl -u spot-solution -f
   ```
   You will observe the outbox worker picking up pending notifications and delivering them.
