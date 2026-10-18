Tax Advisory is a Material Design 3 (M3) system for a **desktop web product for German tax advisors and their clients**. Advisors define _personas_ (types of client, such as Angestellt, Vermietung, Selbstständig) with the documents and information each persona must provide, assign personas to clients, and follow progress. Clients see exactly what they have to provide, upload documents and fill in text fields. Tokens use M3 role names (`primary`, `on-primary-container`, `surface-container-high`, `shape-large`, `label-large`), so any M3 guidance applies unless this book says otherwise.

## Using it with Material Web and Svelte

- Production UI is built with **`@material/web`** web components inside **Svelte**. Where Material Web has a component (button, FAB, chips, text field, select, checkbox, switch, tabs, list, dialog, progress), use it; do not rebuild it. Where it has none (navigation rail, top app bar, card, data table, status chip, steps, file upload, panes), implement a Svelte component using the same tokens and the `md-*` class rules in `components/bundle.css` as the reference rendering. [`usage.md`](usage.md) says, per component, which applies and what is specific to this product; the `md-*` rules in `components/bundle.css` are only the reference rendering used by the mockups.
- `bundle.css` bridges the tokens to Material Web: every `--md-sys-color-*`, `--md-sys-typescale-*` and `--md-sys-shape-corner-*` variable points at this system's tokens, so Material Web components pick up the brand and all three themes without per-component overrides. Load `tokens.css`, then `bundle.css`, then the Material Web scripts.
- Set the theme with `data-theme="light" | "dark" | "high-contrast"` on `<html>`. Default to `high-contrast` when the OS reports `prefers-contrast: more`, and let users override it in Einstellungen.

## Content fundamentals

- **Language is German.** Address clients with **Sie**; use plain words and one idea per sentence. Advisors may see technical terms; clients never see internal words such as "Persona", "Status" or "Veranlagungsstatus". Explain a tax term once, in brackets: "Anlage V (Formular für Mieteinnahmen)".
- **Buttons are verbs, sentence case:** "Datei hochladen", "Einladung senden", "Persona löschen". Never "OK" or "Absenden" for a decision with consequences.
- **Formats follow German convention.** Money: `4.820,00 €` (thousands point, decimal comma, non-breaking space, symbol after) in a `num` element or a `numeric-*` style. Dates: `31.07.2025` in tables, `31. Juli 2025` in prose. Percent: `75 %` with a space. IBAN in groups of four. Steuer-ID as 11 digits. Deadlines always add the days left ("noch 14 Tage").
- **Errors say how to fix it:** "Bitte im Format TT.MM.JJJJ eingeben", not "Ungültige Eingabe".
- **Reassure, don't alarm:** pending work is "Offen" or "Wartet auf Mandant" (`tertiary`); only a missed statutory deadline is "Überfällig" (`error`).
- No emoji, no exclamation marks.

## What is in this folder

- `tokens.json`, `tokens.css`, `components/bundle.css`: the system the app imports.
- [`usage.md`](usage.md): how the standard Material components (button, text field, dialog, rail, …) are used in this product, with links to the M3 and `@material/web` docs. Their specs are not repeated here.
- `components/<Name>/README.md` and `preview.html`: full documentation and a static specimen for the components M3 does not provide (DataTable, StatusChip, FileUpload, ProgressIndicator steps). Open `index.html` through a local server to browse them in all three themes.
- `guidelines/layout.md`: rails, panes and window size classes.
- `brand/cover.html`: the cover.
- `screenshots/`: renders of earlier component previews, kept for reference.

Mockups of product screens are **not** part of the system; they live in [`../mockups`](../mockups/README.md) and are built from it. The three core flows they show are: advisors define **personas** (required documents and information), advisors **assign personas to clients** (the union of requirements becomes the client's checklist; an item that is Pflicht in any persona is Pflicht for the client), and **clients** upload documents and enter information.

Statuses use `StatusChip`: Entwurf, In Prüfung, Wartet auf Mandant, Freigegeben, Eingereicht, Überfällig; document-level Offen, Hochgeladen, Geprüft, Rückfrage, Abgelehnt.

## Visual foundations

- **Color.** Teal `primary` is the only color that means "act here". Use roles in pairs (`primary` with `on-primary`, `secondary-container` with `on-secondary-container`); never put another token's text color on a fill. Rail and page sit on `surface`; panes on `surface-container-low`; detail pane, tables and requirement rows on `surface-container-lowest`; dialogs on `surface-container-high`. `tertiary` (amber) means a deadline or "waiting on the client"; `success` means complete or approved; `error` means overdue or destructive. Every status has an icon and a word: color is never the only cue.
- **Type.** **Roboto** throughout, following the M3 type scale: `headline-*` for page and pane titles, `title-*` for card headings and list headlines, `body-*` for text, `label-*` for buttons, tabs, chips and navigation, `numeric-*` for headline figures (tabular figures). Roboto is loaded from Google Fonts in `bundle.css`; self-host it for production to satisfy GDPR.
- **Shape.** M3 shape scale only: `shape-extra-small` fields, `shape-small` chips, `shape-medium` cards and requirement rows, `shape-large` selected items and FABs, `shape-extra-large` panes and dialogs, `shape-full` buttons and indicators.
- **Elevation.** Tonal surfaces first; `elevation-1` for elevated cards, `elevation-3` for FAB and dialogs.
- **State.** Every interactive element has a `currentColor` state layer: `state-hover` 8%, `state-focus` 10%, `state-pressed` 10%. Keyboard focus draws a solid 3px `focus-ring` outline with 2px offset. Disabled content is `state-disabled-content` (38%) and is explained in text.
- **Space.** 4dp grid, `space-1` to `space-8`. Controls are at least `touch-target` (48px) in their hit area.
- **Motion.** 150ms for state changes, 250ms for pane transitions, M3 standard easing; none under `prefers-reduced-motion`.

## Layout: rails and panes

The product is **desktop only**; the minimum supported width is **1024px**. Use the M3 window size classes **Expanded** (840–1199), **Large** (1200–1599) and **Extra-large** (1600+). See [`guidelines/layout.md`](guidelines/layout.md).

- A **navigation rail** (`nav-rail-width`) holds 3 to 7 destinations. Advisor: Übersicht, Mandanten, Personas, Fristen, Nachrichten. Client: Start, Aufgaben, Unterlagen, Nachrichten. An optional FAB sits on top of the rail.
- **Panes** are rounded containers (`shape-extra-large`) inset by `pane-margin` and separated by `pane-gutter` (24px). The list pane has a fixed `pane-list-width`, the supporting pane a fixed `pane-supporting-width`; only the detail pane flexes.
- Canonical layouts: **list-detail + supporting pane** for advisor work (persona builder, client setup), **detail + supporting pane** for clients (client portal); see the mockups. The list pane stays visible while the detail changes.
- Below 1200px the supporting pane becomes a side sheet opened from the pane header; the rail never disappears.

## Iconography

**Material Symbols Outlined**, 24px, referenced by ligature name (`<md-icon>upload_file</md-icon>` with Material Web). Active navigation items and status icons use the filled variant. An icon next to a tax term always has a label. No logo was provided, so the product name is set in `headline` text.

## Accessibility

- Target **WCAG 2.2 AA** in the light and dark themes: text 4.5:1, controls, borders and focus rings 3:1.
- The **high-contrast theme** (`data-theme="high-contrast"`) is a third, complete theme: all text 7:1 on its grounds, surfaces and controls separated by 2px `outline` borders instead of shade, selected states marked by border and underline as well as fill, thicker tab indicator, 4px focus ring.
- Every status has icon and word; success and error are never told apart by hue alone.
- Full keyboard operation. Upload has a button path, not only drag and drop. Announce progress and results to screen readers (`aria-live="polite"`), and give badge counts an accessible name ("Nachrichten, 4 ungelesen").
- Text must remain usable at 200% zoom and with user font-size changes; do not set fixed heights on text containers.
