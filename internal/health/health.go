package health

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"os"
)

func GetHealthEnvVar() (string, string, string) {
	userID := os.Getenv("CODEC_USER_ID")
	apiToken := os.Getenv("CODEC_API_TOKEN")
	apiURL := os.Getenv("CODEC_API_URL")

	if (userID == "" || apiToken == "" || apiURL == "") {
		return "", "", ""
	}

	return userID, apiToken, apiURL
}

func InitHealthChecks(
	userID string,
	apiToken string,
	apiURL string,
) {
	token := base64.StdEncoding.EncodeToString([]byte(userID + "; " + apiToken))
	healthURL := apiURL + "/api/v1/instance/health"
	lastError := ""

	updateHealth := func() {
		req, err := http.NewRequest("PUT", healthURL, nil)
		if err != nil {
			if lastError != err.Error() {
				log.Println("health check: failed to create request:", err)
			}

			lastError = err.Error()
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			if lastError != err.Error() {
				log.Println("health check: fetch error:", err)
			}

			lastError = err.Error()
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)

			err := fmt.Errorf(
				"failed to get health, status code: %d - %s - %s", 
				resp.StatusCode, 
				resp.Status, 
				string(body),
			)

			if lastError != err.Error() {
				log.Println("health check: wrong status code:", err)
			}

			lastError = err.Error()
			return
		}

		lastError = ""
	}

	updateHealth()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			updateHealth()
		}
	}()
}
