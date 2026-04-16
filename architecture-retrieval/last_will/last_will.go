package last_will
import (
	"os"
	"os/signal"
	"syscall"
	"os/exec"
	"path/filepath"
	"log"
)

func SetupCleanupHandler() {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-c 
        log.Println("Shutdown signal received! Cleaning up environment...")

        downloadsDir := "/api/downloads"
        subdirs, err := os.ReadDir(downloadsDir)
        if err != nil {
            log.Printf("Failed to read downloads directory: %v", err)
            os.Exit(1)
        }

        for _, subdir := range subdirs {
            if subdir.IsDir() {
                dirPath := filepath.Join(downloadsDir, subdir.Name())
                log.Printf("🧹 Cleaning up project: %s", subdir.Name())

                // 1. "docker compose down" is the clean way. 
                // Adding -v removes volumes, --remove-orphans catches stragglers.
                downCmd := exec.Command("docker", "compose", "down", "-v", "--remove-orphans")
                downCmd.Dir = dirPath
                downCmd.Run() // We don't necessarily need to block for logs here

                // 2. Force removal if 'down' failed to stop them (The "Brute Force" clause)
                // We use sh -c here to handle the $(...) subshell
                forceRm := exec.Command("sh", "-c", "docker rm -f $(docker ps -a -q) 2>/dev/null || true")
                forceRm.Run()
            }
        }

        // 3. Final Network Cleanup
        log.Println("Removing shared network...")
        exec.Command("docker", "network", "rm", "shared_network").Run()

        log.Println("Cleanup finished. Goodbye!")
        os.Exit(0) 
    }()
}