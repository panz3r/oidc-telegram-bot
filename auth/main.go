package main

import (
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	common "github.com/panz3r/aws-go-bot"
)

// Request is of type APIGatewayProxyRequest since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Request events.APIGatewayProxyRequest

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

// LambdaHandler is the handler function invoked by the `lambda.Start` function call
type LambdaHandler func(Request) (Response, error)

func getHandler(c common.OIDCConnector, p common.Persistor, b *common.TGBot) LambdaHandler {
	// Handler is our lambda handler invoked by the `lambda.Start` function call
	return func(req Request) (Response, error) {
		cd := req.QueryStringParameters["code"]
		st := req.QueryStringParameters["state"]

		usr, err := p.FindUserByOAuthState(st) // Retrieve user state from DynamoDB
		if err != nil {
			return Response{
				StatusCode: http.StatusInternalServerError,
				Body:       "Failed to retrieve User from Persistor: " + err.Error(),
			}, err
		}

		if st != usr.OAuth2State {
			return Response{
				StatusCode: http.StatusUnauthorized,
				Body:       "Error: State did not match",
			}, nil
		}

		oauth2Token, err := c.Exchange(cd)
		if err != nil {
			return Response{
				StatusCode: http.StatusInternalServerError,
				Body:       "Failed to exchange token: " + err.Error(),
			}, nil
		}

		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			return Response{
				StatusCode: http.StatusInternalServerError,
				Body:       "No id_token field in oauth2 token.",
			}, nil
		}

		if _, err := c.Verify(rawIDToken); err != nil {
			return Response{
				StatusCode: http.StatusInternalServerError,
				Body:       "Failed to verify ID Token: " + err.Error(),
			}, err
		}

		if err := p.StoreUserTokens(usr.TelegramUserID, oauth2Token.AccessToken, oauth2Token.RefreshToken, oauth2Token.Expiry); err != nil {
			return Response{
				StatusCode: http.StatusInternalServerError,
				Body:       "Failed to store User Tokens: " + err.Error(),
			}, err
		}

		// Delete Welcome message
		if err := b.DeleteMessage(usr.TelegramChatID, usr.TelegramWelcomeMsgID); err != nil {
			return Response{
				StatusCode: http.StatusInternalServerError,
				Body:       "Failed to delete welcome message: " + err.Error(),
			}, err
		}

		return Response{
			StatusCode: http.StatusFound,
			Headers: map[string]string{
				"Location": os.Getenv("POST_AUTH_REDIRECT"),
			},
		}, nil
	}
}

func main() {
	// AWS env variables
	awsRegion := os.Getenv("AWS_REGION")
	if len(awsRegion) == 0 {
		awsRegion = "eu-west-1"
	}
	apiID := os.Getenv("AWS_AG_API_ID")
	stage := os.Getenv("AWS_AG_API_STAGE")
	dbTable := os.Getenv("AWS_DB_TABLE")

	// Keycloak env variables
	configURL := os.Getenv("OIDC_CONFIG_URL")
	clientID := os.Getenv("OIDC_CLIENT_ID")
	clientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	redirectURL := "https://" + apiID + ".execute-api." + awsRegion + ".amazonaws.com/" + stage + "/auth/callback"

	// Bot env variables
	tgBotTkn := os.Getenv("TG_BOT_TOKEN")

	// Init Providers
	oidcCnn := common.NewOIDCConnector(configURL, clientID, clientSecret, redirectURL)
	persistor := common.NewPersistor(awsRegion, dbTable, "")
	tgBot, err := common.NewTGBot(tgBotTkn)
	if err != nil {
		panic(err)
	}

	// Inject Providers into Handler
	h := getHandler(oidcCnn, persistor, tgBot)

	// Star Handler
	lambda.Start(h)
}
