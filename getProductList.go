package main

import (
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

type Product struct {
	Name        string
	Collection  []string
	Url         string
	Preis       string
	Sonderpreis string

	Ausverkauft bool
	SaleText    string
}

func getProductList(collection, url string) (map[string]Product, error) {
	data := map[string]Product{}

	doc, err := getRequest(shopUrl + url)
	if err != nil {
		return data, err
	}

	doc.Find("div.grid-product__content").Each(func(i int, item *goquery.Selection) {
		p := Product{}
		link, exists := item.Find("a.grid-product__link").Attr("href")
		if !exists {
			println("kein Link")
			return
		}
		link = strings.TrimPrefix(link, url)
		p.Url = link
		p.Name = item.Find("div.grid-product__title").Text()
		urlName := strings.TrimPrefix(url, "/collections/")

		collectionArray := strings.FieldsFunc(collection, func(c rune) bool { // Der Collection Name von der Webseite
			return !unicode.IsLetter(c) && !unicode.IsNumber(c)
		})
		if collection != urlName {
			urlNameArray := strings.FieldsFunc(urlName, func(c rune) bool { // Die Collection Url von der Webseite
				return !unicode.IsLetter(c) && !unicode.IsNumber(c)
			})
			p.Collection = mergeArray(collectionArray, urlNameArray)
		} else {
			p.Collection = collectionArray
		}

		// Der Tag oben Rechts am Bild
		soldOut := item.Find("div.grid-product__tag.grid-product__tag--sold-out").Text()
		if soldOut != "" {
			p.Ausverkauft = true
		}
		sale := item.Find("div.grid-product__tag.grid-product__tag--sale").Text()
		if sale != "" {
			p.SaleText = strings.TrimSpace(sale)
		}

		preis := item.Find("div.grid-product__price").Text()
		preis = strings.TrimSpace(preis)
		preis = strings.TrimPrefix(preis, "Von ")

		if strings.HasPrefix(preis, "Normaler Preis") {
			normalPreis := strings.TrimPrefix(preis, "Normaler Preis")
			preisArray := strings.Split(normalPreis, "Sonderpreis")
			p.Preis = strings.TrimSpace(preisArray[0])
			p.Sonderpreis = strings.TrimSpace(preisArray[1])
		} else {
			p.Preis = strings.TrimSpace(preis)
		}
		data[p.Name] = p
	})
	return data, nil
}
