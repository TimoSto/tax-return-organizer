Progress indicators show how far a client has come with their requirements.

**Consumer provides:** a value 0 to 100 and a text label with the same information ("9 von 12 Unterlagen").

- Linear (4px) inside a card or list; circular (`md-ring`) for one headline value; steps (`md-steps`) for the filing journey: Unterlagen, Prüfung, Freigabe, Eingereicht.
- Track `secondary-container`, indicator `primary`. Set `--ring-bg` on rings and steps to the surface they sit on.
- Steps show a check when done, an outlined number when current; always label every step.
- Never use an indeterminate loader for known work.

**Implementation:** `<md-linear-progress>` and `<md-circular-progress>` from `@material/web` (set `value`, `max`); the steps are a custom Svelte component (`Steps.svelte`).
