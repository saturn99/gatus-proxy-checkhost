package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Config struct {
	Port          string
	TelegramToken string
	TelegramChat  string
	TelegramBase  string
	ProxyURL      string // Optional: http://user:pass@proxy:port
}

type GatusWebhook struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type CheckHostResponse struct {
	RequestID string `json:"request_id"`
	Permanent string `json:"permanent_link"`
}

func main() {
	config := Config{
		Port:          getEnv("PORT", "8080"),
		TelegramToken: getEnv("TELEGRAM_TOKEN", ""),
		TelegramChat:  getEnv("TELEGRAM_CHAT", ""),
		TelegramBase:  getEnv("TELEGRAM_BASE", "https://api.telegram.org/bot"),
		ProxyURL:      getEnv("PROXY_URL", ""), // e.g., http://user:pass@proxy.example.com:8080
	}

	if config.TelegramToken == "" || config.TelegramChat == "" {
		log.Fatal("TELEGRAM_TOKEN and TELEGRAM_CHAT environment variables are required")
	}

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		handleWebhook(w, r, config)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Starting server on port %s", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}

func handleWebhook(w http.ResponseWriter, r *http.Request, config Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload GatusWebhook
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("Error decoding webhook: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	log.Printf("Received alert: %s (Status: %s)", payload.Name, payload.Status)

	// Only trigger check-host when status is "check"
	if payload.Status == "check" {
		go processAlert(payload, config)
	} else {
		log.Printf("Skipping alert - status is not 'check' (got: %s)", payload.Status)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func processAlert(payload GatusWebhook, config Config) {
	// Extract hostname from service name or use a default
	// You might want to customize this based on your Gatus config
	targetHost := payload.Name
	
	checkHostLink, err := createCheckHostCheck(targetHost, config.ProxyURL)
	if err != nil {
		log.Printf("Error creating check-host check: %v", err)
		sendTelegramMessage(config, fmt.Sprintf("Failed to create check-host report: %v", err))
		return
	}

	if err := sendTelegramMessage(config, checkHostLink); err != nil {
		log.Printf("Error sending Telegram message: %v", err)
	}
}

func createCheckHostCheck(host string, proxyURL string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Configure proxy if provided
	if proxyURL != "" {
		proxyURLParsed, err := url.Parse(proxyURL)
		if err != nil {
			return "", fmt.Errorf("invalid proxy URL: %w", err)
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURLParsed),
		}
		log.Printf("Using proxy: %s", proxyURL)
	}

	// Create check-host check (HTTP check as example)
	apiURL := fmt.Sprintf("https://check-host.net/check-http?host=%s&max_nodes=10", url.QueryEscape(host))
	
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("check-host returned status %d: %s", resp.StatusCode, string(body))
	}

	var checkResp CheckHostResponse
	if err := json.Unmarshal(body, &checkResp); err != nil {
		return "", fmt.Errorf("failed to parse check-host response: %w", err)
	}

	// Wait a bit for results to be ready
	time.Sleep(5 * time.Second)

	// Return the permanent link
	if checkResp.Permanent != "" {
		return checkResp.Permanent, nil
	}

	if checkResp.RequestID != "" {
		return fmt.Sprintf("https://check-host.net/check-result/%s", checkResp.RequestID), nil
	}

	return "https://check-host.net", nil
}

func sendTelegramMessage(config Config, message string) error {
	apiURL := fmt.Sprintf("%s%s/sendMessage", config.TelegramBase , config.TelegramToken)

	payload := map[string]interface{}{
		"chat_id":                  config.TelegramChat,
		"text":                     message,
		"disable_web_page_preview": true,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API returned status %d: %s", resp.StatusCode, string(body))
	}

	log.Println("Telegram message sent successfully")
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}