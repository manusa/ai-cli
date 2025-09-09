package lmstudio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/manusa/ai-cli/internal/test"
	"github.com/stretchr/testify/suite"
)

type LmStudioTestSuite struct {
	suite.Suite
	originalBaseUrl  string
	originalInstance *Provider
	MockServer       *test.MockServer
}

func (s *LmStudioTestSuite) SetupTest() {
	s.originalBaseUrl = defaultBaseURL
	s.originalInstance = test.Clone(instance)
	s.MockServer = test.NewMockServer()
}
func (s *LmStudioTestSuite) TearDownTest() {
	defaultBaseURL = s.originalBaseUrl
	instance = s.originalInstance
	s.MockServer.Close()
}

func (s *LmStudioTestSuite) TestInitializeWithNoServer() {
	instance.Initialize(s.T().Context())
	s.Run("when no server is running, is not available", func() {
		s.False(instance.IsAvailable())
	})
	s.Run("when no server is running, shows reason", func() {
		s.Contains(instance.Reason(), "LM Studio is not accessible at http://localhost:1234")
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
				`"description":"LM Studio local inference provider",`+
				`"local":true,`+
				`"models":null,`+
				`"name":"lmstudio",`+
				`"public":false,`+
				`"reason":"LM Studio is not accessible at http://localhost:1234"`+
				`}`, string(data))
		})
	})
}

func (s *LmStudioTestSuite) TestInitializeWithNoCompatibleServer() {
	defaultBaseURL = s.MockServer.URL()
	instance.Initialize(s.T().Context())
	s.Run("when no server is running, is not available", func() {
		s.False(instance.IsAvailable())
	})
	s.Run("when no server is running, shows reason", func() {
		s.Regexp("The server at http:\\/\\/.+ is accessible but is not LM Studio", instance.Reason())
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
				`"description":"LM Studio local inference provider",`+
				`"local":true,`+
				`"models":null,`+
				`"name":"lmstudio",`+
				`"public":false,`+
				fmt.Sprintf(`"reason":"The server at %s is accessible but is not LM Studio"`, s.MockServer.URL())+
				`}`, string(data))
		})
	})
}

func (s *LmStudioTestSuite) TestInitializeWithCompatibleServer() {
	s.MockServer.Handle(func(w http.ResponseWriter, req *http.Request) (handled bool) {
		if req.Method == http.MethodGet && req.URL.Path == "/v1/models" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"id":"model-1"},{"id":"model-2"}]}`))
			handled = true
		}
		return
	})
	defaultBaseURL = s.MockServer.URL()
	instance.Initialize(s.T().Context())
	s.Run("when a compatible server is running, is available", func() {
		s.True(instance.IsAvailable())
	})
	s.Run("when a compatible server is running, shows reason", func() {
		s.Regexp(fmt.Sprintf("LM Studio is accessible at %s", s.MockServer.URL()), instance.Reason())
	})
	s.Run("when a compatible server is running, has models", func() {
		s.Len(instance.Models(), 2)
		s.Contains(instance.Models(), "model-1")
		s.Contains(instance.Models(), "model-2")
	})
	s.Run("when a compatible server is running, sets first model as default", func() {
		s.NotNil(instance.Model)
		s.Equal("model-1", *instance.Model)

	})
	s.Run("when a compatible server is running, marshaled JSON shows availability fields", func() {
		data, err := json.Marshal(instance)
		s.Run("does not return an error", func() {
			s.Nil(err)
		})
		s.Run("returns expected JSON", func() {
			s.JSONEq(`{`+
				`"description":"LM Studio local inference provider",`+
				`"local":true,`+
				`"models":["model-1","model-2"],`+
				`"name":"lmstudio",`+
				`"public":false,`+
				fmt.Sprintf(`"reason":"LM Studio is accessible at %s"`, s.MockServer.URL())+
				`}`, string(data))
		})
	})
}

func (s *LmStudioTestSuite) TestInheritsSystemPrompt() {
	s.Run("Is empty", func() {
		s.Empty(instance.SystemPrompt())
	})
}

func TestLmStudio(t *testing.T) {
	suite.Run(t, new(LmStudioTestSuite))
}
