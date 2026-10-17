The navigation rail is the primary navigation of the desktop product. It sits at the left edge on `surface`, outside the pane area.

**Consumer provides:** 3 to 7 top-level destinations (icon + one-word German label), the current destination, and optionally a FAB at the top and a utility destination pinned to the bottom.

- Collapsed width `nav-rail-width` (80px): icon in a 56×32 indicator, label under it in `label-medium`. Active: `secondary-container` indicator, filled icon, bolder label (and a border in high contrast).
- Expanded rail (`md-rail--expanded`, 220px): icon and label in a row, for longer German labels or users who prefer it.
- Badges show counts that need the user and carry an accessible name ("Nachrichten, 4 ungelesen").
- Advisor: Übersicht, Mandanten, Personas, Fristen, Nachrichten. Client: Start, Aufgaben, Unterlagen, Nachrichten. Never more than seven.

**Implementation:** Custom Svelte component (`NavigationRail.svelte`); `@material/web` has no navigation rail. Build it from `<md-ripple>`, `<md-focus-ring>` and `<md-icon>` and keep the markup `nav > a[aria-current]`.
