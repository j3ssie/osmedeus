package execution

import (
	"github.com/j3ssie/osmedeus/libs"
	"testing"
)

func TestSlackWebHook(t *testing.T) {

	content := `
[jira-subversion-xss][Tentative-Medium] - https://142.104.128.133:443/plugins/servlet/svnwebclient/error.jsp?errormessage=%27%22%3E%3Cscript%3Ealert(document.domain)%3C%2Fscript%3E&description=test - out/142.104.128.133/jira-subversion-xss-b38aaad1bc0567262e48e374e902bda60d522f94\n
[jira-subversion-xss][Tentative-Medium] - http://58.49.154.201:8080/plugins/servlet/svnwebclient/error.jsp?errormessage=%27%22%3E%3Cscript%3Ealert(document.domain)%3C%2Fscript%3E&description=test - out/58.49.154.201/jira-subversion-xss-2e18d80689f97485af742d8c0368ffd2cdc78d7a
`
	err := SlackWebHook("https://hooks.slack.com/services/T01D9M3RSLA/B01CUUVD98X/xxxxxxxx", content)
	if err != nil {
		t.Error(err)
	}
}

func TestTeleSendMess(t *testing.T) {
	var opt libs.Options
	opt.Noti.TelegramChannel = "-1001166523435"
	opt.Noti.TelegramToken = "1288534500:xxxxxx"

	content := `
[jira-subversion-xss][Tentative-Medium] - https://142.104.128.133:443/plugins/servlet/svnwebclient/error.jsp?errormessage=%27%22%3E%3Cscript%3Ealert(document.domain)%3C%2Fscript%3E&description=test - out/142.104.128.133/jira-subversion-xss-b38aaad1bc0567262e48e374e902bda60d522f94\n
[jira-subversion-xss][Tentative-Medium] - http://58.49.154.201:8080/plugins/servlet/svnwebclient/error.jsp?errormessage=%27%22%3E%3Cscript%3Ealert(document.domain)%3C%2Fscript%3E&description=test - out/58.49.154.201/jira-subversion-xss-2e18d80689f97485af742d8c0368ffd2cdc78d7a
`
	err := TeleSendMess(opt, content, "general", true)
	if err != nil {
		t.Error(err)
	}
}

func TestTeleSendFile(t *testing.T) {
	var opt libs.Options
	opt.Noti.TelegramChannel = "-1001166523435"
	opt.Noti.TelegramToken = "1288534500:xxxx"

	err := TeleSendFile(opt, "/tmp/jtt/out/jaeles-summary.txt", "general")
	if err != nil {
		t.Error(err)
	}
}
