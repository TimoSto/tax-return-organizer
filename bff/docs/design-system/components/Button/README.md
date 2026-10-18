Buttons start an action. Use five emphasis levels and at most one filled button per view.

**Consumer provides:** a label (verb first: "Einladung senden", "Datei hochladen"), optionally a leading icon.

- Filled: the one primary action of a pane or form. Tonal: secondary actions that still matter. Outlined: alternatives of equal weight. Text: tertiary actions and dialog dismissal. Elevated: only on busy surfaces.
- Error button confirms an irreversible action (Persona löschen); never the default focus in a dialog.
- Height 40px, touch target 48px. Labels use `label-large`, sentence case, German ("Speichern", not "Save").
- A disabled button is explained by helper text next to it; do not disable without saying why.
- In high contrast, filled buttons gain an inner border; the focus ring grows to 4px.

**Implementation:** `<md-filled-button>`, `<md-filled-tonal-button>`, `<md-outlined-button>`, `<md-text-button>`, `<md-elevated-button>` from `@material/web`. The error variant is the filled button with `--md-filled-button-container-color: var(--error)` and `--md-filled-button-label-text-color: var(--on-error)`.
