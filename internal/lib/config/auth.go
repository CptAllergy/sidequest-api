package config

import (
	"os"

	"github.com/supertokens/supertokens-golang/recipe/dashboard"
	"github.com/supertokens/supertokens-golang/recipe/emailpassword"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/recipe/userroles"
	"github.com/supertokens/supertokens-golang/supertokens"
)

func getStringPointer(s string) *string {
	return &s
}

func getApiDomain() string {
	return os.Getenv("API_URL")
}

func getWebsiteDomain() string {
	return os.Getenv("WEBSITE_URL")
}

func GetSuperTokensConfig() supertokens.TypeInput {
	return supertokens.TypeInput{
		Supertokens: &supertokens.ConnectionInfo{
			ConnectionURI: os.Getenv("SUPERTOKENS_URL"),
		},
		AppInfo: supertokens.AppInfo{
			AppName:         "Sidequest",
			APIDomain:       getApiDomain(),
			WebsiteDomain:   getWebsiteDomain(),
			APIBasePath:     getStringPointer("/auth"),
			WebsiteBasePath: getStringPointer("/auth"),
		},
		RecipeList: []supertokens.Recipe{
			emailpassword.Init(nil),
			session.Init(nil),
			dashboard.Init(nil),
			userroles.Init(nil),
		},
	}
}
