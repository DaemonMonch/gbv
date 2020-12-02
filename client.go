package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"net/url"
)

type Client struct {
	MetricsUrl string
	c          *http.Client
}

func NewClient(addr, path string) *Client {
	c := &http.Client{Timeout: 2 * time.Second}
	return &Client{fmt.Sprintf("http://%s%s", addr, path), c}
}

func (client *Client) Names() (*MeterNames, error) {
	resp, err := http.Get(client.MetricsUrl)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var names MeterNames
	if err = decoder.Decode(&names); err != nil {
		log.Println(err)
		return nil, err
	}
	return &names, nil
}

func (client *Client) Meter(ctx context.Context, meterName string, tags ...string) (*Meter, error) {
	var tagParams string
	if len(tags) > 0 {
		sb := bytes.NewBufferString(fmt.Sprintf("tag=%s&", url.QueryEscape(tags[0])))
		for i := 1; i < len(tags); i++ {
			sb.WriteString(fmt.Sprintf("tag=%s&", url.QueryEscape(tags[i])))
		}
		tagParams = sb.String()
	}
	
	url := fmt.Sprintf("%s/%s?%s", client.MetricsUrl, meterName, tagParams)
	log.Println("url=", url)
	request, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	resp, err := client.c.Do(request)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var meter Meter
	if err = decoder.Decode(&meter); err != nil {
		log.Println(err)
		return nil, err
	}

	return &meter, nil
}
