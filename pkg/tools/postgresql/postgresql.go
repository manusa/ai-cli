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
	"github.com/manusa/ai-cli/pkg/containers"
	"github.com/manusa/ai-cli/pkg/keyring"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/manusa/ai-cli/pkg/ui/components/password_input"
	"github.com/manusa/ai-cli/pkg/ui/components/selector"
	"github.com/manusa/ai-cli/pkg/utils"
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

	if p.getDatabaseURI() == "" {
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
			if databaseUri := p.getDatabaseURI(); !strings.HasPrefix(databaseUri, "postgresql://") && os.Getenv(pgPasswordEnvVar) != "" {
				uri := fmt.Sprintf(
					"postgresql://%s:%s@%s:%s/%s",
					p.getEnvVarValueOrDefault(pgUserEnvVar, defaultPgUser),
					os.Getenv(pgPasswordEnvVar),
					p.getEnvVarValueOrDefault(pgHostEnvVar, defaultPgHost),
					p.getEnvVarValueOrDefault(pgPortEnvVar, defaultPgPort),
					p.getEnvVarValueOrDefault(pgDatabaseEnvVar, defaultPgDatabase),
				)
				settings.Env = append(settings.Env, fmt.Sprintf("%s=%s", databaseUriEnvVar, uri))
			} else if databaseUri := p.getDatabaseURI(); strings.HasPrefix(databaseUri, "postgresql://") {
				settings.Env = append(settings.Env, fmt.Sprintf("%s=%s", databaseUriEnvVar, databaseUri))
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
	detectExistingInstances := "Detect existing containerized PostgreSQL instances"
	createNewInstance := "Create a new containerized PostgreSQL instance"
	quit := "Terminate PostgreSQL setup"
	choices := []list.Item{
		selector.Item(registerExistingInstance),
		selector.Item(detectExistingInstances),
		selector.Item(createNewInstance),
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
			uri, err := password_input.Prompt()
			if err != nil {
				return err
			}
			err = keyring.SetKey(databaseUriEnvVar, uri)
			if err != nil {
				return err
			}
		case detectExistingInstances:
			instances, err := p.detectExistingInstances()
			if err != nil {
				return err
			}
			instance, err := p.selectInstance(instances)
			if err != nil {
				return err
			}
			prefix := "POSTGRES_"
			environmentVariables, err := containers.GetContainerEnvironmentVariables(instance.ID, &prefix)
			if err != nil {
				return err
			}
			portMapped, err := containers.GetContainerPortMapping(instance.ID, p.getFirstNotEmpty(environmentVariables["POSTGRES_PORT"], defaultPgPort), "tcp")
			if err != nil {
				return err
			}
			uri := p.getContainerURI(environmentVariables, portMapped)
			err = keyring.SetKey(databaseUriEnvVar, uri)
			if err != nil {
				return err
			}
		case createNewInstance:
			envvars := map[string]string{
				"POSTGRES_USER":     "user",
				"POSTGRES_PASSWORD": utils.GenerateRandomPassword(8),
			}
			localPort, err := utils.GetFreePort()
			if err != nil {
				return err
			}
			_, err = containers.CreateContainer(containers.CreateContainerOptions{
				Image: "postgres",
				Env:   envvars,
				TcpPortBindings: map[int]int{
					5432: localPort,
				},
			})
			if err != nil {
				return err
			}
			uri := p.getContainerURI(envvars, fmt.Sprintf("%d", localPort))
			err = keyring.SetKey(databaseUriEnvVar, uri)
			fmt.Printf("✅ A new PostgreSQL instance has been created\n")
			fmt.Printf("✅ The connection string of the new instance has been saved in the keyring\n")
			if err != nil {
				return err
			}
		case quit:
			return nil
		}
	}
}

func (p *Provider) getContainerURI(environmentVariables map[string]string, portMapped string) string {
	pgUser := p.getFirstNotEmpty(environmentVariables["POSTGRES_USER"], defaultPgUser)
	pgPassword := environmentVariables["POSTGRES_PASSWORD"]
	pgDatabase := p.getFirstNotEmpty(environmentVariables["POSTGRES_DB"], environmentVariables["POSTGRES_USER"], defaultPgDatabase)
	return fmt.Sprintf(
		"postgresql://%s:%s@localhost:%s/%s",
		pgUser, pgPassword, portMapped, pgDatabase,
	)
}

func (p *Provider) getFirstNotEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (p *Provider) detectExistingInstances() ([]containers.Container, error) {
	return containers.ListContainers(containers.ListContainersFilters{
		Images: []string{"postgres"},
	})
}

func (p *Provider) selectInstance(instances []containers.Container) (containers.Container, error) {
	choices := []list.Item{}
	for _, instance := range instances {
		choices = append(choices, selector.Item(instance.Name))
	}
	choice, err := selector.Select("Please select an existing PostgreSQL instance:", choices)
	if err != nil {
		return containers.Container{}, err
	}
	for _, instance := range instances {
		if instance.Name == choice {
			return instance, nil
		}
	}
	return containers.Container{}, fmt.Errorf("instance %s not found", choice)
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
