package main

import (
	"crypto"
	"io/ioutil"
	"net/http"
	"time"

	alexa "github.com/mikeflynn/go-alexa/skillserver"

	"bytes"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
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
	c.Bind(&req)

	resp := alexa.NewEchoResponse()

	reqType := req.GetRequestType()

	if reqType == "LaunchRequest" {
		// say hello to a new user
		c.JSON(http.StatusOK, resp.OutputSpeech("Welcome to The super secret innovate player. Say Play to play some innovation!"))
	}

	if reqType != "IntentRequest" {
		// nothing we can do here
		return c.JSON(http.StatusOK, resp)
	}

	intent := req.GetIntentName()

	if intent == "NowPlaying" {
		return c.JSON(http.StatusOK, resp.OutputSpeech("I don't know... ask Google. Okay Google, what song is this"))
	}

	if intent == "Play" || intent == "AMAZON.ResumeIntent" {
		audioResonse := NewAudioStartResponse()
		audioResonse.Response.Directives = []AudioDirective{
			AudioDirective{
				Type:         "AudioPlayer.Play",
				PlayBehavior: "REPLACE_ALL",
				AudioItem: AudoItem{
					Stream: Stream{
						URL:   "https://opencast.radioca.st/streams/320kbps",
						Token: "0",
						ExpectedPreviousToken: nil,
						OffsetInMilliseconds:  0,
					},
				},
			},
		}

		return c.JSON(http.StatusOK, audioResonse)
	}

	if intent == "AMAZON.PauseIntent" {
		audioResonse := NewAudioStartResponse()
		audioResonse.Response.Directives = []AudioDirective{
			AudioDirective{
				Type: "AudioPlayer.Stop",
			},
		}
		return c.JSON(http.StatusOK, audioResonse)
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
		c.Bind(&body)

		reqTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", body.Request.Timestamp)
		if time.Since(reqTimestamp) < time.Duration(150)*time.Second {
			return next(c)
		}

		return c.String(http.StatusBadRequest, "Could not verify request")
	}
}
