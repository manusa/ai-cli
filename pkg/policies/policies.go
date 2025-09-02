package policies

import "github.com/manusa/ai-cli/pkg/api"

type Provider struct{}

var PoliciesProvider api.PoliciesProvider = &Provider{}
