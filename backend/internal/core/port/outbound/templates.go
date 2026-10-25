package outbound

import "github.com/TimoSto/tax-return-organizer/backend/internal/core/personatemplate"

// PersonaTemplateCatalog provides the predefined personas every new collector starts with.
type PersonaTemplateCatalog interface {
	PersonaTemplates() []personatemplate.PersonaTemplate
}
