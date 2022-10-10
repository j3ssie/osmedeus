package libs

// Notification struct define notification method
type Notification struct {
	ClientName string
	// SlacksWebHooks list
	SlacksWebHooks map[string]string
	// TelegramWebHooks list
	TelegramWebHooks map[string]string
	// Telegram part
	TelegramToken            string
	TelegramChannel          string
	TelegramStatusChannel    string
	TelegramReportChannel    string
	TelegramDirbChannel      string
	TelegramSensitiveChannel string
	TelegramMicsChannel      string
	// use this when we want to send a file to channel
	SlackWebHook       string
	SlackToken         string
	SlackReportChannel string
	SlackStatusChannel string
	SlackDiffChannel   string
	// later then
	DiscordToken string
}
