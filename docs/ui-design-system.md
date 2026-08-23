# UI design system

`docs/frontend.md` covers how frontend *code* should be written. This document covers what the
interface should *look like and how it should behave*: the design tokens, the type scale, and the
interaction rules that apply across touch, touchpad, mouse, and keyboard.

Every rule here exists because the opposite was tried first. Where a rule has an automated guard,
the guard is named — those exist for the mistakes that are invisible on screen, which are the ones a
convention alone never catches.

## Theme

The theme is owned by `ThemeProvider` in `web/src/app/ThemeContext.tsx`. Nothing else should decide
what theme is active.

- `data-theme` goes on `<html>`, not on an app-level `div`. That is what lets `color-scheme` reach
  native UI, so scrollbars, form controls, and the overscroll area follow the theme instead of
  staying on the browser default.
- An inline bootstrap script in `web/index.html` resolves the theme before first paint. Persistence
  without it only trades a lost setting for a flash of the wrong theme on every load, and a module
  script runs too late to help.
- The storage key is duplicated between that script and `web/src/app/theme.ts` by necessity. Both
  sides carry a comment pointing at the other; keep them in step.
- An explicit choice is stored and wins. With nothing stored the theme follows the operating system
  and tracks live changes to it.
- Storage access is guarded. Blocked site data must fall back to the system preference, not to a
  hardcoded theme.

## Color

### Two families, split by role

This is the distinction that matters most, because getting it wrong is the difference between 9:1
and 1.8:1.

**`brand-*`** is the vivid palette. It is theme-invariant, and it is used **only as a solid fill
beneath `text-brand-ink`** — primary buttons, the brand mark, filled status dots. Dark ink on a
vivid fill measures 7.4–12.2:1 in either theme, which is why these values never need to change.

**`accent-*`** is the same hues used as **content**: text, borders, focus outlines, SVG strokes, and
low-alpha tints. These sit on canvas or panel, so they are theme-aware. The vivid values measure
1.4–2.5:1 against a white panel, nowhere near the 4.5:1 that body text needs.

The test: **if something sits on a surface, it is `accent`. If text sits on it, it is `brand`.**

Two consequences that are easy to miss:

- A progress bar or meter fill is `accent`, not `brand`. It carries no text, so it has to contrast
  with its *track* rather than with a label. The vivid fill measures about 1.4:1 against the
  light-mode track, which reads as an empty bar.
- A solid 2px accent border or a status-dot rim is `accent`. It is a meaningful graphic, and the
  vivid hue is close to invisible on a white panel.

### Surfaces and text

Use the semantic tokens, never a raw Tailwind palette color:

| role | tokens |
| --- | --- |
| surfaces | `canvas`, `shell`, `shell-nav`, `panel`, `panel-soft`, `panel-muted` |
| nested lift | `raised` |
| text | `ink`, `body`, `muted`, `faint` |
| dividers | `line`, `line-strong` |
| code | `code-surface`, `code-ink`, and the `code-*` syntax tokens |

`raised` replaced a literal `bg-white/5` that appeared at seventeen call sites and was invisible
against a white panel. It lifts in dark mode and recesses in light.

All four text tiers clear AA on every opaque surface in both themes. Do not reach for `faint` to
mean "unimportant enough to be unreadable" — if content does not need to be read, it probably does
not need to be rendered.

`bg-black/65` on the tutor scrim is the one deliberate hardcoded color: a black scrim is correct in
both themes.

### Contrast is measured, not judged

The palette is derived by walking lightness along each hue until it meets a target ratio, so hues
are preserved rather than re-picked by eye. Targets are WCAG AA for normal text — 4.5:1. AAA is not
chased.

**Guard: `web/src/styles.contrast.test.ts`.** It re-measures the stylesheet and asserts every text
and accent token against every opaque surface in both themes, checks `brand-ink` on every vivid
fill, checks every syntax token against the code surface, and asserts the teal hover partner moves
*away* from the surface in both directions. When adding a hue, run it rather than eyeballing it.

## Type scale

There are three tiers and no size below 12px:

| token | size | role |
| --- | --- | --- |
| `text-xs` | 12px | the floor. Uppercase eyebrows, badges, tab-bar labels. **Never a sentence.** |
| `text-sm` | 14px | secondary copy: descriptions, metadata, list rows, controls, buttons |
| `text-base` | 16px | primary reading copy, including authored lesson prose |
| `text-lg`+ | | headings |

A `--text-2xs` of 9px used to carry real content at 119 call sites, which is what made the interface
feel cramped at every viewport. That token is gone, so a stray `text-2xs` now generates **no CSS at
all** — the class looks deliberate in the JSX and silently applies no size.

Sizes are uniform across breakpoints. There is no viewport scaling.

Four exceptions are legitimate, and each one in the codebase carries a comment saying why:

- **tab-bar labels**, which are a label tier rather than body copy — five columns on a narrow phone
  cannot fit "Curriculum" at 14px;
- **a glyph sized by its container**, such as a single initial in a 28px circle;
- **text inside an SVG `viewBox`**, which is in user units and scales with the diagram, so it is not
  on this scale at all — and whose label coordinates are usually hand-tuned around their widths;
- **compact numeric grids**, where a fixed column count and fixed-width values genuinely do not fit
  at 14px on a narrow phone.

If you reach for an exception, state the constraint in a comment. "It looked better" is not one.

**Guard: `web/src/styles.typescale.test.ts`.** It rejects `text-2xs` and any sub-12px arbitrary
value such as `text-[8px]`, naming the offending file.

## Interaction

The interface should feel the same whether it is driven by a finger, a touchpad, a mouse, or a
keyboard. That is one requirement, not four.

### Focus

There is a single global `:focus-visible` rule in `web/src/styles.css`.

Do not add per-component focus utilities, and **do not use `outline-0` or `outline-none`**. Before
the global rule existed, eight of roughly sixty interactive elements had a focus style, and two
places — the curriculum search input and the tutor prompt — actively removed the browser default and
replaced it with nothing. A base rule is the only version of this that a new component cannot
forget, which is why it is deliberately not a utility.

### Touch and pointer

- Expand hit targets with the **`pointer-coarse:`** variant, which resolves to
  `@media (pointer: coarse)`. A finger gets 44px; a mouse keeps the tighter density it reads well
  at. Inflating targets unconditionally loosens the desktop layout for no one's benefit.
- `touch-action: manipulation` is set on controls. Without it, tapping a control repeatedly — the
  review rating buttons especially — can register as double-tap-to-zoom.
- Horizontal scrollers carry `overscroll-x-contain`, so swiping past the end of a wide code block
  does not trigger browser back-navigation.
- **`active:` states matter more than `hover:` does.** A finger never hovers, so the press state is
  the only feedback a tap receives.

### State a control conveys

- **A toggle button must pass `pressed`**, which sets `aria-pressed`. Seventeen toggle groups once
  signalled selection purely through `variant={x === y ? … : …}`, so the selected item was announced
  identically to the unselected ones. Nothing looked wrong on screen, which is why it survived four
  pull requests.
- **A disabled control must look disabled.** The shared `Button` handles this; do not reimplement
  it. A disabled button that keeps full contrast and the pointer cursor while ignoring clicks is
  worse than no disabled state.

**Guard: `web/src/toggleState.test.ts`.** It pairs every conditional `variant` with a `pressed`
prop, and fails naming the exact line.

### Dialogs

A modal owes the keyboard all of this, and `TutorOverlay` is the reference implementation:

- closes on `Escape`;
- moves focus in on open, ideally to where the user is about to type;
- restores focus to the opener on close;
- traps `Tab` inside the dialog, so it cannot walk into a page hidden behind the scrim;
- carries `role="dialog"`, `aria-modal`, and an accessible name — on the **panel**, not the scrim;
- locks background scrolling while open.

### `role="img"` prunes its children

An element with `role="img"` is a leaf in the accessibility tree, so its descendants are dropped.
Putting `aria-label` on the cells inside one does nothing. Summarise the information into the
image's own label and mark the children `aria-hidden` so the pruning is intentional rather than
incidental.

## Language the learner reads

A learner should only ever read authored prose. Identifiers such as `python.execution-model` and
`01-scientific-python` are authoring vocabulary.

- **Never render an identifier as visible text.** Six screens once did.
- **Never de-slugify one into prose.** `moduleId.replaceAll('-', ' ')` with `capitalize` produces
  prose-*shaped* text that is still an identifier, which is worse than showing the raw value because
  it looks intentional.
- **An `aria-label` must be human text, not a state key.** "in-progress" and "not-assessed" are
  internal vocabulary that happen to look like words, and a screen reader will read them out.
- When a title cannot be resolved, say something true — "Defined in another module" — rather than
  falling back to the identifier.
- Resources should be described by something useful. A citation's publisher host answers the
  question its ID was sitting in the space of.

Identifiers are correct in keys, route builders, search matching, persistence keys, and download
filenames, where a slug beats a title with spaces.

**Guard: `web/src/humanReadableLabels.test.ts`.** It inspects expressions in JSX *children*
position, where a value becomes visible text, and ignores attributes — so `key={module.id}` and
`to={lessonPath(course.id, …)}` stay correct.

## Code rendering

Read-only code goes through `HighlightedCode` in `web/src/components/HighlightedCode.tsx`. Lesson
MDX, worksheets, and the exercise test preview all share it rather than each having its own path.

- Highlighting reuses the Lezer Python grammar that already ships with `@codemirror/lang-python`
  for the exercise editor, so it costs no new dependency. Adding a language is one line in the
  parser map.
- An unsupported language degrades to plain code rather than throwing. `text` blocks want no
  highlighting at all.
- Code surfaces stay dark in both themes by design. That is a deliberate choice, and because it is
  tokenised, changing it is a change to two token values.
- If you touch the tokenizer, keep the round-trip property: `highlightTree` only reports *styled*
  ranges, so an implementation that fails to fill the gaps between them silently drops indentation
  and blank lines. In Python that corrupts the code rather than merely under-styling it.

## Why so many guards

Each guard here replaced a rule that had already been written down and then broken, usually because
breaking it produced no visible symptom:

| guard | catches |
| --- | --- |
| `styles.contrast.test.ts` | a token that passes on one surface and fails on another |
| `styles.typescale.test.ts` | a class that generates no CSS and applies no size |
| `toggleState.test.ts` | selection state that exists only as a color |
| `humanReadableLabels.test.ts` | an internal identifier reaching the screen |

The pattern worth keeping: when a mistake is repeated across several components *and* invisible on
screen, write a check rather than a convention.

## Review checklist

Alongside the checklist in `docs/frontend.md`, verify that:

- brand hues are `accent-*` unless they are a solid fill under `text-brand-ink`;
- no Tailwind default palette color is hardcoded;
- nothing renders below 12px, and 12px is a label rather than a sentence;
- no `outline-0`, and no per-component focus styles;
- touch targets reach 44px under `pointer-coarse:`;
- conditional `variant` is paired with `pressed`;
- no identifier is rendered as visible text;
- new colors were measured rather than eyeballed.
