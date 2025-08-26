package postgresql

import (
	"testing"

	"github.com/manusa/ai-cli/pkg/api"
)

func TestIsAvailable(t *testing.T) {
	for _, tt := range []struct {
		name        string
		databaseUri string
		pgPassword  string
		pgDatabase  string
		pgHost      string
		pgPort      string
		pgUser      string
		// expected values
		available         bool
		reason            string
		mcpServerSettings *api.McpSettings
		finalDatabaseUri  string
	}{
		{
			name:              "available via URI",
			databaseUri:       "postgresql://localhost:5432/test",
			pgPassword:        "",
			available:         true,
			reason:            "DATABASE_URI is set with postgresql schema",
			mcpServerSettings: &api.McpSettings{Type: api.McpTypeStdio, Command: "uvx", Args: []string{"postgres-mcp", "--access-mode=unrestricted"}},
			finalDatabaseUri:  "", // not explicitly set, as already set by the caller
		},
		{
			name:              "available via password",
			databaseUri:       "",
			pgPassword:        "SeCrEt",
			available:         true,
			reason:            "PGPASSWORD is set (will also consider PGDATABASE, PGHOST, PGPORT, PGUSER)",
			mcpServerSettings: &api.McpSettings{Type: api.McpTypeStdio, Command: "uvx", Args: []string{"postgres-mcp", "--access-mode=unrestricted"}},
			finalDatabaseUri:  "DATABASE_URI=postgresql://postgres:SeCrEt@localhost:5432/postgres",
		},
		{
			name:              "available via password with more variables",
			databaseUri:       "",
			pgPassword:        "000",
			pgDatabase:        "mydb",
			pgHost:            "myhost",
			pgPort:            "5435",
			pgUser:            "myuser",
			available:         true,
			reason:            "PGPASSWORD is set (will also consider PGDATABASE, PGHOST, PGPORT, PGUSER)",
			mcpServerSettings: &api.McpSettings{Type: api.McpTypeStdio, Command: "uvx", Args: []string{"postgres-mcp", "--access-mode=unrestricted"}},
			finalDatabaseUri:  "DATABASE_URI=postgresql://myuser:000@myhost:5435/mydb",
		},
		{
			name:              "available via password with mysql uri",
			databaseUri:       "mysql://localhost:3306/test",
			pgPassword:        "SeCrEt",
			available:         true,
			reason:            "PGPASSWORD is set (will also consider PGDATABASE, PGHOST, PGPORT, PGUSER)",
			mcpServerSettings: &api.McpSettings{Type: api.McpTypeStdio, Command: "uvx", Args: []string{"postgres-mcp", "--access-mode=unrestricted"}},
			finalDatabaseUri:  "DATABASE_URI=postgresql://postgres:SeCrEt@localhost:5432/postgres",
		},
		{
			name:        "not available",
			databaseUri: "",
			pgPassword:  "",
			available:   false,
			reason:      "DATABASE_URI is not set and PGPASSWORD is not set",
		},
		{
			name:        "not available with mysql database uri",
			databaseUri: "mysql://localhost:3306/test",
			pgPassword:  "",
			available:   false,
			reason:      "DATABASE_URI is not set with postgresql schema and PGPASSWORD is not set",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DATABASE_URI", tt.databaseUri)
			t.Setenv("PGPASSWORD", tt.pgPassword)
			t.Setenv("PGDATABASE", tt.pgDatabase)
			t.Setenv("PGHOST", tt.pgHost)
			t.Setenv("PGPORT", tt.pgPort)
			t.Setenv("PGUSER", tt.pgUser)
			provider := &Provider{}
			available := provider.IsAvailable(nil, nil)
			if available != tt.available {
				t.Errorf("expected postgresql to be %v", tt.available)
			}
			if provider.Reason != tt.reason {
				t.Errorf("expected reason to be %s, but got %s", tt.reason, provider.Reason)
			}
			if available {
				mcpServerSettings, err := provider.findBestMcpServerSettings(false)
				if err != nil {
					t.Errorf("expected no error, but got %v", err)
				}
				if mcpServerSettings != nil && tt.mcpServerSettings == nil {
					t.Errorf("expected mcpServerSettings to be nil, but got %v", mcpServerSettings)
				}
				if mcpServerSettings == nil && tt.mcpServerSettings != nil {
					t.Errorf("expected mcpServerSettings to be %v, but got nil", tt.mcpServerSettings)
				}
				finalDatabaseUri := ""
				if mcpServerSettings != nil && len(mcpServerSettings.Env) == 1 {
					finalDatabaseUri = mcpServerSettings.Env[0]
				}
				if finalDatabaseUri != tt.finalDatabaseUri {
					t.Errorf("expected DATABASE_URI to be %s, but got %s", tt.finalDatabaseUri, finalDatabaseUri)
				}
				if tt.mcpServerSettings != nil && mcpServerSettings != nil {
					if mcpServerSettings.Command != tt.mcpServerSettings.Command {
						t.Errorf("expected mcpServerSettings.Command to be %s, but got %s", tt.mcpServerSettings.Command, mcpServerSettings.Command)
					}
					if mcpServerSettings.Type != tt.mcpServerSettings.Type {
						t.Errorf("expected mcpServerSettings.Type to be %d, but got %d", tt.mcpServerSettings.Type, mcpServerSettings.Type)
					}
					if len(mcpServerSettings.Args) != len(tt.mcpServerSettings.Args) {
						t.Errorf("expected mcpServerSettings.Args len to be %d, but got %d", len(tt.mcpServerSettings.Args), len(mcpServerSettings.Args))
					}
					for i, arg := range mcpServerSettings.Args {
						if arg != tt.mcpServerSettings.Args[i] {
							t.Errorf("expected mcpServerSettings.Args to be %v, but got %v", tt.mcpServerSettings.Args, mcpServerSettings.Args)
						}
					}
				}
			}
		})
	}
}
