package ollama

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalJSON(t *testing.T) {
	provider := &Provider{}
	data, err := provider.MarshalJSON()
	t.Run("MarshalJSON does not return an error", func(t *testing.T) {
		assert.Nil(t, err)
	})
	t.Run("MarshalJSON returns expected JSON", func(t *testing.T) {
		assert.JSONEq(t, `{"local":true,"models":null,"name":"ollama","public":false,"reason":""}`, string(data))
	})
}
