package webhook

import (
	"bytes"
	"log"
	"net/http"
	"os"

	"github.com/larderdev/kubewise/kwrelease"
	"github.com/larderdev/kubewise/presenters"
)

type Webhook struct {
	URL    string
	Method string
}

func (w *Webhook) Init() {
	var url string

	method := "POST"
	if value, ok := os.LookupEnv("KW_WEBHOOK_METHOD"); ok {
		method = value
	}

	if value, ok := os.LookupEnv("KW_WEBHOOK_URL"); ok {
		url = value
	} else {
		log.Fatalln("Missing environment variable KW_WEBHOOK_URL")
	}

	w.Method = method
	w.URL = url
}

func (w *Webhook) HandleEvent(releaseEvent *kwrelease.Event) {
	jsonStr, jsonErr := presenters.ReleaseEventToJSON(releaseEvent)

	if jsonErr != nil {
		// The message should never contain any sensitive data so it's safe to log this err.
		log.Println("Error encoding JSON in webhook event", jsonErr)
		return
	}

	client := &http.Client{}
	req, reqErr := http.NewRequest(w.Method, w.URL, bytes.NewBuffer(jsonStr))

	if reqErr != nil {
		// Safe enough to print this err because any authentication header has not yet been attached.
		// There could be auth tokens in the query string of course. This will need handling later.
		log.Println("Error forming request in webhook event", reqErr)
		return
	}

	req.Header.Add("Content-Type", "application/json")

	if value, ok := os.LookupEnv("KW_WEBHOOK_AUTH_TOKEN"); ok {
		req.Header.Add("Authorization", "Bearer "+value)
	}

	resp, respErr := client.Do(req)

	if respErr != nil {
		// Do NOT print this error. Could leak Authorization header.
		log.Println("Error handling response to webhook event. Response status:", resp.Status)
		return
	}

	log.Println("Successful response received from", w.Method, w.URL, ":", resp.StatusCode)
}
