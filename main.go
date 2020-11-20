package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/spddl/go-twitch-ws"
)

type ShopClient struct {
	cache       *cache.Cache
	ProductList []Product
}

const shopUrl = "https://gronkh.shop"

type Rank struct {
	// Target is the word matched against.
	Target string

	// Index Of Pharse
	IndexOfPharse []int

	// Location of Target in original list
	OriginalIndex int

	Matches int
}

type Ranks []Rank

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile) // https://ispycode.com/GO/Logging/Setting-output-flags

	shopClient := ShopClient{
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}

	// // Lokal Cache
	// file, err := ioutil.ReadFile("./allProducts.json")
	// if err != nil {
	// 	log.Println(err)
	// }

	// data := map[string]Product{}
	// err = json.Unmarshal([]byte(file), &data)
	// if err != nil {
	// 	log.Println(err)
	// }

	// ProductList := []Product{}
	// for _, obj := range data {
	// 	ProductList = append(ProductList, obj)
	// }
	// shopClient.ProductList = ProductList

	//////////////////////////////////////////////

	collections, err := getCollections()
	if err != nil {
		log.Println(err)
	}

	allProducts := map[string]Product{}
	for name, url := range collections {
		productList, err := getProductList(name, url)
		if err != nil {
			log.Println(err)
		}

		for productName, product := range productList {
			thisProduct, exist := allProducts[productName]
			if exist {
				thisProduct.Collection = mergeArray(allProducts[productName].Collection, product.Collection)
				allProducts[productName] = thisProduct
			} else {
				allProducts[productName] = product
			}
		}
	}
	err = ioutil.WriteFile("./allProducts.json", prettyprint(allProducts), 0644) // save debug cache
	if err != nil {
		log.Println(err)
	}

	ProductList := []Product{}
	for _, obj := range allProducts {
		ProductList = append(ProductList, obj)
	}
	shopClient.ProductList = ProductList

	// log.Printf("%s\n", prettyprint(allProducts))

	/////////////////////////

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	bot, err := twitch.NewClient(&twitch.Client{
		Server:  "wss://irc-ws.chat.twitch.tv",
		User:    "gronkhshopbot",
		Oauth:   oauth,
		Channel: []string{"gronkhshopbot", "gronkhtv", "gronkh", "xpandorya"},
		Debug:   true,
	})
	if err != nil {
		panic(err)
	}

	bot.OnNoticeMessage = func(msg twitch.IRCMessage) {
		log.Printf("%s\n", msg)
	}

	bot.OnUnknownMessage = func(msg twitch.IRCMessage) {
		log.Printf("%s\n", msg)
	}
	bot.OnClearChatMessage = func(msg twitch.IRCMessage) {
		targetUser := string(msg.Params[1])
		if targetUser == "gronkhshopbot" {
			log.Printf("OnClearChatMessage: %s\n", msg)
			channel := string(msg.Params[0][1:])
			bot.Part([]string{channel})
		}
	}

	bot.OnWhisperMessage = func(msg twitch.IRCMessage) { // https://github.com/tmijs/tmi.js/issues/333#issuecomment-712474914
		go func(msg twitch.IRCMessage) { // Starte einen Thread um die anderen Nachrichten nicht zu blockieren
			line := strings.ToLower(string(msg.Params[1]))
			nick := string(msg.Tags["display-name"])
			toWhisper, _ := shopClient.SearchEngine(line, "", nick, true)
			bot.Whisper(nick, toWhisper)
		}(msg)
	}

	bot.OnPrivateMessage = func(msg twitch.IRCMessage) {
		line := strings.ToLower(string(msg.Params[1]))
		channel := string(msg.Params[0][1:])
		nick := string(msg.Tags["display-name"])

		if channel == "gronkhshopbot" {
			if strings.Contains(line, "droggelbecher") {
				log.Println(nick+":", line)
				bot.Join([]string{nick})
				return
			}
		}

		if strings.HasPrefix(line, "@gronkhshopbot") { // Wenn die Nachricht mit "@gronkhshopbot" anfängt
			go func(msg, channel, nick string) { // Starte einen Thread um die anderen Nachrichten nicht zu blockieren
				log.Printf("< #%s [%s] %s", channel, nick, strings.TrimSpace(msg))
				toChat, toWhisper := shopClient.SearchEngine(msg, channel, nick, false)
				log.Printf("> #%s %s", channel, toChat)
				bot.Say(channel, toChat, false)
				// bot.Whisper(nick, toWhisper)
				_ = toWhisper
			}(line[15:], channel, nick)

		} else if strings.HasPrefix(line, "gronkhshopbot") { // Wenn die Nachricht mit "gronkhshopbot" anfängt

			go func(msg, channel, nick string) { // Starte einen Thread um die anderen Nachrichten nicht zu blockieren
				log.Printf("< #%s [%s] %s", channel, nick, strings.TrimSpace(msg))
				toChat, toWhisper := shopClient.SearchEngine(msg, channel, nick, false)
				log.Printf("> #%s %s", channel, toChat)
				bot.Say(channel, toChat, false)
				// bot.Whisper(nick, toWhisper)
				_ = toWhisper
			}(line[14:], channel, nick)
		}
	}

	bot.OnConnect = func(status bool) {
		log.Println("Connected")
	}

	bot.Run()

	for { // ctrl - c
		<-interrupt
		bot.Close()
		os.Exit(0)
	}
}

func existArrayString(array []string, target string) (bool, int) {
	for index, value := range array {
		if value == target {
			return true, index
		}
	}
	return false, -1
}

func mergeArray(inputArray, addArray []string) []string {
	for _, inputValue := range inputArray {
		found, index := existArrayString(addArray, inputValue)
		if found {
			addArray[index] = addArray[len(addArray)-1] // https://yourbasic.org/golang/delete-element-slice/
			addArray[len(addArray)-1] = ""
			addArray = addArray[:len(addArray)-1]
		}
	}
	return append(inputArray, addArray...)
}
