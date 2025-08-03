package main

import (
	"fmt"
	"os"
	"path/filepath"

	"mygit/commands"
	"mygit/rbac"
)

func getCurrentUser() string {
	user, err := os.UserHomeDir()
	if err != nil {
		return "unknown"
	}
	return filepath.Base(user)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  gitdock remote add drive")
		fmt.Println("  gitdock <command> <username> [args]")
		return
	}

	// Handle: gitdock remote add drive
	if os.Args[1] == "remote" && len(os.Args) >= 4 && os.Args[2] == "add" && os.Args[3] == "drive" {
		AddRemoteDrive()
		return
	}

	// For all other commands, require at least cmd and username
	cmd := os.Args[1]
	var username string
	if len(os.Args) >= 3 {
		username = os.Args[2]
	} else {
		username = getCurrentUser()
		fmt.Printf("⚠️  Username not provided. Using current OS user: %s\n", username)
	}

	// RBAC check
	if !rbac.HasPermission(username, cmd) {
		fmt.Printf("❌ Access denied: User '%s' is not allowed to run '%s'\n", username, cmd)
		return
	}

	switch cmd {
	case "init":
		commands.InitRepo()

	case "add":
		if len(os.Args) < 4 {
			fmt.Println("Usage: gitdock add <username> <file>")
			return
		}
		commands.AddFile(os.Args[3])

	case "commit":
		if len(os.Args) < 6 {
			fmt.Println("Usage: gitdock commit <username> <author> <message>")
			return
		}
		commands.CommitChanges(username, os.Args[3], os.Args[4])

	case "status":
		commands.Status()

	case "log":
		commands.ShowLog()

	case "branch":
		commands.Branch()

	case "create-branch":
		if len(os.Args) < 4 {
			fmt.Println("Usage: gitdock create-branch <username> <branch-name>")
			return
		}
		commands.CreateBranch(username, os.Args[3])

	case "checkout":
		if len(os.Args) < 4 {
			fmt.Println("Usage: gitdock checkout <username> <branch-name>")
			return
		}
		commands.Checkout(username, os.Args[3])

	case "push":
		// We'll implement gitdock push <username> next!
		PushToDrive(username)

	default:
		fmt.Println("Unknown command:", cmd)
	}
}
