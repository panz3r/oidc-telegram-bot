package common

import (
	"context"
	"net/http"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

// OIDCConnector represents a wrapper to use to communicate with OIDC Providers
type OIDCConnector struct {
	config   oauth2.Config
	context  context.Context
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
}

// NewOIDCConnector returns a new OIDCConnector with the supplied attributes
func NewOIDCConnector(configURL string, clientID string, clientSecret string, redirectURL string) OIDCConnector {
	ctx := context.Background()

	prv, err := oidc.NewProvider(ctx, configURL)
	if err != nil {
		panic(err)
	}

	vrf := prv.Verifier(&oidc.Config{
		ClientID: clientID,
	})

	return OIDCConnector{
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			// Discovery returns the OAuth2 endpoints.
			Endpoint: prv.Endpoint(),
			// "openid" is a required scope for OpenID Connect flows.
			Scopes: []string{
				oidc.ScopeOpenID,
				"email",
				"profile",
				"offline",
			},
		},
		context:  ctx,
		provider: prv,
		verifier: vrf,
	}
}

func (cnn OIDCConnector) Exchange(code string) (*oauth2.Token, error) {
	return cnn.config.Exchange(cnn.context, code)
}

func (cnn OIDCConnector) Verify(rawIDToken string) (*oidc.IDToken, error) {
	return cnn.verifier.Verify(cnn.context, rawIDToken)
}

func (cnn OIDCConnector) AuthCodeURL(state string) string {
	return cnn.config.AuthCodeURL(state)
}

func (cnn OIDCConnector) GetClient(t *oauth2.Token) *http.Client {
	return cnn.config.Client(cnn.context, t)
}
