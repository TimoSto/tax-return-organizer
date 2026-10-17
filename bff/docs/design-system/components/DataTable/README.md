Data tables present figures and line items for comparison and audit. An addition to the Material set, because advisors work in numbers.

**Consumer provides:** column definitions, row data, and which row (if any) is a total.

- Right-align numeric columns, header too, with tabular figures (`num`). German format: `301.215,00 €` (thousands point, decimal comma, symbol after the number with a non-breaking space). Negative amounts use a minus sign, not red.
- Put the source in a `.sub` line instead of adding columns. Row height 52px, header 48px; at most 7 columns.
- Totals use `tr.is-total`. No zebra stripes.

**Implementation:** Custom Svelte component (`DataTable.svelte`); `@material/web` has no data table.
