package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"
)

type CoordinatesAddress struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Address struct {
	Neighborhood string             `json:"neighborhood"`
	StreetName   string             `json:"streetName"`
	StreetNumber string             `json:"streetNumber"`
	Complement   string             `json:"complement"`
	Coordinates  CoordinatesAddress `json:"coordinates"`
}

func ListAddress() []Address {
	baseURL := "https://marketplace.ifood.com.br/v1/customers/me/addresses"
	req, _ := http.NewRequest("GET", baseURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		fmt.Println("err", err)
	}
	defer r.Body.Close()

	content, _ := ioutil.ReadAll(r.Body)

	var result []Address
	json.Unmarshal(content, &result)

	return result
}

func AskOtpCode(email string) string {
	url := "https://marketplace.ifood.com.br/v1/identity-providers/OTP/authorization-codes"
	payload := map[string]string{"tenant_id": "IFO", "email": email, "type": "EMAIL"}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("err", err)
	}

	req.Header.Set("Platform", "Desktop")
	req.Header.Set("accept-language", "pt-BR,pt")
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("err", err)
	}
	defer resp.Body.Close()

	content, _ := ioutil.ReadAll(resp.Body)

	var result map[string]string
	json.Unmarshal(content, &result)

	return result["key"]
}

func ClaimOtpCode(otpCode, token string) string {
	base := "https://marketplace.ifood.com.br/v1/identity-providers/OTP/access-tokens?key=%s&auth_code=%s"
	fullURL := fmt.Sprintf(base, url.QueryEscape(token), otpCode)
	fullURL = strings.Split(fullURL, "\n")[0]

	r, err := http.DefaultClient.Get(fullURL)
	if err != nil {
		fmt.Println("err", err)
	}
	defer r.Body.Close()

	var p map[string]string
	json.NewDecoder(r.Body).Decode(&p)
	return p["access_token"]
}

func Auth(email, token string) (string, string) {
	baseURL := "https://marketplace.ifood.com.br/v2/identity-providers/OTP/authentications"
	payload := map[string]string{
		"tenant_id": "IFO",
		"device_id": "17a538bb-d063-4fd6-a613-74bc9465e09c",
		"email":     email,
		"token":     token,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("err", err)
	}

	req.Header.Set("Platform", "Desktop")
	req.Header.Set("accept-language", "pt-BR,pt")
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("err", err)
	}
	defer resp.Body.Close()

	content, _ := ioutil.ReadAll(resp.Body)

	var result map[string]string
	json.Unmarshal(content, &result)

	return result["access_token"], result["refresh_token"]
}

func GetHome() gjson.Result {
	uri := "https://marketplace.ifood.com.br/v2/home?alias=single_tab_cms&latitude=%s&longitude=%s&channel=IFOOD&size=100"
	uri = fmt.Sprintf(uri, fmt.Sprint(selectedAddress.Coordinates.Latitude), fmt.Sprint(selectedAddress.Coordinates.Longitude))
	fmt.Println(uri)

	payload := map[string][]string{
		"supported-headers": []string{"OPERATION_HEADER"},
		"supported-cards":   []string{"SMALL_BANNER_CAROUSEL"},
		"supported-actions": []string{"catalog-item", "merchant", "page", "card-content", "last-restaurants"},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", uri, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		fmt.Println("err", err)
	}
	defer r.Body.Close()

	content, _ := ioutil.ReadAll(r.Body)

	return gjson.Parse(string(content))
}

func RefreshToken() (string, string) {
	baseURL := "https://marketplace.ifood.com.br/v2/access_tokens"
	payload := map[string]string{
		"refresh_token": refreshToken,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("err", err)
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 201 {
		fmt.Println("err", err)
	}
	defer resp.Body.Close()

	content, _ := ioutil.ReadAll(resp.Body)

	var result map[string]string
	json.Unmarshal(content, &result)

	return result["access_token"], result["refresh_token"]
}
