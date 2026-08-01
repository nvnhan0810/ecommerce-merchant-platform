package postgres_test

import (
	"testing"

	"github.com/nvnhan0810/ecomerce-api/internal/platform/postgres"
)

func TestMigrateHelpers_should_reject_empty_url(t *testing.T) {
	t.Parallel()
	if err := postgres.Up(""); err == nil {
		t.Fatal("expected error for empty database url")
	}
}
