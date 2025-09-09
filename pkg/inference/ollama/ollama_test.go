package ollama

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/manusa/ai-cli/internal/test"
	"github.com/stretchr/testify/suite"
)

type OllamaTestSuite struct {
	suite.Suite
	originalEnv      []string
	originalInstance *Provider
	MockServer       *test.MockServer
}

func (s *OllamaTestSuite) SetupTest() {
	s.originalEnv = os.Environ()
	os.Clearenv()
	s.originalInstance = test.Clone(instance)
	s.MockServer = test.NewMockServer()
}
func (s *OllamaTestSuite) TearDownTest() {
	test.RestoreEnv(s.originalEnv)
	instance = s.originalInstance
	s.MockServer.Close()
}

func (s *OllamaTestSuite) TestInitializeWithNoServer() {
	instance.Initialize(s.T().Context())
	s.Run("when no server is running, is not available", func() {
		s.False(instance.IsAvailable())
	})
	s.Run("when no server is running, shows reason", func() {
		s.Contains(instance.Reason(), "ollama is not accessible at http://localhost:11434")
	})
	s.Run("when no server is running, has no models", func() {
		s.Nil(instance.Models())
	})
	s.Run("when no server is running, marshaled JSON shows availability fields", func() {
		data, err := json.Marshal(instance)
		s.Run("does not return an error", func() {
			s.Nil(err)
		})
		s.Run("returns expected JSON", func() {
			s.JSONEq(`{`+
				`"description":"Ollama local inference provider",`+
				`"local":true,`+
				`"models":null,`+
				`"name":"ollama",`+
				`"public":false,`+
				`"reason":"ollama is not accessible at http://localhost:11434"`+
				`}`, string(data))
		})
	})
}

func (s *OllamaTestSuite) TestInitializeWithNoCompatibleServer() {
	_ = os.Setenv("OLLAMA_HOST", s.MockServer.URL())
	instance.Initialize(s.T().Context())
	s.Run("when no server is running, is not available", func() {
		s.False(instance.IsAvailable())
	})
	s.Run("when no server is running, shows reason", func() {
		s.Regexp("The server at http:\\/\\/.+ defined by the OLLAMA_HOST environment variable is accessible but is not Ollama", instance.Reason())
	})
	s.Run("when no server is running, has no models", func() {
		s.Nil(instance.Models())
	})
	s.Run("when no server is running, marshaled JSON shows availability fields", func() {
		data, err := json.Marshal(instance)
		s.Run("does not return an error", func() {
			s.Nil(err)
		})
		s.Run("returns expected JSON", func() {
			s.JSONEq(`{`+
				`"description":"Ollama local inference provider",`+
				`"local":true,`+
				`"models":null,`+
				`"name":"ollama",`+
				`"public":false,`+
				fmt.Sprintf(`"reason":"The server at %s defined by the OLLAMA_HOST environment variable is accessible but is not Ollama"`, s.MockServer.URL())+
				`}`, string(data))
		})
	})
}

func (s *OllamaTestSuite) TestInitializeWithCompatibleServer() {
	s.MockServer.Handle(func(w http.ResponseWriter, req *http.Request) (handled bool) {
		if req.Method == http.MethodGet && req.URL.Path == "/v1/models" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"id":"model-1"},{"id":"model-2"},{"id":"llama3.2:3b"}]}`))
			handled = true
		}
		return
	})
	_ = os.Setenv("OLLAMA_HOST", s.MockServer.URL())
	instance.Initialize(s.T().Context())
	s.Run("when a compatible server is running, is available", func() {
		s.True(instance.IsAvailable())
	})
	s.Run("when a compatible server is running, shows reason", func() {
		s.Regexp(fmt.Sprintf("ollama is accessible at %s defined by the OLLAMA_HOST environment variable", s.MockServer.URL()), instance.Reason())
	})
	s.Run("when a compatible server is running, has models", func() {
		s.Len(instance.Models(), 3)
		s.Contains(instance.Models(), "model-1")
		s.Contains(instance.Models(), "model-2")
		s.Contains(instance.Models(), "llama3.2:3b")
	})
	s.Run("when a compatible server is running, sets first compatible model as default", func() {
		s.NotNil(instance.Model)
		s.Equal("llama3.2:3b", *instance.Model)

	})
	s.Run("when a compatible server is running, marshaled JSON shows availability fields", func() {
		data, err := json.Marshal(instance)
		s.Run("does not return an error", func() {
			s.Nil(err)
		})
		s.Run("returns expected JSON", func() {
			s.JSONEq(`{`+
				`"description":"Ollama local inference provider",`+
				`"local":true,`+
				`"models":["model-1", "model-2", "llama3.2:3b"],`+
				`"name":"ollama",`+
				`"public":false,`+
				fmt.Sprintf(`"reason":"ollama is accessible at %s defined by the OLLAMA_HOST environment variable"`, s.MockServer.URL())+
				`}`, string(data))
		})
	})
}

func (s *OllamaTestSuite) TestInheritsSystemPrompt() {
	s.Run("Is empty", func() {
		s.Empty(instance.SystemPrompt())
	})
}

func TestOllama(t *testing.T) {
	suite.Run(t, new(OllamaTestSuite))
}
