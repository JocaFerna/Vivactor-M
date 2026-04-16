package emulation

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
)
func RestartEmulationWithNewGraph(graph string) error {
	// First, we need to stop the current emulation if there is one running.
	// We can do this by killing any existing "docker compose watch" processes related to the current project.
	// To find the project, we can parse the graph to get the system name, which corresponds to the folder name in /api/downloads.
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return fmt.Errorf("error parsing json: %w", err)
	}

	// Clean repository name
	repoNameRaw := graphStruct.System.Name
	repoNameClean := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, repoNameRaw)

	graphStruct.System.Name = repoNameClean

	basePath := filepath.Join("/api/downloads", repoNameClean)

	// In case that there is some process "docker compose watch" running from a previous emulation, we should kill it before starting a new one.
	err = CleanProjectLock(basePath)
	if err != nil {
		return fmt.Errorf("error cleaning project lock: %w", err)
	}

	// Also, we must shutdown the existent containers.
	cmd := exec.Command("docker", "compose", "down", "--remove-orphans")
	cmd.Dir = basePath
	cmd.Run()

	// Now we can start the new emulation with the new graph.
	err = EmulateArchitecture(graph)
	if err != nil {
		return fmt.Errorf("error starting new emulation: %w", err)
	}
	return nil
}



// Given a graph in JSON format (stringified), this function will create a folder structure with files and a docker-compose.yml to emulate the architecture
func EmulateArchitecture(graph string) error {
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return fmt.Errorf("error parsing json: %w", err)
	}

	// Clean repository name
	repoNameRaw := graphStruct.System.Name
	repoNameClean := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, repoNameRaw)

	graphStruct.System.Name = repoNameClean

	basePath := filepath.Join("/api/downloads", repoNameClean)
	// Remove any existent files in that folder
	os.RemoveAll(basePath)
	os.MkdirAll(basePath, 0755)

	// In case that there is some process "docker compose watch" running from a previous emulation, we should kill it before starting a new one.
	err = CleanProjectLock(basePath)
	if err != nil {
		return fmt.Errorf("error cleaning project lock: %w", err)
	}
	

	// This will hold the content for the docker-compose.yml file
	var servicesYaml strings.Builder

	// Service Names needed for health checks later
	expectedServices := make(map[string]int)

	for _, node := range graphStruct.Nodes {
		// Name sanitization: remove spaces and special characters for folder names
		node.Label = sanitizeString(node.Label)

		expectedServices[node.Label] = 0

		serviceDir := filepath.Join(basePath, node.Label)
		os.MkdirAll(serviceDir, 0755)
		if node.Type == "DatabaseNode"{
			// Database nodes can be emulated with a simple Dockerfile that uses an official image (e.g., MySQL, PostgreSQL)
			dockerContent, _ := loadDockerfileTemplate("database", node.Label)
        	os.WriteFile(filepath.Join(serviceDir, "Dockerfile"), []byte(dockerContent), 0644)
		}else{

			// 1. Load and manipulate the template
			lang := strings.ToLower(node.Properties.Language)
			log.Println("Language: %s", lang)

			ext, content, err := loadAndProcessTemplate(node, lang)
			if err != nil {
				// Fallback if template doesn't exist
				ext = ".txt"
				content = fmt.Sprintf("Fallback for service: %s", node.Label)
			}

			// 2. Write the main entry file
			mainFilePath := filepath.Join(serviceDir, "main"+ext)
			os.WriteFile(mainFilePath, []byte(content), 0644)

			// 3. Generate magnitude files (10^n)
			generateMagnitudeFiles(serviceDir, node.Properties.OrderOfMagnitudeOfFiles, ext)
			
			// 4. Load and write Dockerfile if template exists
			dockerContent, err := loadDockerfileTemplate(lang, node.Label)
			if err == nil {
				dockerFilePath := filepath.Join(serviceDir, "Dockerfile")
				os.WriteFile(dockerFilePath, []byte(dockerContent), 0644)
			}
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
		
		// TODO: This port configuration is TOO much basic and we will maybe need to refactor
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

		// Also, check if the port matches any potentially edges towards that node.
		check := checkPortMatch(node, graphStruct.Edges)
		if !check {
			log.Printf("Warning: Port %d for service %s does not match any incoming edges. Healthcheck might fail.\n", port, node.Label)
		}

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
	}

	// 6. Generate the root docker-compose.yml file
	err = generateRootCompose(basePath, servicesYaml.String())
	if err != nil {
		return fmt.Errorf("error generating docker-compose.yml: %w", err)
	}

	// 7. Iterate through edges to add network connections
	for _, edge := range graphStruct.Edges {
		source := edge.Source
		target := edge.Target

		// 8. Find the corresponding nodes for source and target
		sourceNode := findNodeByID(graphStruct.Nodes, source)
		targetNode := findNodeByID(graphStruct.Nodes, target)


		// Sanitize the labels
		if sourceNode != nil {
			sourceNode.Label = sanitizeString(sourceNode.Label)
		}
		if targetNode != nil {
			targetNode.Label = sanitizeString(targetNode.Label)
		}

		if sourceNode != nil && targetNode != nil {

			if sourceNode.Type != "DatabaseNode" && targetNode.Type != "DatabaseNode" {
				// 9. Find the language of the source node to determine how to represent the call
				lang_source := strings.ToLower(sourceNode.Properties.Language)
				lang_target := strings.ToLower(targetNode.Properties.Language)

				// 10. Insert the outgoing call into source node's main file
				err := insertOutgoingCall(sourceNode, targetNode, edge, lang_source, basePath)
				if err != nil {
					log.Printf("Error inserting outgoing call from %s to %s: %v", sourceNode.Label, targetNode.Label, err)
				}

				// 11. Insert the incoming call into target node's main file
				err = insertIncomingCall(targetNode, sourceNode, edge, lang_target, basePath)
				if err != nil {
					log.Printf("Error inserting incoming call to %s from %s: %v", targetNode.Label, sourceNode.Label, err)
				}
			} else if targetNode.Type == "DatabaseNode" && sourceNode.Type != "DatabaseNode" {
				// NEW: Handle Database Connection
				lang_source := strings.ToLower(sourceNode.Properties.Language)
				err := handleCallToDBNode(sourceNode, targetNode, edge, lang_source, basePath)
				if err != nil {
					log.Printf("Error handling DB call: %v", err)
				}
			}
			
		}
	}

	// 12. Build and start the Docker containers
	log.Println("Building and starting Docker containers...")
	log.Println("Base path for emulation: ", basePath)
    process, err := startDockerCompose(basePath)
    if err != nil {
        return fmt.Errorf("failed to start architecture: %w", err)
    }

	// 13. Check health status of container.
	// This works by running `docker compose ps --format json` and parsing the output to verify that all expected services are running and healthy. We will retry this check several times with a delay in between, to give the containers time to start up and become healthy.
	maxRetries := 30
	var waitErr error

	for i := 0; i < maxRetries; i++ {
		err := checkComposeHealth(basePath, expectedServices)
		var nonHealthy = false

		if err == nil {
			// Save serviceNames into global value
			log.Println("Architecture is healthy!")
			waitErr = nil
			break
		}

		switch err.(type) {

		case *ServiceUnhealthyError:
			nonHealthy = true
			log.Printf("Service unhealthy: %v", err)

		case *ServiceNotStartedError:
			log.Printf("Service not started yet: %v", err)

		default:
			log.Printf("Unknown error: %v", err)
		}
		if !nonHealthy {
			log.Printf("Waiting for services to start... (attempt %d/%d)\n", i+1, maxRetries)
			// Wait for 10 seconds before retrying
			time.Sleep(10 * time.Second)
			waitErr = err
		} else {
			log.Printf("Services not healthy!")
			//IDEA: This could need a "break" here but I don't want to cause any other bugs now.
		}
		
	}

	if waitErr != nil {
		log.Printf("Wait failed after retries: %s. Killing watch process.\n", waitErr.Error())
		process.Process.Kill()
		return waitErr
	}

    log.Println("Architecture is successfully running!")
    return nil
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
	content = strings.ReplaceAll(content, "{{SERVICE_NAME}}", sourceNode.Label)
	content = strings.ReplaceAll(content, "{{TARGET_LABEL}}", targetNode.Label)
	
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
// Sanitize a string
func sanitizeString(input string) string {
	input = strings.ToLower(input)
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, input)
}

func CleanProjectLock(repoPath string) error {

	// First, let's check if there is a docker compose process running.
	// We check for 'docker compose' specifically to see if a 
    // management process is active.
    cmd := exec.Command("pgrep", "-f", "docker")
    err := cmd.Run()

	if err != nil {
		// No docker-related processes found, so no lock to clean
		log.Println("No docker processes found, no lock to clean.")
		return nil
	}



    // Do "docker compose watch" to get the PID from the output.
	cmd = exec.Command("docker", "compose", "watch")
	cmd.Dir = repoPath
	out := &bytes.Buffer{}
	cmd.Stdout = out
	cmd.Stderr = out

	
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

// Structure matching docker compose ps --format json output
type ComposeContainer struct {
	Service string `json:"Service"`
	Name    string `json:"Name"`
	State   string `json:"State"`
	Health  string `json:"Health"`
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
func checkComposeHealth(repoPath string, expectedServices map[string]int) error {

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

// Find the node based on its ID
func findNodeByID(nodes []graphparsing.Node, id string) *graphparsing.Node {
	for _, node := range nodes {
		if node.Id == id {
			return &node
		}
	}
	return nil
}

// Starts the architecture emulation by running the needed commands.
func startDockerCompose(basePath string) (*exec.Cmd, error) {
	fmt.Println("Reading docker-compose.yml from: ", basePath)
	// 1. Run 'docker compose build'
	buildCmd := exec.Command("docker", "compose", "build")
	buildCmd.Dir = basePath
	buildCmd.Stdout = os.Stdout // Pipe output to see build progress
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return nil, fmt.Errorf("docker compose build failed: %w", err)
	}

	
	// 2. Run 'docker compose watch' to start containers and wait for health
	upCmd := exec.Command("docker", "compose", "watch") // --wait ensures it waits for healthy status if healthchecks are defined
	upCmd.Dir = basePath
	upCmd.Stdout = os.Stdout
	upCmd.Stderr = os.Stderr
	if err := upCmd.Start(); err != nil {
		return nil, fmt.Errorf("docker compose up failed: %w", err)
	}

	return upCmd, nil
}

// Generate the root docker-compose.yml file using the services block
func generateRootCompose(basePath string, servicesBlock string) error {
	templatePath := filepath.Join("public", "templates","docker", "docker-compose.yml.template")
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}

	content := strings.ReplaceAll(string(data), "# {{SERVICES_BLOCK}}", servicesBlock)
	
	composePath := filepath.Join(basePath, "docker-compose.yml")
	return os.WriteFile(composePath, []byte(content), 0644)
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

func checkPortMatch(node graphparsing.Node, edges []graphparsing.Edge) bool {
	for _, edge := range edges {
		if edge.Target == node.Id {
			// Check if any of the call definitions in the source node's outgoing calls contains the port number
			callDef := edge.Properties.CallDefinitionInSource
			if strings.Contains(callDef, node.Properties.Port) {
				return true
			}
		}
	}
	return false
}	