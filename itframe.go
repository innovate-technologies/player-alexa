package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"

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

func getNowPlaying(username string) (string, error) {
	resp, _ := resty.R().Get("https://itframe.innovatete.ch/nowplaying/" + username)

	spew.Dump(resp)

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
