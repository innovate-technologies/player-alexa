package main

import (
	"crypto"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	alexa "github.com/mikeflynn/go-alexa/skillserver"

	"bytes"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/labstack/echo"
)

func main() {
	e := echo.New()

	e.Use(verifySignature)
	e.Use(verifyTimeStamp)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Alexa!")
	})
	e.POST("/alexa/:username", handleAlexa)
	e.Logger.Fatal(e.Start(":80"))
}

func handleAlexa(c echo.Context) error {
	req := alexa.EchoRequest{}
	user := c.Param("username")
	config, err := getConfigForUser(user)

	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, "An error happened")
	}
	c.Bind(&req)

	language := "en" // beta
	var thisLanguageEntry languageEntry
	for _, entry := range config.LanguageEntries {
		if entry.Language == language {
			thisLanguageEntry = entry
		}
	}

	resp := alexa.NewEchoResponse()
	reqType := req.GetRequestType()

	if reqType == "LaunchRequest" {
		// say hello to a new user
		c.JSON(http.StatusOK, resp.EndSession(false).OutputSpeech(thisLanguageEntry.Intro))
	}

	if reqType != "IntentRequest" {
		// nothing we can do here
		return c.JSON(http.StatusOK, resp)
	}

	intent := req.GetIntentName()

	if intent == "NowPlaying" {
		song, err := getNowPlaying(user)
		if err != nil {
			return c.JSON(http.StatusOK, resp.OutputSpeech("I'm sorry, I have no idea which song is playing right now"))
		}
		return c.JSON(http.StatusOK, resp.OutputSpeech(fmt.Sprintf("Currently, %s is playing", song)))
	}

	if intent == "Play" || intent == "AMAZON.ResumeIntent" || intent == "AMAZON.NextIntent" || intent == "AMAZON.PreviousIntent" || intent == "AMAZON.RepeatIntent" || intent == "AMAZON.StartOverIntent" {
		audioResonse := NewAudioStartResponse()
		audioResonse.Response.Directives = []AudioDirective{
			AudioDirective{
				Type:         "AudioPlayer.Play",
				PlayBehavior: "REPLACE_ALL",
				AudioItem: AudoItem{
					Stream: Stream{
						URL:   config.TuneInURL,
						Token: "0",
						ExpectedPreviousToken: nil,
						OffsetInMilliseconds:  0,
					},
				},
			},
		}

		return c.JSON(http.StatusOK, audioResonse)
	}

	if intent == "AMAZON.PauseIntent" || intent == "AMAZON.StopIntent" || intent == "AMAZON.CancelIntent" {
		audioResonse := NewAudioStartResponse()
		audioResonse.Response.Directives = []AudioDirective{
			AudioDirective{
				Type: "AudioPlayer.Stop",
			},
		}
		return c.JSON(http.StatusOK, audioResonse)
	}

	if intent == "AMAZON.HelpIntent" || intent == "AMAZON.FallbackIntent" {
		return c.JSON(http.StatusOK, resp.OutputSpeech(thisLanguageEntry.Help))
	}

	// general fallback
	return c.JSON(http.StatusOK, resp)
}

func verifySignature(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		certURL := c.Request().Header.Get("SignatureCertChainUrl")
		if !isValidCertURL(certURL) {
			return c.String(http.StatusBadRequest, "Could not verify request")
		}
		publicKey, err := getCert(certURL)
		if err != nil {
			return c.String(http.StatusBadRequest, "Could not verify request")
		}
		signature, _ := base64.StdEncoding.DecodeString(c.Request().Header.Get("Signature"))

		var buf bytes.Buffer
		hash := sha1.New()
		_, err = io.Copy(hash, io.TeeReader(c.Request().Body, &buf))
		if err != nil {
			return c.String(http.StatusBadRequest, "Could not verify request")
		}

		c.Request().Body = ioutil.NopCloser(&buf)

		err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA1, hash.Sum(nil), signature)
		if err != nil {
			return c.String(http.StatusBadRequest, "Could not verify request")
		}

		return next(c)
	}
}

func verifyTimeStamp(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		body := alexa.EchoRequest{}

		var bodyBytes []byte
		if c.Request().Body != nil {
			bodyBytes, _ = ioutil.ReadAll(c.Request().Body)
		}

		// Restore the io.ReadCloser to its original state
		c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		json.Unmarshal(bodyBytes, &body)

		reqTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", body.Request.Timestamp)
		if time.Since(reqTimestamp) < time.Duration(150)*time.Second {
			return next(c)
		}

		return c.String(http.StatusBadRequest, "Could not verify request")
	}
}
