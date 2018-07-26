package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func isValidCertURL(in string) bool {
	link, err := url.Parse(in)
	if err != nil {
		return false
	}

	if link.Scheme != "https" {
		return false
	}

	if !strings.HasPrefix(link.Path, "/echo.api/") {
		return false
	}

	if link.Host != "s3.amazonaws.com" && link.Host != "s3.amazonaws.com:443" {
		return false
	}

	return true
}

func getCert(in string) (*rsa.PublicKey, error) {
	req, err := http.Get(in)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	certContents, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(certContents)
	if block == nil {
		return nil, errors.New("Invalid cert")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	if time.Now().Unix() < cert.NotBefore.Unix() || time.Now().Unix() > cert.NotAfter.Unix() {
		return nil, errors.New("Cert expired")
	}

	foundName := false
	for _, altName := range cert.Subject.Names {
		if altName.Value == "echo-api.amazon.com" {
			foundName = true
		}
	}

	if !foundName {
		return nil, errors.New("Not an echo api cert")
	}

	return cert.PublicKey.(*rsa.PublicKey), nil
}
