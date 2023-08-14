package sbervoice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"github.com/google/uuid"
	"crypto/tls"
	"io"
)


type Auth struct{
	AccessToken string `json:"access_token"`
	ExpiresAt    int64 `json:"expires_at"`
}


type SberSpech struct {
	Result   []string `json:"result"`
	Emotions []struct {
		Negative float64 `json:"negative"`
		Neutral  float64 `json:"neutral"`
		Positive float64 `json:"positive"`
	} `json:"emotions"`
	Status int `json:"status"`
}

func Voice2Text(file string, token string) (string, error) {
	// get api key 
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	var data = strings.NewReader(`scope=SALUTE_SPEECH_PERS`)
	req, err := http.NewRequest("POST", "https://ngw.devices.sberbank.ru:9443/api/v2/oauth", data)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("RqUID", uuid.New().String())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+ token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var ctask Auth

	jsonErr := json.Unmarshal(responseData, &ctask)
	if jsonErr != nil {
		log.Fatal(jsonErr)
		return "", fmt.Errorf("error parce json")
	}

	// get voice spech 
	fstr, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer fstr.Close()

	client = &http.Client{Transport: tr}
	req, err = http.NewRequest("POST", "https://smartspeech.sber.ru/rest/v1/speech:recognize", fstr)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer " + ctask.AccessToken)
	req.Header.Set("Content-Type", "audio/ogg;codecs=opus")
	resp, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("from api result: %s\n", bodyText)

	var resVoice SberSpech

	jsonErr = json.Unmarshal(bodyText, &resVoice)
	if jsonErr != nil {
		log.Fatal(jsonErr)
		return "", fmt.Errorf("error parce json")
	}

	if resVoice.Result[0] != "" {
		return resVoice.Result[0], nil
	} else {
		return "", fmt.Errorf("error from api")
	}

}
