package main

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func getRequest(url string) (*goquery.Document, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &goquery.Document{}, err
	}

	req.Header.Set("User-Agent", "GronkhShopBot/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return &goquery.Document{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &goquery.Document{}, err
	}

	r := bytes.NewReader([]byte(body))
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return &goquery.Document{}, err
	}
	return doc, nil
}
