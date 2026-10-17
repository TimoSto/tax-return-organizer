Text fields collect values. **Outlined** is the default; use **filled** in dense toolbars.

**Consumer provides:** a persistent label (a noun, e.g. "Steuer-ID"), the value, optional supporting text, and error text that says how to fix the problem. Clients type in these fields, so every label is in plain German and every format is shown before they type.

- Label floats above the value; never use a placeholder as the only label. In the CSS replica set `placeholder=" "`.
- Height 56px; border `outline`, 2px `primary` on focus, `error` with icon and message on error. In high contrast borders are 2px `outline`.
- Supporting text explains the format ("11 Ziffern, ohne Leerzeichen"); errors replace it.
- Money fields right-align with tabular figures and a trailing `€`. German formats: IBAN with spaces in groups of four, date `TT.MM.JJJJ`, decimal comma.
- Prefill what the advisor already knows; the client confirms.
- Select uses the same container with a trailing arrow.

**Implementation:** `<md-outlined-text-field>` and `<md-filled-text-field>` for text, `<md-outlined-select>` with `<md-select-option>` for choices. Set `--md-outlined-text-field-container-shape` through the shape tokens (already mapped).
