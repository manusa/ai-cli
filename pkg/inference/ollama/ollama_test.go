package ollama

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarshalJSON(t *testing.T) {
	provider := &Provider{}
	data, err := provider.MarshalJSON()
	t.Run("MarshalJSON does not return an error", func(t *testing.T) {
		assert.Nil(t, err)
	})
	t.Run("MarshalJSON returns expected JSON", func(t *testing.T) {
		assert.JSONEq(t, `{"name":"ollama","Distant":false}`, string(data))
	})
}
