package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func getCollections() (map[string]string, error) {
	data := make(map[string]string)

	doc, err := getRequest(shopUrl + "/collections")
	if err != nil {
		return data, err
	}

	doc.Find("div.collections-list div.grid__item").Each(func(i int, item *goquery.Selection) {
		name := item.Find("div.skrim__title").Text()
		name = strings.TrimSpace(name)
		name = strings.ToLower(name)

		a := item.Find("a.skrim__link")
		link, exists := a.Attr("href")
		if exists {
			data[name] = link
		} else {
			print("keinen Link zur Collection")
		}
	})
	return data, nil
}
