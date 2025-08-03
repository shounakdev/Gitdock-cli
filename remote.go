package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func AddRemoteDrive() {
	// Step 1: Authenticate with Google (opens browser)
	token, email, err := authenticateWithGoogle()
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	// Step 2: Get current folder name as repo name
	cwd, _ := os.Getwd()
	repoName := filepath.Base(cwd)

	// Step 3: Create Drive client with access token
	ctx := context.Background()
	driveService, err := drive.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(token)))
	if err != nil {
		log.Fatalf("Failed to create Drive service: %v", err)
	}

	// Step 4: Check or create "GitDockRepos" parent folder
	parentID := getOrCreateFolder(driveService, "GitDockRepos", "")

	// Step 5: Create repo folder inside GitDockRepos
	repoFolderID := getOrCreateFolder(driveService, repoName, parentID)

	// Step 6: Load existing config
	config, _ := LoadConfig()

	// Step 7: Save new repo config
	if config.Repos == nil {
		config.Repos = make(map[string]RepoConfig)
	}

	config.UserEmail = email
	config.Repos[repoName] = RepoConfig{
		DriveFolderID: repoFolderID,
		AccessToken:   token.AccessToken,
		RefreshToken:  token.RefreshToken,
		TokenExpiry:   token.Expiry,
	}

	err = SaveConfig(config)
	if err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}

	fmt.Printf("Remote Drive added for repo '%s'. All files will be pushed to this folder.\n", repoName)
}

func getOrCreateFolder(service *drive.Service, folderName string, parentID string) string {
	query := fmt.Sprintf("mimeType='application/vnd.google-apps.folder' and name='%s' and trashed=false", folderName)
	if parentID != "" {
		query += fmt.Sprintf(" and '%s' in parents", parentID)
	}

	res, err := service.Files.List().Q(query).Fields("files(id, name)").Do()
	if err != nil {
		log.Fatalf("Failed to search for folder '%s': %v", folderName, err)
	}

	if len(res.Files) > 0 {
		return res.Files[0].Id
	}

	// Folder not found, create it
	f := &drive.File{
		Name:     folderName,
		MimeType: "application/vnd.google-apps.folder",
	}
	if parentID != "" {
		f.Parents = []string{parentID}
	}

	created, err := service.Files.Create(f).Do()
	if err != nil {
		log.Fatalf("Failed to create folder '%s': %v", folderName, err)
	}
	return created.Id
}
