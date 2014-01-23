package main

import (
	"flag"
	"fmt"
	"os"
	"net/http"
	"io/ioutil"
	"regexp"
	"strings"
)

const URL = "http://tw.money.yahoo.com/currency_exc_result?amt=%s&from=%s&to=%s"

func usage() {
	fmt.Printf("usage: currency [str]\n")
	os.Exit(1)
	os.Exit(-1)
}

func expand_alias(alias string, aliases map[string]string) string {
	if val,ok := aliases[alias]; ok {
		return val
	}
	return alias
}

func extract(strs []string) (string, string, string, error) {
	COINS := []string {
		"TWD", "CNY", "JPY", "KRW",
		"HKD", "THB", "SGD", "IDR",
		"VND", "MYR", "PHP", "INR",
		"AED", "KWD", "AUD", "NZD",
		"USD", "CAD", "BRL", "MXN",
		"ARS", "CLP", "VEB", "EUR",
		"GBP", "RUB", "CHF", "SEK",
		"ZAR",
	}
	CALIASES := map[string]string {
		"NTD": "TWD", "RMB": "CNY",
	}

	calias := ""
	for k, _ := range CALIASES {
		if calias == "" {
			calias = k
		} else {
			calias = calias + "|" + k
		}
	}

	text := strings.Join(strs, " ")
	coins := strings.Join(COINS, "|") + "|" + calias

	// First regex
	pattern := fmt.Sprintf("(?i)^([\\d\\.\\+\\-\\*\\/]+)\\s*(%s)$", coins)
	regex, _ := regexp.Compile(pattern)
	finded := regex.FindSubmatch([]byte(text))
	if finded != nil {
		money, to := string(finded[1]), "TWD"
		from := expand_alias(strings.ToUpper(string(finded[2])), CALIASES)
		return money, from, to, nil
	}

	// Second regex
	pattern = fmt.Sprintf("(?i)^([\\d\\.\\+\\-\\*\\/]+)\\s*(%s)\\s+to\\s+(%s)$", coins, coins)
	regex, _ = regexp.Compile(pattern)
	finded = regex.FindSubmatch([]byte(text))
	if finded != nil {
		money := string(finded[1])
		from := expand_alias(strings.ToUpper(string(finded[2])), CALIASES)
		to := expand_alias(strings.ToUpper(string(finded[3])), CALIASES)
		return money, from, to, nil
	}

	return "", "", "", fmt.Errorf("No match string found: %s", text)
}

func currency(strs []string) {
	money, from, to, err := extract(strs)
	if err != nil {
		fmt.Println(err)
		return
	}

	url := fmt.Sprintf(URL, money, from, to)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error http get")
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Read content failed")
		return
	}

	regex, _ := regexp.Compile("經過計算後， (.+)<div")
	finded := regex.FindSubmatch(body)
	if finded == nil {
		fmt.Println("Not find")
		return
	}

	regForSub, _ := regexp.Compile("</?em>")
	out := regForSub.ReplaceAll(finded[1], []byte(""))
	fmt.Println(string(out))
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		usage()
		os.Exit(1)
	}

	currency(os.Args[1:])
}
