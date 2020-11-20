package main

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/patrickmn/go-cache"
)

const MaxChatLine = 240

var fullOfStock = []string{"Alle Größen vorhanden", "Lieferbestände voll", "Alle Größen auf Lager"}

func (sc *ShopClient) SearchEngine(search, channel, nick string, whisper bool) (string, string) {
	Ranking := Ranks{}
	search = strings.ToLower(search)
	searchArray := strings.FieldsFunc(search, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})

	wordsToTest := []string{}
	for p := range sc.ProductList {
		lowerCaseName := strings.ToLower(sc.ProductList[p].Name)
		lowerCaseNameArray := strings.Split(lowerCaseName, " ")
		wordsToTest = append(wordsToTest, fmt.Sprintf("%s %s %s %s", lowerCaseName, channel, strings.Join(sc.ProductList[p].Collection, " "), fillWords(lowerCaseNameArray)))
	}

	var maxMatches int
	for searchIndex := range wordsToTest {
		r := Rank{}
		for line := range searchArray {
			found := strings.Index(wordsToTest[searchIndex], searchArray[line])
			if found != -1 {
				r.Target = wordsToTest[searchIndex]
				r.IndexOfPharse = append(r.IndexOfPharse, found)
				lenIndexOfPharse := len(r.IndexOfPharse)
				r.Matches = lenIndexOfPharse
				r.OriginalIndex = searchIndex
				if maxMatches < lenIndexOfPharse {
					maxMatches = lenIndexOfPharse
				}
			}
		}
		if r.Target != "" {
			Ranking = append(Ranking, r)
		}
	}

	temp := []Rank{}
	for i := range Ranking { // das Produkt mit den meisten treffern wird angezeigt
		if Ranking[i].Matches == maxMatches {
			temp = append(temp, Ranking[i])
		}
	}
	Ranking = temp

	sort.SliceStable(Ranking, func(i, j int) bool {
		var sum1 int
		for sumi := 0; sumi < len(Ranking[i].IndexOfPharse); sumi++ {
			sum1 += Ranking[i].IndexOfPharse[sumi]
		}
		var sum2 int
		for sumj := 0; sumj < len(Ranking[j].IndexOfPharse); sumj++ {
			sum2 += Ranking[j].IndexOfPharse[sumj]
		}
		return sum1 > sum2
	})

	// Pre Message
	chatmsg := fmt.Sprintf("@%s %d Treffer: ", nick, len(Ranking))

	// Post Message
	var postMsg string
	// if !whisper { // TODO: wenn der Bot flüstern darf
	// 	postMsg = ", link per flüster"
	// }

	if len(Ranking) == 0 {
		chatmsg = "kein Treffer"
	} else if len(Ranking) < 3 {
		for line := range Ranking {
			preMsg := createDescription(sc.ProductList[Ranking[line].OriginalIndex]) + ", "
			if len(preMsg)+len(chatmsg)+len(postMsg) < MaxChatLine {
				chatmsg += preMsg
				var produktMessage string

				variantsString, found := sc.cache.Get(sc.ProductList[Ranking[0].OriginalIndex].Url)
				if found {
					log.Printf("From Cache: %s\n", sc.ProductList[Ranking[0].OriginalIndex].Url)
					produktMessage = variantsString.(string)
				} else {
					variants, err := getData(sc.ProductList[Ranking[0].OriginalIndex].Url)
					if err != nil {
						log.Println(err)
					}

					var AusverkauftCounter []string
					for groesse := range variants {
						if variants[groesse] == "Ausverkauft" {
							AusverkauftCounter = append(AusverkauftCounter, groesse)
						}
					}
					lenAusverkauftCounter := len(AusverkauftCounter)
					if lenAusverkauftCounter == 0 {
						rand.Seed(time.Now().UnixNano())
						produktMessage += fullOfStock[rand.Intn(len(fullOfStock))] // einen zufälligen Text
					} else if lenAusverkauftCounter == 1 {
						produktMessage += fmt.Sprintf("%s ist ausverkauft", AusverkauftCounter[0])
					} else {
						produktMessage += fmt.Sprintf("%s sind ausverkauft", strings.Join(AusverkauftCounter, ", "))
					}
					sc.cache.Set(sc.ProductList[Ranking[0].OriginalIndex].Url, produktMessage, cache.DefaultExpiration)
				}
				chatmsg += produktMessage
			}
		}
		chatmsg += postMsg

	} else {
		for line := range Ranking {
			var separator string
			if line != 0 {
				separator = ", "
			}
			preMsg := separator + createDescription(sc.ProductList[Ranking[line].OriginalIndex])
			if len(preMsg)+len(chatmsg)+len(postMsg) < MaxChatLine {
				chatmsg += preMsg
			}
		}
		chatmsg += postMsg
	}

	if whisper {
		return "", chatmsg
	}
	return chatmsg, ""
}

func fillWords(line []string) string {
	var result []string
	for i := range line {
		switch line[i] {
		case "pants":
			result = append(result, "hosen")
		case "shorts":
			result = append(result, "kurzehosen")
		case "collection":
			result = append(result, "kollektion")
		case "kollektion":
			result = append(result, "collection")
		case "t-shirt":
			result = append(result, "tshirts")
		case "hoodie":
			result = append(result, "kapuzenpullover")
		case "zip-hoodie":
			result = append(result, "kapuzenpullover")
			result = append(result, "reißverschluss")
		case "sweatshirt":
			result = append(result, "pullover")
		case "herren":
			result = append(result, "man")
		case "damen":
			result = append(result, "woman")
		case "mausmatte":
			result = append(result, "mauspad")
		case "lurch":
			result = append(result, "lurche")

			// Farben
		case "yellowSub":
			fallthrough
		case "yellow":
			result = append(result, "gelb")
		case "whitewalker":
			fallthrough
		case "white":
			result = append(result, "weiss")
			result = append(result, "weiß")
		case "black\u0026white":
			result = append(result, "schwarz")
			result = append(result, "weiss")
			result = append(result, "weiß")
		}
	}

	return strings.Join(result, "#")
}
