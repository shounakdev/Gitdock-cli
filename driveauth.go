package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"golang.org/x/oauth2"
)

func authenticateWithGoogle() (*oauth2.Token, string, error) {
	oauthConfig := GetOAuthConfig()

	// Step 1: Generate auth URL
	authURL := oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("Open this URL in your browser to authenticate:\n", authURL)

	// Step 2: Auto open in browser
	exec.Command("xdg-open", authURL).Start()

	// Step 3: Start local HTTP server for OAuth callback
	codeCh := make(chan string)
	srv := &http.Server{Addr: ":8080"}

	http.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		fmt.Fprintf(w, "Authentication successful. You can close this window.")
		codeCh <- code
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	code := <-codeCh
	srv.Shutdown(context.Background())

	// Step 4: Exchange code for tokens
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, "", err
	}

	// Step 5: Get user email
	client := oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
	}
	json.NewDecoder(resp.Body).Decode(&userInfo)

	return token, userInfo.Email, nil
}
