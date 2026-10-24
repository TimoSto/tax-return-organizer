package pgdb_test

import (
	"os"
	"testing"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/outbound/postgres/postgrestest"
)

// One container is shared by all tests in this package.
func TestMain(m *testing.M) {
	os.Exit(postgrestest.Run(m))
}
