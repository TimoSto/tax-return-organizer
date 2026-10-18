The persona builder is where advisors define **what a type of client has to provide**. A persona (Angestellt, Vermietung, Selbstständig, Rente, …) is a named bundle of required **documents** and required **information fields**. This is a **list-detail layout with a supporting pane**.

**Structure** (left to right): navigation rail | **list pane** (all personas, with counts of documents, fields and clients using them) | **detail pane** (the selected persona: name, description, tabs Unterlagen / Angaben / Mandanten / Einstellungen, sortable rows with a "Pflicht" switch) | **supporting pane** (preview as the client sees it, usage, warning).

- Rows are ordered by drag (`drag_indicator`); the order is the order the client sees. Each document has a name, a hint for the client ("Vom Vermieter oder der Hausverwaltung") and Pflicht on/off. Each field has a name, a type (Zahl, Text, Datum, Ja/Nein, Auswahl) and Pflicht.
- Suggest common items as assist chips ("Grundsteuerbescheid") so advisors do not type everything.
- Changing a persona affects every client who has it: always show the count in the supporting pane and confirm with the Dialog.
- The only filled button in the detail pane is **Speichern**; Duplizieren is tonal.
- The supporting pane must show the client-facing wording, so advisors can write hints in plain German.

**Implementation:** Custom Svelte layout (`ListDetailLayout.svelte`, `Pane.svelte`) on CSS grid with the `pane-*` and `nav-rail-width` tokens. Rows use `<md-list-item>`, switches `<md-switch>`, chips `<md-assist-chip>`, tabs `<md-tabs>`; drag sorting is custom.
