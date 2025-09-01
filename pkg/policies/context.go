package policies

import "context"

type policiesCtxKeyType struct{}

var policiesCtxKey = policiesCtxKeyType{}

func WithPolicies(ctx context.Context, policies *Policies) context.Context {
	return context.WithValue(ctx, policiesCtxKey, policies)
}

func GetPolicies(ctx context.Context) *Policies {
	policies, ok := ctx.Value(policiesCtxKey).(*Policies)
	if !ok {
		return nil
	}
	return policies
}
