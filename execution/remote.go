package execution

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/parnurzeal/gorequest"
)

var remoteOptions libs.Options

// RemoteLogin get to get JWT to run
func RemoteLogin(username string, password string, URL string, options libs.Options) {
	u, err := url.Parse(URL)
	if err != nil {
		return
	}
	URL = fmt.Sprintf("%v://%v:%v", u.Scheme, u.Hostname(), u.Port())

	request := gorequest.New()
	_, body, _ := request.Post(fmt.Sprintf("%v/auth/login", URL)).
		Set("Content-Type", "application/json").
		Send(fmt.Sprintf(`{"username":"%v", "password":"%v"}`, username, password)).
		End()

	jsonParsed, _ := gabs.ParseJSON([]byte(body))

	token := strings.Trim(jsonParsed.S("token").String(), `"`)
	if token == "null" {
		utils.WarnF("Login fail at %v", URL)
	}
	options.Client.JWT = fmt.Sprintf("Osmedeus %v", token)
	options.Client.URL = URL
	utils.DebugF("Got token: %v at %v", options.Client.JWT, options.Client.URL)
	remoteOptions = options
}

// RemoteUpload upload data to remote url
func RemoteUpload(src string, options libs.Options) {
	if options.Client.JWT == "" || options.Client.URL == "" {
		utils.WarnF("JWT Token was not set")
		return
	}
	// keep \n not escape from JSON body
	data := strings.Join(utils.ReadingFileUnique(utils.NormalizePath(src)), "\\n")
	filename := path.Base(src)
	jsonBody := fmt.Sprintf(`{"data":"%v", "filename":"%v"}`, data, filename)

	request := gorequest.New()
	_, body, _ := request.Post(fmt.Sprintf("%v/api/upload/data", options.Client.URL)).
		Set("Content-Type", "application/json").
		Set("Authorization", options.Client.JWT).
		Send(jsonBody).
		End()
	// sample --> /tmp/data-osm-sample
	jsonParsed, _ := gabs.ParseJSON([]byte(body))
	remoteFile := strings.Trim(jsonParsed.S("content").String(), `"`)

	if remoteFile == "null" {
		utils.WarnF("Fail to Upload %v at %v", src, options.Client.URL)
	}
	utils.DebugF("Sucessfully uploaded: %v", remoteFile)
}

// RemoteExec run command on a remote client
func RemoteExec(command string, options libs.Options) {
	if options.Client.JWT == "" || options.Client.URL == "" {
		utils.WarnF("JWT Token was not set")
		return
	}

	// default master password is blank now
	mpassword := ""
	request := gorequest.New()
	_, body, _ := request.Post(fmt.Sprintf("%v/api/task/new", options.Client.URL)).
		Set("Content-Type", "application/json").
		Set("Authorization", options.Client.JWT).
		Send(fmt.Sprintf(`{"command": "%v","password": "%v"}`, command, mpassword)).
		End()

	jsonParsed, _ := gabs.ParseJSON([]byte(body))
	message := jsonParsed.S("content").String()
	if message == "null" {
		utils.WarnF("Run Command Fail at %v", options.Client.URL)
	}
	utils.DebugF("Sucessfully run Command: %v", command)

}

// RemoteExecSchedule run command on a remote client with schedule
func RemoteExecSchedule(command string, schedule string, options libs.Options) {
	if options.Client.JWT == "" || options.Client.URL == "" {
		utils.WarnF("JWT Token was not set")
		return
	}

	// default master password is blank now
	mpassword := ""
	request := gorequest.New()
	_, body, _ := request.Post(fmt.Sprintf("%v/api/task/new", options.Client.URL)).
		Set("Content-Type", "application/json").
		Set("Authorization", options.Client.JWT).
		Send(fmt.Sprintf(`{"command":"%v","seconds":%v,"password":"%v"}`, command, schedule, mpassword)).
		End()

	jsonParsed, _ := gabs.ParseJSON([]byte(body))
	message := jsonParsed.S("content").String()
	if message == "null" {
		utils.WarnF("Run Schedule Command Fail at %v", options.Client.URL)
	}
	utils.DebugF("Sucessfully run Schedule Command: %v", command)

	/*
		{
			"name": "example",
		    "command": "id",
		    "minutes": 1,
		    "password": "321"
		}
	*/
}
