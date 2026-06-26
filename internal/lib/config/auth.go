package config

import (
	"fmt"

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
	apiPort := "8080"
	apiUrl := fmt.Sprintf("http://localhost:%s", apiPort)
	return apiUrl
}

func getWebsiteDomain() string {
	websitePort := "3000"
	websiteUrl := fmt.Sprintf("http://localhost:%s", websitePort)
	return websiteUrl
}

var SuperTokensConfig = supertokens.TypeInput{
	Supertokens: &supertokens.ConnectionInfo{
		ConnectionURI: "https://try.supertokens.com",
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
