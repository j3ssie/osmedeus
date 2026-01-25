package notify

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/j3ssie/osmedeus/v5/internal/config"
)

// TelegramClient wraps the Telegram bot API
type TelegramClient struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

// NewTelegramClient creates a new Telegram client from config
func NewTelegramClient(cfg *config.TelegramConfig) (*TelegramClient, error) {
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}
	if cfg.ChatID == 0 {
		return nil, fmt.Errorf("telegram chat ID is required")
	}

	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	return &TelegramClient{
		bot:    bot,
		chatID: cfg.ChatID,
	}, nil
}

// ResolveTelegramChannelID resolves "#channel_name" or numeric ID string to int64
// If channel starts with "#", looks up in telegram_channel_map config
// Otherwise, parses as numeric chat ID
func ResolveTelegramChannelID(channel string) (int64, error) {
	if strings.HasPrefix(channel, "#") {
		// Look up in channel map
		channelName := strings.TrimPrefix(channel, "#")
		cfg := config.Get()
		if cfg == nil {
			return 0, fmt.Errorf("global config not loaded")
		}
		if cfg.Notification.Telegram.TelegramChannelMap == nil {
			return 0, fmt.Errorf("telegram_channel_map not configured")
		}
		chatID, ok := cfg.Notification.Telegram.TelegramChannelMap[channelName]
		if !ok {
			return 0, fmt.Errorf("channel '%s' not found in telegram_channel_map", channelName)
		}
		return chatID, nil
	}

	// Parse as numeric ID
	chatID, err := strconv.ParseInt(channel, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid channel ID '%s': %w", channel, err)
	}
	return chatID, nil
}

// NewTelegramClientFromGlobal creates a client from global config
func NewTelegramClientFromGlobal() (*TelegramClient, error) {
	cfg := config.Get()
	if cfg == nil {
		return nil, fmt.Errorf("global config not loaded")
	}
	if !cfg.IsTelegramConfigured() {
		return nil, fmt.Errorf("telegram not configured")
	}
	return NewTelegramClient(&cfg.Notification.Telegram)
}

// SendMessage sends a text message to the configured chat
func (c *TelegramClient) SendMessage(text string) error {
	msg := tgbotapi.NewMessage(c.chatID, text)
	_, err := c.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

// SendMessageToChatID sends a text message to a specific chat ID
func (c *TelegramClient) SendMessageToChatID(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := c.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to chat %d: %w", chatID, err)
	}
	return nil
}

// SendMarkdownToChatID sends a markdown message to a specific chat ID
func (c *TelegramClient) SendMarkdownToChatID(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := c.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send markdown message to chat %d: %w", chatID, err)
	}
	return nil
}

// SendFileToChatID sends a file (document) to a specific chat ID
func (c *TelegramClient) SendFileToChatID(chatID int64, filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(filePath))
	_, err := c.bot.Send(doc)
	if err != nil {
		return fmt.Errorf("failed to send file to chat %d: %w", chatID, err)
	}
	return nil
}

// SendFileToChatIDWithCaption sends a file with a caption to a specific chat ID
func (c *TelegramClient) SendFileToChatIDWithCaption(chatID int64, filePath, caption string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(filePath))
	doc.Caption = caption
	_, err := c.bot.Send(doc)
	if err != nil {
		return fmt.Errorf("failed to send file to chat %d: %w", chatID, err)
	}
	return nil
}

// SendMessagef sends a formatted text message
func (c *TelegramClient) SendMessagef(format string, args ...interface{}) error {
	return c.SendMessage(fmt.Sprintf(format, args...))
}

// SendMarkdown sends a message with Markdown formatting
func (c *TelegramClient) SendMarkdown(text string) error {
	msg := tgbotapi.NewMessage(c.chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := c.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send markdown message: %w", err)
	}
	return nil
}

// SendHTML sends a message with HTML formatting
func (c *TelegramClient) SendHTML(text string) error {
	msg := tgbotapi.NewMessage(c.chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := c.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send HTML message: %w", err)
	}
	return nil
}

// SendFile sends a file (document) to the configured chat
func (c *TelegramClient) SendFile(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	doc := tgbotapi.NewDocument(c.chatID, tgbotapi.FilePath(filePath))
	_, err := c.bot.Send(doc)
	if err != nil {
		return fmt.Errorf("failed to send file: %w", err)
	}
	return nil
}

// SendFileWithCaption sends a file with a caption
func (c *TelegramClient) SendFileWithCaption(filePath, caption string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	doc := tgbotapi.NewDocument(c.chatID, tgbotapi.FilePath(filePath))
	doc.Caption = caption
	_, err := c.bot.Send(doc)
	if err != nil {
		return fmt.Errorf("failed to send file: %w", err)
	}
	return nil
}

// SendPhoto sends a photo to the configured chat
func (c *TelegramClient) SendPhoto(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	photo := tgbotapi.NewPhoto(c.chatID, tgbotapi.FilePath(filePath))
	_, err := c.bot.Send(photo)
	if err != nil {
		return fmt.Errorf("failed to send photo: %w", err)
	}
	return nil
}

// SendPhotoWithCaption sends a photo with a caption
func (c *TelegramClient) SendPhotoWithCaption(filePath, caption string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	photo := tgbotapi.NewPhoto(c.chatID, tgbotapi.FilePath(filePath))
	photo.Caption = caption
	_, err := c.bot.Send(photo)
	if err != nil {
		return fmt.Errorf("failed to send photo: %w", err)
	}
	return nil
}

// GetBotInfo returns information about the bot
func (c *TelegramClient) GetBotInfo() (string, error) {
	return fmt.Sprintf("Bot: @%s (ID: %d)", c.bot.Self.UserName, c.bot.Self.ID), nil
}

// Convenience functions for quick use without creating a client

// SendTelegramMessage sends a message using global config with Markdown formatting
func SendTelegramMessage(text string) error {
	client, err := NewTelegramClientFromGlobal()
	if err != nil {
		return err
	}
	return client.SendMarkdown(text)
}

// SendTelegramFile sends a file using global config
func SendTelegramFile(filePath string) error {
	client, err := NewTelegramClientFromGlobal()
	if err != nil {
		return err
	}
	return client.SendFile(filePath)
}

// SendTelegramNotification sends a formatted notification
func SendTelegramNotification(title, message string) error {
	client, err := NewTelegramClientFromGlobal()
	if err != nil {
		return err
	}
	text := fmt.Sprintf("*%s*\n\n%s", title, message)
	return client.SendMarkdown(text)
}

// SendTelegramMessageToChannel sends a message to a specific channel with Markdown formatting
// channel can be "#channel_name" (looked up in config) or numeric chat ID string
func SendTelegramMessageToChannel(channel, text string) error {
	chatID, err := ResolveTelegramChannelID(channel)
	if err != nil {
		return err
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("global config not loaded")
	}
	if cfg.Notification.Telegram.BotToken == "" {
		return fmt.Errorf("telegram bot token is required")
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Notification.Telegram.BotToken)
	if err != nil {
		return fmt.Errorf("failed to create telegram bot: %w", err)
	}

	client := &TelegramClient{bot: bot, chatID: chatID}
	return client.SendMarkdownToChatID(chatID, text)
}

// SendTelegramFileToChannel sends a file to a specific channel
// channel can be "#channel_name" (looked up in config) or numeric chat ID string
func SendTelegramFileToChannel(channel, filePath string) error {
	chatID, err := ResolveTelegramChannelID(channel)
	if err != nil {
		return err
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("global config not loaded")
	}
	if cfg.Notification.Telegram.BotToken == "" {
		return fmt.Errorf("telegram bot token is required")
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Notification.Telegram.BotToken)
	if err != nil {
		return fmt.Errorf("failed to create telegram bot: %w", err)
	}

	client := &TelegramClient{bot: bot, chatID: chatID}
	return client.SendFileToChatID(chatID, filePath)
}

// SendTelegramFileToChannelWithCaption sends a file with caption to a specific channel
// channel can be "#channel_name" (looked up in config) or numeric chat ID string
func SendTelegramFileToChannelWithCaption(channel, filePath, caption string) error {
	chatID, err := ResolveTelegramChannelID(channel)
	if err != nil {
		return err
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("global config not loaded")
	}
	if cfg.Notification.Telegram.BotToken == "" {
		return fmt.Errorf("telegram bot token is required")
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Notification.Telegram.BotToken)
	if err != nil {
		return fmt.Errorf("failed to create telegram bot: %w", err)
	}

	client := &TelegramClient{bot: bot, chatID: chatID}
	return client.SendFileToChatIDWithCaption(chatID, filePath, caption)
}
