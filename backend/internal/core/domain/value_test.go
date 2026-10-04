package domain_test

import (
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/core/domain"
)

func TestValueAndRequirementKinds(t *testing.T) {
	pairs := []struct {
		req  domain.Requirement
		val  domain.Value
		kind domain.ValueKind
	}{
		{domain.DocumentRequirement{}, domain.DocumentValue{}, domain.KindDocument},
		{domain.TextRequirement{}, domain.TextValue{}, domain.KindText},
		{domain.NumberRequirement{}, domain.NumberValue{}, domain.KindNumber},
		{domain.YearRequirement{}, domain.YearValue{}, domain.KindYear},
		{domain.BoolRequirement{}, domain.BoolValue{}, domain.KindBool},
	}
	for _, p := range pairs {
		if p.req.Kind() != p.kind || p.val.Kind() != p.kind {
			t.Errorf("%T/%T kinds = %v/%v, want %v", p.req, p.val, p.req.Kind(), p.val.Kind(), p.kind)
		}
	}
}
