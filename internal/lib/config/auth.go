package config

import (
	"context"
	"net/http"
	"os"

	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/http/middleware"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
)

func CreateZitadelMiddleware(ctx context.Context) (func(http.Handler) http.Handler, error) {
	clientId := os.Getenv("ZITADEL_CLIENT_ID")
	url := os.Getenv("ZITADEL_URL")

	// Initiate the authorization with zitadel JWT verifier
	authZ, err := authorization.New(ctx, zitadel.New(url), oauth.DefaultJWTAuthorization(clientId))
	if err != nil {
		return nil, err
	}

	mw := middleware.New(authZ)
	return mw.RequireAuthorization(), nil
}
