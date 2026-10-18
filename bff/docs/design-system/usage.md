# Using Material components in this product

The standard components follow **Material Design 3** and are implemented with **`@material/web`** in Svelte. This page does not repeat their specs. For each one it links the M3 guidance and the Material Web docs, and lists only what is specific to this product. Theme tokens come from `tokens.css`; `components/bundle.css` maps them to the `--md-sys-*` variables Material Web reads.

Components that M3 does not provide are documented in full under [`components/`](components/): DataTable, StatusChip, FileUpload and ProgressIndicator (steps). Page layout is in [`guidelines/layout.md`](guidelines/layout.md).

Common rules for everything below: German UI text, **Sie** for clients, sentence case, minimum 48px hit area, state layers from the `state-*` tokens, visible 3px focus ring (4px in high contrast), and no information carried by color alone.

## Button

[M3](https://m3.material.io/components/buttons/overview) · [Material Web](https://github.com/material-components/material-web/blob/main/docs/components/button.md) · `<md-filled-button>`, `<md-filled-tonal-button>`, `<md-outlined-button>`, `<md-text-button>`, `<md-elevated-button>`

- At most **one filled button per view** (Speichern, Einladung senden). Secondary actions are tonal; dialog dismissal is text.
- Labels are verbs: "Datei hochladen", "Persona löschen", never "OK" or "Absenden" for consequential actions.
- Destructive confirm: filled button with `--md-filled-button-container-color: var(--error)` and `--md-filled-button-label-text-color: var(--on-error)`. It is never the default focus in a dialog.
- A disabled button is explained by helper text next to it.

## FAB

[M3](https://m3.material.io/components/floating-action-button/overview) · [Material Web](https://github.com/material-components/material-web/blob/main/docs/components/fab.md) · `<md-fab>`

- One per screen, at the top of the navigation rail (icon only) or at the end of a list-pane header (extended).
- Advisor: **Neue Persona** on the persona page, **Mandant anlegen** on the client list. The client portal has no FAB; its primary action is the upload button in each requirement.

## Chips

[M3](https://m3.material.io/components/chips/overview) · [Material Web](https://github.com/material-components/material-web/blob/main/docs/components/chip.md) · `<md-chip-set>`, `<md-filter-chip>`, `<md-input-chip>`, `<md-assist-chip>`, `<md-suggestion-chip>`

- Filter chips filter lists (Alle, Unterlagen fehlen, Frist diesen Monat). Input chips show assigned personas and are removable. Suggestion chips offer common documents in the persona builder ("Grundsteuerbescheid").
- **Status is not a chip.** Use StatusChip, which is not interactive.

## Text fields and select

[M3](https://m3.material.io/components/text-fields/overview) · [Material Web: text field](https://github.com/material-components/material-web/blob/main/docs/components/text-field.md), [select](https://github.com/material-components/material-web/blob/main/docs/components/select.md) · `<md-outlined-text-field>`, `<md-filled-text-field>`, `<md-outlined-select>`

- Outlined by default; filled only in dense toolbars. Clients type in these fields, so labels are plain German nouns ("Steuer-ID") and the **format is shown before they type** in the supporting text ("11 Ziffern, steht auf Ihrem Steuerbescheid"). Errors replace it and say how to fix the problem ("Bitte im Format TT.MM.JJJJ eingeben").
- German formats: decimal comma, `€` as a trailing suffix, IBAN in groups of four, dates `TT.MM.JJJJ`. Money fields are right-aligned with tabular figures.
- Prefill what the advisor already knows; the client confirms. Save per field and show "Gespeichert" in the supporting text.

## Switch and checkbox

[M3](https://m3.material.io/components/switch/overview) · [Material Web: switch](https://github.com/material-components/material-web/blob/main/docs/components/switch.md), [checkbox](https://github.com/material-components/material-web/blob/main/docs/components/checkbox.md) · `<md-switch>`, `<md-checkbox>`

- A switch applies immediately; a checkbox is confirmed with a button. In the persona builder a switch marks a document or field as **Pflicht**.
- Always a visible label beside the control, never inside; the state is also shown by thumb position and icon.

## Tabs

[M3](https://m3.material.io/components/tabs/overview) · [Material Web](https://github.com/material-components/material-web/blob/main/docs/components/tabs.md) · `<md-tabs>`, `<md-primary-tab>`

- Primary tabs under a pane header, 2 to 6 tabs, one or two German words ("Unterlagen", "Angaben", "Mandanten", "Einstellungen"). A badge only for counts that need attention.
- Tabs switch views of the same record; navigation between sections of the app is the rail's job.

## Dialog

[M3](https://m3.material.io/components/dialogs/overview) · [Material Web](https://github.com/material-components/material-web/blob/main/docs/components/dialog.md) · `<md-dialog>`

- Use for deletes, sending an invitation, and **changing a persona used by many clients** (state the number of affected clients). Not for validation errors or long forms.
- The title states the decision; the confirm label names the consequence ("Persona löschen", "Änderung übernehmen"), the dismiss is a text button "Abbrechen".

## List

[M3](https://m3.material.io/components/lists/overview) · [Material Web](https://github.com/material-components/material-web/blob/main/docs/components/list.md) · `<md-list>`, `<md-list-item>`

- Headline is a name, not an ID. Two-line items for clients and personas; trailing element is a StatusChip or a count.
- Exactly one selected item in a list-detail layout (`secondary-container`; in high contrast also bordered and underlined).
- Order by what needs action first unless a search is active. Truncate with an ellipsis.

## Progress indicators

[M3](https://m3.material.io/components/progress-indicators/overview) · [Material Web](https://github.com/material-components/material-web/blob/main/docs/components/progress.md) · `<md-linear-progress>`, `<md-circular-progress>`

- Always pair with text that says the same in words ("9 von 12 Unterlagen"). Never indeterminate for known work.
- The filing-journey **steps** are custom: see [`components/ProgressIndicator`](components/ProgressIndicator/README.md).

## Card

[M3](https://m3.material.io/components/cards/overview) · no Material Web component; build `Card.svelte`.

- Filled by default. **Tertiary** is reserved for deadlines and "waiting on the client"; **primary** for the single most important result; an `error-container` card only for overdue.
- Do not nest cards. In high contrast every card has a 2px `outline` border.

## Top app bar

[M3](https://m3.material.io/components/top-app-bar/overview) · no Material Web component; build `TopAppBar.svelte`.

- Only on full-width pages without panes (settings). Pane-based pages use a pane header instead.

## Navigation rail

[M3](https://m3.material.io/components/navigation-rail/overview) · no Material Web component; build `NavigationRail.svelte` from `<md-ripple>`, `<md-focus-ring>` and `<md-icon>` with markup `nav > a[aria-current]`.

- The desktop product's primary navigation, 3 to 7 destinations, German one-word labels. **Advisor:** Übersicht, Mandanten, Personas, Fristen, Nachrichten. **Client:** Start, Aufgaben, Unterlagen, Nachrichten. Settings pinned at the bottom.
- Badges show counts that need the user and have an accessible name ("Nachrichten, 4 ungelesen").
- The expanded rail (220px, labels beside icons) is an opt-in for users and long labels.

## Not used

Navigation bar (bottom navigation), compact and medium window layouts, bottom sheets and phone patterns: the product is desktop only, minimum 1024px.
