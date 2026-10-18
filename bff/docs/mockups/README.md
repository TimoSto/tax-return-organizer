# Mockups

Static mockups of how the product screens are intended to look, built from the [design system](../design-system/README.md). They use sample data and are a reference for implementation, not a source of truth: when a screen is built in Svelte, the real component wins.

| Mockup | What it shows |
|---|---|
| [Advisor home](advisor-home/README.md) | Landing page: what needs attention today |
| [Persona builder](persona-builder/README.md) | Advisors define required documents and information per persona |
| [Client setup](client-setup/README.md) | Advisors create a client and assign personas |
| [Client portal](client-portal/README.md) | Clients upload documents and enter information |

Each folder has `mockup.html`, a `README.md` with the structure and rules of the screen, and screenshots (`screenshot.png`, plus `-dark` and `-high-contrast` where useful).

## Browsing

```bash
cd docs && python3 -m http.server 8080
```

Open `http://localhost:8080/mockups/` for a viewer with a theme switcher. It needs network access for the Google Fonts and icons.

`mockups.css` holds styles used only by these mockups (requirement rows, selectable persona cards, KPI grid). If one of them becomes a real component, move it into `design-system/components/<Name>/` with a README.

Dependency direction: mockups depend on the design system, never the reverse.
