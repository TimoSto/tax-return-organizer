The client portal is where clients **see what they must provide, upload documents and enter information**. Desktop only. It uses the same rail and panes, with fewer destinations and one task at a time.

**Structure:** navigation rail (Start, Aufgaben, Unterlagen, Nachrichten) | **detail pane** (greeting, overall progress, requirements grouped as "Unterlagen hochladen" and "Angaben ergänzen") | **supporting pane** (deadline, progress, advisor).

- Every requirement is a row (`md-req`) with an icon, a title, a plain-language hint, and its state. Done requirements are collapsed with a green check and the file name; open ones show the upload zone or the field right there, never on another page.
- Group by persona topic only if it helps ("Ihre Mieteinnahmen"); never show the word "Persona" to clients.
- Address clients with "Sie". One primary action per requirement ("Datei hochladen"). Explain tax words once: "Anlage V (Formular für Mieteinnahmen)".
- Saving is automatic per field; show "Gespeichert" in the field's supporting text. A single **Angaben abschicken** button tells the advisor the section is complete.
- Body copy is `body-large`; page title `headline-medium`. Never show advisor-only data (notes, status codes, fees).

**Implementation:** Custom Svelte layout (`DetailSupportLayout.svelte`). Fields use `<md-outlined-text-field>` (label, supporting-text, error-text), the upload uses `FileUpload.svelte`, status uses `StatusChip.svelte`, rings and progress use `<md-circular-progress>` / `<md-linear-progress>`.
