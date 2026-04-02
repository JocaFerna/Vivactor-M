package utils

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func GetFilesInDirectory(dirPath string) ([]string, error) {
	var files []string
	
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Do not consider files in "target" and "test" directories
		// Only consider java and yaml files (config files)
		if !info.IsDir() && !strings.Contains(path, "/target/") && !strings.Contains(path, "/test/") {
			files = append(files, path)
		}
		return nil
	},
	)
	if err != nil {
		return nil, fmt.Errorf("error walking the path %q: %v", dirPath, err)
	}
	return files, nil
}

func ReadFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file %s: %v", filePath, err)
	}
	return string(content), nil
}

func WriteFileContent(filePath string, content string) error {
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("error writing to file %s: %v", filePath, err)
	}
	return nil
}

// Given a graph, it returns the filepath
func GetBasePathOfGraph(graphStruct graphparsing.Graph) string {
	// Get repo
	repoNameRaw := graphStruct.System.Name
	repoNameClean := SanitizeName(repoNameRaw)

	graphStruct.System.Name = repoNameClean

	basePath := filepath.Join("/api/downloads", repoNameClean)
	return basePath
}

// Given a node, it returns the file path of the node, if it is a file node.
func GetFilePathFromNode(graphStruct graphparsing.Graph, nodeId string) (string, error) {
	node, err := graphparsing.GetNodeById(graphStruct, nodeId)
	if err != nil {
		return "", fmt.Errorf("error getting node by id %s: %v", nodeId, err)
	}
	// Get repo
	repoNameRaw := graphStruct.System.Name
	repoNameClean := SanitizeName(repoNameRaw)

	graphStruct.System.Name = repoNameClean

	basePath := filepath.Join("/api/downloads", repoNameClean)

	basePath = filepath.Join(basePath, SanitizeName(node.Label))
	
	// Configuration: mapping languages to their actual code extensions
	extensions := map[string]string{
		"java":       ".java",
		"python":     ".py",
		"javascript": ".js",
		"html":       ".html",
		"golang":     ".go",
	}
	
	// Get the extension for the node's language
	extension, ok := extensions[strings.ToLower(node.Properties.Language)]
	if !ok {
		return "", fmt.Errorf("unsupported language: %s", node.Properties.Language)
	}
	return filepath.Join(basePath, "main"+extension), nil
}

// SanitizeName removes all characters from the name that are not letters or numbers, to create a valid file path.
func SanitizeName(name string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, name)
}
// Structure matching docker compose ps --format json output
type ComposeContainer struct {
    Service   string `json:"Service"`
    Name      string `json:"Name"`
    State     string `json:"State"`
    Health    string `json:"Health"`
    StartedAt string `json:"StartedAt"` // Add this!
}

// Error when a service is not started
type ServiceNotStartedError struct {
	Service string
	State   string
}

func (e *ServiceNotStartedError) Error() string {
	return fmt.Sprintf("service %s not started (state: %s)", e.Service, e.State)
}

// Error when a service is unhealthy
type ServiceUnhealthyError struct {
	Service string
	Health  string
}
func (e *ServiceUnhealthyError) Error() string {
	return fmt.Sprintf("service %s unhealthy (health: %s)", e.Service, e.Health)
}

// checkComposeHealth runs `docker compose ps --format json`,
// parses the output line-by-line, and verifies that all expected
// services are running and healthy.
func CheckComposeHealth(repoPath string, expectedServices map[string]int, mitigationStartTime time.Time) error {

	cmd := exec.Command("docker", "compose", "ps", "--format", "json")
	cmd.Dir = repoPath

	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run docker compose ps: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))

	foundServices := map[string]bool{}

	for scanner.Scan() {

		var container ComposeContainer
		line := scanner.Bytes()

		if err := json.Unmarshal(line, &container); err != nil {
			return fmt.Errorf("failed to parse compose JSON: %w", err)
		}

		foundServices[container.Service] = true

		// Check if container is running
		if container.State != "running" || container.Health == "starting" {
			return &ServiceNotStartedError{
				Service: container.Service,
				State:   container.State,
			}
		}

		// Check health status if healthcheck exists
		if container.Health != "" && container.Health != "healthy" {
			return &ServiceUnhealthyError{
				Service: container.Service,
				Health:  container.Health,
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Verify all expected services appeared
	for service := range expectedServices {
		if !foundServices[service] {
			return &ServiceNotStartedError{
				Service: service,
				State:   "missing",
			}
		}
	}

	return nil
}
