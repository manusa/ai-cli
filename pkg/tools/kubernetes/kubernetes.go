package kubernetes

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/manusa/ai-cli/pkg/tools/utils/eino"
)

type Provider struct {
	api.BasicToolsProvider
	ReadOnly           bool `json:"-"`
	DisableDestructive bool `json:"-"`
}

var _ api.ToolsProvider = &Provider{}

const (
	RecommendedConfigPathEnvVar = "KUBECONFIG"
	RecommendedHomeDir          = ".kube"
	RecommendedFileName         = "config"
)

var (
	RecommendedConfigDir = filepath.Join(homedir(), RecommendedHomeDir)
	RecommendedHomeFile  = filepath.Join(RecommendedConfigDir, RecommendedFileName)
	supportedMcpSettings = map[string]api.McpSettings{
		"uvx": {
			Type:    api.McpTypeStdio,
			Command: "uvx",
			Args:    []string{"kubernetes-mcp-server@latest"},
		},
		"npx": {
			Type:    api.McpTypeStdio,
			Command: "npx",
			Args:    []string{"-y", "kubernetes-mcp-server@latest"},
		},
	}
)

// copied from https://github.com/kubernetes/client-go/blob/d99dd130a2fc7519c0bc2bd7271447b2a16c04a2/util/homedir/homedir.go#L31
func homedir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOME")
		homeDriveHomePath := ""
		if homeDrive, homePath := os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"); len(homeDrive) > 0 && len(homePath) > 0 {
			homeDriveHomePath = homeDrive + homePath
		}
		userProfile := os.Getenv("USERPROFILE")

		// Return first of %HOME%, %HOMEDRIVE%/%HOMEPATH%, %USERPROFILE% that contains a `.kube\config` file.
		// %HOMEDRIVE%/%HOMEPATH% is preferred over %USERPROFILE% for backwards-compatibility.
		for _, p := range []string{home, homeDriveHomePath, userProfile} {
			if len(p) == 0 {
				continue
			}
			if _, err := config.FileSystem.Stat(filepath.Join(p, ".kube", "config")); err != nil {
				continue
			}
			return p
		}

		firstSetPath := ""
		firstExistingPath := ""

		// Prefer %USERPROFILE% over %HOMEDRIVE%/%HOMEPATH% for compatibility with other auth-writing tools
		for _, p := range []string{home, userProfile, homeDriveHomePath} {
			if len(p) == 0 {
				continue
			}
			if len(firstSetPath) == 0 {
				// remember the first path that is set
				firstSetPath = p
			}
			info, err := config.FileSystem.Stat(p)
			if err != nil {
				continue
			}
			if len(firstExistingPath) == 0 {
				// remember the first path that exists
				firstExistingPath = p
			}
			if info.IsDir() && info.Mode().Perm()&(1<<(uint(7))) != 0 {
				// return first path that is writeable
				return p
			}
		}

		// If none are writeable, return first location that exists
		if len(firstExistingPath) > 0 {
			return firstExistingPath
		}

		// If none exist, return first location that is set
		if len(firstSetPath) > 0 {
			return firstSetPath
		}

		// We've got nothing
		return ""
	}
	return os.Getenv("HOME")
}

func (p *Provider) Initialize(_ context.Context) {
	var err error
	p.McpSettings, err = findBestMcpServerSettings(p.ReadOnly, p.DisableDestructive)
	if err != nil {
		p.IsAvailableReason = err.Error()
		return
	}

	// using the same logic as kubectl to find the config files
	// https://github.com/kubernetes/client-go/blob/d99dd130a2fc7519c0bc2bd7271447b2a16c04a2/tools/clientcmd/loader.go#L159
	var allFiles []string
	envVarFiles := os.Getenv(RecommendedConfigPathEnvVar)
	if len(envVarFiles) != 0 {
		allFiles = filepath.SplitList(envVarFiles)
	} else {
		allFiles = []string{RecommendedHomeFile}
	}

	// return true if any of the files exist
	for _, file := range allFiles {
		if _, err := config.FileSystem.Stat(file); err == nil {
			if len(envVarFiles) == 0 {
				p.IsAvailableReason = "default kubeconfig file found"
			} else {
				p.IsAvailableReason = "kubeconfig file found in the locations specified by the KUBECONFIG environment variable"
			}
			p.Available = true
			return
		}
	}
	if len(envVarFiles) == 0 {
		p.IsAvailableReason = "no kubeconfig file found in the default location"
	} else {
		p.IsAvailableReason = "no kubeconfig file found in the locations specified by the KUBECONFIG environment variable"
	}
}

func (p *Provider) GetTools(ctx context.Context) ([]*api.Tool, error) {
	mcpSettings, err := findBestMcpServerSettings(p.ReadOnly, p.DisableDestructive)
	if err != nil || mcpSettings.Type != api.McpTypeStdio {
		return nil, err
	}

	cli, err := eino.StartMcp(ctx, mcpSettings.Env, slices.Concat([]string{mcpSettings.Command}, mcpSettings.Args))
	if err != nil {
		return nil, err
	}
	return eino.GetTools(ctx, cli)
}

func findBestMcpServerSettings(readOnly bool, disableDestructive bool) (*api.McpSettings, error) {
	for command, settings := range supportedMcpSettings {
		if config.CommandExists(command) {
			if readOnly {
				settings.Args = append(settings.Args, "--read-only")
			}
			if disableDestructive && !readOnly {
				settings.Args = append(settings.Args, "--disable-destructive")
			}
			return &settings, nil
		}
	}
	// TODO support manual download and installation of kubernetes-mcp-server as a last resort
	return nil, errors.New("no suitable MCP settings found for the Kubernetes MCP server")
}

var instance = &Provider{
	BasicToolsProvider: api.BasicToolsProvider{
		BasicToolsAttributes: api.BasicToolsAttributes{
			BasicFeatureAttributes: api.BasicFeatureAttributes{
				FeatureName:        "kubernetes",
				FeatureDescription: "Provides access to Kubernetes clusters, allowing management and interaction with cluster resources.",
			},
		},
	},
}

func init() {
	tools.Register(instance)
}
