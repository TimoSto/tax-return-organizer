Tabs switch between views of the same subject inside a pane: a persona's Unterlagen, Angaben and Einstellungen.

**Consumer provides:** 2 to 6 tab labels (one or two words), and a badge only for counts that need attention.

- Primary tabs under a pane header. The active tab uses `primary` text and a 3px indicator (4px in high contrast).
- Tabs never navigate between sections of the app; that is the rail's job.

**Implementation:** `<md-tabs>` with `<md-primary-tab>` from `@material/web`.
