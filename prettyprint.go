package main

import (
	"encoding/json"
	"log"
)

// log.Printf("%s\n", prettyprint(data))
func prettyprint(data interface{}) []byte {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Println("error:", err)
	}
	return b
}
