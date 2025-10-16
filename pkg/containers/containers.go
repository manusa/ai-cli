package containers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/manusa/ai-cli/pkg/config"
)

type Container struct {
	ID   string
	Name string
}

const (
	command = "podman"
)

type PortMapping struct {
	HostPort string
	HostIP   string
}

type ListContainersFilters struct {
	// container runs one of these images (excluding tags)
	Images []string
}

func ListContainers(filters ListContainersFilters) ([]Container, error) {
	args := []string{"container", "list", "--format", "{{.ID}} {{.Names}}"}
	if len(filters.Images) > 0 {
		for _, image := range filters.Images {
			args = append(args, "--filter", fmt.Sprintf("ancestor=%s", image))
		}
	}
	cmd := config.ExecCommand(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}
	var containers []Container
	for info := range strings.SplitSeq(string(output), "\n") {
		if info == "" {
			continue
		}
		parts := strings.Split(info, " ")
		if len(parts) != 2 {
			continue
		}
		id := parts[0]
		name := parts[1]
		containers = append(containers, Container{ID: id, Name: name})
	}
	return containers, nil
}

func GetContainerEnvironmentVariables(id string, prefix *string) (map[string]string, error) {
	args := []string{"container", "inspect", id, "--format", "{{range .Config.Env}}{{.}}\n{{end}}"}
	cmd := config.ExecCommand(command, args...)
	fmt.Printf("cmd/args: %s %v\n", command, args)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get container environment variables: %w", err)
	}
	environmentVariables := make(map[string]string)
	for env := range strings.SplitSeq(string(output), "\n") {
		if env == "" {
			continue
		}
		if prefix != nil && !strings.HasPrefix(env, *prefix) {
			continue
		}
		key, value, _ := strings.Cut(env, "=")
		environmentVariables[key] = value
	}
	return environmentVariables, nil
}

func GetContainerPortMapping(id string, port string, protocol string) (string, error) {
	args := []string{"container", "inspect", id, "--format", "{{json .NetworkSettings.Ports}}"}
	cmd := config.ExecCommand(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get container port mapping: %w", err)
	}
	mapping := make(map[string][]PortMapping)
	err = json.Unmarshal(output, &mapping)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal container port mapping: %w", err)
	}
	fmt.Printf("mapping: %+v\n", mapping)
	key := fmt.Sprintf("%s/%s", port, protocol)
	if _, ok := mapping[key]; !ok {
		return "", fmt.Errorf("port mapping not found for %s", key)
	}
	if len(mapping[key]) == 0 {
		return "", fmt.Errorf("port mapping not found for %s", key)
	}
	return mapping[key][0].HostPort, nil
}
