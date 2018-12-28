package alimns

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// PeekMessage 查看消息
type PeekMessage struct {
	MessageID        string `xml:"MessageId"`
	MessageBody      string `xml:"MessageBody"`
	MessageBodyMD5   string `xml:"MessageBodyMD5"`
	EnqueueTime      int64  `xml:"EnqueueTime"`
	FirstDequeueTime int64  `xml:"FirstDequeueTime"`
	DequeueCount     int    `xml:"DequeueCount"`
	Priority         int    `xml:"Priority"`
}

// PeekMessageResponse 查看消息回复
type PeekMessageResponse struct {
	PeekMessage
}

// PeekMessage 查看消息
func (c *Client) PeekMessage(name string) (*PeekMessageResponse, error) {
	requestLine := fmt.Sprintf(mnsPeekMessage, name)
	req, err := http.NewRequest(http.MethodGet, c.endpoint+requestLine, nil)
	if err != nil {
		return nil, err
	}

	c.finalizeHeader(req, nil)

	contextLogger.
		WithField("method", req.Method).
		WithField("url", req.URL.String()).
		Info("查看消息请求")

	ctx, cancel := context.WithCancel(context.TODO())
	_ = time.AfterFunc(time.Second*timeout, func() {
		cancel()
	})
	req = req.WithContext(ctx)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	contextLogger.
		WithField("status", resp.Status).
		WithField("body", string(body)).
		WithField("url", req.URL.String()).
		Info("查看消息回复")

	switch resp.StatusCode {
	case http.StatusOK:
		var peekMessageResponse PeekMessageResponse
		if err := xml.Unmarshal(body, &peekMessageResponse); err != nil {
			return nil, err
		}
		return &peekMessageResponse, nil
	default:
		var respErr RespErr
		if err := xml.Unmarshal(body, &respErr); err != nil {
			return nil, err
		}
		return nil, errors.New(respErr.Code)
	}
}
