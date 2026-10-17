In client setup the advisor **creates a client and assigns personas**; the system then derives the client's checklist. List-detail with a supporting pane.

**Structure:** navigation rail | **list pane** (clients, with a status chip) | **detail pane** (client data form, then persona picker as selectable cards, then the send action) | **supporting pane** (the resulting requirements, live, grouped by persona).

- Persona cards are multi-select (`md-select-card`, selected = `secondary-container` + checked box + border in high contrast). A client may have any number of personas.
- The supporting pane updates as the advisor selects: total documents and fields, a group per persona, and a note when duplicates were merged ("Personalausweis ist in 2 Personas enthalten und wird nur einmal angefragt").
- A requirement that is Pflicht in at least one persona is Pflicht for the client.
- Actions: **Einladung senden** (filled) emails the client; **Als Entwurf speichern** is text. Disable "Einladung senden" until name, e-mail and at least one persona are set, and say why.

**Implementation:** Custom Svelte layout (`ListDetailLayout.svelte`). Form fields use `<md-outlined-text-field>` and `<md-outlined-select>`; persona cards are a custom `SelectCard.svelte` wrapping `<md-checkbox>`; summary cards are `Card.svelte`.
