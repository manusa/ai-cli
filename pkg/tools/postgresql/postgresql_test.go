package postgresql

import (
	"testing"
)

func TestIsAvailable(t *testing.T) {
	for _, tt := range []struct {
		name        string
		databaseUri string
		pgPassword  string
		// expected values
		available bool
		reason    string
	}{
		{name: "available via URI", databaseUri: "postgresql://localhost:5432/test", pgPassword: "", available: true, reason: "DATABASE_URI is set"},
		{name: "available via password", databaseUri: "", pgPassword: "SeCrEt", available: true, reason: "PGPASSWORD is set (will also consider PGDATABASE, PGHOST, PGPORT, PGUSER)"},
		{name: "not available", databaseUri: "", pgPassword: "", available: false, reason: "DATABASE_URI and PGPASSWORD are not set"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DATABASE_URI", tt.databaseUri)
			t.Setenv("PGPASSWORD", tt.pgPassword)
			provider := &Provider{}
			available := provider.IsAvailable(nil, nil)
			if available != tt.available {
				t.Errorf("expected postgresql to be %v", tt.available)
			}
			if provider.Reason != tt.reason {
				t.Errorf("expected reason to be %s, but got %s", tt.reason, provider.Reason)
			}
		})
	}
}
