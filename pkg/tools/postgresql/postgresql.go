package postgresql

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/v2/list"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/keyring"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/manusa/ai-cli/pkg/ui/components/password_input"
	"github.com/manusa/ai-cli/pkg/ui/components/selector"
)

type Provider struct {
	api.BasicToolsProvider
}

var _ api.ToolsProvider = &Provider{}

const (
	databaseUriEnvVar = "DATABASE_URI"

	pgDatabaseEnvVar  = "PGDATABASE"
	defaultPgDatabase = "postgres"

	pgHostEnvVar  = "PGHOST"
	defaultPgHost = "localhost"

	pgPortEnvVar  = "PGPORT"
	defaultPgPort = "5432"

	pgUserEnvVar  = "PGUSER"
	defaultPgUser = "postgres"

	pgPasswordEnvVar = "PGPASSWORD"
)

var (
	supportedMcpSettings = map[string]api.McpSettings{
		"uvx": {
			Type:    api.McpTypeStdio,
			Command: "uvx",
			Args: []string{
				"postgres-mcp",
			},
		},
	}
)

func (p *Provider) Initialize(ctx context.Context) {
	// TODO: probably move to features.Discover orchestration
	if cfg := config.GetConfig(ctx); cfg != nil {
		p.ToolsParameters = cfg.ToolsParameters(p.Attributes().Name())
	}

	var err error
	p.McpSettings, err = p.findBestMcpServerSettings(*p.ReadOnly)
	if err != nil {
		p.IsAvailableReason = err.Error()
		return
	}

	if available := strings.HasPrefix(p.getDatabaseURI(), "postgresql://"); available {
		p.Available = true
		p.IsAvailableReason = fmt.Sprintf("%s is set with postgresql schema", databaseUriEnvVar)
		return
	}

	if pgpassword := os.Getenv(pgPasswordEnvVar); pgpassword != "" {
		p.Available = true
		p.IsAvailableReason = fmt.Sprintf("%s is set (will also consider %s)", pgPasswordEnvVar, strings.Join([]string{pgDatabaseEnvVar, pgHostEnvVar, pgPortEnvVar, pgUserEnvVar}, ", "))
		return
	}

	if os.Getenv(databaseUriEnvVar) == "" {
		p.IsAvailableReason = fmt.Sprintf("%s is not set and %s is not set", databaseUriEnvVar, pgPasswordEnvVar)
	} else {
		p.IsAvailableReason = fmt.Sprintf("%s is not set with postgresql schema and %s is not set", databaseUriEnvVar, pgPasswordEnvVar)
	}
}

func (p *Provider) getDatabaseURI() string {
	if key, err := keyring.GetKey(databaseUriEnvVar); err == nil && len(key) > 0 {
		return key
	}
	return os.Getenv(databaseUriEnvVar)
}

func (p *Provider) findBestMcpServerSettings(readOnly bool) (*api.McpSettings, error) {
	for command, settings := range supportedMcpSettings {
		if config.CommandExists(command) {
			if readOnly {
				settings.Args = append(settings.Args, "--access-mode=restricted")
			} else {
				settings.Args = append(settings.Args, "--access-mode=unrestricted")
			}

			// Get or build URI
			if databaseUri := os.Getenv(databaseUriEnvVar); !strings.HasPrefix(databaseUri, "postgresql://") && os.Getenv(pgPasswordEnvVar) != "" {
				uri := fmt.Sprintf(
					"postgresql://%s:%s@%s:%s/%s",
					p.getEnvVarValueOrDefault(pgUserEnvVar, defaultPgUser),
					os.Getenv(pgPasswordEnvVar),
					p.getEnvVarValueOrDefault(pgHostEnvVar, defaultPgHost),
					p.getEnvVarValueOrDefault(pgPortEnvVar, defaultPgPort),
					p.getEnvVarValueOrDefault(pgDatabaseEnvVar, defaultPgDatabase),
				)
				settings.Env = append(settings.Env, fmt.Sprintf("%s=%s", databaseUriEnvVar, uri))
			}
			return &settings, nil
		}
	}
	return nil, errors.New("no suitable MCP settings found for the PostgreSQL MCP server")
}

func (p *Provider) getEnvVarValueOrDefault(envVar string, defaultValue string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return defaultValue
}

func (p *Provider) InstallHelp() error {
	registerExistingInstance := "Register an existing PostgreSQL instance using complete connection string (postgresql://user:password@host:port/database)"
	quit := "Terminate PostgreSQL setup"
	choices := []list.Item{
		selector.Item(registerExistingInstance),
		selector.Item(quit),
	}
	for {
		choice, err := selector.Select("Please select a step:", choices)
		if err != nil {
			return err
		}
		switch choice {
		case registerExistingInstance:
			fmt.Printf("Paste your connection string below (postgresql://user:password@host:port/database):\n")
			apiKey, err := password_input.Prompt()
			if err != nil {
				return err
			}
			err = keyring.SetKey(databaseUriEnvVar, apiKey)
			if err != nil {
				return err
			}
		case quit:
			return nil
		}
	}
}

func (p *Provider) Clear(ctx context.Context) (done bool, err error) {
	return keyring.DeleteKey(databaseUriEnvVar)
}

var instance = &Provider{
	BasicToolsProvider: api.BasicToolsProvider{
		BasicToolsAttributes: api.BasicToolsAttributes{
			BasicFeatureAttributes: api.BasicFeatureAttributes{
				FeatureName:        "postgresql",
				FeatureDescription: "Provides access to a PostgreSQL database, allowing execution of SQL queries and retrieval of data.",
				SupportsSetupAttr:  true,
			},
		},
	},
}

func init() {
	tools.Register(instance)
}
