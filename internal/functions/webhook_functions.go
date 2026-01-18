package functions

import (
	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/notify"
	"go.uber.org/zap"
)

// notifyWebhook sends a message to all configured webhooks
// Usage: notifyWebhook("message") -> bool
func (vf *vmFunc) notifyWebhook(call goja.FunctionCall) goja.Value {
	message := call.Argument(0).String()
	logger.Get().Debug("Calling notifyWebhook", zap.Int("msgLength", len(message)))

	if message == "undefined" || message == "" {
		logger.Get().Warn("notifyWebhook: empty message provided")
		return vf.vm.ToValue(false)
	}

	err := notify.SendWebhookMessage(message)
	if err != nil {
		logger.Get().Warn("notifyWebhook: failed to send message", zap.Error(err))
	} else {
		logger.Get().Debug("notifyWebhook result", zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// sendWebhookEvent sends a structured event to all configured webhooks
// Usage: sendWebhookEvent("event_type", {key: value}) -> bool
func (vf *vmFunc) sendWebhookEvent(call goja.FunctionCall) goja.Value {
	eventType := call.Argument(0).String()
	logger.Get().Debug("Calling sendWebhookEvent", zap.String("eventType", eventType))

	if eventType == "undefined" || eventType == "" {
		logger.Get().Warn("sendWebhookEvent: empty event type provided")
		return vf.vm.ToValue(false)
	}

	// Parse data argument as map
	data := make(map[string]interface{})
	if !goja.IsUndefined(call.Argument(1)) {
		dataObj := call.Argument(1).ToObject(vf.vm)
		if dataObj != nil {
			for _, key := range dataObj.Keys() {
				val := dataObj.Get(key)
				if val != nil {
					// Convert to Go value
					data[key] = val.Export()
				}
			}
		}
	}

	err := notify.SendWebhookEvent(eventType, data)
	if err != nil {
		logger.Get().Warn("sendWebhookEvent: failed to send event",
			zap.String("eventType", eventType),
			zap.Error(err))
	} else {
		logger.Get().Debug("sendWebhookEvent result",
			zap.String("eventType", eventType),
			zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}
