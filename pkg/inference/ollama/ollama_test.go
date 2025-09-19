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
	s.MockServer = test.NewMockServer()
	s.originalEnv = os.Environ()
	os.Clearenv()
	s.originalInstance = test.Clone(instance)
}
func (s *OllamaTestSuite) TearDownTest() {
	s.MockServer.Close()
	test.RestoreEnv(s.originalEnv)
	instance = s.originalInstance
}

func (s *OllamaTestSuite) TestInitializeWithNoServer() {
	s.Run("no server running", func() {
		instance.Initialize(s.T().Context())
		s.Run("is not available", func() {
			s.False(instance.IsAvailable())
		})
		s.Run("shows reason", func() {
			s.Contains(instance.Reason(), "ollama is not accessible at http://localhost:11434")
		})
		s.Run("has no models", func() {
			s.Nil(instance.Models())
		})
		s.Run("marshaled JSON shows availability fields", func() {
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
	})
}

func (s *OllamaTestSuite) TestInitializeWithNoCompatibleServer() {
	s.Run("incompatible server running", func() {
		_ = os.Setenv("OLLAMA_HOST", s.MockServer.URL())
		instance.Initialize(s.T().Context())
		s.Run("is not available", func() {
			s.False(instance.IsAvailable())
		})
		s.Run("shows reason", func() {
			s.Regexp("The server at http:\\/\\/.+ defined by the OLLAMA_HOST environment variable is accessible but is not Ollama", instance.Reason())
		})
		s.Run("has no models", func() {
			s.Nil(instance.Models())
		})
		s.Run("marshaled JSON shows availability fields", func() {
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
	})
}

func (s *OllamaTestSuite) TestInitializeWithCompatibleServerMissingModel() {
	s.Run("compatible server running with unmatched models", func() {
		s.MockServer.Handle(func(w http.ResponseWriter, req *http.Request) (handled bool) {
			if req.Method == http.MethodGet && req.URL.Path == "/v1/models" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":[{"id":"model-1"},{"id":"model-2"}]}`))
				handled = true
			}
			return
		})
		_ = os.Setenv("OLLAMA_HOST", s.MockServer.URL())
		instance.Initialize(s.T().Context())
		s.Run("is not available", func() {
			s.False(instance.IsAvailable())
		})
		s.Run("shows reason", func() {
			s.Equal(fmt.Sprintf("ollama is accessible at %s defined by the OLLAMA_HOST environment variable but no preferred models (llama3.2:3b, granite3.3:latest, mistral:7b) are served", s.MockServer.URL()), instance.Reason())
		})
		s.Run("has models", func() {
			s.Len(instance.Models(), 2)
			s.Contains(instance.Models(), "model-1")
			s.Contains(instance.Models(), "model-2")
		})
		s.Run("marshaled JSON shows availability fields", func() {
			data, err := json.Marshal(instance)
			s.Run("does not return an error", func() {
				s.Nil(err)
			})
			s.Run("returns expected JSON", func() {
				s.JSONEq(`{`+
					`"description":"Ollama local inference provider",`+
					`"local":true,`+
					`"models":["model-1", "model-2"],`+
					`"name":"ollama",`+
					`"public":false,`+
					fmt.Sprintf(`"reason":"ollama is accessible at %s defined by the OLLAMA_HOST environment variable but no preferred models (llama3.2:3b, granite3.3:latest, mistral:7b) are served"`, s.MockServer.URL())+
					`}`, string(data))
			})
		})
	})
}

func (s *OllamaTestSuite) TestInitializeWithCompatibleServer() {
	s.Run("compatible server running", func() {
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
		s.Run("is available", func() {
			s.True(instance.IsAvailable())
		})
		s.Run("shows reason", func() {
			s.Regexp(fmt.Sprintf("ollama is accessible at %s defined by the OLLAMA_HOST environment variable", s.MockServer.URL()), instance.Reason())
		})
		s.Run("has models", func() {
			s.Len(instance.Models(), 3)
			s.Contains(instance.Models(), "model-1")
			s.Contains(instance.Models(), "model-2")
			s.Contains(instance.Models(), "llama3.2:3b")
		})
		s.Run("sets first compatible model as default", func() {
			s.NotNil(instance.Model)
			s.Equal("llama3.2:3b", *instance.Model)

		})
		s.Run("marshaled JSON shows availability fields", func() {
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
