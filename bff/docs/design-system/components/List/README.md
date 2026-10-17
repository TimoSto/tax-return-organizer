Lists present collections people scan and select: clients, personas, documents, messages.

**Consumer provides:** a headline per item (name, not ID), optional supporting text, optional leading element (avatar, icon, checkbox) and trailing element (status chip, count, date).

- One line = 56px, two lines = 72px. Headline `title-medium`, supporting text `body-medium`, trailing metadata `label-medium`.
- The selected item uses `secondary-container` and `shape-large`; in a list-detail layout exactly one item is selected. In high contrast it also gets a border and an underlined headline.
- Order by what needs action first, unless a search is active. Truncate with an ellipsis; the detail pane has the full text.

**Implementation:** `<md-list>` and `<md-list-item>` for rows (`headline`, `supporting-text` slots, `start` and `end` slots); `<md-checkbox>` for the check; avatar is custom.
