Switches toggle one setting immediately. In the persona builder a switch marks a document or field as **Pflicht** (required).

**Consumer provides:** a visible label next to the switch and the state. State is also conveyed by the thumb position and icon, not color alone.

- Use a switch for settings that apply at once; use a checkbox for choices confirmed with a button.
- Track 52×32; the label is always to the left or right, never inside.

**Implementation:** `<md-switch>` and `<md-checkbox>` from `@material/web`; wrap in a `<label>`.
