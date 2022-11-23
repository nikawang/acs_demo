package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ACSMail struct {
	Content    Content    `json:"content"`
	Sender     string     `json:"sender"`
	Importance string     `json:"importance"`
	Recipients Recipients `json:"recipients"`
}
type Content struct {
	Subject   string `json:"subject"`
	PlainText string `json:"plainText"`
	HTML      string `json:"html"`
}
type To struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}
type Cc struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}
type Recipients struct {
	To []To `json:"to"`
	Cc []Cc `json:"CC"`
}

func getAuth(method string, body string, secret string, urlPath string, hostStr string, dateStr string) (string, string) {

	digestBody := sha256.New()
	digestBody.Write([]byte(body))
	digestBodyContent := base64.StdEncoding.EncodeToString(digestBody.Sum(nil))

	signingString := method + "\n" + urlPath + "\n" + dateStr + ";" + hostStr + ";" + digestBodyContent

	key, _ := base64.StdEncoding.DecodeString(secret)
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(signingString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	authorization := "HMAC-SHA256 SignedHeaders=date;host;x-ms-content-sha256&Signature=" + signature

	return digestBodyContent, authorization
}

func main() {

	secret := "2DGffI36Z0B4UjdjflwrRLpaGflNWLpWD/MiNoeNmLxaqdGHrzj0VvzJlq+45ijEvqo0UhdB5RbqKWHedPwS1g=="
	endpoint := "https://lilith-acs-poc.communication.azure.com"
	hostStr := strings.Split(endpoint, "//")[1]

	sendPath := "/emails:send?api-version=2021-10-01-preview"

	// request path
	url := endpoint + sendPath
	uuidStr := uuid.New().String()
	fmt.Println(uuidStr)
	// date
	loc, _ := time.LoadLocation("GMT")
	date := time.Now().In(loc)
	layout := "Mon, 02 Jan 2006 15:04:05 GMT"
	dateFormat := date.Format(layout)

	// body
	acsMail := &ACSMail{
		Content: Content{
			Subject:   "An exciting offer especially for you!",
			PlainText: "This exciting offer was created especially for you, our most loyal customer.",
			HTML:      "<html><head><title>Exciting offer!</title></head><body><h1>This exciting offer was created especially for you, our most loyal customer.</h1></body></html>",
		},
		Sender:     "DoNotReply@email.farlightgames.com",
		Importance: "normal",
		Recipients: Recipients{
			To: []To{
				{
					Email:       "danielwzhg@gmail.com",
					DisplayName: "Daniel Wang",
				},
			},
			Cc: []Cc{
				{
					Email:       "zhaw@microsoft.com",
					DisplayName: "Zhanggui Wang",
				},
			},
		},
	}
	requestBytes, _ := json.Marshal(acsMail)
	emailBody := string(requestBytes)
	digestEmailBodyContent, sendAuthorization := getAuth("POST", emailBody, secret, sendPath, hostStr, dateFormat)

	client := &http.Client{}

	sendReq, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(emailBody)))
	if err != nil {
		log.Fatalln(err)
	}

	sendReq.Header.Set("x-ms-content-sha256", digestEmailBodyContent)
	sendReq.Header.Set("Authorization", sendAuthorization)
	sendReq.Header.Set("Date", dateFormat)
	sendReq.Header.Set("repeatability-request-id", uuidStr)
	sendReq.Header.Set("repeatability-first-sent", dateFormat)
	sendReq.Header.Set("Content-Type", "application/json")

	requestDump, err := httputil.DumpRequest(sendReq, true)

	if err == nil {
		log.Println("Send Email Request:", string(requestDump))
	}

	sendResp, _ := client.Do(sendReq)
	body, err := ioutil.ReadAll(sendResp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Send Email Response header:")
	messageId := ""
	for k, v := range sendResp.Header {
		fmt.Println(k, ":", v)
		if strings.EqualFold("X-Ms-Request-Id", k) {
			messageId = v[0]
		}

	}

	log.Println("Send Email Response:", string(body))
	log.Println("=====================")
	// =========Get Send Status
	getStatusPath := "/emails/" + messageId + "/status?api-version=2021-10-01-preview"
	digestGetStatusContent, getAuthorization := getAuth("GET", "", secret, getStatusPath, hostStr, dateFormat)
	getStatusUrl := endpoint + getStatusPath
	// fmt.Println(getUrl)

	getSendStatus, _ := http.NewRequest("GET", getStatusUrl, bytes.NewBuffer([]byte("")))
	getSendStatus.Header.Set("Authorization", getAuthorization)
	getSendStatus.Header.Set("Date", dateFormat)
	getSendStatus.Header.Set("x-ms-content-sha256", digestGetStatusContent)

	getStatusResp, _ := client.Do(getSendStatus)

	getStatusBody, err1 := ioutil.ReadAll(getStatusResp.Body)
	if err1 != nil {
		log.Fatalln(err1)
	}

	log.Println("Get status Response header:")

	for k, v := range getStatusResp.Header {
		fmt.Println(k, ":", v)
	}

	log.Println("Get status Response:", string(getStatusBody))
}
