package emulation

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"log"
	"os/exec"
)


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

	// This will hold the content for the docker-compose.yml file
	var servicesYaml strings.Builder

	for _, node := range graphStruct.Nodes {
		// Name sanitization: remove spaces and special characters for folder names
		node.Label = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
		}, node.Label)

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
		servicesYaml.WriteString(fmt.Sprintf("    container_name: %s\n", node.Label))
		servicesYaml.WriteString("    networks:\n")
		servicesYaml.WriteString("      - shared_network\n")
		
		// Optional: Add environment variables for edges
		if node.Properties.Language == "python" || node.Properties.Language == "javascript" {
			servicesYaml.WriteString("    environment:\n")
			servicesYaml.WriteString(fmt.Sprintf("      - SERVICE_NAME=%s\n", node.Label))
		}
		
		// TODO: This port configuration is TOO much basic and we will maybe need to refactor
		var port int
		if node.Type == "DatabaseNode" {
			port = 5432
		} else if node.Properties.Language == "html" {
			port = 80
		} else {
			port = 8080
		}

		servicesYaml.WriteString("    healthcheck:\n")
		servicesYaml.WriteString(fmt.Sprintf("      test: [\"CMD-SHELL\", \"nc -z localhost %d || exit 1\"]\n", port))
		servicesYaml.WriteString("      interval: 5s\n")
		servicesYaml.WriteString("      timeout: 5s\n")
		servicesYaml.WriteString("      retries: 5\n")
		

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
			}
			// TODO: For now, we are ignoring edges from/to DatabaseNodes in terms of call representation, as they might not have "main" files or typical incoming/outgoing call logic.
		}
	}

	// 12. Build and start the Docker containers
	log.Println("Building and starting Docker containers...")
    err = startDockerCompose(basePath)
    if err != nil {
        return fmt.Errorf("failed to start architecture: %w", err)
    }

    log.Println("Architecture is successfully running!")
    return nil
}

// Sanitize a string
func sanitizeString(input string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, input)
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
	safeName = strings.Trim(safeName, "_") // Remove leading/trailing underscores

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
func startDockerCompose(basePath string) error {
	// 1. Run 'docker compose build'
	buildCmd := exec.Command("docker", "compose", "build")
	buildCmd.Dir = basePath
	//buildCmd.Stdout = os.Stdout // Pipe output to see build progress
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("docker compose build failed: %w", err)
	}

	// 2. Run 'docker compose up -d'
	// We use -d (detached) so the Go orchestrator doesn't hang 
    // while waiting for the microservices to finish.
	// We also add --wait to ensure it waits for healthy status if healthchecks are defined in the Dockerfiles.
	upCmd := exec.Command("docker", "compose", "up", "-d", "--wait") // --wait ensures it waits for healthy status if healthchecks are defined
	upCmd.Dir = basePath
	//upCmd.Stdout = os.Stdout
	upCmd.Stderr = os.Stderr
	if err := upCmd.Run(); err != nil {
		return fmt.Errorf("docker compose up failed: %w", err)
	}

	return nil
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

