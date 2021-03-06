package httpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

type Client struct {
	http.Client
	header map[string]string
}

func (c *Client) SetHeader(k, v string) *Client {
	c.header[k] = v
	return c
}

func (c *Client) Get(url string) ([]byte, error) {
	return c.request(http.MethodGet, url, nil)
}

func (c *Client) PostJson(url string, params map[string]interface{}) ([]byte, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	c.SetHeader("Content-Type", "application/json")
	return c.request(http.MethodPost, url, bytes.NewReader(body))
}

func (c *Client) PostRemoteFile() {

}

func (c *Client) PostFile(filed, filename, url string) ([]byte, error) {
	pr, pw := io.Pipe()
	bodyWriter := multipart.NewWriter(pw)
	go func() {
		fw, err := bodyWriter.CreateFormFile(filed, filename)
		if err != nil {
			log.Println(err)
		}
		fr, err := os.Open(filename)
		if err != nil {
			log.Println(err)
		}
		_, err = io.Copy(fw, fr)
		if err != nil {
			log.Println(err)
		}
		err = fr.Close()
		if err != nil {
			log.Println(err)
		}
		err = bodyWriter.Close()
		if err != nil {
			log.Println(err)
		}
		err = pw.Close()
		if err != nil {
			log.Println(err)
		}

	}()
	c.SetHeader("Content-Type", bodyWriter.FormDataContentType())
	c.SetHeader("Transfer-Encoding", "chunked")
	body := io.NopCloser(pr)
	return c.request(http.MethodPost, url, body)
}

func (c *Client) request(method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if len(c.header) > 0 {
		for k, v := range c.header {
			req.Header.Set(k, v)
		}
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respBody, nil

}
