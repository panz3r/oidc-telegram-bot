package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/uuid"
	"github.com/machinebox/graphql"
	"golang.org/x/oauth2"

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

type doorsResponse struct {
	Doors []struct {
		Name        string
		DisplayName string
	}
}

type doorOpenResponse struct {
	DoorOpen struct {
		Name   string
		Status string
	}
}

func getHandler(c common.OIDCConnector, p common.Persistor, b *common.TGBot) LambdaHandler {
	// Handler is our lambda handler invoked by the `lambda.Start` function call
	return func(req Request) (Response, error) {
		var update tgbotapi.Update

		if err := json.Unmarshal([]byte(req.Body), &update); err != nil {
			fmt.Println("Error decoding Telegram Update:", err)
			return Response{
				StatusCode: http.StatusInternalServerError,
				Body:       "Error decoding Telegram Update:" + err.Error(),
			}, err
		}

		var chat *tgbotapi.Chat
		var tgUserID int

		actionType := "msg"
		if update.Message != nil {
			chat = update.Message.Chat
			tgUserID = update.Message.From.ID
		} else if update.CallbackQuery != nil {
			actionType = "cb"
			chat = update.CallbackQuery.Message.Chat
			tgUserID = update.CallbackQuery.From.ID
		}

		if chat == nil {
			return Response{
				StatusCode: http.StatusOK,
				Body:       "No chat to reply to... aborting",
			}, nil
		}

		usr, err := p.FindUserByTelegramID(tgUserID)
		if err != nil {
			fmt.Printf("Failed to retrieve Telegram User (%d) from Persistor\n", tgUserID)

			return Response{
				StatusCode: http.StatusOK,
				Body:       "Failed to retrieve Telegram User:" + err.Error(),
			}, err
		}

		if usr == nil || usr.AccessToken == "" {
			fmt.Println("Authentication required for Telegram user: " + chat.UserName)
			state := uuid.New().String()

			welcomeMsg := tgbotapi.NewMessage(chat.ID, "Login to allow me to better help you")
			welcomeMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				[]tgbotapi.InlineKeyboardButton{
					tgbotapi.NewInlineKeyboardButtonURL("Login", c.AuthCodeURL(state)),
				},
			)

			msg, err := b.SendMessage(chat.ID, welcomeMsg)
			if err != nil {
				fmt.Println("Error sending message", err.Error())

				return Response{
					StatusCode: http.StatusOK,
					Body:       "Error sending message:" + err.Error(),
				}, err
			}

			if err := p.StoreUserOAuthState(tgUserID, chat.ID, msg.MessageID, state); err != nil {
				return Response{
					StatusCode: http.StatusOK,
					Body:       "Failed to store User OAuth2 State:" + err.Error(),
				}, err
			}

			return Response{
				StatusCode: http.StatusOK,
			}, nil
		}

		// User is logged in

		exp, err := time.Parse("2006-01-02T15:04:05.000Z", usr.TokenExpiry)
		if err != nil {
			fmt.Println("Error parsing stored token time", err.Error())
			return Response{
				StatusCode: http.StatusOK,
				Body:       "Error parsing stored token time: " + err.Error(),
			}, err
		}

		token := new(oauth2.Token)
		token.AccessToken = usr.AccessToken
		token.RefreshToken = usr.RefreshToken
		token.Expiry = exp
		token.TokenType = "bearer"

		// Get an HTTPClient with token
		clt := c.GetClient(token)

		return Response{
			StatusCode: http.StatusOK,
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
	tgBotAPIToken := os.Getenv("TG_BOT_TOKEN")

	// Init Providers
	oidcCnn := common.NewOIDCConnector(configURL, clientID, clientSecret, redirectURL)
	persistor := common.NewPersistor(awsRegion, dbTable, "")
	tgBot, err := common.NewTGBot(tgBotAPIToken)
	if err != nil {
		panic(err)
	}

	// Inject Providers into Handler
	h := getHandler(oidcCnn, persistor, tgBot)

	// Star Handler
	lambda.Start(h)
}
