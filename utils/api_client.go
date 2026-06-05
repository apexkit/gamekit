package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

type ApiClient struct {
	httpClient *http.Client
}

func NewApiClient(maxIdleConns, maxIdleConnsPerHost int, timeout time.Duration) *ApiClient {
	client := &ApiClient{}
	client.httpClient = &http.Client{
		Transport: &http.Transport{
			// 保留默认的 Dialer 配置
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          maxIdleConns,
			MaxIdleConnsPerHost:   maxIdleConnsPerHost,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: timeout,
	}
	return client
}

func (c *ApiClient) Do(appId, accessSecret, apiURL string, data []byte) ([]byte, error) {
	nonce := generateNonce(8)
	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)
	signature := generateSign(appId, accessSecret, nonce, timestamp, string(data))

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("Api Callback Error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-appid", appId)
	req.Header.Set("x-sign", signature)
	req.Header.Set("x-nonce", nonce)
	req.Header.Set("x-timestamp", timestampStr)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Api Callback Error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Api Callback Error reading response body: %w", err)
	}

	return body, nil
}
