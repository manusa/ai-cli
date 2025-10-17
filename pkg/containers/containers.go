package containers

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
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

type commandExecutor interface {
	Output() ([]byte, error)
	// Other methods of the exec.Cmd struct could be added
}

var shellCommandFunc = func(name string, arg ...string) commandExecutor {
	return exec.Command(name, arg...)
}

type CreateContainerOptions struct {
	Image           string
	Env             map[string]string
	TcpPortBindings map[int]int
}

func ListContainers(filters ListContainersFilters) ([]Container, error) {
	args := []string{"container", "list", "--format", "{{.ID}} {{.Names}}"}
	if len(filters.Images) > 0 {
		for _, image := range filters.Images {
			args = append(args, "--filter", fmt.Sprintf("ancestor=%s", image))
		}
	}
	cmd := shellCommandFunc(command, args...)
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
	cmd := shellCommandFunc(command, args...)
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
	cmd := shellCommandFunc(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get container port mapping: %w", err)
	}
	mapping := make(map[string][]PortMapping)
	err = json.Unmarshal(output, &mapping)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal container port mapping: %w", err)
	}
	key := fmt.Sprintf("%s/%s", port, protocol)
	if _, ok := mapping[key]; !ok {
		return "", fmt.Errorf("port mapping not found for %s", key)
	}
	if len(mapping[key]) == 0 {
		return "", fmt.Errorf("port mapping not found for %s", key)
	}
	return mapping[key][0].HostPort, nil
}

func CreateContainer(options CreateContainerOptions) (string, error) {
	args := []string{"container", "run", "-d", "--rm"}
	for key, value := range options.Env {
		args = append(args, "--env", fmt.Sprintf("%s=%s", key, value))
	}
	for port, hostPort := range options.TcpPortBindings {
		args = append(args, "-p", fmt.Sprintf("%d:%d", hostPort, port))
	}
	args = append(args, options.Image)
	cmd := shellCommandFunc(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}
	return string(output), nil
}
