package utils

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
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

func RestartDockerComposeWithoutTraefik(repoPath string) error {
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

func CreateDatabaseFromNode(graphStruct graphparsing.Graph, node graphparsing.Node, basePath string, databaseOwner graphparsing.Node) (graphparsing.Node, graphparsing.Graph, error) {
	// Create a new database node for the service.
	sanitizedLabel := SanitizeName(node.Label)
	serviceDir := filepath.Join(basePath, sanitizedLabel)
	// Create the directory for the service if it doesn't exist
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		err := os.Mkdir(serviceDir, 0755)
		if err != nil {
			return graphparsing.Node{}, graphparsing.Graph{}, fmt.Errorf("error creating service directory: %v", err)
		}
	}
	dockerContent, _ := loadDockerfileTemplate("database", sanitizedLabel)	
	os.WriteFile(filepath.Join(serviceDir, "Dockerfile"), []byte(dockerContent), 0644)

	// Buid the docker compose entry for this node
	var servicesYaml strings.Builder
	servicesYaml.WriteString(fmt.Sprintf("  %s:\n", sanitizedLabel))
	servicesYaml.WriteString(fmt.Sprintf("    build: ./%s\n", sanitizedLabel))
	servicesYaml.WriteString("    networks:\n")
	servicesYaml.WriteString("      - shared_network\n")

	port := 5432
	servicesYaml.WriteString("    healthcheck:\n")
	servicesYaml.WriteString(fmt.Sprintf("      test: [\"CMD-SHELL\", \"nc -z localhost %d || exit 1\"]\n", port))
	servicesYaml.WriteString("      interval: 5s\n")
	servicesYaml.WriteString("      timeout: 5s\n")
	servicesYaml.WriteString("      retries: 5\n")

	watchAction := "rebuild"
	watchTarget := "/app"
	servicesYaml.WriteString("    develop:\n")
	servicesYaml.WriteString("      watch:\n")
	servicesYaml.WriteString(fmt.Sprintf("        - action: %s\n", watchAction))
	servicesYaml.WriteString(fmt.Sprintf("          path: ./%s\n", sanitizedLabel))
	servicesYaml.WriteString(fmt.Sprintf("          target: %s\n", watchTarget))
	servicesYaml.WriteString("\n")

	// Add the servicesYaml content to the compose file
	composeFilePath := filepath.Join(basePath, "docker-compose.yml")
	composeContent, err := ReadFileContent(composeFilePath)
	if err != nil {
		return graphparsing.Node{}, graphparsing.Graph{}, fmt.Errorf("error reading compose file: %v", err)
	}

	// Insert servicesYaml under the "services:" section
	lines := strings.Split(composeContent, "\n")
	var newComposeContent strings.Builder
	inserted := false
	for _, line := range lines {
		newComposeContent.WriteString(line + "\n")
		if strings.TrimSpace(line) == "services:" && !inserted {
			newComposeContent.WriteString(servicesYaml.String())
			inserted = true
		}
	}

	err = WriteFileContent(composeFilePath, newComposeContent.String())
	if err != nil {
		return graphparsing.Node{}, graphparsing.Graph{}, fmt.Errorf("error writing updated compose file: %v", err)
	}

	// Add the edge to the database owner
	newEdge := graphparsing.Edge{
		Source: databaseOwner.Id,
		Target: node.Id,
		Endpoint: "/db/owned",
		Properties: graphparsing.EdgeProperties{
			CallDefinitionInSource: "jdbc:postgresql://" + node.Label + ":5432/mydb",
			Method: "SQL",
		},
	}
	graphStruct.Edges = append(graphStruct.Edges, newEdge)

	// Create the connection in the file of the database owner.
	databaseOwner.Label = SanitizeName(databaseOwner.Label)
	err = handleCallToDBNode(&databaseOwner, &node, newEdge, strings.ToLower(databaseOwner.Properties.Language), basePath)
	if err != nil {
		return graphparsing.Node{}, graphparsing.Graph{}, fmt.Errorf("error handling DB call: %v", err)
	}

	return node,graphStruct, nil

}

// This function handles the special case of edges towards DatabaseNodes. Since these nodes might not have typical "main" files or incoming/outgoing call logic, we can represent the DB call by injecting environment variables or configuration files into the source node's service to represent the database connection details (e.g., host, port, credentials). This is a simplified representation and can be expanded based on specific requirements.
func handleCallToDBNode(sourceNode *graphparsing.Node, targetNode *graphparsing.Node, edge graphparsing.Edge, lang string, basePath string) error {
	// 1. Configuration: mapping languages to their actual code extensions
	extensions := map[string]string{
		"java":       ".java",
		"python":     ".py",
		"javascript": ".js",
		"golang":     ".go",
	}

	ext, ok := extensions[lang]
	if !ok {
		// If it's HTML or plain text, a DB call doesn't make sense, so we skip
		return nil
	}

	// 2. Read the existing main file content of the SOURCE service
	mainFilePath := filepath.Join(basePath, sourceNode.Label, "main"+ext)
	mainContent, err := os.ReadFile(mainFilePath)
	if err != nil {
		return fmt.Errorf("error reading source main file for DB call: %w", err)
	}

	// 3. Load the Database-specific template for the source language
	// File expected at: public/templates/[lang]/[lang].outgoing_call_db.template
	templatePath := filepath.Join("public", "templates", lang, lang+".outgoing_call_db.template")
	data, err := os.ReadFile(templatePath)
	if err != nil {
		log.Printf("Warning: DB template not found for %s at %s. Skipping DB call injection.", lang, templatePath)
		return nil
	}

	// 4. Manipulate the template content
	// We use the targetNode.Label as the hostname for the DB connection
	content := string(data)
	content = strings.ReplaceAll(content, "{{SERVICE_NAME}}", SanitizeName(sourceNode.Label))
	content = strings.ReplaceAll(content, "{{TARGET_LABEL}}", SanitizeName(targetNode.Label))
	
	// Default credentials based on your Postgres Dockerfile setup
	content = strings.ReplaceAll(content, "{{DB_NAME}}", "emulation_db")
	content = strings.ReplaceAll(content, "{{DB_USER}}", "user")
	content = strings.ReplaceAll(content, "{{DB_PASS}}", "pass")

	// 5. Inject the DB call into the outgoing calls placeholder
	newContent := string(mainContent)
	if lang != "python" {
		// Use JS/Java style comments
		newContent = strings.ReplaceAll(newContent, "//{{OUTGOING_CALLS}}", content+"\n//{{OUTGOING_CALLS}}")
	} else {
		// Use Python style comments
		newContent = strings.ReplaceAll(newContent, "#{{OUTGOING_CALLS}}", content+"\n#{{OUTGOING_CALLS}}")
	}

	// 6. Write the updated content back to the source service's main file
	err = os.WriteFile(mainFilePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing DB call to main file: %w", err)
	}

	log.Printf("Successfully injected DB call logic from %s to %s", sourceNode.Label, targetNode.Label)
	return nil
}



// Load the dockerfile template for the given language and replace placeholders
func loadDockerfileTemplate(lang string, serviceName string) (string, error) {
    templatePath := filepath.Join("public", "templates", lang, lang+".dockerfile.template")
    data, err := os.ReadFile(templatePath)
    if err != nil {
        return "", err
    }
    
    content := string(data)
    content = strings.ReplaceAll(content, "{{SERVICE_NAME}}", serviceName)
    return content, nil
}

func RemoveCallToDBNode(sourceNode *graphparsing.Node, targetNode *graphparsing.Node, basePath string) error {
	// 1. Configuration: mapping languages to their actual code extensions
	extensions := map[string]string{
		"java":       ".java",
		"python":     ".py",
		"javascript": ".js",
		"golang":     ".go",
	}

	ext, ok := extensions[strings.ToLower(sourceNode.Properties.Language)]
	if !ok {
		// If it's HTML or plain text, a DB call doesn't make sense, so we skip
		return nil
	}

	mainFilePath := filepath.Join(basePath, SanitizeName(sourceNode.Label), "main"+ext)
	mainContent, err := os.ReadFile(mainFilePath)
	if err != nil {
		return fmt.Errorf("error reading source main file for DB call removal: %w", err)
	}

	content := string(mainContent)

	var startMarker, endMarker string
	if ext == ".py" {
		startMarker = fmt.Sprintf("# BEGIN DB CALL TO %s", SanitizeName(targetNode.Label))
		endMarker = fmt.Sprintf("# END DB CALL TO %s", SanitizeName(targetNode.Label))
	} else {
		startMarker = fmt.Sprintf("// BEGIN DB CALL TO %s", SanitizeName(targetNode.Label))
		endMarker = fmt.Sprintf("// END DB CALL TO %s", SanitizeName(targetNode.Label))
	}
	
	startIdx := strings.Index(content, startMarker)
	endIdx := strings.Index(content, endMarker)
	if startIdx == -1 || endIdx == -1 || endIdx < startIdx {
		log.Printf("DB call markers not found for %s in %s. Skipping DB call removal.", targetNode.Label, sourceNode.Label)
		return nil
	}
	
	// Remove the content between the markers, including the markers themselves
	newContent := content[:startIdx] + content[endIdx+len(endMarker):]
	err = os.WriteFile(mainFilePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing updated main file after DB call removal: %w", err)
	}
	
	log.Printf("Successfully removed DB call logic from %s to %s", sourceNode.Label, targetNode.Label)
	return nil
}

func CreateBasicFileFromNode(node graphparsing.Node, basePath string) error {
	extensions := map[string]string{
		"java":       ".java",
		"python":     ".py",
		"javascript": ".js",
		"html":       ".html",
		"golang":     ".go",
	}
	var servicesYaml strings.Builder
	
	ext, ok := extensions[strings.ToLower(node.Properties.Language)]
	if !ok {
		return fmt.Errorf("unsupported language: %s", node.Properties.Language)
	}
	
	node.Label = SanitizeName(node.Label)
	serviceDir := filepath.Join(basePath, node.Label)
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		err := os.Mkdir(serviceDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating service directory: %v", err)
		}
	}

	// 1. Load template content for the node's language
	ext, content, err := loadAndProcessTemplate(node, strings.ToLower(node.Properties.Language))
	if err != nil {
		return fmt.Errorf("error loading template: %v", err)
	}

	//2. Write the main entry file
	mainFilePath := filepath.Join(serviceDir, "main"+ext)
	err = os.WriteFile(mainFilePath, []byte(content), 0644)

	//3. Generate magnitude files.
	generateMagnitudeFiles(serviceDir, node.Properties.OrderOfMagnitudeOfFiles, ext)


	//4. Load and write Dockerfile template
	dockerContent, err := loadDockerfileTemplate(strings.ToLower(node.Properties.Language), node.Label)
	if err != nil {
		log.Printf("Warning: Dockerfile template not found for %s. Skipping Dockerfile creation.", node.Properties.Language)
	} else {
		dockerFilePath := filepath.Join(serviceDir, "Dockerfile")
		os.WriteFile(dockerFilePath, []byte(dockerContent), 0644)
	}
	// 5. Build the docker compose entry for this node
	servicesYaml.WriteString(fmt.Sprintf("  %s:\n", node.Label))
	servicesYaml.WriteString(fmt.Sprintf("    build: ./%s\n", node.Label))
	servicesYaml.WriteString("    networks:\n")
	servicesYaml.WriteString("      - shared_network\n")

	// Optional: Add environment variables for edges
	if node.Properties.Language == "python" || node.Properties.Language == "javascript" {
		servicesYaml.WriteString("    environment:\n")
		servicesYaml.WriteString(fmt.Sprintf("      - SERVICE_NAME=%s\n", node.Label))
	}
	var port int
	if node.Properties.Port != "" {
		port, _ = strconv.Atoi(node.Properties.Port)
	} else if node.Type == "DatabaseNode" {
		port = 5432
	} else if node.Properties.Language == "html" {
		port = 80
	} else {
		port = 8080
	}

	// Check if port is valid, otherwise skip healthcheck
	if port <= 0 || port > 65535 {
		return fmt.Errorf("Invalid port %d for service %s, skipping healthcheck\n", port, node.Label)
	}

	//TODO: Lacks checkPortMatch.

	servicesYaml.WriteString("    healthcheck:\n")
	servicesYaml.WriteString(fmt.Sprintf("      test: [\"CMD-SHELL\", \"nc -z localhost %d || exit 1\"]\n", port))
	servicesYaml.WriteString("      interval: 5s\n")
	servicesYaml.WriteString("      timeout: 5s\n")
	servicesYaml.WriteString("      retries: 5\n")

	// 5.5. Generate the Docker Compose "Watch" block (develop)
	lang := strings.ToLower(node.Properties.Language)
	var watchAction, watchTarget, watchIgnore string

	switch lang {
		case "html":
			watchAction = "sync"
			watchTarget = "/usr/share/nginx/html"
		case "python":
			watchAction = "sync+restart"
			watchTarget = "/app"
			// Standardized indentation: 12 spaces for each list item
			watchIgnore = "            - \"**/__pycache__/**\"\n            - \"**/.pytest_cache/**\""
		case "javascript", "js":
			watchAction = "sync+restart"
			watchTarget = "/app"
			watchIgnore = "            - \"**/node_modules/**\""
		case "java":
			watchAction = "sync+restart"
			watchTarget = "/app"
		default:
			watchAction = "rebuild"
			watchTarget = "/app"
	}

	// Write the develop/watch section to the YAML buffer
	servicesYaml.WriteString("    develop:\n")
	servicesYaml.WriteString("      watch:\n")
	servicesYaml.WriteString(fmt.Sprintf("        - action: %s\n", watchAction))
	servicesYaml.WriteString(fmt.Sprintf("          path: ./%s\n", node.Label))
	servicesYaml.WriteString(fmt.Sprintf("          target: %s\n", watchTarget))

	if watchIgnore != "" {
		// We write the key 'ignore:' and then a newline before pasting the items
		servicesYaml.WriteString("          ignore:\n")
		servicesYaml.WriteString(fmt.Sprintf("%s\n", watchIgnore))
	}
	servicesYaml.WriteString("\n")

	// 6. Add the servicesYaml content to the compose file
	composeFilePath := filepath.Join(basePath, "docker-compose.yml")
	composeContent, err := ReadFileContent(composeFilePath)
	if err != nil {
		return fmt.Errorf("error reading compose file: %v", err)
	}

	// Insert servicesYaml under the "services:" section
	lines := strings.Split(composeContent, "\n")
	var newComposeContent strings.Builder
	inserted := false
	for _, line := range lines {
		newComposeContent.WriteString(line + "\n")
		if strings.TrimSpace(line) == "services:" && !inserted {
			newComposeContent.WriteString(servicesYaml.String())
			inserted = true
		}
	}

	err = WriteFileContent(composeFilePath, newComposeContent.String())
	if err != nil {
		return fmt.Errorf("error writing updated compose file: %v", err)
	}
	return err
}



func HandleCallFromNode(sourceNode graphparsing.Node, targetNode graphparsing.Node, edge graphparsing.Edge, basePath string) error {
	// Guarantee that label's are sanitized for file paths
	sourceNode.Label = SanitizeName(sourceNode.Label)
	targetNode.Label = SanitizeName(targetNode.Label)
	
	if sourceNode.Type != "DatabaseNode" && targetNode.Type != "DatabaseNode" {
		// Find the language of the source node to determine how to represent the call
		lang_source := strings.ToLower(sourceNode.Properties.Language)
		lang_target := strings.ToLower(targetNode.Properties.Language)

		// Insert the outgoing call into source node's main file
		err := insertOutgoingCall(&sourceNode, &targetNode, edge, lang_source, basePath)
		if err != nil {
			log.Printf("Error inserting outgoing call from %s to %s: %v", sourceNode.Label, targetNode.Label, err)
			return err
		}

		// Insert the incoming call into target node's main file
		err = insertIncomingCall(&targetNode, &sourceNode, edge, lang_target, basePath)
		if err != nil {
			log.Printf("Error inserting incoming call to %s from %s: %v", targetNode.Label, sourceNode.Label, err)
			return err
		}
	}
	return nil
}
// Insert the incoming call representation into the target node's main file
func insertIncomingCall(targetNode *graphparsing.Node, sourceNode *graphparsing.Node, edge graphparsing.Edge, lang string, basePath string) error {
	// Configuration: mapping languages to their actual code extensions
	extensions := map[string]string{
		"java":       ".java",
		"python":     ".py",
		"javascript": ".js",
		"html":       ".html",
		"golang":     ".go",
	}
	

	ext, ok := extensions[lang]
	if !ok {
		ext = ".txt"
	}

	if ext == ".html" || ext == ".txt" {
		// For non-code files, we won't insert incoming call logic
		return nil
	}
	// Read the existing main file content
	mainFilePath := filepath.Join(basePath, targetNode.Label, "main"+ext)
	mainContent, err := os.ReadFile(mainFilePath)
	if err != nil {
		return fmt.Errorf("error reading main file: %w", err)
	}
	

	// Read from your public/templates folder
	templatePath := filepath.Join("public", "templates",lang, lang+".incoming_call.template")
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}

	// Get safe name for the endpoint (e.g., replace slashes with underscores)
	safeName := strings.ReplaceAll(edge.Endpoint, "/", "_")
	safeName = strings.ReplaceAll(safeName, "-", "") // Remove leading/trailing hyphens


	// Manipulate the template content
	content := string(data)
	content = strings.ReplaceAll(content, "{{ENDPOINT}}", edge.Endpoint)
	content = strings.ReplaceAll(content, "{{SAFE_NAME}}", safeName)
	content = strings.ReplaceAll(content, "{{SERVICE_NAME}}", targetNode.Label)

	// Append the outgoing call content to the main file
	newContent := string(mainContent)
	if lang != "python" {
		newContent = strings.ReplaceAll(newContent, "//{{CUSTOM_ROUTES}}", content+"\n//{{CUSTOM_ROUTES}}") // Insert before the placeholder
	} else {
		newContent = strings.ReplaceAll(newContent, "#{{CUSTOM_ROUTES}}", content+"\n#{{CUSTOM_ROUTES}}") // Insert before the placeholder
	}

	// Write the updated content back to the main file
	err = os.WriteFile(mainFilePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing main file: %w", err)
	}
	return nil
}
// Insert the outgoing call representation into the source node's main file
func insertOutgoingCall(sourceNode *graphparsing.Node, targetNode *graphparsing.Node, edge graphparsing.Edge, lang string, basePath string) error {
	// Configuration: mapping languages to their actual code extensions
	extensions := map[string]string{
		"java":       ".java",
		"python":     ".py",
		"javascript": ".js",
		"html":       ".html",
		"golang":     ".go",
	}

	ext, ok := extensions[lang]
	if !ok {
		ext = ".txt"
	}


	if ext == ".html" || ext == ".txt" {
		// For non-code files, we won't insert incoming call logic
		return nil
	}

	// Read the existing main file content
	mainFilePath := filepath.Join(basePath, sourceNode.Label, "main"+ext)
	mainContent, err := os.ReadFile(mainFilePath)
	if err != nil {
		return fmt.Errorf("error reading main file: %w", err)
	}
	

	// Read from your public/templates folder
	templatePath := filepath.Join("public", "templates",lang, lang+".outgoing_call.template")
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}

	// Sanitize the call definition from hyphens
	callDefSafe := strings.ReplaceAll(edge.Properties.CallDefinitionInSource, "-", "")
	// Manipulate the template content
	content := string(data)
	content = strings.ReplaceAll(content, "{{SERVICE_NAME}}", sourceNode.Label)
	content = strings.ReplaceAll(content, "{{TARGET_LABEL}}", targetNode.Label)
	content = strings.ReplaceAll(content, "[{{CALL_DEFINITION}}]", callDefSafe)

	// Append the outgoing call content to the main file
	newContent := string(mainContent)
	if lang != "python" {
		newContent = strings.ReplaceAll(newContent, "//{{OUTGOING_CALLS}}", content+"\n//{{OUTGOING_CALLS}}") // Insert before the placeholder
	} else {
		newContent = strings.ReplaceAll(newContent, "#{{OUTGOING_CALLS}}", content+"\n#{{OUTGOING_CALLS}}") // Insert before the placeholder
	}

	// Write the updated content back to the main file
	err = os.WriteFile(mainFilePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing main file: %w", err)
	}
	return nil
}

func generateMagnitudeFiles(dir string, magnitudeStr string, ext string) {
	parts := strings.Split(magnitudeStr, "^")
	if len(parts) < 2 { return }
	
	exponent, _ := strconv.Atoi(parts[1])
	count := int(math.Pow(10, float64(exponent)))

	for i := 1; i < count; i++ {
		name := filepath.Join(dir, fmt.Sprintf("module_%d%s", i, ext))
		_ = os.WriteFile(name, []byte("// Supporting component"), 0644)
	}
}

func loadAndProcessTemplate(node graphparsing.Node, lang string) (string, string, error) {
	// Configuration: mapping languages to their actual code extensions
	extensions := map[string]string{
		"java":       ".java",
		"python":     ".py",
		"javascript": ".js",
		"html":       ".html",
		"golang":     ".go",
	}

	ext, ok := extensions[lang]
	if !ok {
		ext = ".txt"
	}

	// Read from your public/templates folder
	templatePath := filepath.Join("public", "templates",lang, lang+".template")
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return ext, "", err
	}

	// Manipulate the template content
	content := string(data)
	// Inside your Go code:
	content = strings.ReplaceAll(content, "{{SERVICE_NAME}}", node.Label)
	content = strings.ReplaceAll(content, "{{LANGUAGE}}", node.Properties.Language)
	content = strings.ReplaceAll(content, "{{NODE_ID}}", string(node.Id))
	content = strings.ReplaceAll(content, "{{MAGNITUDE}}", node.Properties.OrderOfMagnitudeOfFiles)
	if node.Properties.Port != "" {
		content = strings.ReplaceAll(content, "{{PORT}}", node.Properties.Port)
	} else {
		content = strings.ReplaceAll(content, "{{PORT}}", "8080") // Default port if not specified
	}
	return ext, content, nil
}