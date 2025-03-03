package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Representerar svaret från authorization servern (https://test-user.stim.se/token)
type AuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Hämta en access token från authorization server
func GetJWT() (string, error) {
	// Hämta credentials (Ska ej ligga här. Lägg till i .env-fil senare!!)
	clientID := "6e8ae61d-a721-4efa-a0ab-0c469269f6e8"
	clientSecret := "dnRFPIv1OsHAmEo7"
	authURL := "https://test-user.stim.se/token"

	// Data-form payload
	form := url.Values{}
	form.Add("grant_type", "client_credentials")
	form.Add("client_id", clientID)
	form.Add("client_secret", clientSecret)
	form.Add("scope", "backstage.reporters.list")

	// HTTP-anropet
	req, err := http.NewRequest("POST", authURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	// Headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Skicka request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Kolla att statuskoden är 200/OK
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authentication failed, status code: %d", resp.StatusCode)
	}

	// Läs svaret
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Omvandla JSON-svaret till Go-struktur
	var authResp AuthResponse
	err = json.Unmarshal(body, &authResp)
	if err != nil {
		return "", err
	}

	// Returnera JWT
	return authResp.AccessToken, nil
}
