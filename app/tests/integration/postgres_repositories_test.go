//go:build integration

package integration_test

import "testing"

func TestIntegrationPlaceholder(t *testing.T) {
	t.Skip("set DATABASE_URL and run with -tags=integration to execute integration tests")
}
