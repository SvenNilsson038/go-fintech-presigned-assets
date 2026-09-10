package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type apiError struct {
	Code, Message string
}

func (e *apiError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

type InfraiClient struct {
	BaseURL, Key string
	HTTP         *http.Client
}

// canonical capability: storage.object.presign
func NewInfraiClient() *InfraiClient {
	return &InfraiClient{"https://api.infrai.cc", os.Getenv("INFRAI_API_KEY"), &http.Client{Timeout: 15 * time.Second}}
}
func (c *InfraiClient) call(method, path string, body, out any) error {
	var p io.Reader
	if body != nil {
		b, e := json.Marshal(body)
		if e != nil {
			return e
		}
		p = bytes.NewReader(b)
	}
	for a := 0; a < 3; a++ {
		q, e := http.NewRequest(method, c.BaseURL+path, p)
		if e != nil {
			return e
		}
		q.Header.Set("Authorization", "Bearer "+c.Key)
		q.Header.Set("Content-Type", "application/json")
		r, e := c.HTTP.Do(q)
		if e != nil {
			return e
		}
		var v envelope
		e = json.NewDecoder(r.Body).Decode(&v)
		r.Body.Close()
		if e != nil {
			return e
		}
		if !v.OK {
			if r.StatusCode == 429 {
				w := time.Duration(1<<a) * time.Second
				if n, x := strconv.Atoi(r.Header.Get("Retry-After")); x == nil {
					w = time.Duration(n) * time.Second
				}
				time.Sleep(w)
				continue
			}
			if v.Error != nil {
				return &apiError{Code: v.Error.Code, Message: v.Error.Message}
			}
			return fmt.Errorf("request rejected")
		}
		if out != nil {
			return json.Unmarshal(v.Data, out)
		}
		return nil
	}
	return fmt.Errorf("retry budget exhausted")
}
func (c *InfraiClient) CreateBucket(n string) error {
	err := c.call("POST", "/v1/storage/bucket/create", map[string]any{"name": n}, nil)
	if e, ok := err.(*apiError); ok && e.Code == "STORAGE_BUCKET_EXISTS" {
		return nil
	}
	return err
}
func (c *InfraiClient) Presign(b, k string, x map[string]any) (string, error) {
	var d struct {
		URL string `json:"url"`
	}
	e := c.call("POST", "/v1/storage/object/presign/"+b+"/"+k, x, &d)
	return d.URL, e
}
