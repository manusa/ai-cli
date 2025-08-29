package mcpconfig

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

type McpConfigTestSuite struct {
	suite.Suite
}

func (s *McpConfigTestSuite) SetupTest() {
	config.FileSystem = afero.NewMemMapFs()
}

func (s *McpConfigTestSuite) TearDownTest() {}

type TestProvider struct{}

func (p *TestProvider) GetFile() string {
	return "/path/to/config"
}

func (p *TestProvider) GetConfig(tools []api.ToolsProvider) ([]byte, error) {
	return []byte("a config"), nil
}

func (s *McpConfigTestSuite) TestSaveNewFile() {
	s.Run("Save creates a new file if it does not exist", func() {
		out, errOut, err := captureStdouts(func() error { return Save(&TestProvider{}, nil) })
		s.NoError(err)

		if _, err := config.FileSystem.Stat("/path/to/config"); err != nil && !os.IsNotExist(err) {
			s.Fail("file does not exist")
		}
		content, err := readFile(config.FileSystem, "/path/to/config")
		s.NoError(err)
		s.Equal("a config", string(content))
		s.Equal("MCP config file /path/to/config has been created\n", out)
		s.Equal("", errOut)
	})
}

func (s *McpConfigTestSuite) TestSaveStdout() {
	s.Run("Save outputs to stdoutif config file exists", func() {
		err := createFile(config.FileSystem, "/path/to/config", "content before")
		s.NoError(err)
		out, errOut, err := captureStdouts(func() error { return Save(&TestProvider{}, nil) })
		s.NoError(err)
		if _, err := config.FileSystem.Stat("/path/to/config"); err != nil && !os.IsNotExist(err) {
			s.Fail("file does not exist")
		}
		content, err := readFile(config.FileSystem, "/path/to/config")
		s.NoError(err)
		// content should not have changed
		s.Equal("content before", string(content))
		s.Equal("MCP config file /path/to/config already exists, outputting config to stdout\n", errOut)
		s.Equal("a config\n", out)
	})
}
func TestMcpConfig(t *testing.T) {
	suite.Run(t, new(McpConfigTestSuite))
}

func createFile(fs afero.Fs, path string, content string) error {
	err := fs.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}
	file, err := fs.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	_, err = file.Write([]byte(content))
	return err
}

func readFile(fs afero.Fs, path string) ([]byte, error) {
	file, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()
	return io.ReadAll(file)
}

func captureStdouts(f func() error) (string, string, error) {
	originalOut := os.Stdout
	originalErr := os.Stderr
	defer func() {
		os.Stdout = originalOut
		os.Stderr = originalErr
	}()
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr
	err := f()
	_ = wOut.Close()
	_ = wErr.Close()
	outBytes, _ := io.ReadAll(rOut)
	errBytes, _ := io.ReadAll(rErr)
	return string(outBytes), string(errBytes), err
}
