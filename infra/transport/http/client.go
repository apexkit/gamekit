package http

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

type HttpClient struct {
	client *http.Client
}

func InitHttpClient(timeOut time.Duration) *HttpClient {
	client := &HttpClient{}
	client.Init(timeOut)
	return client
}

func (client *HttpClient) Init(timeOut time.Duration) {
	roots := x509.NewCertPool()

	ok := roots.AppendCertsFromPEM(GetCacert())
	if !ok {
		zap.L().Info("failed to parse root certificate")
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: roots},
	}

	client.client = &http.Client{Timeout: timeOut, Transport: tr}
}

const maxBytes int64 = 10 * 1024 * 1024

func (client *HttpClient) Get(url string, values url.Values) ([]byte, error) {
	resp, err := client.client.Get(url + "?" + values.Encode())

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := &io.LimitedReader{R: resp.Body, N: maxBytes}
	bytes, err := io.ReadAll(r)
	return bytes, err
}

func (client *HttpClient) Post(url string, contentType string, data []byte) ([]byte, error) {
	body := bytes.NewReader(data)

	resp, err := client.client.Post(url, contentType, body)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := &io.LimitedReader{R: resp.Body, N: maxBytes}
	bytes, err := io.ReadAll(r)
	return bytes, err
}
