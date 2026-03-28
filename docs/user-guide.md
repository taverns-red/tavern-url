# Tavern URL — User Guide

> Get started with creating, managing, and tracking your short links.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Creating Links](#creating-links)
3. [Managing Links](#managing-links)
4. [Analytics](#analytics)
5. [QR Codes](#qr-codes)
6. [API Keys](#api-keys)
7. [Organizations](#organizations)
8. [Browser Extension](#browser-extension)
9. [Dark Mode](#dark-mode)

---

## Getting Started

### Create an Account

1. Visit your Tavern URL instance (e.g., `https://go.yourorg.org`)
2. Click **Sign up** in the header
3. Fill in your name, email, and password
4. Click **Create account**

You'll be redirected to your dashboard.

### Log In with Google

If your instance has Google OAuth enabled, click **Continue with Google** on the login page. An account is created automatically on first use.

---

## Creating Links

### From the Dashboard

1. Click **+ New Link** at the top of the dashboard
2. Enter the **Destination URL** — the full URL you want to shorten
3. Optionally enter a **Custom slug** (e.g., `spring-gala`) — must be 3–64 characters, letters, numbers, and hyphens only
4. Click **Create**

The link appears in your list. If you leave the slug blank, a random 6-character slug is generated automatically.

### Bulk Creation

Use the API to create up to 100 links at once:

```bash
curl -X POST https://go.yourorg.org/api/v1/links/bulk \
  -H "Authorization: Bearer tvn_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{"urls": ["https://example.com/page1", "https://example.com/page2"]}'
```

---

## Managing Links

### View Your Links

Your dashboard shows all your links with:
- The short URL (clickable, opens in a new tab)
- The original destination URL
- Creation date
- **Copy** button to copy the short URL to your clipboard
- **Delete** button to remove the link

### Edit a Link

Use the API to update a link's destination URL or slug:

```bash
curl -X PUT https://go.yourorg.org/api/v1/links/42 \
  -H "Authorization: Bearer tvn_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://new-destination.org", "slug": "new-slug"}'
```

### Delete a Link

Click the **Delete** button next to any link. You'll be asked to confirm before the link is removed. Deleted links return a 404 — visitors will no longer be redirected.

### Search Links

Use the search box on the dashboard or the `?q=` parameter with the API:

```bash
curl https://go.yourorg.org/api/v1/links?q=gala \
  -H "Authorization: Bearer tvn_your_api_key"
```

---

## Analytics

Click on any link in your dashboard to view its detail page with aggregate analytics.

### What's Tracked

| Metric | Detail |
|--------|--------|
| **Total clicks** | Lifetime count |
| **Clicks over time** | Daily bar chart (7, 30, 90, or 365 day view) |
| **Countries** | Country-level geo from request IP (IP discarded after lookup) |
| **Devices** | Mobile, Desktop, Tablet, Bot |
| **Referrers** | Referring domain (e.g., "twitter.com", "direct") |

### Privacy Guarantee

- **No cookies** are set on redirect visitors
- **No PII** is stored — IPs are used ephemerally for geo lookup and discarded
- **User-Agent** is parsed to a device category, raw string is never stored
- **Referrer URLs** are truncated to domain only

### CSV Export

Download analytics data as CSV for board reports:

```bash
curl https://go.yourorg.org/api/v1/links/42/analytics/export \
  -H "Authorization: Bearer tvn_your_api_key" \
  -o analytics.csv
```

---

## QR Codes

Every link gets a QR code for print materials (flyers, posters, event programs).

### Download a QR Code

- **From the link detail page:** Click the **Download QR** button
- **Via the API:**

```bash
# Default size (256×256 px)
curl https://go.yourorg.org/api/v1/links/42/qr -o qr.png

# Custom size (512×512 px)
curl https://go.yourorg.org/api/v1/links/42/qr?size=512 -o qr.png
```

The QR code is returned as a PNG image.

---

## API Keys

API keys let you automate link creation from scripts, CI/CD pipelines, or other tools.

### Create an API Key

Use the API:

```bash
curl -X POST https://go.yourorg.org/api/v1/keys \
  -H "Cookie: session=your_session_cookie" \
  -H "Content-Type: application/json" \
  -d '{"name": "CI Pipeline"}'
```

The response includes your key (format: `tvn_...`). **Save it immediately** — the raw key is only shown once.

### Use an API Key

Include your key in the `Authorization` header:

```bash
curl https://go.yourorg.org/api/v1/links \
  -H "Authorization: Bearer tvn_your_api_key"
```

### Security

- Keys are **SHA-256 hashed** in the database — even a database breach won't expose raw keys
- Keys are **scoped to your user** — they can only access your links and organizations
- **Rotate regularly** — delete old keys and create new ones

---

## Organizations

Organizations let teams share links and analytics.

### Create an Organization

```bash
curl -X POST https://go.yourorg.org/api/v1/orgs \
  -H "Authorization: Bearer tvn_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{"name": "Habitat for Humanity", "slug": "habitat"}'
```

### Invite Members

```bash
curl -X POST https://go.yourorg.org/api/v1/orgs/habitat/invite \
  -H "Authorization: Bearer tvn_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{"email": "volunteer@habitat.org", "role": "member"}'
```

### Roles

| Role | Permissions |
|------|-------------|
| **Owner** | Full control, manage members, delete org |
| **Admin** | Create/edit/delete links, invite members |
| **Member** | Create links, view analytics |

---

## Browser Extension

The Chrome extension lets you shorten any URL with one click.

### Setup

1. Open `chrome://extensions`
2. Enable **Developer mode** (toggle in top-right)
3. Click **Load unpacked** and select the `extension/` directory from the Tavern URL repo
4. Click the Tavern URL extension icon in your toolbar
5. Enter your **instance URL** and **API key**, then click **Save**

### Usage

On any webpage, click the extension icon and then **Shorten**. The short URL is copied to your clipboard.

---

## Dark Mode

Tavern URL automatically matches your system theme. To switch manually:

1. Look for the theme toggle in the UI
2. Your preference is saved in your browser and persists across sessions

Dark mode uses a carefully tuned color palette designed for readability and reduced eye strain.
