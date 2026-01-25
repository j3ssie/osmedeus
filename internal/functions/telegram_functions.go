package functions

import (
	"os"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/installer"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/notify"
	"go.uber.org/zap"
)

// notifyTelegram sends a message to Telegram
// Usage: notifyTelegram("message") -> bool
func (vf *vmFunc) notifyTelegram(call goja.FunctionCall) goja.Value {
	message := call.Argument(0).String()
	logger.Get().Debug("Calling notifyTelegram", zap.Int("msgLength", len(message)))

	if message == "undefined" || message == "" {
		logger.Get().Warn("notifyTelegram: empty message provided")
		return vf.vm.ToValue(false)
	}

	err := notify.SendTelegramMessage(message)
	if err != nil {
		logger.Get().Warn("notifyTelegram: failed to send message", zap.Error(err))
	} else {
		logger.Get().Debug("notifyTelegram result", zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// sendTelegramFile sends a file to Telegram
// Usage: sendTelegramFile(path) -> bool
// Usage: sendTelegramFile(path, caption) -> bool
func (vf *vmFunc) sendTelegramFile(call goja.FunctionCall) goja.Value {
	filePath := call.Argument(0).String()
	logger.Get().Debug("Calling sendTelegramFile", zap.String("filePath", filePath))

	if filePath == "undefined" || filePath == "" {
		logger.Get().Warn("sendTelegramFile: empty file path provided")
		return vf.vm.ToValue(false)
	}

	// Expand path (handles ~, $HOME, etc.)
	filePath = installer.ExpandPath(filePath)

	caption := ""
	if !goja.IsUndefined(call.Argument(1)) {
		caption = call.Argument(1).String()
	}

	var err error
	if caption != "" {
		client, e := notify.NewTelegramClientFromGlobal()
		if e != nil {
			logger.Get().Warn("sendTelegramFile: failed to create Telegram client", zap.Error(e))
			return vf.vm.ToValue(false)
		}
		err = client.SendFileWithCaption(filePath, caption)
	} else {
		err = notify.SendTelegramFile(filePath)
	}

	if err != nil {
		logger.Get().Warn("sendTelegramFile: failed to send file", zap.String("filePath", filePath), zap.Error(err))
	} else {
		logger.Get().Debug("sendTelegramFile result", zap.String("filePath", filePath), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// notifyTelegramChannel sends a message to a specific Telegram channel
// Usage: notifyTelegramChannel(channel, message) -> bool
// channel can be "#channel_name" (looked up in telegram_channel_map) or numeric chat ID string
func (vf *vmFunc) notifyTelegramChannel(call goja.FunctionCall) goja.Value {
	channel := call.Argument(0).String()
	message := call.Argument(1).String()
	logger.Get().Debug("Calling notifyTelegramChannel", zap.String("channel", channel), zap.Int("msgLength", len(message)))

	if channel == "undefined" || channel == "" {
		logger.Get().Warn("notifyTelegramChannel: empty channel provided")
		return vf.vm.ToValue(false)
	}
	if message == "undefined" || message == "" {
		logger.Get().Warn("notifyTelegramChannel: empty message provided")
		return vf.vm.ToValue(false)
	}

	err := notify.SendTelegramMessageToChannel(channel, message)
	if err != nil {
		logger.Get().Warn("notifyTelegramChannel: failed to send message", zap.String("channel", channel), zap.Error(err))
	} else {
		logger.Get().Debug("notifyTelegramChannel result", zap.String("channel", channel), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// sendTelegramFileChannel sends a file to a specific Telegram channel
// Usage: sendTelegramFileChannel(channel, path) -> bool
// Usage: sendTelegramFileChannel(channel, path, caption) -> bool
// channel can be "#channel_name" (looked up in telegram_channel_map) or numeric chat ID string
func (vf *vmFunc) sendTelegramFileChannel(call goja.FunctionCall) goja.Value {
	channel := call.Argument(0).String()
	filePath := call.Argument(1).String()
	logger.Get().Debug("Calling sendTelegramFileChannel", zap.String("channel", channel), zap.String("filePath", filePath))

	if channel == "undefined" || channel == "" {
		logger.Get().Warn("sendTelegramFileChannel: empty channel provided")
		return vf.vm.ToValue(false)
	}
	if filePath == "undefined" || filePath == "" {
		logger.Get().Warn("sendTelegramFileChannel: empty file path provided")
		return vf.vm.ToValue(false)
	}

	// Expand path (handles ~, $HOME, etc.)
	filePath = installer.ExpandPath(filePath)

	caption := ""
	if !goja.IsUndefined(call.Argument(2)) {
		caption = call.Argument(2).String()
	}

	var err error
	if caption != "" {
		err = notify.SendTelegramFileToChannelWithCaption(channel, filePath, caption)
	} else {
		err = notify.SendTelegramFileToChannel(channel, filePath)
	}

	if err != nil {
		logger.Get().Warn("sendTelegramFileChannel: failed to send file", zap.String("channel", channel), zap.String("filePath", filePath), zap.Error(err))
	} else {
		logger.Get().Debug("sendTelegramFileChannel result", zap.String("channel", channel), zap.String("filePath", filePath), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// notifyMessageAsFileTelegram reads file content and sends as markdown message to default Telegram channel
// Usage: notifyMessageAsFileTelegram(path) -> bool
func (vf *vmFunc) notifyMessageAsFileTelegram(call goja.FunctionCall) goja.Value {
	filePath := call.Argument(0).String()
	logger.Get().Debug("Calling notifyMessageAsFileTelegram", zap.String("filePath", filePath))

	if filePath == "undefined" || filePath == "" {
		logger.Get().Warn("notifyMessageAsFileTelegram: empty file path provided")
		return vf.vm.ToValue(false)
	}

	// Expand path (handles ~, $HOME, etc.)
	filePath = installer.ExpandPath(filePath)

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		logger.Get().Warn("notifyMessageAsFileTelegram: failed to read file", zap.String("filePath", filePath), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Send content as markdown message
	err = notify.SendTelegramMessage(string(content))
	if err != nil {
		logger.Get().Warn("notifyMessageAsFileTelegram: failed to send message", zap.String("filePath", filePath), zap.Error(err))
	} else {
		logger.Get().Debug("notifyMessageAsFileTelegram result", zap.String("filePath", filePath), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// notifyMessageAsFileTelegramChannel reads file content and sends as markdown message to specific Telegram channel
// Usage: notifyMessageAsFileTelegramChannel(channel, path) -> bool
// channel can be "#channel_name" (looked up in telegram_channel_map) or numeric chat ID string
func (vf *vmFunc) notifyMessageAsFileTelegramChannel(call goja.FunctionCall) goja.Value {
	channel := call.Argument(0).String()
	filePath := call.Argument(1).String()
	logger.Get().Debug("Calling notifyMessageAsFileTelegramChannel", zap.String("channel", channel), zap.String("filePath", filePath))

	if channel == "undefined" || channel == "" {
		logger.Get().Warn("notifyMessageAsFileTelegramChannel: empty channel provided")
		return vf.vm.ToValue(false)
	}
	if filePath == "undefined" || filePath == "" {
		logger.Get().Warn("notifyMessageAsFileTelegramChannel: empty file path provided")
		return vf.vm.ToValue(false)
	}

	// Expand path (handles ~, $HOME, etc.)
	filePath = installer.ExpandPath(filePath)

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		logger.Get().Warn("notifyMessageAsFileTelegramChannel: failed to read file", zap.String("filePath", filePath), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Send content as markdown message to channel
	err = notify.SendTelegramMessageToChannel(channel, string(content))
	if err != nil {
		logger.Get().Warn("notifyMessageAsFileTelegramChannel: failed to send message", zap.String("channel", channel), zap.String("filePath", filePath), zap.Error(err))
	} else {
		logger.Get().Debug("notifyMessageAsFileTelegramChannel result", zap.String("channel", channel), zap.String("filePath", filePath), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}
