package common

import (
	tgapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TGBot struct {
	bot *tgapi.BotAPI
}

func NewTGBot(tkn string) (*TGBot, error) {
	tgbot, err := tgapi.NewBotAPI(tkn)
	if err != nil {
		return nil, err
	}

	return &TGBot{
		bot: tgbot,
	}, nil
}

func (b TGBot) SendMessage(chatID int64, message tgapi.Chattable) (tgapi.Message, error) {
	return b.bot.Send(message)
}

func (b TGBot) DeleteMessage(cID int64, msgID int) error {
	_, err := b.bot.DeleteMessage(
		tgapi.NewDeleteMessage(cID, msgID),
	)
	return err
}

func (b TGBot) AnswerCallback(cqID string, text string) error {
	cbAnswer := tgapi.NewCallback(cqID, text)
	_, err := b.bot.AnswerCallbackQuery(cbAnswer)
	return err
}
