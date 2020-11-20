package main

import (
	"strings"
)

func createDescription(item Product) string {
	var result []string
	result = append(result, "\""+strings.ReplaceAll(item.Name, " - ", "-")+"\"")
	if item.SaleText != "" {
		saletext := strings.Replace(item.SaleText, "Jetzt sparen €", "Spare ", 1)
		saletext = strings.TrimSuffix(saletext, ",00")
		result = append(result, saletext+"€")
	}

	result = append(result, "für nur")
	var preis string
	if item.Sonderpreis != "" {
		preis = item.Sonderpreis
	} else {
		preis = item.Preis
	}
	preis = strings.TrimPrefix(preis, "€")
	preis = strings.TrimSuffix(preis, ",00")
	result = append(result, preis+"€")
	return strings.Join(result, " ")
}
