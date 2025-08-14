package api

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type McpTypeTestSuite struct {
	suite.Suite
}

func (s *McpTypeTestSuite) TestString() {
	cases := []struct {
		Type     McpType
		Expected string
	}{
		{McpTypeStdio, "stdio"},
		{McpTypeSse, "sse"},
		{McpTypeStreamableHttp, "http"},
	}
	for _, c := range cases {
		s.Run(fmt.Sprintf("%v is converted to %s", c.Type, c.Expected), func() {
			s.Equal(c.Expected, c.Type.String())
		})
	}
}

func (s *McpTypeTestSuite) TestMarshalJSON() {
	cases := []struct {
		Type     McpType
		Expected string
	}{
		{McpTypeStdio, `["stdio"]`},
		{McpTypeSse, `["sse"]`},
		{McpTypeStreamableHttp, `["http"]`},
	}
	for _, c := range cases {
		data, err := json.Marshal([]McpType{c.Type})
		s.Run(fmt.Sprintf("Marshaling %v returns no error", c.Type), func() {
			s.NoError(err, "Expected no error when marshaling McpType")
		})
		s.Run(fmt.Sprintf("%v is marshaled to %s", c.Type, c.Expected), func() {
			s.Equal(c.Expected, string(data))
		})
	}
}

func (s *McpTypeTestSuite) TestUnmarshalJSON() {
	cases := []struct {
		Input    string
		Expected McpType
	}{
		{`"stdio"`, McpTypeStdio},
		{`"sse"`, McpTypeSse},
		{`"http"`, McpTypeStreamableHttp},
		{`"stdIO"`, McpTypeStdio},
		{`"sSe"`, McpTypeSse},
		{`"hTTp"`, McpTypeStreamableHttp},
	}
	for _, c := range cases {
		var t McpType
		err := json.Unmarshal([]byte(c.Input), &t)
		s.Run(fmt.Sprintf("Unmarshaling %s returns no error", c.Input), func() {
			s.NoError(err, "Expected no error when unmarshalling McpType")
		})
		s.Run(fmt.Sprintf("%s is unmarshalled to %v", c.Input, c.Expected), func() {
			s.Equal(c.Expected, t)
		})
	}
	s.Run("Unmarshalling invalid McpType returns error", func() {
		var t []McpType
		err := json.Unmarshal([]byte(`["invalid"]`), &t)
		s.Error(err, "Expected error when unmarshalling invalid McpType")
		s.Equal([]McpType{McpTypeStdio}, t, "Expected slice to contain default McpTypeStdio when unmarshalling invalid type")
	})
}

func TestMcpType(t *testing.T) {
	suite.Run(t, new(McpTypeTestSuite))
}
