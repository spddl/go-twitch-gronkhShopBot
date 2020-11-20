package main

import (
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func getData(url string) (map[string]string, error) {
	doc, err := getRequest(shopUrl + url)
	if err != nil {
		log.Println(err)
		return map[string]string{}, err
	}

	var variants = make(map[string]string)
	doc.Find("select option").Each(func(i int, s *goquery.Selection) {
		textil := strings.TrimSpace(s.Text())
		data := strings.Split(textil, " - ")
		variants[strings.Replace(data[0], " / ", "/", -1)] = data[1]
	})

	return variants, nil
}
