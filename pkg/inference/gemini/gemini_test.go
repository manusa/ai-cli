package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/test"
	"github.com/stretchr/testify/suite"
)

type GeminiTestSuite struct {
	suite.Suite
	originalEnv []string
	ctx         context.Context
}

func (s *GeminiTestSuite) SetupTest() {
	s.originalEnv = os.Environ()
	os.Clearenv()
	s.ctx = config.WithConfig(s.T().Context(), config.New())
}
func (s *GeminiTestSuite) TearDownTest() {
	test.RestoreEnv(s.originalEnv)
}

func (s *GeminiTestSuite) TestInitializeWithNoAPIKey() {
	instance.Initialize(s.ctx)
	s.Run("when GEMINI_API_KEY is not set, is not available", func() {
		s.False(instance.IsAvailable())
	})
	s.Run("when GEMINI_API_KEY is not set, shows reason", func() {
		s.Equal("GEMINI_API_KEY is not set", instance.Reason())
	})
	s.Run("when GEMINI_API_KEY is not set, has models", func() {
		s.Len(instance.Models(), 1)
		s.Contains(instance.Models(), "gemini-2.0-flash")
	})
	s.Run("when GEMINI_API_KEY is set, marshaled JSON shows availability fields", func() {
		data, err := json.Marshal(instance)
		s.Run("does not return an error", func() {
			s.Nil(err)
		})
		s.Run("returns expected JSON", func() {
			s.JSONEq(`{`+
				`"description":"Google Gemini inference provider",`+
				`"local":false,`+
				`"models":["gemini-2.0-flash"],`+
				`"name":"gemini",`+
				`"public":true,`+
				`"reason":"GEMINI_API_KEY is not set"`+
				`}`, string(data))
		})
	})
}

func (s *GeminiTestSuite) TestInitializeWithAPIKey() {
	_ = os.Setenv("GEMINI_API_KEY", "A_VALID_KEY")
	ctxWithApiKey := config.WithConfig(s.ctx, config.New())
	instance.Initialize(ctxWithApiKey)
	s.Run("when GEMINI_API_KEY is set, is available", func() {
		s.True(instance.IsAvailable())
	})
	s.Run("when GEMINI_API_KEY is set, shows reason", func() {
		s.Equal("GEMINI_API_KEY is set", instance.Reason())
	})
	s.Run("when GEMINI_API_KEY is set, has models", func() {
		s.Len(instance.Models(), 1)
		s.Contains(instance.Models(), "gemini-2.0-flash")
	})
	s.Run("when GEMINI_API_KEY is set, marshaled JSON shows availability fields", func() {
		data, err := json.Marshal(instance)
		s.Run("does not return an error", func() {
			s.Nil(err)
		})
		s.Run("returns expected JSON", func() {
			s.JSONEq(`{`+
				`"description":"Google Gemini inference provider",`+
				`"local":false,`+
				`"models":["gemini-2.0-flash"],`+
				`"name":"gemini",`+
				`"public":true,`+
				`"reason":"GEMINI_API_KEY is set"`+
				`}`, string(data))
		})
	})
}

func (s *GeminiTestSuite) TestHasCustomSystemPrompt() {
	s.Run("Is not empty", func() {
		s.NotEmpty(instance.SystemPrompt())
	})
	s.Run("Contains today's date", func() {
		s.Contains(instance.SystemPrompt(), fmt.Sprintf("Today is %s.", time.Now().Format("January 2, 2006")))
	})
}

func TestGemini(t *testing.T) {
	suite.Run(t, new(GeminiTestSuite))
}
