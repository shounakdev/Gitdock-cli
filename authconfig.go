package main

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Central place for your Google OAuth credentials
const (
	CLIENT_ID     = ""
	CLIENT_SECRET = ""
	REDIRECT_URI  = "http://localhost:8080/oauth2callback"
)

var SCOPES = []string{
	"https://www.googleapis.com/auth/drive.file",
	"https://www.googleapis.com/auth/userinfo.email",
}

// GetOAuthConfig returns reusable oauth2.Config
func GetOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     CLIENT_ID,
		ClientSecret: CLIENT_SECRET,
		RedirectURL:  REDIRECT_URI,
		Scopes:       SCOPES,
		Endpoint:     google.Endpoint,
	}
}
