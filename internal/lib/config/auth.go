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
	port := os.Getenv("ZITADEL_INSECURE_PORT")

	var opts []zitadel.Option

	if os.Getenv("ZITADEL_INSECURE") == "true" {
		opts = append(opts, zitadel.WithInsecure(port))
	}

	// Initiate the authorization with zitadel JWT verifier
	authZ, err := authorization.New(ctx, zitadel.New(url, opts...), oauth.DefaultJWTAuthorization(clientId))
	if err != nil {
		return nil, err
	}

	mw := middleware.New(authZ)
	return mw.RequireAuthorization(), nil
}
