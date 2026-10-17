File upload is how clients provide documents. An addition to the Material set.

**Consumer provides:** the requirement name, accepted types and size ("PDF, JPG oder PNG, max. 10 MB"), and callbacks for add and remove.

- Empty state: dashed drop zone with an upload icon, one sentence ("Datei hierher ziehen oder auswählen") and an outlined button. Drag-over uses `primary-container`.
- Uploaded files become rows (`md-file`): icon, name, size and date, and a remove icon button. A failed file turns `error-container` with the reason and a retry button; never lose the file name.
- Uploading shows a linear progress bar and a "Abbrechen" action. Announce completion to screen readers ("Datei hochgeladen: Mietvertrag.pdf").
- Provide a keyboard path: the button opens the system file dialog; the drop zone alone is never the only way.
- Do not ask for a document twice: once uploaded, the requirement shows "Hochgeladen", and the advisor sets "Geprüft".

**Implementation:** Custom Svelte component (`FileUpload.svelte`) wrapping a visually hidden `<input type="file">`; use `<md-outlined-button>` for the trigger, `<md-icon-button>` for remove and `<md-linear-progress>` for progress.
