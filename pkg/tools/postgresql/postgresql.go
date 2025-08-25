package postgresql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/policies"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/manusa/ai-cli/pkg/tools/utils/eino"
)

type Provider struct {
	tools.BasicToolsProvider
	ReadOnly bool
}

type PostgresqlPolicies struct {
	policies.ToolPolicies
}

var _ tools.Provider = &Provider{}

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
		"podman": {
			Type:    api.McpTypeStdio,
			Command: "podman", // TODO: Note that this is platform dependent (on windows this is podman.exe)
			Args: []string{
				"run",
				"-i",
				"--rm",
				"-e",
				"DATABASE_URI",
				"crystaldba/postgres-mcp",
			},
		},
	}
)

func (p *Provider) Attributes() tools.Attributes {
	return tools.Attributes{
		BasicFeatureAttributes: api.BasicFeatureAttributes{
			FeatureName: "postgresql",
		},
	}
}

func (p *Provider) Data() tools.Data {
	data := tools.Data{
		BasicFeatureData: api.BasicFeatureData{
			Reason: p.Reason,
		},
	}
	settings, err := p.findBestMcpServerSettings(p.ReadOnly)
	if err == nil {
		data.McpSettings = settings
	}
	return data
}

func (p *Provider) IsAvailable(_ *config.Config, toolPolicies any) bool {
	if !policies.IsEnabledByPolicies(toolPolicies) {
		p.Reason = "postgresql is not authorized by policies"
		return false
	}

	if policies.IsReadOnlyByPolicies(toolPolicies) {
		p.ReadOnly = true
	}

	if available := os.Getenv(databaseUriEnvVar) != ""; available {
		p.Reason = fmt.Sprintf("%s is set", databaseUriEnvVar)
		return true
	}

	if pgpassword := os.Getenv(pgPasswordEnvVar); pgpassword != "" {
		p.Reason = fmt.Sprintf("%s is set (will also consider %s)", pgPasswordEnvVar, strings.Join([]string{pgDatabaseEnvVar, pgHostEnvVar, pgPortEnvVar, pgUserEnvVar}, ", "))
		return true
	}

	p.Reason = fmt.Sprintf("%s and %s are not set", databaseUriEnvVar, pgPasswordEnvVar)

	return false
}

func (p *Provider) GetTools(ctx context.Context, _ *config.Config) ([]*api.Tool, error) {
	mcpSettings, err := p.findBestMcpServerSettings(p.ReadOnly)
	if err != nil || mcpSettings.Type != api.McpTypeStdio {
		return nil, err
	}

	cli, err := eino.StartMcp(ctx, slices.Concat([]string{mcpSettings.Command}, mcpSettings.Args))
	if err != nil {
		return nil, err
	}
	return eino.GetTools(ctx, cli)
}

func (p *Provider) MarshalJSON() ([]byte, error) {
	return json.Marshal(tools.Report{
		Attributes: p.Attributes(),
		Data:       p.Data(),
	})
}

func (p *Provider) findBestMcpServerSettings(readOnly bool) (*api.McpSettings, error) {
	for command, settings := range supportedMcpSettings {
		if commandExists(command) {
			if readOnly {
				settings.Args = append(settings.Args, "--access-mode=restricted")
			} else {
				settings.Args = append(settings.Args, "--access-mode=unrestricted")
			}

			// Get or build URI
			uri := ""
			if databaseUri := os.Getenv(databaseUriEnvVar); databaseUri != "" {
				uri = databaseUri
			} else if os.Getenv(pgPasswordEnvVar) != "" {
				uri = fmt.Sprintf(
					"postgresql://%s:%s@%s:%s/%s",
					p.getEnvVarValueOrDefault(pgUserEnvVar, defaultPgUser),
					os.Getenv(pgPasswordEnvVar),
					p.getEnvVarValueOrDefault(pgHostEnvVar, defaultPgHost),
					p.getEnvVarValueOrDefault(pgPortEnvVar, defaultPgPort),
					p.getEnvVarValueOrDefault(pgDatabaseEnvVar, defaultPgDatabase),
				)
			}

			// replace localhost with host.containers.internal in the DATABASE_URI environment variable
			for i, arg := range settings.Args {
				if arg == "DATABASE_URI" {
					settings.Args[i] = fmt.Sprintf("DATABASE_URI=%s", strings.Replace(uri, "localhost", "host.containers.internal", 1))
				}
			}
			return &settings, nil
		}
	}
	return nil, errors.New("no suitable MCP settings found for the PostgreSQL MCP server")
}

func commandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func (p *Provider) GetDefaultPolicies() map[string]any {
	var policies = PostgresqlPolicies{}
	jsonBody, err := json.Marshal(policies)
	if err != nil {
		return nil
	}
	var policiesMap map[string]any
	err = json.Unmarshal(jsonBody, &policiesMap)
	if err != nil {
		return nil
	}
	return policiesMap
}

func (p *Provider) getEnvVarValueOrDefault(envVar string, defaultValue string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return defaultValue
}

var instance = &Provider{}

func init() {
	tools.Register(instance)
}
