package last_will
import (
	"os"
	"os/signal"
	"syscall"
	"os/exec"
	"log"
)

func SetupCleanupHandler() {
	// Create a channel to listen for OS signals
	c := make(chan os.Signal, 1)
	
	// We want to catch SIGTERM (Docker stop) and SIGINT (Ctrl+C)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c // This blocks until a signal is received
		log.Println("Shutdown signal received! Executing Last-Will...")

		// EXECUTE YOUR MAKEFILE
		// Change the directory to where your Makefile is located
		cmd := exec.Command("docker", "compose", "down") // or whatever your target is
		
		// Execute this command in every subdirectory of /api/downloads.
		downloadsDir := "/api/downloads"
		subdirs, err := os.ReadDir(downloadsDir)
		if err != nil {
			log.Printf("Failed to read downloads directory: %v", err)
			os.Exit(1)
		}
		
		for _, subdir := range subdirs {
			if subdir.IsDir() {
				cmd.Dir = downloadsDir + "/" + subdir.Name()
				// For every repository, execute docker compose down to stop the containers and remove them
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				if err := cmd.Run(); err != nil {
					log.Printf("Last-Will Makefile failed: %v", err)
				} else {
					log.Println("Last-Will Makefile executed successfully.")
				}
			}
		}
		os.Exit(0) // Now we can safely exit
	}()
}


