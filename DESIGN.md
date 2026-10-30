# Spot Solutions - UI/UX Design System & Specification

This document details the visual design system, component architecture, layout rules, and accessibility standards implemented for the **Spot Solutions** web application.

---

## 1. Brand Identity & Design Direction

- **Brand Character:** Modern Enterprise, Trustworthy, Technical Precision, High-Performance B2B Consulting.
- **Visual Style:** Deep Navy paired with Electric Cyan and Mint accents, featuring subtle glassmorphism (`backdrop-filter: blur(14px)`), clean elevation shadows, and sharp typography.
- **Reference Inspiration:** Replicates the corporate structure of [Spot Solutions](https://spotsolutions.com/) with modern aesthetic standards.

---

## 2. Color Palette (Actual CSS Variables)

Source: [`web/static/css/main.css`](web/static/css/main.css#L6-L43)

### Primary Theme & Surface Colors
| CSS Variable | Hex Value | Role & Usage |
|---|---|---|
| `--color-primary-navy` | `#07172c` | Main dark background, header background, footer background |
| `--color-surface-navy` | `#0c2340` | Elevated dark container surfaces, dark card backgrounds |
| `--spot-dark` | `#0f172a` | Primary heading text color (`h1`–`h6`), high-contrast labels |
| `--spot-black` | `#07172c` | Brand black alias |

### Accent & Feedback Colors
| CSS Variable | Hex Value | Role & Usage |
|---|---|---|
| `--color-accent-blue` | `#0084ff` | Primary CTA buttons, hyperlink default color, logo brand mark |
| `--color-accent-blue-hover` | `#006ed6` | Hover state for primary buttons and interactive links |
| `--color-accent-cyan` | `#00b4d8` | Electric Cyan accent, badge highlights, active tab indicator |
| `--color-cyan-glow` | `#38bdf8` | Focus ring glow, gradient stops, secondary accents |
| `--color-mint` | `#10b981` | Success states, qualified lead pills, live status tags |
| `--spot-green` | `#10b981` | Success accent alias |
| `--spot-green-border` | `#34d399` | Success alert borders |

### Neutral & Background Colors
| CSS Variable | Hex Value | Role & Usage |
|---|---|---|
| `--spot-text` | `#475569` | Body text color (WCAG AA compliant against light backgrounds) |
| `--spot-light-bg` | `#f8fafc` | Alternate section background, table header background |
| `--spot-gray-bg` | `#f1f5f9` | Admin CRM page background, modal backdrop color |
| `--spot-border` | `#e2e8f0` | Standard card border, divider lines, table row borders |
| `#ffffff` | `#ffffff` | Primary content card surface, modal dialog background |

---

## 3. Typography

Source: [`web/static/css/main.css`](web/static/css/main.css#L28-L33)

- **Headings Font:** `'Montserrat', sans-serif` (`--font-heading`, `--font-montserrat`)
  - Weight: 700 (Bold) / 800 (Extra Bold)
  - Line-height: `1.3`
  - Sizes:
    - Hero H1: `2.75rem` – `3.25rem`
    - Section H2: `2.0rem` – `2.25rem`
    - Card H3: `1.25rem` – `1.4rem`
    - Subsection H4: `1.0rem` – `1.15rem`
- **Body Font:** `'Open Sans', 'Inter', -apple-system, BlinkMacSystemFont, sans-serif` (`--font-body`)
  - Weight: 400 (Regular) / 500 (Medium) / 600 (Semi-Bold)
  - Size: `15px` (`1rem` base on `html`)
  - Line-height: `1.7`
- **Secondary Display Font:** `'Oswald', sans-serif` (`--font-oswald`)

---

## 4. Layout, Spacing & Breakpoints

### Container & Elevation
- **Container Max-Width:** `1240px` with `padding: 0 20px` (`.container`).
- **Border Radii:**
  - Small: `--radius-sm: 8px` (inputs, buttons, pills)
  - Medium: `--radius-md: 14px` (cards, dialogs)
  - Large: `--radius-lg: 20px` (hero banners, section containers)
  - Pill: `--radius-full: 9999px` (status badges, category pills)
- **Shadow Tokens:**
  - `--shadow-sm: 0 2px 8px rgba(15, 23, 42, 0.04)`
  - `--shadow-md: 0 10px 25px -5px rgba(15, 23, 42, 0.06), 0 4px 6px -2px rgba(15, 23, 42, 0.03)`
  - `--shadow-lg: 0 20px 40px -10px rgba(0, 132, 255, 0.12)`
  - `--shadow-hover: 0 16px 36px rgba(15, 23, 42, 0.09)`

### Responsive Breakpoints
- **Desktop Extra Wide (`> 1200px`):** Full 4-column service grids, sticky sidebar navigation.
- **Standard Desktop (`992px` – `1199px`):** 3-column service cards, 2-column testimonial cards.
- **Tablet (`768px` – `991px`):**
  - Navigation switches to mobile hamburger drawer (`.nav-toggle`).
  - Table horizontally scrolls inside `.crm-table-container`.
  - 2-column form grids collapse to single columns.
- **Mobile (`< 600px`):**
  - Hero header downscales to `1.85rem`.
  - Full-width stacked buttons.
  - Search and filter toolbars wrap into vertical flow.

---

## 5. UI Component Specifications

### 1. Header & Navigation (`.header-corporate`)
- Semi-transparent white surface (`rgba(255, 255, 255, 0.95)`) with `backdrop-filter: blur(14px)`.
- Sticky positioning (`position: sticky; top: 0; z-index: 1000`).
- Multi-level dropdown menus for 10 service categories with hover transitions.
- High-visibility "Get a Free Consultation" CTA button.

### 2. Buttons (`.btn`)
- `.btn-primary`: Solid blue (`#0084ff`), white text, `border-radius: 8px`, hover transition to `#006ed6`.
- `.btn-outline`: Transparent surface, 1.5px solid border, hover fill transition.
- Micro-interactions: `transition: all 0.25s ease`, transform lift on hover (`translateY(-2px)`).

### 3. Cards
- `.service-card`: White background, 1px solid `#e2e8f0`, `--radius-md: 14px`, padding `28px`. Hover elevates with `--shadow-hover` and top border accent `#00b4d8`.
- `.feature-card`: Icon badge with cyan glow background, heading, and body copy.

### 4. Forms (`.ajax-contact-form`)
- Input fields (`.form-control`): 1px solid `#cbd5e1`, padding `10px 14px`, `--radius-sm: 8px`.
- Focus state: Outline border transitions to `--color-accent-cyan: #00b4d8` with `box-shadow: 0 0 0 3px rgba(0, 180, 216, 0.15)`.
- State alerts: `.alert-success-msg` (green `#10b981`), `.alert-error-msg` (red `#ef4444`).

### 5. CRM Management Components
- **Filter Toolbar (`.crm-controls-bar`):** Status pills (`.crm-status-pill`) with active highlight indicator.
- **Search Box (`.crm-search-box`):** Embedded search icon, inline input, submit arrow button.
- **Details Drawer (`.crm-drawer-dialog`):** Slide-in panel (`z-index: 1060`, width `90%`, max `580px`, `height: 100vh`) with contact card, inquiry message, and notes interface.
- **Add Lead Modal (`.crm-modal-dialog`):** Centered overlay (`max-width: 620px`, `max-height: 90vh`).

### 6. Footer (`.footer-corporate`)
- Deep navy background (`#07172c`) with 4-column layout: Company Bio, Core Services, Quick Links, and Canadian Office Contact info (Vancouver & Calgary).
- Sub-footer copyright with Privacy and Terms links.

---

## 6. Page Layouts & User Flows

### Public Marketing Flow
1. **Home (`/`):** Hero value proposition $\to$ Service capabilities $\to$ Enterprise metrics $\to$ Testimonials $\to$ Lead capture CTA.
2. **Services (`/services` & `/services/:slug`):** Categorized sidebar navigation $\to$ Service scope $\to$ Case studies $\to$ Consultation booking.
3. **Contact (`/contact`):** Office details $\to$ Interactive form $\to$ AJAX submission $\to$ Immediate feedback alert.

### CRM Operator Flow
1. **Sign-In (`/admin/login`):** Admin/Manager credentials verification.
2. **Dashboard (`/admin/leads`):** View inbound leads $\to$ Filter by status $\to$ Search keyword $\to$ Sort columns $\to$ Paginate.
3. **Lead Inspection & Note Update:** Click "Details" $\to$ Drawer opens $\to$ Edit notes/status $\to$ Save updates with zero status regression $\to$ Table row updates.
4. **Lead Deletion (Admin Only):** Click trash icon $\to$ Confirmation dialog $\to$ Single-dispatch deletion $\to$ Row removed from DOM.

---

## 7. Accessibility (a11y) & Performance Checks

- **Contrast Ratios:** Primary body text (`#475569`) against white achieves `4.8:1` (exceeds WCAG AA minimum of `4.5:1`). Headings (`#0f172a`) against white achieve `14.2:1`.
- **Keyboard Navigation:** Form inputs and interactive buttons retain visible outline states on `:focus-visible`.
- **Semantic Structure:** Single `<h1>` per page with hierarchical `<h2>` and `<h3>` tags; semantic `<header>`, `<main>`, `<nav>`, `<aside>`, and `<footer>` elements.
- **Asset Optimization:** Native SVG icons and responsive image scaling without horizontal overflow.

---

## 8. Architecture Diagram Reference

For full system architecture, database transaction flows, and background email outbox worker diagrams, refer to **Section 1 of [`README.md`](README.md#L9-L46)**.
