package utils

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
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

func SanitizeURL(name string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '/' || r == ':' || r == '?' || r == '&' || r == '=' {
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

func RestartDockerCompose(repoPath string) error {

	if err := CleanProjectLock(repoPath); err != nil {
        log.Printf("Lock cleanup warning: %v", err)
    }

	downCmd := exec.Command("docker", "compose", "down", "--remove-orphans")
	downCmd.Dir = repoPath
	//downCmd.Stdout = os.Stdout
	downCmd.Stderr = os.Stderr
	if err := downCmd.Run(); err != nil {
		return fmt.Errorf("failed to down compose: %w", err)
	}

	// Clean recreate of traefik_internal
	exec.Command("docker", "network", "rm", "traefik_internal").Run()
	if out, err := exec.Command("docker", "network", "create", "traefik_internal").CombinedOutput(); err != nil {
		log.Printf("traefik_internal create: %s", string(out))
	}

	buildCmd := exec.Command("docker", "compose", "build")
	buildCmd.Dir = repoPath
	//buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build docker compose: %w", err)
	}

	upCmd := exec.Command("docker", "compose", "watch")
	upCmd.Dir = repoPath
	//upCmd.Stdout = os.Stdout
	upCmd.Stderr = os.Stderr
	if err := upCmd.Start(); err != nil {
		return fmt.Errorf("failed to start docker compose watch: %w", err)
	}

	// Wait for traefik container to actually exist before connecting networks.
	// compose watch is non-blocking so containers may not be created yet.
	log.Printf("Waiting for traefik container to be created...")
	var traefikContainerName string
	for i := 0; i < 30; i++ {
		name, err := getTraefikContainerName(repoPath)
		if err == nil && name != "" {
			traefikContainerName = name
			log.Printf("Traefik container found: %s", traefikContainerName)
			break
		}
		if i == 29 {
			return fmt.Errorf("timed out waiting for traefik container to be created")
		}
		log.Printf("Traefik container not ready yet (attempt %d/30): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	// Only connect to shared_network if traefik is present in the compose file
	aliases, err := getTraefikSharedNetworkAliases(repoPath)
	if err != nil {
		log.Printf("Warning: could not read traefik aliases: %v", err)
		aliases = []string{}
	}

	if len(aliases) > 0 {
		connectArgs := []string{"network", "connect"}
		for _, alias := range aliases {
			connectArgs = append(connectArgs, "--alias", alias)
		}
		connectArgs = append(connectArgs, "shared_network", traefikContainerName)

		if out, err := exec.Command("docker", connectArgs...).CombinedOutput(); err != nil {
			return fmt.Errorf("failed to connect traefik to shared_network: %w\n%s", err, string(out))
		}
		log.Printf("Connected traefik to shared_network with aliases: %v", aliases)
	}

	// Read compose file to determine which services to health-check
	composeFilePath := filepath.Join(repoPath, "docker-compose.yml")
	composeContent, err := ReadFileContent(composeFilePath)
	if err != nil {
		return fmt.Errorf("failed to read compose file: %w", err)
	}

	var yamlContent map[string]interface{}
	if err := yaml.Unmarshal([]byte(composeContent), &yamlContent); err != nil {
		return fmt.Errorf("failed to parse compose YAML: %w", err)
	}

	services, ok := yamlContent["services"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid compose file format: missing services")
	}

	expectedServices := make(map[string]int)
	for serviceName := range services {
		expectedServices[serviceName] = 1
	}

	if err := LoopUntilHealthy(repoPath, expectedServices, time.Now()); err != nil {
		return fmt.Errorf("failed to wait for healthy state: %w", err)
	}

	return nil
}


func CleanProjectLock(repoPath string) error {
    // Do "docker compose watch" to get the PID from the output.
	cmd := exec.Command("docker", "compose", "watch")
	cmd.Dir = repoPath
	out := &bytes.Buffer{}
	cmd.Stdout = out
	cmd.Stderr = out
	// We know it will give error, thats what we want.
	cmd.Run()
	
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		if strings.Contains(line, "PID") {
			// The output is something like this:
			// cannot take exclusive lock for project "ecommercesystem": process with PID 105 is still running
			parts := strings.Split(line, "PID")
			if len(parts) < 2 {
				continue
			}
			pidPart := strings.TrimSpace(parts[1])
			pidStr := strings.Split(pidPart, " ")[0]
			pid, err := strconv.Atoi(pidStr)
			if err != nil {
				log.Printf("Failed to parse PID from compose output: %v", err)
				continue
			}
			log.Printf("Found stale compose lock with PID %d. Attempting to kill process.", pid)
			proc, err := os.FindProcess(pid)
			if err != nil {
				log.Printf("Failed to find process with PID %d: %v", pid, err)
				continue
			}
			if err := proc.Kill(); err != nil {
				log.Printf("Failed to kill process with PID %d: %v", pid, err)
				continue
			}
			log.Printf("Successfully killed process with PID %d. Lock should be cleared now.", pid)
			break
		}
	}
	return nil	
}

func getTraefikContainerName(repoPath string) (string, error) {
    cmd := exec.Command("docker", "compose", "ps", "--format", "json")
    cmd.Dir = repoPath
    out, err := cmd.Output()
    if err != nil {
        return "", err
    }

    // Try unmarshaling as an array first (Modern Docker behavior)
    var containers []ComposeContainer
    if err := json.Unmarshal(out, &containers); err == nil {
        for _, c := range containers {
            if c.Service == "traefik" {
                return c.Name, nil
            }
        }
    }

    // Fallback: Try line-by-line (Older/Some environments)
    scanner := bufio.NewScanner(bytes.NewReader(out))
    for scanner.Scan() {
        var c ComposeContainer
        if err := json.Unmarshal(scanner.Bytes(), &c); err == nil {
            if c.Service == "traefik" {
                return c.Name, nil
            }
        }
    }
    return "", fmt.Errorf("traefik container not found")
}

// getTraefikSharedNetworkAliases reads the compose file and extracts the aliases
// that were intended for traefik's traefik_internal network (i.e. Traefik-fronted
// service names). These are replayed as --alias flags when connecting shared_network
// post-up, since the container can only be created with a single network.
func getTraefikSharedNetworkAliases(repoPath string) ([]string, error) {
	composeFilePath := filepath.Join(repoPath, "docker-compose.yml")
	composeContent, err := ReadFileContent(composeFilePath)
	if err != nil {
		return nil, err
	}

	var root map[string]interface{}
	if err := yaml.Unmarshal([]byte(composeContent), &root); err != nil {
		return nil, err
	}

	services, ok := root["services"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no services found")
	}

	traefik, ok := services["traefik"].(map[string]interface{})
	if !ok {
		// No traefik service — not an error, just nothing to do
		return nil, nil
	}

	networks, ok := traefik["networks"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	// Aliases are stored on traefik_internal by addAliasToTraefik.
	// They are replayed onto shared_network at runtime via network connect.
	internalNet, ok := networks["traefik_internal"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	rawAliases, exists := internalNet["aliases"]
	if !exists {
		return nil, nil
	}

	var aliases []string
	switch v := rawAliases.(type) {
	case []interface{}:
		for _, a := range v {
			aliases = append(aliases, fmt.Sprint(a))
		}
	case []string:
		aliases = v
	}

	return aliases, nil
}

// CheckComposeHealth runs `docker compose ps --format json`,
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

func LoopUntilHealthy(repoPath string, expectedServices map[string]int, mitigationStartTime time.Time) error {
	maxRetries := 15

	for i := 0; i < maxRetries; i++ {
		err := CheckComposeHealth(repoPath, expectedServices, mitigationStartTime)
		if err == nil {
			log.Println("Architecture is healthy!")
			return nil
		}

		switch err.(type) {
		case *ServiceUnhealthyError:
			log.Printf("Service unhealthy (attempt %d/%d): %v", i+1, maxRetries, err)
		case *ServiceNotStartedError:
			log.Printf("Service not started yet (attempt %d/%d): %v", i+1, maxRetries, err)
		default:
			log.Printf("Unknown error (attempt %d/%d): %v", i+1, maxRetries, err)
		}

		log.Printf("Retrying in 10 seconds...")
		time.Sleep(10 * time.Second)
	}

	return fmt.Errorf("failed to achieve healthy state after %d retries", maxRetries)
}