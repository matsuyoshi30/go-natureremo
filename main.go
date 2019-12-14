package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
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

const URL = "https://api.nature.global/1/"

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

type Appliance struct {
	ID string `json:"id"`
	DeviceCore
	Tv TV `json:"tv"`
}

type TV struct {
	StateTV TVState  `json:"state"`
	Buttons []Button `json:"buttons"`
}

type TVState struct {
	Input string `json:"input"`
}

type Button struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	Label string `json:"label"`
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

	devices := flag.Bool("d", false, "devices")
	// appliances := flag.Bool("a", false, "appliances")
	volup := flag.Bool("vu", false, "tv volume up")
	voldown := flag.Bool("vb", false, "tv volume down")
	flag.Parse()
	if !*devices && !*volup && !*voldown {
		fmt.Println("Choose flag -d or -a")
		return
	}

	var url string
	if *devices {
		url = URL + "devices"
	} else if *volup || *voldown {
		url = URL + "appliances"
	}

	req, err := http.NewRequest("GET", url, nil)
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

	if *devices {
		showDevices(resp.Body)
	} else if *volup || *voldown {
		applianceID := checkAppliances(resp.Body)
		url = url + "/" + applianceID + "/tv"

		req, err := http.NewRequest("POST", url, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		req.Header.Add("Authorization", "Bearer "+accessToken)

		params := req.URL.Query()
		if *volup {
			params.Add("button", "vol-up")
		} else if *voldown {
			params.Add("button", "vol-down")
		}
		req.URL.RawQuery = params.Encode()

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Println("failed")
			return
		}
		fmt.Println("success")
	}
}

func showDevices(reader io.Reader) {
	var devices []Device
	if err := json.NewDecoder(reader).Decode(&devices); err != nil {
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

func checkAppliances(reader io.Reader) string {
	var appliances []Appliance
	if err := json.NewDecoder(reader).Decode(&appliances); err != nil {
		fmt.Println(err)
		return ""
	}

	for _, a := range appliances {
		return a.ID
		// for _, b := range a.Tv.Buttons {
		// 	fmt.Printf("Name: %s, image: %s, label: %s\n", b.Name, b.Image, b.Label)
		// }
	}

	return ""
}
