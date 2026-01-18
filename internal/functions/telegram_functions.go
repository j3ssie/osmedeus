package functions

import (
	"github.com/dop251/goja"
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
