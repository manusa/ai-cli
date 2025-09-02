package policies

import (
	"context"

	"github.com/manusa/ai-cli/pkg/api"
)

type policiesCtxKeyType struct{}

var policiesCtxKey = policiesCtxKeyType{}

func WithPolicies(ctx context.Context, policies *api.Policies) context.Context {
	return context.WithValue(ctx, policiesCtxKey, policies)
}

func GetPolicies(ctx context.Context) *api.Policies {
	policies, ok := ctx.Value(policiesCtxKey).(*api.Policies)
	if !ok {
		return nil
	}
	return policies
}
