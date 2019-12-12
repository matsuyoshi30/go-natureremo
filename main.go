package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

type Device struct {
	DeviceCore
	Users        []User `json:"users"`
	NewestEvents Events `json:"newest_events"`
}

type DeviceCore struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	TemperatureOffset int       `json:"temperature_offset"`
	HumidityOffset    int       `json:"humidity_offset"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	FirmwareVersion   string    `json:"firmware_version"`
	MacAddress        string    `json:"mac_address"`
	SerialNumber      string    `json:"serial_number"`
}

type User struct {
	ID        string `json:"id"`
	Nickname  string `json:"nickname"`
	Superuser bool   `json:"superuser"`
}

type Events struct {
	HU SensorValue `json:"hu"`
	IL SensorValue `json:"il"`
	MO SensorValue `json:"mo"`
	TE SensorValue `json:"te"`
}

type SensorValue struct {
	Val       float64 `json:"val"`
	CreatedAt string  `json:"created_at"`
}

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

	var devices []Device
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		fmt.Println(err)
		return
	}

	for _, d := range devices {
		fmt.Printf("[Temperature]  %f°C\n", d.NewestEvents.TE.Val)
		fmt.Printf("[Humidity]     %f%%\n", d.NewestEvents.HU.Val)
		fmt.Printf("[Illumination] %f\n", d.NewestEvents.IL.Val)
		fmt.Printf("[movement]     %f\n", d.NewestEvents.MO.Val)
	}
}
