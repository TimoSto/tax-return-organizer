Status chips show the state of a filing, document or task in lists, tables and headers. An addition to the Material set.

**Consumer provides:** one of six states. Every state has an icon **and** a word; color is never the only signal.

| State | Class | Meaning |
|---|---|---|
| Entwurf | (default) | In preparation |
| In Prüfung | `md-status--info` | With the advisor |
| Wartet auf Mandant | `md-status--wait` | Waiting for the client |
| Freigegeben | `md-status--ok` | Client and advisor agree |
| Eingereicht | `md-status--done` | Filed with the Finanzamt |
| Überfällig | `md-status--late` | Past a deadline; the most prominent state |

Document-level states reuse the same chips: Offen (default), Hochgeladen (info), Geprüft (ok), Rückfrage (wait), Abgelehnt (late). Height 24px; one chip per row.

**Implementation:** Custom Svelte component (`StatusChip.svelte`) with a `status` prop mapping to the classes above. Do not use `<md-assist-chip>`: it is interactive.
