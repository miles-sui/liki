Liki Design System v1.0

0. Design Direction

Core Metaphor

晨光映照静水。Morning light over still water.

Warm gold light above. Deep cool water below. A quiet horizon between. This single image governs every design decision that follows.

Key Decisions

- Light base — morning, not night. Deep-sea color anchors the bottom, not the whole page.
- Typography is the primary visual medium — the Sage communicates through words.
- Names are the imagery — a well-set Chinese character carries more weight than any illustration.
- One accent per view — one beam of morning light on the water.
- Modern 70%, traditional 30% — modern is the container (layout, spacing, interaction). Traditional is the accent (amber, serif names, DaoLiTi brand moment, paper surfaces).
- Spring motion — natural physics, not mechanical easing. Fade + gentle rise, no pop.

Sources

| Surface & material | Apple, Aesop |
| Color atmosphere | Natural scene (warm gold → deep blue), Stripe |
| Typography & editorial | Anthropic, Notion |
| Structure & precision | Linear |
| Cultural expression | 观夏, 誠品 |
| Trust & authority | Mayo Clinic |
| Motion | Apple, Stripe |

Anti-Principles

No dark mode. No stock Chinese culture photography. No AI/tech visual clichés. No heavy shadows or neon. No fortune/luck visual language. No AI-generated images. No decorative complexity.

---

1. Tokens

1.1 Color

Semantic naming. No philosophical names in token keys.

Surface

| Token | Value | Usage |
|---|---|---|
| `--color-surface-primary` | `#f6f8fb` | Page background. Mist white — warm-ish gray-white. |
| `--color-surface-secondary` | `#ffffff` | Card background. Clean paper white. |
| `--color-surface-tertiary` | `#edf0f5` | Hover state, alternating section background. |
| `--color-surface-dark` | `#0c1a2e` | Footer, dark header. Deep sea. |
| `--color-surface-darker` | `#060f1c` | Deepest anchor. |

Text

| Token | Value | Usage |
|---|---|---|
| `--color-text-primary` | `#121820` | Headings, body text. Near-black with blue depth. |
| `--color-text-secondary` | `#515967` | Secondary text, labels, meta. |
| `--color-text-tertiary` | `#99a0ae` | Placeholder, disabled, faint. |
| `--color-text-on-dark` | `#e2e8f0` | Text on dark surfaces. |
| `--color-text-on-dark-soft` | `#8899aa` | Secondary text on dark surfaces. |

Accent

| Token | Value | Usage |
|---|---|---|
| `--color-accent` | `#b86b0e` | Primary CTA, links, active state. Morning amber. |
| `--color-accent-hover` | `#92400e` | Hover, pressed state. |
| `--color-accent-soft` | `#fffbeb` | Highlight background, tags. Amber glow. |
| `--color-accent-soft-border` | `#fde68a` | Soft accent border. |
| `--color-accent-focus` | `rgba(184,107,14,0.15)` | Focus ring. |

Border

| Token | Value | Usage |
|---|---|---|
| `--color-border` | `#dfe3ea` | Card border, input border. Hairline. |
| `--color-border-hover` | `#c8ced9` | Border on hover or focus. |

Semantic

| Token | Value | Usage |
|---|---|---|
| `--color-success` | `#15803d` | Success text, icon. |
| `--color-success-soft` | `#f0fdf4` | Success background. |
| `--color-error` | `#dc2626` | Error text, icon. |
| `--color-error-soft` | `#fef2f2` | Error background. |

Accent usage rule: one accent element per view. If the primary CTA is amber, no other amber element competes.

---

1.2 Typography

Font Families

| Token | Stack | Usage |
|---|---|---|
| `--font-ui` | `"PingFang HK", "PingFang SC", "Noto Sans TC", -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif` | Body, labels, inputs, navigation, chat. |
| `--font-reading` | `"Noto Serif TC", "Songti SC", "Times New Roman", serif` | Name displays, classical quotes, report headings. |
| `--font-brand` | `"Noto Serif TC", "Songti SC", "Times New Roman", serif` | Hero title, footer brand name. |

Type Scale

| Step | Size | Line Height | Weight | Usage |
|---|---|---|---|---|
| `--text-xs` | `0.75rem` | `1.5` | 400 | Labels, tags, meta |
| `--text-sm` | `0.875rem` | `1.5` | 400 | Secondary body, captions |
| `--text-base` | `1rem` | `1.75` | 400 | Body text |
| `--text-lg` | `1.125rem` | `1.75` | 400 | Lead paragraphs |
| `--text-xl` | `1.25rem` | `1.6` | 600 | Section headings |
| `--text-2xl` | `1.5rem` | `1.5` | 600 | Page headings |
| `--text-3xl` | `2rem` | `1.4` | 700 | Display headings |
| `--text-4xl` | `2.5rem` | `1.3` | 700 | Hero heading |
| `--text-5xl` | `3rem` | `1.2` | 700 | Hero heading (large) |

Headings use `--color-text-primary` with weight 600-700. Body uses `--color-text-primary` or `--color-text-secondary` with weight 400. All headings start at `--text-xl` and above. Nothing between `--text-lg` and `--text-xl` should function as a heading — use a lead paragraph instead.

Font Loading

`--font-brand` loads as `font-display: swap` via `@font-face`. Fallback to `--font-ui` while loading. The swap should be imperceptible on fast connections.

---

1.3 Spacing

Base unit: `0.25rem` (4px).

| Token | Value | Usage |
|---|---|---|
| `--space-1` | `0.25rem` | Inline gap, icon-text gap |
| `--space-2` | `0.5rem` | Tight internal padding |
| `--space-3` | `0.75rem` | Card padding, button padding |
| `--space-4` | `1rem` | Standard gap, section internal |
| `--space-6` | `1.5rem` | Section gap, card grid gap |
| `--space-8` | `2rem` | Section padding (mobile) |
| `--space-12` | `3rem` | Section padding (desktop) |
| `--space-16` | `4rem` | Major section separation |
| `--space-24` | `6rem` | Hero to content separation |

Page max-width: `56rem` (`--max-w-prose`). Wide enough for comfortable reading, narrow enough for focus.

---

1.4 Radius

| Token | Value | Usage |
|---|---|---|
| `--radius-sm` | `0.25rem` | Tags, kbd, small labels |
| `--radius-md` | `0.5rem` | Buttons, inputs |
| `--radius-lg` | `0.75rem` | Cards, chat bubbles |
| `--radius-xl` | `1rem` | Modals, large cards |
| `--radius-full` | `9999px` | Pills, badges |

No sharp corners (`0`) anywhere. The minimum radius is `--radius-sm`.

---

1.5 Shadow

Minimal. Reserved for elevation that needs it.

| Token | Value | Usage |
|---|---|---|
| `--shadow-card` | `0 2px 12px rgba(0,0,0,0.06)` | Card hover lift. Subtle, barely there. |
| `--shadow-float` | `0 4px 24px rgba(0,0,0,0.10)` | Dropdown, modal, floating element. |
| `--shadow-glow` | `0 0 0 3px var(--color-accent-focus)` | Focus ring (uses box-shadow for accessibility). |

Cards at rest: no shadow. Only border. Shadow appears on hover with the 2px lift.

---

1.6 Motion

Durations

| Token | Value | Usage |
|---|---|---|
| `--duration-micro` | `150ms` | Hover state toggle, focus transition |
| `--duration-fast` | `200ms` | Exit animation, dismiss |
| `--duration-base` | `300ms` | Entrance, page transition, reveal |
| `--duration-slow` | `500ms` | Hero entrance, major transition, breathing |

Easing

| Token | Value | Usage |
|---|---|---|
| `--ease-spring` | `cubic-bezier(0.4, 0, 0.2, 1)` | Entrance, reveal. Natural deceleration. |
| `--ease-spring-soft` | `cubic-bezier(0.2, 0, 0, 1)` | Very gentle entrance, breathing. |
| `--ease-exit` | `cubic-bezier(0.4, 0, 0.6, 0.2)` | Exit, dismiss. Faster acceleration. |

All motion respects `prefers-reduced-motion: reduce` — durations collapse to `0.01ms`, displacements to `0`.

---

1.7 Breakpoints

| Token | Value | Usage |
|---|---|---|
| `--bp-sm` | `640px` | Mobile landscape, large phones |
| `--bp-md` | `768px` | Tablet, small desktop |
| `--bp-lg` | `1024px` | Desktop |
| `--bp-xl` | `1280px` | Wide desktop |

Mobile-first. All base styles target viewports below `--bp-sm`. Overrides layer upward.

---

2. Components

Each component is defined by its visual parameters: which tokens it consumes, its states, and its variants. Implementation (CSS class names, HTML structure) is downstream of this spec.

2.1 Button

Primary

| Property | Value |
|---|---|
| Background | `linear-gradient(135deg, var(--color-accent), #a15c0a)` |
| Text | `#ffffff`, `--text-base`, weight 500 |
| Radius | `--radius-md` |
| Padding | `--space-3` `--space-6` |
| Border | none |

States:
- Default: as above
- Hover: background darkens to `var(--color-accent-hover)` / `#7c2d12`
- Active: `transform: scale(0.98)`
- Focus: `--shadow-glow`
- Disabled: `opacity: 0.35`, `cursor: not-allowed`

Secondary

| Property | Value |
|---|---|
| Background | transparent |
| Text | `var(--color-accent)` |
| Border | `1px solid var(--color-accent)` |
| Radius | `--radius-md` |
| Padding | `--space-3` `--space-6` |

States: hover fills background with `var(--color-accent-soft)`.

Ghost

| Property | Value |
|---|---|
| Background | transparent |
| Text | `var(--color-text-secondary)` |
| Border | none |

States: hover text becomes `var(--color-accent)`.

One primary button per view. Use secondary or ghost for additional actions.

2.2 Card

Default

| Property | Value |
|---|---|
| Background | `var(--color-surface-secondary)` |
| Border | `1px solid var(--color-border)` |
| Radius | `--radius-lg` |
| Padding | `--space-6` |
| Shadow | none at rest |

States:
- Default: as above
- Hover: `transform: translateY(-2px)`, `box-shadow: var(--shadow-card)`

Variants:
- Card (interactive): the default, used when the card is clickable.
- Card (static): no hover lift. Used for content display without interaction.
- Card (dark): `background: var(--color-surface-dark)`, `color: var(--color-text-on-dark)`. Used for contrast sections.

2.3 Input

| Property | Value |
|---|---|
| Background | `var(--color-surface-secondary)` |
| Border | `1px solid var(--color-border)` |
| Radius | `--radius-md` |
| Padding | `--space-3` `--space-4` |
| Text | `--text-base`, `var(--color-text-primary)` |
| Placeholder | `var(--color-text-tertiary)` |

States:
- Default: as above
- Focus: `border-color: var(--color-accent)`, `box-shadow: var(--shadow-glow)`
- Error: `border-color: var(--color-error)`
- Disabled: `opacity: 0.5`, `cursor: not-allowed`

Full-width by default within its container. Min-height for touch: `44px`.

2.4 Tag

| Property | Value |
|---|---|
| Background | `var(--color-accent-soft)` |
| Text | `var(--color-accent)`, `--text-xs`, weight 500 |
| Radius | `--radius-full` |
| Padding | `--space-1` `--space-3` |
| Border | `1px solid var(--color-accent-soft-border)` |

Variants:
- Tag (amber): the default, for cultural/metadata labels.
- Tag (green): `background: var(--color-success-soft)`, `color: var(--color-success)`, border matching. For positive indicators.
- Tag (neutral): `background: var(--color-surface-tertiary)`, `color: var(--color-text-secondary)`, border matching. For neutral metadata.

Inline. Never stacked vertically unless in a wrapping flex container.

2.5 Divider

Two variants:

Section divider (narrative separator):
- `width: 1.5rem`, `height: 2px`, `background: var(--color-accent-soft-border)`
- Centered. Used between narrative text blocks.
- Represents the horizon line.

Content divider:
- `width: 100%`, `height: 1px`, `background: var(--color-border)`
- Used to separate list items or content sections within a card.

2.6 Chat Bubble

User bubble

| Property | Value |
|---|---|
| Background | `var(--color-accent-soft)` |
| Text | `var(--color-text-primary)` |
| Radius | `--radius-lg`, bottom-right `--radius-sm` |
| Padding | `--space-3` `--space-4` |
| Max width | `80%` of chat column |

Assistant bubble

| Property | Value |
|---|---|
| Background | `var(--color-surface-secondary)` |
| Text | `var(--color-text-primary)` |
| Border | `1px solid var(--color-border)` |
| Radius | `--radius-lg`, bottom-left `--radius-sm` |
| Padding | `--space-3` `--space-4` |
| Max width | `80%` of chat column |

Typing indicator: three dots, `6px` diameter, `--color-text-tertiary`, sequential opacity pulse (1.4s cycle, 160ms stagger). Uses `--font-ui` for message content, `--font-reading` for name displays within messages.

2.7 Header

Two variants:

Header (brand) — landing page:
- `background: var(--color-surface-dark)` with warm radial gradient overlay from `var(--color-accent)` at 8-10% opacity
- `color: var(--color-text-on-dark)`
- Padding: `--space-16` top, `--space-24` bottom
- Typography: `--font-brand` for title, `--font-ui` for tagline
- No border, no shadow. Depth comes from color.

Header (app) — chat and inner pages:
- `background: var(--color-surface-dark)`
- `color: var(--color-text-on-dark)`
- Padding: `--space-4` `--space-6`
- Typography: `--font-ui` only. No brand font in app context.
- Fixed or sticky position.

2.8 Footer

| Property | Value |
|---|---|
| Background | `var(--color-surface-darker)` |
| Text | `var(--color-text-on-dark-soft)` |
| Padding | `--space-8` top and bottom |
| Typography | `--text-sm`, `--font-ui` |

Contains: copyright, five navigation links, brand name. Centered. Links use `var(--color-text-on-dark)`, hover to `var(--color-accent)`.

2.9 Status

Three visual states for non-content views:

Loading:
- Centered spinner (`--color-accent` ring, `--color-border` track, 2rem diameter, 0.75s rotation)
- Optional status text below: `--text-sm`, `--color-text-secondary`

Error:
- `--color-error-soft` background card
- `--color-error` icon and heading
- Action button: use Secondary or Primary depending on recoverability

Empty:
- `--color-text-tertiary` icon or illustration placeholder
- `--text-base`, `--color-text-secondary` heading
- `--text-sm`, `--color-text-tertiary` description
- No aggressive empty state personality. Calm, not apologetic.

---

3. Patterns

3.1 Page Layout

Landing page:
```
Header (brand) — full width, dark, deep padding
  Content sections — max-w-prose, centered, alternating bg (primary / tertiary)
    Narrative text blocks — centered, generous spacing
    Card grids — 2-4 columns
    CTA band — max-w-prose, centered
Footer — full width, darker, anchor
```

Content page (about, contact, terms, etc.):
```
Header (app) — full width, dark, compact
  Content — max-w-prose, centered, single column
    Section cards — static card variant, sequential
Footer — full width
```

App page (chat, report):
```
Header (app) — fixed top
  Content area — full width, no max-width constraint, fills viewport
Footer — hidden or minimal
```

3.2 Grid

Mobile (< `--bp-md`): single column, full width. Name previews: horizontal snap-scroll.

Desktop (>= `--bp-md`): 
- 2 columns for method cards, trust cards, product cards
- 3 columns for name preview grid
- Max content width: `--max-w-prose` (56rem) for reading, wider for grids

Grid gap: `--space-6`.

3.3 Responsive

Rules:
- Mobile-first CSS.
- Typography scales down one step on mobile (e.g., `--text-4xl` → `--text-3xl`).
- Section padding: `--space-8` on mobile, `--space-12` on desktop.
- Grid columns collapse from 3→2→1 as viewport narrows.
- Name horizontal scroll on mobile only. Desktop shows full grid.
- Touch targets: minimum `44px` for all interactive elements.
- Header (brand) padding reduces on mobile: `--space-12` top, `--space-16` bottom.

3.4 Motion Patterns

Fade In — entrance for content sections, cards, text blocks.
- `opacity: 0 → 1`
- `--duration-base`, `--ease-spring`
- No translation. Pure opacity.

Gentle Rise — entrance for hero elements, CTA bands, key messaging.
- `opacity: 0 → 1`, `transform: translateY(8px) → translateY(0)`
- `--duration-slow`, `--ease-spring-soft`
- 4-8px rise only. Never more.

Ripple — press feedback on interactive cards, buttons.
- Scale pulse from press point: `scale(0.98) → scale(1.02) → scale(1.0)`
- `--duration-micro`
- Subtle. Not a Material ripple. A water ripple — one gentle pulse.

Breathing — ambient animation on brand surfaces.
- Slow `opacity` pulse: 0.08 → 0.12 → 0.08
- `--duration-slow` × 4 (2s cycle), `--ease-spring-soft`
- Used on gradient overlays and hero visual elements.
- Respects `prefers-reduced-motion`.

Sequence — staggered entrance for grids and lists.
- Each child delayed by 60-80ms.
- Max stagger group: 12 items. Beyond that, no stagger.
- Use Fade In or Gentle Rise per item.

3.5 States

Every interactive element must handle:
- Default (rest)
- Hover (pointer over)
- Active (pressing)
- Focus (keyboard navigation)
- Disabled (not available)
- Loading (waiting — only for actionable elements like buttons and inputs)

Non-interactive content areas must handle:
- Loading (data not yet available)
- Empty (no data to show)
- Error (data failed to load)

Empty states: calm, informative, never apologetic. "Nothing here yet" is a fact, not a failure. Provide a clear next action where applicable.
