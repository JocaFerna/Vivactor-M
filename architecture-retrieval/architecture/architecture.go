package architecture

import (
	//"context"
	//"log"
	//"fmt"
	//gogithub "github.com/google/go-github/v65/github"
	//"io"
	"os"
	//"path/filepath"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5"
)

func CloneRepository(url string, path string, name string) error {
	log.Printf(url)
	log.Printf(path)

	// To ensure the path is empty
	os.RemoveAll(path)

	finalURL := url
	if !strings.HasSuffix(finalURL, ".git") {
		finalURL += ".git"
	}
	log.Printf("Cloning repository from %s to %s\n", finalURL, path)

	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      finalURL,
		Progress: os.Stdout,
	})
	return err
}

func StartArchitecture(repoName string, make_instructions string) {
	log.Println("Starting architecture...")
	err := dockerComposeHandler(repoName, make_instructions)
	if err != nil {
		log.Printf("Error starting architecture: %s\n", err.Error())
	} else {
		log.Println("Architecture started successfully")
	}
}

func createNetworkOverride(repoPath string) error {
	// This override tells the new containers to join 'thesis_shared_network'
	// and treats it as an external network that already exists.
	overrideContent := `
networks:
  default:
    name: thesis_app-network
    external: true
`
	overridePath := filepath.Join(repoPath, "docker-compose.override.yml")
	return os.WriteFile(overridePath, []byte(overrideContent), 0644)
}

func dockerComposeHandler(repoName string, instructions_to_start string) error {
	// 1. Define the path inside your Go container
	// This must match the volume you mounted in your manager's compose file
	cleanName := strings.TrimPrefix(repoName, "/")
	repoPath := filepath.Join("/api/downloads", cleanName)

	// 1. Create the bridge to the existing network
	if err := createNetworkOverride(repoPath); err != nil {
		return fmt.Errorf("failed to create network override: %w", err)
	}

	// Separate the instructions to start the architecture into a list of commands by \n
	commands := strings.Split(instructions_to_start, "\n")

	// Execute each command in the list
	for _, cmdStr := range commands {

		cmdStr = strings.TrimSpace(cmdStr)
		if cmdStr == "" {
			continue // Skip empty lines
		}
		log.Printf("Executing command: %s\n", cmdStr)
		cmdParts := strings.Fields(cmdStr)
		log.Printf("Command parts: %v\n", cmdParts)
		cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
		cmd.Dir = repoPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Printf("Error executing command '%s': %s\n", cmdStr, err.Error())
			return err
		}
	}

	return nil
}

/**func ProcessRepositories(accessToken string, repos []*gogithub.Repository) {
    // 1. Initialize the client with the token you just got
    client := gogithub.NewClient(nil).WithAuthToken(accessToken)
    ctx := context.Background()
	repo := repos[0]

	owner := repo.GetOwner().GetLogin()
	name := repo.GetName()

	log.Printf("Fetching content for: %s/%s\n", owner, name)

	// 2. Get the root directory contents
	// Passing an empty string for "path" gets the root of the repo
	fileContent, directoryContents, _, err := client.Repositories.GetContents(ctx, owner, name, "", nil)
	if err != nil {
		log.Printf("Error fetching contents for %s: %v\n", name, err)
		return
	}

	// 3. Handle the result
	if fileContent != nil {
		// This happens if the path was a single file
		content, _ := fileContent.GetContent()
		fmt.Printf("File Content: %s\n", content)
	} else if directoryContents != nil {
		// This happens for the root ("") or any folder
		fmt.Printf("Found %d items in the root of %s\n", len(directoryContents), name)
		for _, item := range directoryContents {
			fmt.Printf("- [%s] %s\n", item.GetType(), item.GetName())
		}
		// Download the content of each file in the root directory
		for _, item := range directoryContents {
			if item.GetType() == "file" {
				reader, _, err := client.Repositories.DownloadContents(ctx, owner, name, item.GetPath(), nil)
				if err != nil {
					log.Printf("Error downloading content for %s: %v\n", item.GetName(), err)
					continue
				}
				contentBytes, err := io.ReadAll(reader)
				if err != nil {
					log.Printf("Error reading content for %s: %v\n", item.GetName(), err)
					continue
				}

				path1 := filepath.Join(os.TempDir(),name,"downloads", item.GetName())
				log.Printf("Saving content for %s to %s\n", item.GetName(), path1)
				err = os.MkdirAll(filepath.Dir(path1), os.ModePerm)
				if err != nil {
					log.Printf("Error creating directories for %s: %v\n", item.GetName(), err)
					continue
				}
				f, err := os.Create(path1)
				if err != nil {
					log.Printf("Error creating file for %s: %v\n", item.GetName(), err)
					continue
				}
				defer f.Close()
				_, err2 := f.Write(contentBytes)
				if err2 != nil {
				 	log.Printf("Error saving content for %s: %v\n", item.GetName(), err)
				 	continue
				}
				//io.Close(reader)
				fmt.Printf("Downloaded content for %s\n", item.GetName())

			}
			else if item.GetType() == "dir" {
				// Get the contents of the directory
				subFileContent, subDirectoryContents, _, err := client.Repositories.GetContents(ctx, owner, name, item.GetPath(), nil)
				if err != nil {
					log.Printf("Error fetching contents for directory %s: %v\n", item.GetName(), err)
					continue
				}
				if subFileContent != nil {
					content, _ := subFileContent.GetContent()
					fmt.Printf("File Content in directory %s: %s\n", item.GetName(), content)
				} else if subDirectoryContents != nil {
					fmt.Printf("Found %d items in directory %s\n", len(subDirectoryContents), item.GetName())
					for _, subItem := range subDirectoryContents {
						fmt.Printf("- [%s] %s\n", subItem.GetType(), subItem.GetName())
					}
				}
			}
		}
	}


}*/
