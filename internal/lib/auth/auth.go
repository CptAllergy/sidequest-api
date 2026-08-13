package auth

import (
	"context"
	"net/http"

	"github.com/cptallergy/sidequest-api/internal/lib/config"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/http/middleware"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
)

type Identity struct {
	Id    string
	OrgID string
}

type ctxKey string

const identityKey ctxKey = "auth_identity_ctx"

func NewZitadelMiddleware(ctx context.Context, auth config.Auth) (func(http.Handler) http.Handler, error) {
	var opts []zitadel.Option

	if auth.ZitadelInsecure {
		opts = append(opts, zitadel.WithInsecure(auth.ZitadelInsecurePort))
	}

	// Initiate the authorization with zitadel JWT verifier
	authZ, err := authorization.New(ctx, zitadel.New(auth.ZitadelUrl, opts...), oauth.DefaultJWTAuthorization(auth.ZitadelClientId))
	if err != nil {
		return nil, err
	}

	mw := middleware.New(authZ)
	requireAuth := mw.RequireAuthorization()

	return extractIdentity(requireAuth, mw), nil
}

// Extract JWT claims and put them on standard context key for use in handlers
func extractIdentity(requireAuth func(next http.Handler) http.Handler, mw *middleware.Interceptor[*oauth.IntrospectionContext]) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx := mw.Context(r.Context())
			if authCtx != nil && authCtx.UserID() != "" {
				identity := Identity{
					Id:    authCtx.UserID(),
					OrgID: authCtx.OrganizationID(),
					// Can add extra fields here
				}
				ctx := context.WithValue(r.Context(), identityKey, identity)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		}))
	}
}

func GetIdentityFromContext(ctx context.Context) (Identity, bool) {
	u, ok := ctx.Value(identityKey).(Identity)
	return u, ok
}
