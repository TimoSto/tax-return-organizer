Dialogs interrupt for a decision that cannot be undone or needs a short answer.

**Consumer provides:** a title that states the decision, one or two sentences naming the consequence, and one or two actions.

- Container `surface-container-high`, radius `shape-extra-large`, padding `space-5`, max width 560px; backdrop `scrim` at 32%.
- Actions end-aligned: text button to dismiss, filled (or error) to confirm. Name the consequence in the confirm label ("Persona löschen"), not "OK".
- Use for deleting, sending an invitation, changing a persona used by many clients. Not for validation errors (inline) or long forms.

**Implementation:** `<md-dialog>` from `@material/web` with `<md-text-button>` actions in the `actions` slot.
