package postgresql

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/manusa/ai-cli/pkg/tools/utils/eino"
)

type Provider struct {
	api.BasicToolsProvider
	ReadOnly bool `json:"-"`
}

var _ api.ToolsProvider = &Provider{}

type PostgresqlPolicies struct{}

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

func (p *Provider) Initialize(_ context.Context) {
	var err error
	p.McpSettings, err = p.findBestMcpServerSettings(p.ReadOnly)
	if err != nil {
		p.IsAvailableReason = err.Error()
		return
	}

	if available := strings.HasPrefix(os.Getenv(databaseUriEnvVar), "postgresql://"); available {
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

func (p *Provider) GetTools(ctx context.Context) ([]*api.Tool, error) {
	mcpSettings, err := p.findBestMcpServerSettings(p.ReadOnly)
	if err != nil || mcpSettings.Type != api.McpTypeStdio {
		return nil, err
	}

	cli, err := eino.StartMcp(ctx, mcpSettings.Env, slices.Concat([]string{mcpSettings.Command}, mcpSettings.Args))
	if err != nil {
		return nil, err
	}
	return eino.GetTools(ctx, cli)
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

var instance = &Provider{
	BasicToolsProvider: api.BasicToolsProvider{
		BasicToolsAttributes: api.BasicToolsAttributes{
			BasicFeatureAttributes: api.BasicFeatureAttributes{
				FeatureName:        "postgresql",
				FeatureDescription: "Provides access to a PostgreSQL database, allowing execution of SQL queries and retrieval of data.",
			},
		},
	},
}

func init() {
	tools.Register(instance)
}
