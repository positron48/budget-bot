package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestNewTestBot_Exists(t *testing.T) {
	// Test that the function exists and can be called
	assert.NotNil(t, NewTestBot)
}

func TestNewTestBot_ReturnsBot(t *testing.T) {
	// Test that NewTestBot returns a bot
	bot := NewTestBot(t)
	assert.NotNil(t, bot)
	assert.IsType(t, &tgbotapi.BotAPI{}, bot)
}
