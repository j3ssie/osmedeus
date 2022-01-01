package execution

import (
    "bufio"
    "crypto/tls"
    "encoding/base64"
    "fmt"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"
    "io/ioutil"
    "net/http"
    "net/url"
    "strconv"
    "strings"
    "time"

    "github.com/go-resty/resty/v2"
    "github.com/sirupsen/logrus"
)

func SendGET(token string, url string) (libs.Response, error) {
    client := BuildClient(token, 2)
    res, err := JustSend(url, client)
    return res, err
}

// BuildClient build base HTTP client
func BuildClient(token string, retry int) *resty.Client {
    headers := map[string]string{
        "UserAgent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.132 Safari/537.36",
        "Content-Type": "application/json",
    }
    if token != "" {
        headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
    }

    timeout := 15

    //if len(options.Headers) > 0 {
    //	for _, head := range options.Headers {
    //		if strings.Contains(head, ":") {
    //			data := strings.Split(head, ":")
    //			if len(data) < 2 {
    //				continue
    //			}
    //			headers[data[0]] = strings.Join(data[1:], "")
    //		}
    //	}
    //}

    // disable log when retry
    logger := logrus.New()
    logger.Out = ioutil.Discard

    client := resty.New()
    client.SetLogger(logger)
    client.SetTransport(&http.Transport{
        MaxIdleConns:          100,
        MaxConnsPerHost:       1000,
        IdleConnTimeout:       time.Duration(timeout) * time.Second,
        ExpectContinueTimeout: time.Duration(timeout) * time.Second,
        ResponseHeaderTimeout: time.Duration(timeout) * time.Second,
        TLSHandshakeTimeout:   time.Duration(timeout) * time.Second,
        DisableCompression:    true,
        TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
    })

    client.SetHeaders(headers)
    client.SetCloseConnection(true)
    client.SetRetryCount(retry)
    client.SetTimeout(time.Duration(timeout) * time.Second)
    client.SetRetryWaitTime(time.Duration(timeout/2) * time.Second)
    client.SetRetryMaxWaitTime(time.Duration(timeout) * time.Second)
    return client
}

// JustSend just sending request
func JustSend(url string, client *resty.Client) (res libs.Response, err error) {
    method := "GET"
    // redirect policy

    var resp *resty.Response
    // really sending things here
    method = strings.ToLower(strings.TrimSpace(method))
    switch method {
    case "get":
        resp, err = client.R().
            Get(url)
        break
    case "post":
        resp, err = client.R().
            Post(url)
        break
    }

    // in case we want to get redirect stuff
    if res.StatusCode != 0 {
        return res, nil
    }

    if err != nil || resp == nil {
        utils.ErrorF("%v %v", url, err)
        return libs.Response{}, err
    }

    return ParseResponse(*resp), nil
}

// ParseResponse field to Response
func ParseResponse(resp resty.Response) (res libs.Response) {
    // var res libs.Response
    resLength := len(string(resp.Body()))
    // format the headers
    var resHeaders []map[string]string
    for k, v := range resp.RawResponse.Header {
        if k == "Content-Type" {
            res.ContentType = strings.Join(v[:], "")
        }
        if k == "Location" {
            res.Location = strings.Join(v[:], "")
        }
        element := make(map[string]string)
        element[k] = strings.Join(v[:], "")
        resLength += len(fmt.Sprintf("%s: %s\n", k, strings.Join(v[:], "")))
        resHeaders = append(resHeaders, element)
    }
    // response time in second
    resTime := float64(resp.Time()) / float64(time.Second)
    resHeaders = append(resHeaders,
        map[string]string{"Total Length": strconv.Itoa(resLength)},
        map[string]string{"Response Time": fmt.Sprintf("%f", resTime)},
    )

    // set some variable
    res.Headers = resHeaders
    res.StatusCode = resp.StatusCode()
    res.Status = fmt.Sprintf("%v %v", resp.Status(), resp.RawResponse.Proto)
    res.Body = string(resp.Body())
    res.ResponseTime = resTime
    res.Length = resLength
    // beautify
    res.Beautify = BeautifyResponse(res)
    res.BeautifyHeader = BeautifyHeaders(res)
    return res
}

// BeautifyRequest beautify request
func BeautifyRequest(req libs.Request) string {
    var beautifyReq string
    // hardcoded HTTP/1.1 for now
    beautifyReq += fmt.Sprintf("%v %v HTTP/1.1\n", req.Method, req.URL)

    for _, header := range req.Headers {
        for key, value := range header {
            if key != "" && value != "" {
                beautifyReq += fmt.Sprintf("%v: %v\n", key, value)
            }
        }
    }
    if req.Body != "" {
        beautifyReq += fmt.Sprintf("\n%v\n", req.Body)
    }
    return beautifyReq
}

// BeautifyHeaders beautify headers
func BeautifyHeaders(res libs.Response) string {
    beautifyHeader := fmt.Sprintf("< %v \n", res.Status)
    for _, header := range res.Headers {
        for key, value := range header {
            beautifyHeader += fmt.Sprintf("< %v: %v\n", key, value)
        }
    }
    return beautifyHeader
}

// BeautifyResponse beautify response
func BeautifyResponse(res libs.Response) string {
    var beautifyRes string
    beautifyRes += fmt.Sprintf("%v \n", res.Status)

    for _, header := range res.Headers {
        for key, value := range header {
            beautifyRes += fmt.Sprintf("%v: %v\n", key, value)
        }
    }

    beautifyRes += fmt.Sprintf("\n%v\n", res.Body)
    return beautifyRes
}

// ParseBurpRequest parse burp style request
func ParseBurpRequest(raw string) string {
    rawDecoded, err := base64.StdEncoding.DecodeString(raw)
    if err != nil {
        return ""
    }

    var realReq libs.Request
    reader := bufio.NewReader(strings.NewReader(string(rawDecoded)))
    parsedReq, err := http.ReadRequest(reader)
    if err != nil {
        return ""
    }
    realReq.Method = parsedReq.Method
    // URL part
    if parsedReq.URL.Host == "" {
        realReq.Host = parsedReq.Host
        parsedReq.URL.Host = parsedReq.Host
    }
    if parsedReq.URL.Scheme == "" {
        if parsedReq.Referer() == "" {
            realReq.Scheme = "https"
            parsedReq.URL.Scheme = "https"
        } else {
            u, err := url.Parse(parsedReq.Referer())
            if err == nil {
                realReq.Scheme = u.Scheme
                parsedReq.URL.Scheme = u.Scheme
            }
        }
    }
    realReq.URL = parsedReq.URL.String()
    realReq.Path = parsedReq.RequestURI

    return realReq.URL
}
