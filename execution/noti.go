package execution

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cast"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/robertkrimen/otto"
	"github.com/slack-go/slack"
)

// StatusNoti send to when module is done with report file
func StatusNoti(notiType string, options libs.Options) {
	SendAttachment(notiType, "", options)
}

// ReportNoti send to notification when module is done with report file
func ReportNoti(arguments []otto.Value, options libs.Options) {
	if len(arguments) >= 1 {
		for _, argument := range arguments {
			SendFile(argument.String(), options.Noti.SlackReportChannel, options)
		}
		return
	}
	// if doesn't provide report file send all file in noti section
	for _, file := range options.Module.Report.Noti {
		SendFile(file, options.Noti.SlackReportChannel, options)
	}
}

// DiffNoti send to notification based on diff content
func DiffNoti(arguments []otto.Value, options libs.Options) {
	if len(arguments) >= 1 {
		for _, argument := range arguments {
			filename := argument.String()
			if !utils.FileExists(filename) {
				continue
			}
			data := getNewContent(utils.ReadingLines(filename))
			if strings.TrimSpace(data) == "" {
				continue
			}
			// messageContent := fmt.Sprintf("%v \n ```%v```", filename, data)
			// SendAttachment("diff", messageContent, options)
			SendFile(filename, options.Noti.SlackDiffChannel, options)
		}
		return
	}
	// if doesn't provide report file send all file in noti section
	for _, filename := range options.Module.Report.Diff {
		if !utils.FileExists(filename) {
			continue
		}
		data := getNewContent(utils.ReadingLines(filename))
		if strings.TrimSpace(data) == "" {
			continue
		}
		// messageContent := fmt.Sprintf("%v \n ```%v```", filename, data)
		// SendAttachment("diff", messageContent, options)
		SendFile(filename, options.Noti.SlackDiffChannel, options)
	}
}

func getNewContent(data []string) string {
	var result string
	for _, line := range data {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "++") {
			result += fmt.Sprintf("%v\n", line)
		}
	}
	return result
}

// SendAttachment send attach message to specific channel
func SendAttachment(messType string, messContent string, options libs.Options) error {
	if options.Noti.SlackToken == "" {
		return errors.New("Slack config improperly")
	}

	// choose method
	var channel, color, mess string
	switch messType {
	case "start":
		channel = options.Noti.SlackStatusChannel
		color = "#005b9f"
		mess = fmt.Sprintf("%v Start to run *%v* on *%v*", GetEmoji(), options.Module.Name, options.Scan.ROptions["Workspace"])
		break
	case "done":
		channel = options.Noti.SlackStatusChannel
		color = "#32cb00"
		mess = fmt.Sprintf("%v Done run *%v* on *%v*", GetEmoji(), options.Module.Name, options.Scan.ROptions["Workspace"])
		break
	case "diff":
		channel = options.Noti.SlackDiffChannel
		color = "#5E35B1"
		mess = fmt.Sprintf("%v Diff content on %v: \n %v", GetEmoji(), options.Scan.ROptions["Workspace"], messContent)
		break
	case "custom":
		channel = options.Noti.SlackStatusChannel
		color = "#1ABC9C"
		mess = messContent
		break
	}

	if channel == "" || mess == "" {
		return errors.New("Slack channel config improperly")
	}
	utils.DebugF("Sending %v message to %v", messType, channel)

	api := slack.New(options.Noti.SlackToken)
	// message config
	attachment := slack.Attachment{
		Color: color,
		Text:  mess,
		// sender name
		Footer:     options.Noti.ClientName,
		FooterIcon: GetIcon(),
		Ts:         json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
	}

	_, _, err := api.PostMessage(channel, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return err
	}
	return nil
}

// SlackWebHook send message with webhook
func SlackWebHook(webhookURL string, content string) error {
	content = fmt.Sprintf("```%s```", content)
	attachment := slack.Attachment{
		Color: "#1ABC9C",
		Text:  content,
		// sender name
		//Footer:     options.Noti.ClientName,
		FooterIcon: GetIcon(),
		Ts:         json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
	}

	msg := slack.WebhookMessage{
		Attachments: []slack.Attachment{attachment},
	}
	err := slack.PostWebhook(webhookURL, &msg)
	return err
}

// WebHookSendAttachment send attach message to specific channel
func WebHookSendAttachment(options libs.Options, messType string, messContent string) error {
	if options.NoNoti {
		return fmt.Errorf("noti disabled")
	}
	if options.Noti.SlackWebHook == "" {
		return errors.New("slack webhook config improperly")
	}

	// choose method
	var color, mess string
	switch messType {
	case "start":
		color = "#005b9f"
		mess = fmt.Sprintf("%v Start to run *%v* on *%v*", GetEmoji(), options.Module.Name, options.Scan.ROptions["Workspace"])
		break
	case "done":
		color = "#32cb00"
		mess = fmt.Sprintf("%v Done run *%v* on *%v*", GetEmoji(), options.Module.Name, options.Scan.ROptions["Workspace"])
		break
	case "diff":
		color = "#5E35B1"
		mess = fmt.Sprintf("%v Diff content on %v: \n %v", GetEmoji(), options.Scan.ROptions["Workspace"], messContent)
		break
	case "custom":
		color = "#1ABC9C"
		mess = messContent
		break
	default:
		color = "#1ABC9C"
		mess = messContent
	}

	// message config
	attachment := slack.Attachment{
		Color: color,
		Text:  mess,
		// sender name
		Footer:     options.Noti.ClientName,
		FooterIcon: GetIcon(),
		Ts:         json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
	}

	msg := slack.WebhookMessage{
		Attachments: []slack.Attachment{attachment},
	}
	err := slack.PostWebhook(options.Noti.SlackWebHook, &msg)
	if err != nil {
		return err
	}
	return nil
}

// SendFile send file to specific channel
func SendFile(filename string, channel string, options libs.Options) error {
	if options.Noti.SlackToken == "" || options.Noti.SlackReportChannel == "" {
		return fmt.Errorf("Slack config improperly")
	}

	if !utils.FileExists(filename) {
		return fmt.Errorf("report file not found: %v", filename)
	}

	baseName := filepath.Base(filename)
	mess := fmt.Sprintf("%v - %v - Report file for *%v* on *%v*", GetEmoji(), baseName, options.Module.Name, options.Scan.ROptions["Workspace"])
	utils.DebugF("Sending %v message to %v", filename, channel)

	// sending file
	api := slack.New(options.Noti.SlackToken)
	params := slack.FileUploadParameters{
		Channels: []string{channel},
		Title:    mess,
		Filetype: "txt",
		File:     filename,
	}
	_, err := api.UploadFile(params)
	if err != nil {
		return err
	}
	return nil
}

// TeleSendMess send message to telegram
func TeleSendMess(options libs.Options, content string, channel string, wrap bool) error {

	if options.NoNoti {
		return fmt.Errorf("noti disabled")
	}
	bot, err := tgbotapi.NewBotAPI(options.Noti.TelegramToken)
	if wrap {
		content = fmt.Sprintf("```\n%s\n```", content)
	}
	if err != nil {
		utils.DebugF("error init telegram: %v", err)
		return err
	}

	if channel == "" || channel == "general" {
		channel = options.Noti.TelegramChannel
	}
	switch channel {
	case "#status":
		channel = options.Noti.TelegramStatusChannel
	case "#r", "#report", "#reports", "#vuln":
		channel = options.Noti.TelegramReportChannel
	case "#s", "#sensitive", "#sen":
		channel = options.Noti.TelegramSensitiveChannel
	case "#dirb", "#dirscan":
		channel = options.Noti.TelegramDirbChannel
	case "#m", "#mics":
		channel = options.Noti.TelegramMicsChannel
	case "#default", "#general":
		channel = options.Noti.TelegramChannel
	}
	telechannel := cast.ToInt64(channel)
	utils.DebugF("send message to channel %v", channel)
	msg := tgbotapi.NewMessage(telechannel, content)
	msg.ParseMode = "markdown"
	_, err = bot.Send(msg)
	if err != nil {
		utils.DebugF("error sending telegram to %v -- %v", channel, err)
	}
	return err
}

// TeleSendFile send message to telegram
func TeleSendFile(options libs.Options, filename string, channel string) error {
	if options.NoNoti {
		return fmt.Errorf("noti disabled")
	}
	bot, err := tgbotapi.NewBotAPI(options.Noti.TelegramToken)
	if err != nil {
		utils.DebugF("error init telegram: %v", err)
		return err
	}

	if channel == "" || channel == "general" {
		channel = options.Noti.TelegramChannel
	}
	switch channel {
	case "#status":
		channel = options.Noti.TelegramStatusChannel
	case "#r", "#report", "#reports", "#vuln":
		channel = options.Noti.TelegramReportChannel
	case "#sensitive", "#sen":
		channel = options.Noti.TelegramSensitiveChannel
	case "#dirb", "#dirscan":
		channel = options.Noti.TelegramDirbChannel
	case "#m", "#mics":
		channel = options.Noti.TelegramMicsChannel
	case "#default":
		channel = options.Noti.TelegramChannel
	}
	telechannel := cast.ToInt64(channel)

	filename = utils.NormalizePath(filename)
	msg := tgbotapi.NewDocumentUpload(telechannel, filename)
	utils.DebugF("send file %v to channel %v", filename, channel)
	_, err = bot.Send(msg)
	if err != nil {
		utils.DebugF("error sending telegram to %v -- %v", channel, err)
	}
	return err
}

/////// utils for slack message

// GetEmoji get random emoji
func GetEmoji() string {
	rand.Seed(time.Now().Unix())
	emojis := []string{
		":robot_face:",
		":alien:",
		":gift:",
		":gun:",
		":diamond_shape_with_a_dot_inside:",
		":rocket:",
		":bug:",
		":broccoli:",
		":shamrock:",
	}
	n := rand.Int() % len(emojis)
	return emojis[n]
}

// GetIcon get random emoji
func GetIcon() string {
	rand.Seed(time.Now().Unix())
	emojis := []string{
		"https://platform.slack-edge.com/img/default_application_icon.png",
	}
	n := rand.Int() % len(emojis)
	return emojis[n]
}
