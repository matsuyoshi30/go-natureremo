package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Token TokenConfig
}

type TokenConfig struct {
	AccessToken string
}

const URL = "https://api.nature.global/1/devices"

var config Config

func main() {
	_, err := toml.DecodeFile("./config.toml", &config)
	if err != nil {
		fmt.Println(err)
		return
	}

	accessToken := config.Token.AccessToken
	if accessToken == "" {
		fmt.Println("Empty AccessToken")
		return
	}

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(body))
}
