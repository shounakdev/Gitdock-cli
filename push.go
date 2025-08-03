package main

import (
	"context"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// PushToDrive uploads all files/folders in current repo to Drive
func PushToDrive(username string) {
	// Step 1: Get repo name from folder
	cwd, _ := os.Getwd()
	repoName := filepath.Base(cwd)

	// Step 2: Load config and repo data
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	repoConfig, ok := config.Repos[repoName]
	if !ok {
		log.Fatalf("No remote found for repo '%s'. Run 'gitdock remote add drive' first.", repoName)
	}

	// Step 3: Refresh token if needed
	token := &oauth2.Token{
		AccessToken:  repoConfig.AccessToken,
		RefreshToken: repoConfig.RefreshToken,
		Expiry:       repoConfig.TokenExpiry,
	}
	oauthCfg := GetOAuthConfig()
	tokenSource := oauthCfg.TokenSource(context.Background(), token)
	refreshedToken, err := tokenSource.Token()
	if err != nil {
		log.Fatalf("Failed to refresh token: %v", err)
	}

	// Save refreshed token
	repoConfig.AccessToken = refreshedToken.AccessToken
	repoConfig.TokenExpiry = refreshedToken.Expiry
	config.Repos[repoName] = repoConfig
	SaveConfig(config)

	// Step 4: Create Drive client
	driveService, err := drive.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		log.Fatalf("Failed to create Drive service: %v", err)
	}

	// Step 5: Upload everything recursively
	fmt.Println("Uploading files to Google Drive...")
	err = uploadFolderRecursive(driveService, ".", repoConfig.DriveFolderID)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}

	fmt.Println("✅ Upload complete!")
}

// Map to store created folder IDs
var folderCache = make(map[string]string)

func uploadFolderRecursive(service *drive.Service, localPath string, parentID string) error {
	return filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}

		relPath, _ := filepath.Rel(".", path)

		if info.IsDir() {
			// Create folder in Drive
			driveFolderID := createDriveFolder(service, relPath, parentID)
			folderCache[relPath] = driveFolderID
		} else {
			// Upload file to correct Drive folder
			dir := filepath.Dir(relPath)
			parent := parentID
			if id, ok := folderCache[dir]; ok {
				parent = id
			}
			uploadFile(service, relPath, parent)
		}
		return nil
	})
}

func createDriveFolder(service *drive.Service, name string, parentID string) string {
	f := &drive.File{
		Name:     filepath.Base(name),
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentID},
	}
	res, err := service.Files.Create(f).Fields("id").Do()
	if err != nil {
		log.Fatalf("Failed to create folder '%s': %v", name, err)
	}
	return res.Id
}

func uploadFile(service *drive.Service, path string, parentID string) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Failed to open file '%s': %v", path, err)
		return
	}
	defer f.Close()

	mimeType := mime.TypeByExtension(filepath.Ext(path))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	fileMetadata := &drive.File{
		Name:    filepath.Base(path),
		Parents: []string{parentID},
	}

	_, err = service.Files.Create(fileMetadata).Media(f, googleapi.ContentType(mimeType)).Do()
	if err != nil {
		log.Printf("Failed to upload file '%s': %v", path, err)
	} else {
		fmt.Printf("Uploaded: %s\n", path)
	}
}
