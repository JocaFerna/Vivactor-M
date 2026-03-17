package sharedLibraries
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)
// --- Types for JSON Parsing ---

type LibraryDetail struct {
	Library       string   `json:"library"`
	Microservices []string `json:"microservices"`
	Count         int      `json:"count"`
}

// SharedLibraryContext now matches the root of your JSON
type SharedLibraryContext struct {
	SharedLibraries map[string]LibraryDetail `json:"sharedLibraries"`
}

// --- Strategy Constants ---

type ExtractionStrategy string

const (
	SidecarStrategy ExtractionStrategy = "SIDECAR"
	GatewayStrategy ExtractionStrategy = "GATEWAY"
	Internal        ExtractionStrategy = "INTERNAL"
)

// --- Core Refactoring Function ---

func MitigateSharedLibrarySmells(repoName string, jsonData string) error {
    // DEBUG: Print the first 50 characters to see what we actually received
    if len(jsonData) > 50 {
        fmt.Printf("Debug JSON prefix: %s\n", jsonData[:50])
    } else {
        fmt.Printf("Debug JSON content: %s\n", jsonData)
    }

    var context SharedLibraryContext // Use the root struct we fixed earlier

    // Attempt to unmarshal
    err := json.Unmarshal([]byte(jsonData), &context)
    if err != nil {
        return fmt.Errorf("failed to parse JSON: %w", err)
    }

    // Check if we actually got data
    if len(context.SharedLibraries) == 0 {
        return fmt.Errorf("JSON parsed successfully but sharedLibraries map is empty")
    }

	cleanName := strings.TrimPrefix(repoName, "/")
	repoName = filepath.Join("/api/downloads", cleanName)

	fmt.Printf("Starting Mitigation for Repository: %s\n", repoName)
	fmt.Println(strings.Repeat("-", 40))

	// Iterate through the map
	for _, detail := range context.SharedLibraries {
		strategy := classifyLibrary(detail.Library)

		if strategy == Internal {
			fmt.Printf("[SKIP] %s is an internal utility. Strategy: Accept Redundancy.\n", detail.Library)
			continue
		}

		fmt.Printf("[ACT] %s identified for extraction. Strategy: %s\n", detail.Library, strategy)

		for _, msRaw := range detail.Microservices {
			parts := strings.Split(msRaw, ":")
			if len(parts) < 2 {
				continue
			}
			msName := parts[1]

			pomPath := filepath.Join(repoName, msName, "pom.xml")
			
			// Note: ensure removeDependencyFromPom is defined in your package
			err := removeDependencyFromPom(pomPath, detail.Library)
			if err != nil {
				fmt.Printf("   - Warning: Could not update POM for %s: %v\n", msName, err)
			} else {
				fmt.Printf("   - Success: Removed library from %s/pom.xml\n", msName)
			}
		}

		// Note: ensure generateSharedServiceManifest is defined in your package
		err = generateSharedServiceManifest(repoName, detail.Library, strategy)
		if err != nil {
			fmt.Printf("   - Error generating shared service: %v\n", err)
		}
	}

	return nil
}

// --- Helper Functions ---

// classifyLibrary determines if the library is "passible" to be a service
func classifyLibrary(lib string) ExtractionStrategy {
	lib = strings.ToLower(lib)
	if strings.Contains(lib, "mongodb") || strings.Contains(lib, "config") {
		return SidecarStrategy
	}
	if strings.Contains(lib, "security") || strings.Contains(lib, "oauth2") {
		return GatewayStrategy
	}
	return Internal
}

// removeDependencyFromPom strips the XML block from the target file
func removeDependencyFromPom(path string, fullLib string) error {
	input, err := os.ReadFile(path)
	if err != nil { return err }

	parts := strings.Split(fullLib, ":")
	group, artifact := parts[0], parts[1]

	lines := strings.Split(string(input), "\n")
	var output []string
	inTargetBlock := false
	tempBlock := []string{}

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		
		if strings.Contains(line, "<dependency>") {
			inTargetBlock = true
			tempBlock = append(tempBlock, line)
			continue
		}

		if inTargetBlock {
			tempBlock = append(tempBlock, line)
			if strings.Contains(line, "</dependency>") {
				// Check if the block we just collected matches our target
				blockStr := strings.Join(tempBlock, " ")
				if !strings.Contains(blockStr, group) || !strings.Contains(blockStr, artifact) {
					output = append(output, tempBlock...)
				}
				tempBlock = []string{}
				inTargetBlock = false
			}
			continue
		}
		output = append(output, line)
	}

	return os.WriteFile(path, []byte(strings.Join(output, "\n")), 0644)
}

// generateSharedServiceManifest creates the infrastructure config
func generateSharedServiceManifest(repo string, lib string, strategy ExtractionStrategy) error {
	dir := filepath.Join(repo, "extracted-services")
	os.MkdirAll(dir, 0755)

	filename := fmt.Sprintf("%s-config.yaml", strings.ReplaceAll(lib, ":", "-"))
	content := ""

	if strategy == SidecarStrategy {
		content = fmt.Sprintf("# Extracted Shared Service for %s\napiVersion: dapr.io/v1alpha1\nkind: Component\nmetadata:\n  name: shared-%s\nspec:\n  type: state.mongodb\n", lib, lib)
	} else {
		content = fmt.Sprintf("# Centralized Gateway Logic for %s\nroute: /auth\ntarget: shared-auth-service\n", lib)
	}

	return os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644)
}
