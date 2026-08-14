package auth

import "context"

type Identity struct {
	Id                  string
	OrgID               string
	Username            string
	HasZitadelAccount   bool
	HasSidequestAccount bool
}

type ctxKey string

const identityKey ctxKey = "auth_identity_ctx"

func GetIdentityFromContext(ctx context.Context) (Identity, bool) {
	u, ok := ctx.Value(identityKey).(Identity)
	return u, ok
}

func SetIdentityToContext(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, identityKey, identity)
}
