package client

import (
	"fmt"

	"golang.org/x/oauth2"
)

type OAuthConfigBuilderArgs struct {
	AuthServerBaseURL string
	Port              int
	// the config of the github application.
	ClientID     string
	ClientSecret string
}

func OauthConfigBuilder(args OAuthConfigBuilderArgs) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     args.ClientID,
		ClientSecret: args.ClientSecret,
		Scopes:       []string{"repo"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/authorize", args.AuthServerBaseURL),
			TokenURL: fmt.Sprintf("%s/access_token", args.AuthServerBaseURL),
		},
		RedirectURL: fmt.Sprintf("http://localhost:%d/callback", args.Port),
	}
}
