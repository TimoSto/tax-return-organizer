The floating action button is the single most important creation action of a screen: **Neue Persona** on the persona page, **Mandant anlegen** on the client list. The client portal has none; its primary action is the upload button in each requirement.

**Consumer provides:** one icon, and a label for the extended variant.

- Place it at the top of the navigation rail (icon only) or at the end of the list-pane header (extended).
- One FAB per screen. Do not combine it with a filled button for the same action.

**Implementation:** `<md-fab>` from `@material/web` (`size`, `variant`, `label` attributes).
