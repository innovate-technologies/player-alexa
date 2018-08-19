package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	resty "gopkg.in/resty.v0"
)

type nowplaying struct {
	ID         string    `json:"_id"`
	Username   string    `json:"username"`
	Song       string    `json:"song"`
	Artist     string    `json:"artist"`
	Cover      string    `json:"cover"`
	Wiki       string    `json:"wiki"`
	Buy        string    `json:"buy"`
	V          int       `json:"__v"`
	CreateDate time.Time `json:"createDate"`
}

type config struct {
	ID              string          `json:"_id"`
	Logo            string          `json:"logo"`
	Username        string          `json:"username"`
	V               int             `json:"__v"`
	Name            string          `json:"name"`
	LanguageEntries []languageEntry `json:"languageEntries"`
	Status          string          `json:"status"`
	TuneInURL       string          `json:"tuneInURL"`
}

type languageEntry struct {
	ID               string        `json:"_id"`
	Language         string        `json:"language"`
	ShortDescription string        `json:"shortDescription"`
	Description      string        `json:"description"`
	Help             string        `json:"help"`
	Intro            string        `json:"intro"`
	InvocationName   string        `json:"invocationName"`
	Keywords         []interface{} `json:"keywords"`
}

func getNowPlaying(username string) (string, error) {
	resp, _ := resty.R().Get("https://itframe.innovatete.ch/nowplaying/" + username)

	if resp.StatusCode() != http.StatusOK {
		return "", errors.New("Not 200 ok")
	}

	data := []nowplaying{}
	json.Unmarshal(resp.Body(), &data)

	if len(data) < 1 {
		return "", errors.New("No records")
	}

	current := data[0]
	if current.Song != "" && current.Artist != "" {
		return fmt.Sprintf("%s by %s", current.Song, current.Artist), nil
	} else if current.Song != "" {

		return current.Song, nil
	}

	return "", errors.New("No song meta")
}

func getITFrameConfig(username string) (config, error) {
	resp, _ := resty.R().Get("https://itframe.innovatete.ch/alexa/" + username)

	if resp.StatusCode() != http.StatusOK {
		return config{}, errors.New("Not 200 ok")
	}

	data := config{}
	json.Unmarshal(resp.Body(), &data)

	return data, nil
}

func getTuneIn(username string) (string, error) {
	resp, _ := resty.R().Get("https://itframe.innovatete.ch/tunein/" + username)

	if resp.StatusCode() != http.StatusOK {
		return "", errors.New("Not 200 ok")
	}

	data := map[string]string{}
	json.Unmarshal(resp.Body(), &data)

	return data["streamUrl"], nil
}
