module aws-golang-bot

go 1.13

require (
	github.com/aws/aws-lambda-go v1.13.3
	github.com/aws/aws-sdk-go v1.27.4
	github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/google/uuid v1.1.1
	github.com/machinebox/graphql v0.2.2
	github.com/matryer/is v1.2.0 // indirect
	github.com/panz3r/aws-go-bot v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.8.1 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
)

replace github.com/panz3r/aws-go-bot => ./common
