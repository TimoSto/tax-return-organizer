# Screenshots

Static renders of the design-system previews (`../components/<Name>/preview.html`), made with Playwright and Chromium at 1x scale. They show how the app is intended to look; the data is sample data. Light theme unless the file name says otherwise (`-dark`, `-high-contrast`).

For live, themeable previews, run `python3 -m http.server 8080` in `docs/design-system` and open `http://localhost:8080` (see `index.html`).

To regenerate, render each `preview.html` with `../tokens.css` and `../components/bundle.css` loaded and `data-theme` set on `<html>`. Fonts and icons load from Google Fonts, so rendering needs network access.

## Screens

### Advisor home (Übersicht)

![Advisor home, light](AdvisorHome.png)

| Dark | High contrast |
|---|---|
| ![Advisor home, dark](AdvisorHome-dark.png) | ![Advisor home, high contrast](AdvisorHome-high-contrast.png) |

### Persona builder

Advisors define which documents and information a type of client has to provide.

![Persona builder, light](PersonaBuilder.png)

![Persona builder, high contrast](PersonaBuilder-high-contrast.png)

### Client setup

Advisors create a client and assign personas; the checklist is derived from them.

![Client setup](ClientSetup.png)

### Client portal

Clients see what they have to provide, upload documents and enter information.

![Client portal, light](ClientPortal.png)

| Dark | High contrast |
|---|---|
| ![Client portal, dark](ClientPortal-dark.png) | ![Client portal, high contrast](ClientPortal-high-contrast.png) |

### Window size classes

![Window size classes](WindowSizeClasses.png)

## Components

| | |
|---|---|
| ![Buttons](Button.png) | ![Status chips](StatusChip.png) |
| ![Text fields](TextField.png) | ![File upload](FileUpload.png) |
| ![Data table](DataTable.png) | ![Navigation rail](NavigationRail.png) |
| ![Dialog](Dialog.png) | ![Cover](Cover.png) |
