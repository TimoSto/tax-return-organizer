## Design system

The design system lives in `docs/design-system/`. Read `docs/design-system/README.md` first, then the matching `components/<Name>/README.md` before building a component.

- Use tokens only (`tokens.css`); never hard-code colors, radii, spacing or fonts.
- Use `@material/web` components where a component README says so; otherwise build a Svelte component using the `md-*` rules in `components/bundle.css`.
- UI text is German (Sie for clients) and goes through the existing inlang messages. Money `4.820,00 €`, dates `31.07.2025`.
- Desktop only, minimum 1024px. Support `data-theme` light, dark and high-contrast.
