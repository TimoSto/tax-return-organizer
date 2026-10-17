Chips filter, suggest or tag. They are compact and sit in rows above lists or inside forms.

**Consumer provides:** a short label (1 to 3 words), an optional leading icon.

- Filter chips toggle a list filter and show a check when selected. Input chips represent entered values (an assigned persona) and are removable. Assist and suggestion chips trigger an action or prefill; they are outlined and never selected.
- Height 32px, radius `shape-small`. Keep rows to one line.
- Status in lists and tables is shown with **StatusChip**, not these chips.

**Implementation:** `<md-chip-set>` with `<md-filter-chip>`, `<md-input-chip>`, `<md-assist-chip>`, `<md-suggestion-chip>` from `@material/web`.
