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

	// 7. Build and start the Docker containers
	log.Println("Building and starting Docker containers...")
    err = startDockerCompose(basePath)
    if err != nil {
        return fmt.Errorf("failed to start architecture: %w", err)
    }

    log.Println("Architecture is successfully running!")
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

