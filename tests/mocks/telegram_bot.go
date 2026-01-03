package mocks

import (
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MockTelegramBot struct {
	API              *tgbotapi.BotAPI
	sentMessages     []tgbotapi.MessageConfig
	sentCallbacks    []tgbotapi.CallbackConfig
	mu               sync.RWMutex
	messageCounter   int
	updatesChan      chan tgbotapi.Update
}

func NewMockTelegramBot() *MockTelegramBot {
	return &MockTelegramBot{
		sentMessages:   make([]tgbotapi.MessageConfig, 0),
		sentCallbacks:  make([]tgbotapi.CallbackConfig, 0),
		messageCounter: 0,
		updatesChan:    make(chan tgbotapi.Update),
	}
}

func (m *MockTelegramBot) GetSentMessagesCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sentMessages)
}

func (m *MockTelegramBot) GetLastSentMessage() *tgbotapi.MessageConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.sentMessages) == 0 {
		return nil
	}
	return &m.sentMessages[len(m.sentMessages)-1]
}

func (m *MockTelegramBot) GetAllSentMessages() []tgbotapi.MessageConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	messages := make([]tgbotapi.MessageConfig, len(m.sentMessages))
	copy(messages, m.sentMessages)
	return messages
}

func (m *MockTelegramBot) RecordMessage(msg tgbotapi.MessageConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentMessages = append(m.sentMessages, msg)
	m.messageCounter++
}

func (m *MockTelegramBot) RecordCallback(callback tgbotapi.CallbackConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentCallbacks = append(m.sentCallbacks, callback)
}

func (m *MockTelegramBot) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentMessages = make([]tgbotapi.MessageConfig, 0)
	m.sentCallbacks = make([]tgbotapi.CallbackConfig, 0)
	m.messageCounter = 0
}

func (m *MockTelegramBot) GetCallbackCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sentCallbacks)
}
