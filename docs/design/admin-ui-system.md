# Forge Admin — Design System Spec v3 ("Instrument")

Authoritative source of truth for the admin UI redesign. Implementers MUST read this
file before changing any UI code, and MUST NOT invent values that are not in it.

---

## 1. Direction

**Forge Instrument** — a Swiss-industrial operator console.

Reference frame: the discipline of a trading terminal, the typographic restraint of
Linear, the data density of the Stripe dashboard. It is a tool for people who use it
for six hours a day, not a marketing surface.

Five rules, in priority order:

1. **Typography carries hierarchy — color does not.** Size and weight tell you what
   matters. Color is reserved for state and action.
2. **Density is assigned, never uniform.** Three densities exist (§4) and each surface
   is explicitly assigned one. Chrome is tight, decisions are airy, tables are dense.
3. **Separation by hairline, not by shadow.** Surfaces are flat. Elevation exists only
   for things that float above the page (menus, dialogs, popovers).
4. **Numerals are mono and tabular, always.** Every id, count, metric, currency,
   timestamp and percentage. Columns must align down the page.
5. **One memorable moment:** the command palette over a dimmed grid. Everything else
   is quiet on purpose.

Explicitly rejected: glassmorphism, gradient fills, decorative accent color, drop
shadows on cards, route-level page transitions, uniform card grids.

---

## 2. Typography

Two faces, one job each.

| Role | Family | Why |
|---|---|---|
| UI / prose | **Inter** (already loaded) | Neutral, excellent at 11–14px, `cv02 cv03 cv04 cv11` already enabled |
| Numerals / ids / code | **JetBrains Mono** | Tabular by default; makes columns align; gives the console character |

Self-host both via `@fontsource`. No CDN, no external stylesheet (CSP).
`font-display: swap`. Subset latin. Preload only Inter 400 + 500 and JetBrains Mono 400.

### 2.1 Scale

The current UI tops out around 13px. That is the core failure — there is no contrast.
This scale introduces a real jump.

```css
--text-micro:     0.6875rem;  /* 11px — uppercase eyebrow labels, tracking +0.08em */
--text-meta:      0.75rem;    /* 12px — secondary meta, counts, helper text */
--text-body:      0.8125rem;  /* 13px — THE workhorse: table cells, form values */
--text-ui:        0.875rem;   /* 14px — buttons, inputs, nav items */
--text-lead:      1rem;       /* 16px — card titles, section leads */
--text-title:     1.375rem;   /* 22px — page titles */
--text-display:   2rem;       /* 32px — dashboard section headings */
--text-metric:    2.5rem;     /* 40px — stat tile values */
--text-metric-xl: 3.25rem;    /* 52px — hero metric, dashboard only */
```

### 2.2 Weight & tracking

| Size band | Weight | Tracking |
|---|---|---|
| `--text-micro` | 600 | `+0.08em`, `text-transform: uppercase` |
| `--text-meta` → `--text-ui` | 400 body / 500 emphasis | `0` |
| `--text-lead` | 600 | `-0.01em` |
| `--text-title` and above | 600 | `-0.02em` |
| `--text-metric*` | 700, **JetBrains Mono**, `font-variant-numeric: tabular-nums` | `-0.03em` |

Line heights: `1.2` for ≥22px, `1.45` for body, `1` for metrics.

### 2.3 Hard rules

- Never use a raw arbitrary size (`text-[13px]`, `text-[11px]`). Use the token classes.
- Every page has exactly one `--text-title`. Every dashboard section gets one
  `--text-micro` eyebrow above its heading.
- Any element rendering a number, id, timestamp, or code uses the mono stack with
  `tabular-nums`.

---

## 3. Color

Accent stays **iris indigo** (`243 75% 58%` light / `244 80% 70%` dark). The existing
`primaries` list in `src/lib/themes.ts` is kept as-is.

### 3.1 Accent discipline

Iris is allowed in exactly four places: primary action buttons, the active-nav
indicator, the focus ring, and inline links. Nowhere else. It is never a decorative
fill, never a chart default, never a badge background for non-primary state.

### 3.2 Surfaces — the dark-mode bug

Today `--card` and `--background` are the **identical** value in `.dark`
(`240 10% 3.9%`). Cards have zero fill separation in dark mode. This is the single
biggest cause of dark mode reading as a flat sheet. Fixed by a real 3-step surface ramp:

```css
/* light */
--surface-1: 40 18% 97%;    /* page ground — warm paper, not stock zinc */
--surface-2: 0 0% 100%;     /* cards, table body */
--surface-3: 0 0% 100%;     /* overlays: menus, dialogs, popovers */
--surface-sunken: 40 14% 94%; /* table header, inset wells, code blocks */

/* dark */
--surface-1: 240 12% 6%;    /* page ground */
--surface-2: 240 10% 9.5%;  /* cards — MUST be lighter than surface-1 */
--surface-3: 240 9% 13%;    /* overlays — lighter still */
--surface-sunken: 240 14% 4.5%; /* darker than the page: table headers, wells */
```

`--background` maps to `--surface-1`, `--card`/`--popover` to `--surface-2`/`-3`.

### 3.3 Borders

```css
--border-subtle: /* hairlines inside a surface: table rows, list dividers */
--border:        /* default component border */
--border-strong: /* focused / selected / emphasized */
--grid-line:     /* table column rules — weakest of all, ~50% of --border-subtle */
```

### 3.4 Semantic state

Today `StatsCard.tsx` hardcodes `#10b981`, `#f59e0b`, `#ef4444`, and components
scatter `emerald-500` / `rose-500` / `amber-500` utilities. All of it is replaced by
tokens. No raw hex and no Tailwind palette color may appear in application code.

```css
--success / --success-fg / --success-surface   /* positive state, upward delta */
--warning / --warning-fg / --warning-surface
--danger  / --danger-fg  / --danger-surface    /* aliases existing --destructive */
--info    / --info-fg    / --info-surface
--neutral / --neutral-fg / --neutral-surface   /* the DEFAULT badge — most badges */
```

Each `*-surface` is the tinted background used for badges and pills; each base is the
text/icon color on that surface. Both must clear 4.5:1 against their surface in both
themes.

### 3.5 Data visualization

A single ordered categorical ramp, `--chart-1` … `--chart-6`, defined once and used by
every chart. Iris is `--chart-1`. The rest are distinguishable in both themes and to
deuteranopes. Charts never use `--primary` directly and never use raw hex.

---

## 4. Density

Three densities. Every surface is explicitly assigned one.

| Token set | Applies to | Row/control height | Inline padding | Gap |
|---|---|---|---|---|
| **command** | sidebar, topbar, toolbars, filter bar | 32px | 10px | 8px |
| **grid** | data tables, list rows, pickers | 40px (compact) / 52px (comfortable) | 12px | 0 |
| **canvas** | dashboard, forms, detail views, cards | — | card pad 20px | 24px |

```css
--row-command: 2rem;      --pad-command: 0.625rem;  --gap-command: 0.5rem;
--row-grid: 2.5rem;       --row-grid-comfy: 3.25rem; --pad-grid: 0.75rem;
--pad-canvas: 1.25rem;    --gap-canvas: 1.5rem;
--space-section: 2rem;
--page-gutter: 1.5rem;    /* 1rem below 640px */
--page-max: 96rem;
```

The grid density is user-togglable (compact ↔ comfortable), persisted per user in
`localStorage` under `forge.admin.density`.

Radii tighten — this is an instrument, not a pill: `--radius: 0.375rem` (6px),
`--radius-sm: 0.25rem`, `--radius-lg: 0.5rem`. Nothing is `rounded-xl` or fully round
except avatars and status dots.

---

## 5. Elevation

Only three, and only for floating things:

```css
--elev-overlay: 0 4px 16px -2px rgb(0 0 0 / 0.12), 0 2px 4px -2px rgb(0 0 0 / 0.08);
--elev-dialog:  0 16px 48px -8px rgb(0 0 0 / 0.24);
--elev-sticky:  0 1px 0 0 hsl(var(--border-subtle));  /* sticky headers: a rule, not a shadow */
```

Cards get **no** shadow — a `1px` `--border-subtle` hairline only. Delete
`premium-shadow`, `glass-lite`, and `glass-thick` from `index.css`; nothing may use
`backdrop-blur` except the mobile nav scrim and the command palette backdrop.

---

## 6. Motion

```css
--dur-fast: 120ms; --dur-base: 180ms; --dur-slow: 240ms;
--ease-out: cubic-bezier(0.25, 1, 0.5, 1);
--ease-in-out: cubic-bezier(0.65, 0, 0.35, 1);
```

Motion is allowed in exactly five places:

1. Active-nav indicator gliding between items (`framer-motion` `layoutId`).
2. Overlay enter/exit (dialog, popover, dropdown, palette) — scale `0.98 → 1` + fade.
3. Table row hover / selection background.
4. Skeleton shimmer while loading.
5. Toast enter/exit.

**No route-level page transitions.** Operators navigate hundreds of times a day; a
staged fade makes the tool feel slower. The existing `prefers-reduced-motion` block in
`index.css` is kept and must keep working.

---

## 7. Component contracts

New or rebuilt primitives. Each is its own file, each under 200 lines.

| Component | Contract |
|---|---|
| `PageHeader` | `{ eyebrow?, title, description?, actions?, meta? }`. Exactly one per page. Owns the `--text-title` + breadcrumb slot. |
| `StatTile` | Replaces `StatsCard`. Label at `--text-micro`, value at `--text-metric` in mono, delta as a semantic pill, optional sparkline. No hover scale, no gradient. |
| `DataGrid` | Table shell: sticky header on `--surface-sunken`, sticky first identity column, `--grid-line` column rules, zebra off, hover row, density prop, skeleton rows, empty state, column-visibility menu. |
| `FilterBar` | `command` density. One control per filter **typed by `filter.type`** (see §8 VB-05). |
| `EmptyState` | `{ icon, title, description, action? }`. Used by grid, dashboard widgets, search. |
| `StatusBadge` | Takes a semantic tone (`success`/`warning`/`danger`/`info`/`neutral`), never a raw color. |
| `CommandPalette` | `cmdk`. Replaces the hand-rolled `GlobalSearch`. Groups: Navigation, Models, Records, Actions. |
| `Toaster` | `sonner`. Replaces `use-toast` + `toast.tsx`. |
| `Tooltip` | `@radix-ui/react-tooltip`. Required for every icon-only button and every collapsed-sidebar item. Replaces `title=` attributes. |
| `Field` | Label + control + help + error, one consistent vertical rhythm for all form widgets. |

---

## 8. Confirmed defects to fix

Verified by reading the source, with locations.

| ID | Defect | Location | Fix |
|---|---|---|---|
| **VB-01** | Sidebar explodes into ~25 groups plus a giant "OTHER" | `AdminLayout.tsx` `groupByModel`, the `derivedGroup` IIFE | Use `config.model_groups` when present. Fallback: only derive a group from a real `.`/`__` separator. Single-word models go into one ungrouped list with **no** heading. Suppress group headings entirely when only one group results. |
| **VB-02** | Foreign keys render the raw id (`0`) | `ModelListPage.tsx` cell renderer, `val?.toString()` | Resolve FK display labels from `metadata.relations` via the autocomplete cache; render label + muted id. Fall back to `#id` if unresolved, `—` if null. Same fix on `ModelViewPage`. |
| **VB-03** | Literal `null` / inconsistent empties | list + view cells | One shared `<EmptyValue/>` rendering an em-dash at `--text-meta` muted. Note `val?.toString() \|\| —` also swallows empty strings — make the null check explicit. |
| **VB-04** | Chart axes cramped, dark-mode contrast poor | `chart-widget.tsx`, `ChartWidget.tsx` | Shared chart theme: `--chart-*` ramp, shared tooltip component, min-height 280px, axis tick `--text-meta`, grid lines `--grid-line`. |
| **VB-05** | Numeric filters render as free-text inputs | `ModelListPage.tsx` filter block — everything non-choice/date/boolean falls through to `<Input>` | Branch on `filter.type === "number"` → min/max numeric pair; `"relation"` → searchable select. |
| **VB-06** | `created_at`/`updated_at` editable on create | **Backend, not UI.** `ModelUpsertPage.tsx` already does `if (field.read_only && mode === "create") return null` | The Go metadata emitter is not setting `read_only` on auto timestamps. Fix in `forge/admin` Go code. |
| **VB-07** | Palette shows raw model keys | `GlobalSearch.tsx` | Use `verbose_name_plural` from the models cache. Resolved by the `CommandPalette` rebuild. |
| **VB-08** | Mobile table overflows with no affordance | list page table wrapper | Sticky identity column + scroll-shadow affordance on the `overflow-x` container. |
| **VB-09** | Toasts stack/overlap on rapid actions | `use-toast.ts` | `sonner` migration. |
| **VB-10** | Ad-hoc dark classes, inconsistent edges | everywhere | Resolved by the §3 token migration. |
| **VB-11** | `SidebarItem` is **defined inside** `AdminLayout`'s render body | `AdminLayout.tsx` | New component identity every render ⇒ whole nav subtree remounts and per-item `expanded` state resets. Hoist to module scope (or its own file) and pass what it needs via props. **Correctness bug, not cosmetics.** |
| **VB-12** | Files exceed the repo's 800-line limit | `ModelUpsertPage` 1389, `ModelListPage` 1249, `AdminLayout` 830 | Split by responsibility as part of the phases below. |
| **VB-13** | Raw `<select>` in the filter panel while `@radix-ui/react-select` is a dependency | `ModelListPage.tsx` | Use the `Select` primitive. |

---

## 9. Dependencies

Add only these four. Each replaces something hand-rolled and reduces net code.

| Package | Replaces |
|---|---|
| `cmdk` | the ~350-line hand-rolled `GlobalSearch` |
| `sonner` | `use-toast.ts` + `toast.tsx` |
| `@radix-ui/react-tooltip` | `title=` attributes (an a11y requirement) |
| `date-fns` | scattered inline `toLocaleDateString` blocks |
| `framer-motion` | **only** for the nav `layoutId` indicator |
| `@fontsource/inter`, `@fontsource/jetbrains-mono` | self-hosted fonts |

**Explicitly not added:** `@radix-ui/react-avatar`, `-separator`, `-scroll-area`,
`-checkbox`, `-switch`. They are cosmetic swaps producing a large diff and zero
user-visible improvement. Add one only when a specific need appears (e.g. genuine
tristate). No router migration. No `next-themes`. No virtualizer (lists are paginated).

---

## 10. Quality gate

A phase is not done until all of these pass:

1. `npx tsc -b` — no errors.
2. `npx vitest run` — all green, no reduction in test count.
3. `npx eslint .` — no new errors.
4. No raw hex colors and no Tailwind palette colors (`emerald-500`, `zinc-900`, …) in
   `src/` outside `index.css`.
5. No arbitrary font sizes (`text-[13px]`) in `src/`.
6. Every touched file is under 800 lines.
7. Playwright sweep: every page renders in light **and** dark, at 375px and 1440px,
   with no horizontal body scroll and no console errors.

---

## Appendix A — Normative token values

These are exact. Implementers copy them verbatim; they do not adjust, round, or
substitute values. All colors are bare HSL channel triplets (no `hsl()` wrapper) so
they compose with Tailwind's `hsl(var(--x) / <alpha>)` pattern.

### A.1 Compatibility requirement

The existing shadcn alias tokens **must keep existing** and must be defined in terms
of the new ones. ~12,000 lines of components reference `bg-card`, `text-muted-foreground`,
`border-border`, etc. Breaking them is not acceptable. Aliases:

```
--background       → --surface-1
--card, --popover  → --surface-2 (light) / --surface-2 (dark)
--secondary,--muted,--accent → --surface-sunken
--destructive      → --danger
--input            → --border
```

`--primary`, `--primary-foreground`, and `--ring` are written at runtime by
`themeStore`/`ThemeCustomizer` from `src/lib/themes.ts`. Do not hardcode them in a way
that defeats that; keep the iris values as the static default.

### A.2 Light theme (`:root`)

```
--surface-1: 40 18% 97%;      --surface-2: 0 0% 100%;
--surface-3: 0 0% 100%;       --surface-sunken: 40 14% 94%;
--foreground: 240 12% 12%;    --muted-foreground: 240 6% 45%;
--border-subtle: 40 8% 90%;   --border: 40 8% 85%;
--border-strong: 40 8% 72%;   --grid-line: 40 8% 93%;

--success: 152 62% 32%;  --success-fg: 0 0% 100%;  --success-surface: 152 55% 94%;
--warning:  32 88% 40%;  --warning-fg: 0 0% 100%;  --warning-surface:  38 90% 93%;
--danger:  356 68% 46%;  --danger-fg:  0 0% 100%;  --danger-surface:  356 80% 95%;
--info:    214 78% 44%;  --info-fg:    0 0% 100%;  --info-surface:    214 85% 94%;
--neutral: 240  6% 38%;  --neutral-fg: 0 0% 100%;  --neutral-surface: 240  8% 93%;

--chart-1: 243 75% 58%;  --chart-2: 190 78% 40%;  --chart-3:  32 88% 50%;
--chart-4: 330 65% 55%;  --chart-5: 152 52% 40%;  --chart-6: 265 40% 55%;
```

### A.3 Dark theme (`.dark`)

```
--surface-1: 240 12% 6%;      --surface-2: 240 10% 9.5%;
--surface-3: 240  9% 13%;     --surface-sunken: 240 14% 4.5%;
--foreground: 240 15% 95%;    --muted-foreground: 240 8% 62%;
--border-subtle: 240 8% 16%;  --border: 240 8% 20%;
--border-strong: 240 8% 30%;  --grid-line: 240 8% 13%;

--success: 152 55% 58%;  --success-fg: 152 70%  8%;  --success-surface: 152 40% 14%;
--warning:  38 88% 62%;  --warning-fg:  32 80% 10%;  --warning-surface:  34 45% 15%;
--danger:  356 78% 66%;  --danger-fg:  356 70% 10%;  --danger-surface:  356 40% 17%;
--info:    214 85% 66%;  --info-fg:    214 70% 10%;  --info-surface:    214 45% 16%;
--neutral: 240  8% 68%;  --neutral-fg: 240 10% 10%;  --neutral-surface: 240  8% 18%;

--chart-1: 244 80% 70%;  --chart-2: 190 70% 55%;  --chart-3:  35 85% 62%;
--chart-4: 330 70% 68%;  --chart-5: 152 50% 55%;  --chart-6: 265 50% 68%;
```

### A.4 Theme-independent (declare once on `:root`)

```
--font-sans: "Inter", system-ui, -apple-system, sans-serif;
--font-mono: "JetBrains Mono", ui-monospace, SFMono-Regular, Menlo, monospace;

--text-micro: 0.6875rem;  --text-meta: 0.75rem;     --text-body: 0.8125rem;
--text-ui: 0.875rem;      --text-lead: 1rem;        --text-title: 1.375rem;
--text-display: 2rem;     --text-metric: 2.5rem;    --text-metric-xl: 3.25rem;

--row-command: 2rem;      --pad-command: 0.625rem;  --gap-command: 0.5rem;
--row-grid: 2.5rem;       --row-grid-comfy: 3.25rem; --pad-grid: 0.75rem;
--pad-canvas: 1.25rem;    --gap-canvas: 1.5rem;
--space-section: 2rem;    --page-gutter: 1.5rem;    --page-max: 96rem;

--radius-sm: 0.25rem;     --radius: 0.375rem;       --radius-lg: 0.5rem;

--dur-fast: 120ms;        --dur-base: 180ms;        --dur-slow: 240ms;
--ease-out: cubic-bezier(0.25, 1, 0.5, 1);
--ease-in-out: cubic-bezier(0.65, 0, 0.35, 1);

--elev-overlay: 0 4px 16px -2px rgb(0 0 0 / 0.12), 0 2px 4px -2px rgb(0 0 0 / 0.08);
--elev-dialog:  0 16px 48px -8px rgb(0 0 0 / 0.24);
```

Under `@media (max-width: 640px)`: `--page-gutter: 1rem;`

### A.5 Tailwind exposure (`tailwind.config.js`)

Every token above gets a Tailwind utility. Colors follow the existing
`"hsl(var(--x))"` pattern already used in the file:

- `colors`: `surface: {1,2,3,sunken}`, `success/warning/danger/info/neutral` each with
  `DEFAULT`, `fg`, `surface`; `chart: {1..6}`; `border-subtle`, `border-strong`, `grid-line`.
- `fontSize`: `micro, meta, body, ui, lead, title, display, metric, metric-xl`, each
  with its line-height and tracking from §2.2 as the tuple's second element.
- `fontFamily`: `sans: var(--font-sans)`, `mono: var(--font-mono)`.
- `spacing`: `command, grid, grid-comfy, canvas, section, gutter`.
- `borderRadius`: `sm/DEFAULT/lg` from the new radius vars.
- `boxShadow`: `overlay, dialog`. **Remove** any card shadow.
- `transitionTimingFunction`: `out, in-out`. `transitionDuration`: `fast, base, slow`.

---

## Appendix B — VB-02 foreign-key labels: architecture decision

**Constraint discovered during implementation.** The only relation-aware endpoint is
`GET /{model}/autocomplete?field=&q=&limit=` (`src/api/client.ts`). It is a *text
search* — `useAutocomplete` is even gated on `query.length > 0`. There is no way to
ask it "resolve these 25 ids to labels". So the frontend **cannot** fix VB-02 alone
without an N+1 request per row, which is unacceptable.

Three options were considered:

| Option | Verdict |
|---|---|
| Frontend fetches the whole related model and builds an id→label map | Rejected — unbounded, breaks on any large table |
| New batch endpoint `GET /{model}/labels?ids=…` | Workable, but adds an endpoint and still costs one request per FK column |
| **List serializer emits the related object's display value inline** | **Chosen** — one query, no N+1, and it is what Django admin does |

### Decision: two phases, shipped independently

**Phase 1 — frontend only, no backend change (task 4.2a).**
An FK cell stops rendering a bare `0`. It renders `#<id>` in the mono/tabular style as
a link to the related record, using `metadata.relations` to find `related_model`. Null
renders `<EmptyValue/>`. This removes the "why does it say 0" confusion and adds real
navigation value immediately, at zero backend risk.

**Phase 2 — backend (task 6.2), then frontend upgrade (task 4.2b).**
The list serializer adds a sibling display key for each relation — `"<field>__display"`
— carrying the related object's human label. `FieldMetadata`/`RelationMetadata` already
exist in `forge/admin/core/metadata.go`; the serialization path is
`forge/api/serializers/typed_serializer.go` → `Serialize`. Note the blast radius:
`core.Metadata` has 6 callers across `metadata_builder.go`, `openapi.go`,
`admin/api/rest/router.go` and `admin/core/admin.go`, with tests in
`registry_test.go` and `router_test.go`. The change must be **additive** — a new
optional key, never a change to existing field serialization.

Frontend then prefers `row["<field>__display"]` when present and falls back to the
Phase-1 `#<id>` rendering when absent. The two phases are therefore compatible in
either deployment order, which is why Phase 1 ships first.

---

## Appendix C — cmdk deferral (revises §9)

§9 lists `cmdk` as replacing the hand-rolled `GlobalSearch`. That still holds as the
end state, but it is **deferred behind the cheaper fixes**, deliberately.

`GlobalSearch.tsx` is 352 lines and is covered by four behavioural tests in
`GlobalSearch.test.tsx`: it opens and lists models, filters as you type, shows record
results from the API and navigates on Enter, and closes on Escape. Those tests pin
`data-testid="global-search-trigger"`, `data-testid="global-search-input"`, and
`role="dialog"` with the accessible name "Global search".

Handing a 350-line rewrite of a tested, working component to a weak model is the
highest-risk change in this plan, and the payoff is keyboard/grouping polish rather
than a defect fix. So:

- **Now (task 3.4a):** fix VB-07 and restyle to the new tokens, in place. VB-07 is
  precisely `GlobalSearch.tsx:146`, `sub: group.model` — the record subtitle renders
  the raw model key (`order_item`) instead of the verbose name. Models themselves are
  already correct at line 156 (`verbose_name_plural`).
- **Later (task 3.4b):** swap the internals to `cmdk`, with the four existing tests as
  the safety net and a hard requirement that every `data-testid` and the dialog's
  accessible name survive unchanged.

This ordering means the user-visible defect is gone immediately and the risky refactor
lands against a green, meaningful test suite rather than in place of one.

---

## Appendix D — chart ramp, validated (supersedes the `--chart-*` values in A.2/A.3)

The ramp originally written in Appendix A was **eyeballed and it failed**. Running the
`dataviz` skill's `validate_palette.js` against it surfaced a genuine accessibility
defect: the original `--chart-4` (magenta `#d7428c`) and `--chart-5` (green `#319b6a`)
sat at **ΔE 4.0 under deuteranopia** — indistinguishable to red-green colorblind
readers — plus the orange failed the 3:1 contrast floor against the light surface.

Both ramps below were re-stepped and now **pass all six checks** (lightness band,
chroma floor, CVD separation, normal-vision floor, contrast vs surface) in their own
mode. Dark is *selected* against the dark surface, not an automatic flip of light.

**These values replace the `--chart-1..6` lines in A.2 and A.3.**

```
/* light — validated against surface #fcfcfb, worst adjacent ΔE 10.9 (protan) */
--chart-1: 243 75% 58%;     /* #4c44e4 iris   */
--chart-2:  31 94% 41%;     /* #c96a06 amber  */
--chart-3: 192 85% 37%;     /* #0e8fae teal   */
--chart-4: 348 70% 60%;     /* #e0516e rose   */
--chart-5: 264 42% 52%;     /* #7a52b8 violet */
--chart-6: 153 51% 37%;     /* #2e8f63 green  */

/* dark — validated against surface #1a1a19 */
--chart-1: 245 80% 67%;     /* #7369ee */
--chart-2:  34 63% 45%;     /* #bc7d2a */
--chart-3: 193 65% 43%;     /* #2695b5 */
--chart-4: 348 59% 61%;     /* #d65f76 */
--chart-5: 260 43% 61%;     /* #8c6fc6 */
--chart-6: 151 42% 43%;     /* #3f9a6e */
```

### Binding rules that come with this ramp

- **Assign in fixed order, never cycled.** A 7th series is not a generated hue — it
  folds into "Other", or the chart becomes small multiples.
- **Color follows the entity, not its rank.** Filtering out a series must never
  repaint the survivors.
- **Never a dual-axis chart.** Two measures of different scale become two charts or
  are indexed to a common base.
- **Legend always present for ≥2 series**; ≤4 series are also directly labelled, so
  identity is never carried by color alone. A single series needs no legend.
- **Status colors (`--success`/`--warning`/`--danger`) are reserved** and must never
  be reused as "series 4". They always ship with an icon or label, never color alone.
- **Text wears text tokens, never the series color.** Values and labels stay in
  `--foreground` / `--muted-foreground`; the colored mark beside them carries identity.

### Re-validating after any change

```bash
cd <dataviz skill dir>
node scripts/validate_palette.js "<hex,hex,...>" --mode light
node scripts/validate_palette.js "<hex,hex,...>" --mode dark
```
Never adjust a chart color by eye — re-run this and require ALL CHECKS PASS.
