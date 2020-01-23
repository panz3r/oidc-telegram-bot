package common

import (
	"errors"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Persistor struct {
	db        *dynamodb.DynamoDB
	tableName string
}

type User struct {
	TelegramUserID       int
	TelegramChatID       int64
	TelegramWelcomeMsgID int
	OAuth2State          string
	AccessToken          string
	RefreshToken         string
	TokenExpiry          string
}

func NewPersistor(region string, tableName string, endpoint string) Persistor {
	// Connect DynamoDB
	sess := session.Must(session.NewSession())
	config := aws.NewConfig().WithRegion(region)

	if len(endpoint) > 0 {
		config = config.WithEndpoint(endpoint)
	}

	return Persistor{
		db:        dynamodb.New(sess, config),
		tableName: tableName,
	}
}

func (p Persistor) FindUserByTelegramID(telegramUserID int) (*User, error) {
	response, err := p.db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(p.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"TelegramUserID": {
				N: aws.String(strconv.Itoa(telegramUserID)),
			},
		},
		AttributesToGet: []*string{
			aws.String("TelegramUserID"),
			aws.String("TelegramChatID"),
			aws.String("TelegramWelcomeMsgID"),
			aws.String("OAuth2State"),
			aws.String("AccessToken"),
			aws.String("RefreshToken"),
			aws.String("TokenExpiry"),
		},
		ConsistentRead:         aws.Bool(true),
		ReturnConsumedCapacity: aws.String("NONE"),
	})

	if err != nil {
		return nil, err
	}

	user := User{}
	if err := dynamodbattribute.Unmarshal(&dynamodb.AttributeValue{M: response.Item}, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (p Persistor) FindUserByOAuthState(oauthState string) (*User, error) {
	res, err := p.db.Scan(&dynamodb.ScanInput{
		TableName: aws.String(p.tableName),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":OAuth2State": {
				S: aws.String(oauthState),
			},
		},
		FilterExpression: aws.String("OAuth2State = :OAuth2State"),
	})

	if err != nil {
		return nil, err
	}

	users := []User{}
	if err := dynamodbattribute.UnmarshalListOfMaps(res.Items, &users); err != nil {
		return nil, err
	}

	if len(users) < 1 {
		return nil, errors.New("User not found for oauthState " + oauthState)
	}

	return &users[0], nil
}

func (p Persistor) StoreUserOAuthState(userID int, chatID int64, welcomeMsgID int, oauthState string) error {
	_, err := p.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(p.tableName),
		Item: map[string]*dynamodb.AttributeValue{
			"TelegramUserID": {
				N: aws.String(strconv.Itoa(userID)),
			},
			"TelegramChatID": {
				N: aws.String(strconv.FormatInt(chatID, 10)),
			},
			"TelegramWelcomeMsgID": {
				N: aws.String(strconv.Itoa(welcomeMsgID)),
			},
			"OAuth2State": {
				S: aws.String(oauthState),
			},
		},
	})

	return err
}

func (p Persistor) StoreUserTokens(userID int, accessToken string, refreshToken string, tokenExpiry time.Time) error {
	_, err := p.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(p.tableName),
		Item: map[string]*dynamodb.AttributeValue{
			"TelegramUserID": {
				N: aws.String(strconv.Itoa(userID)),
			},
			"AccessToken": {
				S: aws.String(accessToken),
			},
			"RefreshToken": {
				S: aws.String(refreshToken),
			},
			"TokenExpiry": {
				S: aws.String(tokenExpiry.Format("2006-01-02T15:04:05.000Z")),
			},
		},
	})

	return err
}
