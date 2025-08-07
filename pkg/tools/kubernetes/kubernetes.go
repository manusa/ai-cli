package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/manusa/ai-cli/pkg/tools/utils/eino"
)

type Provider struct {
	tools.BasicToolsProvider
}

var _ tools.Provider = &Provider{}

const (
	RecommendedConfigPathEnvVar = "KUBECONFIG"
	RecommendedHomeDir          = ".kube"
	RecommendedFileName         = "config"
)

var (
	RecommendedConfigDir = filepath.Join(homedir(), RecommendedHomeDir)
	RecommendedHomeFile  = filepath.Join(RecommendedConfigDir, RecommendedFileName)
)

func (p *Provider) Attributes() tools.Attributes {
	return tools.Attributes{
		BasicFeatureAttributes: api.BasicFeatureAttributes{
			FeatureName: "kubernetes",
		},
	}
}

func (p *Provider) Data() tools.Data {
	return tools.Data{
		BasicFeatureData: api.BasicFeatureData{
			Reason: p.Reason,
		},
	}
}

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
			if _, err := os.Stat(filepath.Join(p, ".kube", "config")); err != nil {
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
			info, err := os.Stat(p)
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

func (p *Provider) IsAvailable(_ *config.Config) bool {
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
		if _, err := os.Stat(file); err == nil {
			if len(envVarFiles) == 0 {
				p.Reason = "default kubeconfig file found"
			} else {
				p.Reason = "kubeconfig file found in the locations specified by the KUBECONFIG environment variable"
			}
			return true
		}
	}
	if len(envVarFiles) == 0 {
		p.Reason = "no kubeconfig file found in the default location"
	} else {
		p.Reason = "no kubeconfig file found in the locations specified by the KUBECONFIG environment variable"
	}
	return false
}

func (p *Provider) GetTools(ctx context.Context, _ *config.Config) ([]*api.Tool, error) {
	commandAndArgs, err := getBestMcpServerCommand()
	if err != nil {
		return nil, err
	}
	cli, err := eino.StartMcp(ctx, commandAndArgs)
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

func getBestMcpServerCommand() ([]string, error) {
	if commandExists("npx") {
		return []string{"npx", "-y", "kubernetes-mcp-server@latest"}, nil
	} else if commandExists("uvx") {
		return []string{"uvx", "kubernetes-mcp-server@latest"}, nil
	}
	// TODO support manual download and installation of kubernetes-mcp-server as a last resort
	return nil, errors.New("no command found to start the Kubernetes MCP server")
}

func commandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

var instance = &Provider{}

func init() {
	tools.Register(instance)
}
