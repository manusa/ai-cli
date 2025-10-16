package containers

import (
	"reflect"
	"testing"
)

type MockCommand struct {
	OutputBytes []byte
}

func (m *MockCommand) Output() ([]byte, error) {
	return m.OutputBytes, nil
}

func TestListContainers(t *testing.T) {
	tests := []struct {
		name           string
		filters        ListContainersFilters
		commandResult  string
		expectedArgs   []string
		expectedOutput []Container
	}{
		{
			name:           "no filters",
			filters:        ListContainersFilters{},
			commandResult:  "1234567890 container-1234567890\n1234567891 container-1234567891\n",
			expectedArgs:   []string{"container", "list", "--format", "{{.ID}} {{.Names}}"},
			expectedOutput: []Container{{ID: "1234567890", Name: "container-1234567890"}, {ID: "1234567891", Name: "container-1234567891"}},
		},
		{
			name: "single image filter",
			filters: ListContainersFilters{
				Images: []string{"postgres"},
			},
			commandResult:  "\n",
			expectedArgs:   []string{"container", "list", "--format", "{{.ID}} {{.Names}}", "--filter", "ancestor=postgres"},
			expectedOutput: []Container{},
		},
		{
			name: "multiple image filters",
			filters: ListContainersFilters{
				Images: []string{"postgres", "redis"},
			},
			commandResult:  "1234567891 container-1234567891\n",
			expectedArgs:   []string{"container", "list", "--format", "{{.ID}} {{.Names}}", "--filter", "ancestor=postgres", "--filter", "ancestor=redis"},
			expectedOutput: []Container{{ID: "1234567891", Name: "container-1234567891"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			origShellCommandFunc := shellCommandFunc
			defer func() { shellCommandFunc = origShellCommandFunc }()

			var capturedCommand string
			var capturedArgs []string
			shellCommandFunc = func(name string, args ...string) commandExecutor {
				capturedCommand = name
				capturedArgs = args
				return &MockCommand{OutputBytes: []byte(tt.commandResult)}
			}

			// Call the function under test
			output, err := ListContainers(tt.filters)
			if err != nil {
				t.Fatalf("ListContainers() error = %v", err)
			}

			// Verify the command name
			if capturedCommand != command {
				t.Errorf("Expected command %q, got %q", command, capturedCommand)
			}

			// Verify the arguments
			if !reflect.DeepEqual(capturedArgs, tt.expectedArgs) {
				t.Errorf("Expected args %v, got %v", tt.expectedArgs, capturedArgs)
			}

			// Verify the output
			if len(output) != len(tt.expectedOutput) {
				t.Errorf("Expected %d containers, got %d", len(tt.expectedOutput), len(output))
			}
			for i, container := range output {
				if container.ID != tt.expectedOutput[i].ID || container.Name != tt.expectedOutput[i].Name {
					t.Errorf("Expected container %q, got %q", tt.expectedOutput[i], container)
				}
			}
		})
	}
}

func TestGetContainerEnvironmentVariables(t *testing.T) {
	prefix := "POSTGRES_"
	tests := []struct {
		name           string
		id             string
		prefix         *string
		commandResult  string
		expectedArgs   []string
		expectedOutput map[string]string
	}{
		{
			name:           "no prefix",
			id:             "1234567890",
			prefix:         nil,
			commandResult:  "POSTGRES_USER=postgres\nPOSTGRES_PASSWORD=password\nPOSTGRES_DB=postgres\nSHELL=/bin/bash\n",
			expectedArgs:   []string{"container", "inspect", "1234567890", "--format", "{{range .Config.Env}}{{.}}\n{{end}}"},
			expectedOutput: map[string]string{"POSTGRES_USER": "postgres", "POSTGRES_PASSWORD": "password", "POSTGRES_DB": "postgres", "SHELL": "/bin/bash"},
		},
		{
			name:           "POSTGRES_ prefix",
			id:             "1234567890",
			prefix:         &prefix,
			commandResult:  "POSTGRES_USER=postgres\nPOSTGRES_PASSWORD=password\nPOSTGRES_DB=postgres\nSHELL=/bin/bash\n",
			expectedArgs:   []string{"container", "inspect", "1234567890", "--format", "{{range .Config.Env}}{{.}}\n{{end}}"},
			expectedOutput: map[string]string{"POSTGRES_USER": "postgres", "POSTGRES_PASSWORD": "password", "POSTGRES_DB": "postgres"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origShellCommandFunc := shellCommandFunc
			defer func() { shellCommandFunc = origShellCommandFunc }()

			var capturedCommand string
			var capturedArgs []string
			shellCommandFunc = func(name string, args ...string) commandExecutor {
				capturedCommand = name
				capturedArgs = args
				return &MockCommand{OutputBytes: []byte(tt.commandResult)}
			}

			// Call the function under test
			output, err := GetContainerEnvironmentVariables(tt.id, tt.prefix)
			if err != nil {
				t.Fatalf("GetContainerEnvironmentVariables() error = %v", err)
			}
			if capturedCommand != command {
				t.Errorf("Expected command %q, got %q", command, capturedCommand)
			}
			if !reflect.DeepEqual(capturedArgs, tt.expectedArgs) {
				t.Errorf("Expected args %v, got %v", tt.expectedArgs, capturedArgs)
			}
			if !reflect.DeepEqual(output, tt.expectedOutput) {
				t.Errorf("Expected output %v, got %v", tt.expectedOutput, output)
			}
		})
	}
}

func TestGetContainerPortMapping(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		port           string
		protocol       string
		commandResult  string
		expectedArgs   []string
		expectedOutput string
		expectedError  string
	}{
		{
			name:           "no mapping",
			id:             "1234567890",
			port:           "5432",
			protocol:       "tcp",
			commandResult:  "{}",
			expectedArgs:   []string{"container", "inspect", "1234567890", "--format", "{{json .NetworkSettings.Ports}}"},
			expectedOutput: "",
			expectedError:  "port mapping not found for 5432/tcp",
		},
		{
			name:           "mapping",
			id:             "1234567890",
			port:           "5432",
			protocol:       "tcp",
			commandResult:  "{\"5432/tcp\":[{\"HostPort\":\"5433\",\"HostIp\":\"0.0.0.0\"}]}",
			expectedArgs:   []string{"container", "inspect", "1234567890", "--format", "{{json .NetworkSettings.Ports}}"},
			expectedOutput: "5433",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origShellCommandFunc := shellCommandFunc
			defer func() { shellCommandFunc = origShellCommandFunc }()

			var capturedCommand string
			var capturedArgs []string
			shellCommandFunc = func(name string, args ...string) commandExecutor {
				capturedCommand = name
				capturedArgs = args
				return &MockCommand{OutputBytes: []byte(tt.commandResult)}
			}

			output, err := GetContainerPortMapping(tt.id, tt.port, tt.protocol)
			if capturedCommand != command {
				t.Errorf("Expected command %q, got %q", command, capturedCommand)
			}
			if !reflect.DeepEqual(capturedArgs, tt.expectedArgs) {
				t.Errorf("Expected args %v, got %v", tt.expectedArgs, capturedArgs)
			}
			if err != nil {
				if err.Error() != tt.expectedError {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err.Error())
				}
			} else {
				if !reflect.DeepEqual(output, tt.expectedOutput) {
					t.Errorf("Expected output %v, got %v", tt.expectedOutput, output)
				}
			}
		})
	}
}
