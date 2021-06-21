package grt

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigFastest

const g_GRTUniV2 string = "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v2"

type PairData struct {
	tok0Addr string
	tok1Addr string
	tok0Sym  string
	tok1Sym  string
	txCount  int
}

func GetPairCount() int {
	jsonData := map[string]string{"query": `{ 
		uniswapFactory(id: "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f") {
			pairCount
		}
	}`}
	jsonValue, _ := jsoniter.ConfigFastest.Marshal(jsonData)
	resp, err := http.Post(g_GRTUniV2, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Println("IN ERROR")
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(bodyBytes))
	pairCount, _ := strconv.Atoi(json.Get(bodyBytes, "data", "uniswapFactory", "pairCount").ToString())
	return pairCount
}

func QueryAllPairs() map[string]PairData {
	// Maps pair address to pair data
	var pairsData map[string]PairData = make(map[string]PairData)
	totalPairs := GetPairCount()
	remainingPairs := totalPairs
	firstStep := true
	var pairs []string
	var lastID string
	for {
		if firstStep {
			firstStep = false
			jsonData := map[string]string{"query": `{ 
			pairs(first: 1000) {
			    id
				reserve0 
				reserve1
				txCount
				token0 {
					id
					symbol
				}
				token1 {
					id
					symbol
				}
			  }
			}`}
			jsonValue, _ := jsoniter.ConfigFastest.Marshal(jsonData)
			resp, err := http.Post(g_GRTUniV2, "application/json", bytes.NewBuffer(jsonValue))
			if err != nil {
				fmt.Println("IN ERROR")
				log.Fatalln(err)
			}
			defer resp.Body.Close()
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			pairsItr := jsoniter.ConfigFastest.Get(bodyBytes, "data", "pairs")
			remainingPairs -= 1000
			count := 0
			for {
				pairAddr := pairsItr.Get(count, "id").ToString()
				if pairAddr == "" {
					break
				} else {
					pairs = append(pairs, pairAddr)
					txCount, err := strconv.Atoi(pairsItr.Get(count, "txCount").ToString())
					reserve0, err := strconv.ParseFloat(pairsItr.Get(count, "reserve0").ToString(), 64)
					reserve1, err := strconv.ParseFloat(pairsItr.Get(count, "reserve1").ToString(), 64)
					if txCount < 500 {
						count++
						continue
					}
					if err != nil {
						fmt.Println(pairsItr.Get(count).ToString())
						log.Fatalf("failed to convert txcount to int with err: %v\n", err)
					}
					tok0Addr := pairsItr.Get(count, "token0", "id").ToString()
					tok0Sym := pairsItr.Get(count, "token0", "symbol").ToString()
					tok1Addr := pairsItr.Get(count, "token1", "id").ToString()
					tok1Sym := pairsItr.Get(count, "token1", "symbol").ToString()
					if tok0Sym != "WETH" && tok1Sym != "WETH" {
						count++
						continue
					}
					if tok0Sym == "WETH" {
						if reserve0 < 5 || reserve1 < 1000 {
							count++
							continue
						}
					} else if tok1Sym == "WETH" {
						if reserve1 < 5 || reserve0 < 1000 {
							count++
							continue
						}
					}
					pairsData[pairAddr] = PairData{
						tok0Addr: tok0Addr,
						tok0Sym:  tok0Sym,
						tok1Addr: tok1Addr,
						tok1Sym:  tok1Sym,
						txCount:  txCount,
					}
				}
				count++
			}
			lastID = pairs[len(pairs)-1]
		} else {
			if remainingPairs < 1000 {
				jsonData := map[string]string{"query": fmt.Sprintf(`{ 
					pairs(first: %v, where: { id_gt: "%v" }) {
						id
						txCount
						reserve0 
						reserve1
						token0 {
							id
							symbol
						}
						token1 {
							id
							symbol
						}
					  }
				}`, remainingPairs, lastID)}
				jsonValue, _ := jsoniter.ConfigFastest.Marshal(jsonData)
				resp, err := http.Post(g_GRTUniV2, "application/json", bytes.NewBuffer(jsonValue))
				if err != nil {
					fmt.Println("IN ERROR")
					log.Fatalln(err)
				}
				defer resp.Body.Close()
				bodyBytes, _ := ioutil.ReadAll(resp.Body)
				remainingPairs -= 1000
				fmt.Println("remaining pairs", remainingPairs)
				pairsItr := jsoniter.ConfigFastest.Get(bodyBytes, "data", "pairs")
				count := 0
				for {
					pairAddr := pairsItr.Get(count, "id").ToString()
					if pairAddr == "" {
						break
					} else {
						pairs = append(pairs, pairAddr)
						txCount, err := strconv.Atoi(pairsItr.Get(count, "txCount").ToString())
						reserve0, err := strconv.ParseFloat(pairsItr.Get(count, "reserve0").ToString(), 64)
						reserve1, err := strconv.ParseFloat(pairsItr.Get(count, "reserve1").ToString(), 64)
						if err != nil {
							fmt.Println(string(bodyBytes))
							log.Fatalf("failed to convert txcount to int with err: %v\n", err)
						}
						if txCount < 500 {
							count++
							continue
						}
						tok0Addr := pairsItr.Get(count, "token0", "id").ToString()
						tok0Sym := pairsItr.Get(count, "token0", "symbol").ToString()
						tok1Addr := pairsItr.Get(count, "token1", "id").ToString()
						tok1Sym := pairsItr.Get(count, "token1", "symbol").ToString()
						if tok0Sym != "WETH" && tok1Sym != "WETH" {
							count++
							continue
						}
						if tok0Sym == "WETH" {
							if reserve0 < 5 || reserve1 < 1000 {
								count++
								continue
							}
						} else if tok1Sym == "WETH" {
							if reserve1 < 5 || reserve0 < 1000 {
								count++
								continue
							}
						}
						pairsData[pairAddr] = PairData{
							tok0Addr: tok0Addr,
							tok0Sym:  tok0Sym,
							tok1Addr: tok1Addr,
							tok1Sym:  tok1Sym,
							txCount:  txCount,
						}
					}
					count++
				}
				break
			}
			jsonData := map[string]string{"query": fmt.Sprintf(`{ 
			pairs(first: 1000, where: { id_gt: "%v" }) {
			    id
				txCount
				reserve0 
				reserve1
				token0 {
					id
					symbol
				}
				token1 {
					id
					symbol
				}
			  }
			}`, lastID)}
			jsonValue, _ := jsoniter.ConfigFastest.Marshal(jsonData)
			resp, err := http.Post(g_GRTUniV2, "application/json", bytes.NewBuffer(jsonValue))
			if err != nil {
				fmt.Println("IN ERROR")
				log.Fatalln(err)
			}
			defer resp.Body.Close()
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			remainingPairs -= 1000
			fmt.Println("remaining pairs", remainingPairs)
			pairsItr := jsoniter.ConfigFastest.Get(bodyBytes, "data", "pairs")
			count := 0
			for {
				pairAddr := pairsItr.Get(count, "id").ToString()
				if pairAddr == "" {
					break
				} else {
					pairs = append(pairs, pairAddr)
					txCount, err := strconv.Atoi(pairsItr.Get(count, "txCount").ToString())
					reserve0, err := strconv.ParseFloat(pairsItr.Get(count, "reserve0").ToString(), 64)
					reserve1, err := strconv.ParseFloat(pairsItr.Get(count, "reserve1").ToString(), 64)
					fmt.Println(reserve0, reserve1)
					if err != nil {
						fmt.Println(string(bodyBytes))
						log.Fatalf("failed to convert txcount to int with err: %v\n", err)
					}
					if txCount < 500 {
						count++
						continue
					}
					tok0Addr := pairsItr.Get(count, "token0", "id").ToString()
					tok0Sym := pairsItr.Get(count, "token0", "symbol").ToString()
					tok1Addr := pairsItr.Get(count, "token1", "id").ToString()
					tok1Sym := pairsItr.Get(count, "token1", "symbol").ToString()
					if tok0Sym != "WETH" && tok1Sym != "WETH" {
						count++
						continue
					}
					if tok0Sym == "WETH" {
						if reserve0 < 5 || reserve1 < 1000 {
							count++
							continue
						}
					} else if tok1Sym == "WETH" {
						if reserve1 < 5 || reserve0 < 1000 {
							count++
							continue
						}
					}
					pairsData[pairAddr] = PairData{
						tok0Addr: tok0Addr,
						tok0Sym:  tok0Sym,
						tok1Addr: tok1Addr,
						tok1Sym:  tok1Sym,
						txCount:  txCount,
					}
				}
				count++
			}
			lastID = pairs[len(pairs)-1]
			fmt.Printf("LASTID: %v\n", lastID)
			fmt.Println(len(pairs))
		}
	}
	return pairsData
}

func WritePairsData() {
	pairsData := QueryAllPairs()
	f, err := os.Create("/path/to/pairsData.json")
	defer f.Close()
	if err != nil {
		fmt.Printf("failed to create a file in WritePairsData:261 with err: %v\n", err)
		return
	}
	f.WriteString(`{"data": [`)
	f.Sync()
	for k, v := range pairsData {
		toWrite := fmt.Sprintf(`{"pairAddress": "%v","txCount": %v,"token0": "%v","token0Sym": "%v","token1": "%v","token1Sym": "%v"},`,
			k, v.txCount, v.tok0Addr, v.tok0Sym, v.tok1Addr, v.tok1Sym)
		f.WriteString(toWrite)
		f.Sync()
	}
	f.WriteString("]}")
	f.Sync()
	return
}

func QueryGRT(tokenAddress string) int {
	jsonData := map[string]string{
		"query": fmt.Sprintf(`
            {
                token(id: "%v") {
					txCount
                }
            }
        `, strings.ToLower(tokenAddress)),
	}
	jsonValue, _ := jsoniter.ConfigFastest.Marshal(jsonData)
	resp, err := http.Post(g_GRTUniV2, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Println("IN ERROR")
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	txCount, _ := strconv.Atoi(jsoniter.ConfigFastest.Get(bodyBytes, "data", "token", "txCount").ToString())
	return txCount
}
