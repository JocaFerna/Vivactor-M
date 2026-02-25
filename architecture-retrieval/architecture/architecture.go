package architecture

import (
	//"context"
	//"log"
	//"fmt"
	//gogithub "github.com/google/go-github/v65/github"
	//"io"
	"os"
	//"path/filepath"
	//"architecture-retrieval/hashset"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5"
	"gopkg.in/yaml.v3"
)

// ComposeConfig now stores services with their internal ports
type ComposeConfig struct {
	Services map[string]ServiceConfig `yaml:"services"`
}

// ServiceConfig stores ports/expose for a service
type ServiceConfig struct {
	Ports  []string `yaml:"ports"`
	Expose []string `yaml:"expose"`
}

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

// getServiceNames parses a compose file and returns service configs
func getServiceNames(filePath string) (ComposeConfig, error) {
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return ComposeConfig{}, fmt.Errorf("could not read file: %w", err)
	}

	var config ComposeConfig
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return ComposeConfig{}, fmt.Errorf("could not unmarshal yaml: %w", err)
	}

	// Log service names and their internal ports
	for name, svc := range config.Services {
		internalPorts := getInternalPorts(svc)
		log.Printf("Service: %s, Internal Ports: %v\n", name, internalPorts)
	}

	return config, nil
}

// getInternalPorts extracts internal ports from a ServiceConfig
func getInternalPorts(svc ServiceConfig) []string {
	var ports []string

	// First check 'ports' section
	for _, p := range svc.Ports {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// Format can be host:container or just container
		parts := strings.Split(p, ":")
		internal := parts[len(parts)-1]
		ports = append(ports, internal)
	}

	// Then check 'expose' section if no ports found
	for _, e := range svc.Expose {
		e = strings.TrimSpace(e)
		if e != "" {
			ports = append(ports, e)
		}
	}

	return ports
}

// retrieveServicesFromComposeFile now returns map[serviceName]internalPort
func retrieveServicesFromComposeFile(repoPath string) (map[string][]string, error) {
	var files []string

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}

		if !info.IsDir() && (filepath.Base(path) == "docker-compose.yml" || filepath.Base(path) == "docker-compose.yaml" || filepath.Base(path) == "compose.yaml" || filepath.Base(path) == "compose.yml") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	services := make(map[string][]string)

	for _, file := range files {
		log.Printf("Found compose file: %s\n", file)

		// Also, add networks: default: external: true name: shared_network to the NORMAL compose file to ensure that the services can connect to the shared network. We will do this by creating a docker-compose.override.yml file with the necessary network configuration and ensuring that it is included in the commands we run to start the architecture.
		


		composeConfig, err := getServiceNames(file)
		if err != nil {
			log.Printf("Error parsing %s: %s\n", file, err.Error())
			continue
		}

		// Merge services (last one wins)
		for name, svc := range composeConfig.Services {
			internalPorts := getInternalPorts(svc)
			if len(internalPorts) > 0 {
				services[name] = internalPorts
			}
		}
	}
	log.Println("All services with their internal ports:", services)
	return services, nil
}

// addSharedNetworkToComposeFiles updates all docker-compose files in the repo
// - adds `shared_network` to every service
// - clears host ports (keeps only internal ports in override if needed)
// - ensures networks section at the end with default pointing to shared_network
func addSharedNetworkToComposeFiles(repoPath string) error {
	// Find all compose files
	var files []string
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() &&
			(strings.HasPrefix(info.Name(), "docker-compose") || 
			strings.HasPrefix(info.Name(),"compose")) &&
			strings.HasSuffix(info.Name(), ".yml") ||
			strings.HasSuffix(info.Name(), ".yaml") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, file := range files {
		yamlBytes, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Error reading %s: %s", file, err)
			continue
		}

		var compose map[string]interface{}
		err = yaml.Unmarshal(yamlBytes, &compose)
		if err != nil {
			log.Printf("Error parsing %s: %s", file, err)
			continue
		}

		// Update services
		if servicesRaw, ok := compose["services"].(map[string]interface{}); ok {
			for svcName, svcValue := range servicesRaw {
				if svcMap, ok := svcValue.(map[string]interface{}); ok {
					// Add 'networks' if not present
					networks, ok := svcMap["networks"].([]interface{})
					if !ok {
						networks = []interface{}{}
					}
					found := false
					for _, n := range networks {
						if str, ok := n.(string); ok && str == "shared_network" {
							found = true
							break
						}
					}
					if !found {
						networks = append(networks, "shared_network")
					}
					svcMap["networks"] = networks

					// **Clear host ports** to avoid conflicts; leave only internal ports in override if needed
					if _, ok := svcMap["ports"]; ok {
						svcMap["ports"] = []interface{}{}
					}

					servicesRaw[svcName] = svcMap
				}
			}
		}

		// Ensure networks section exists
		if _, ok := compose["networks"].(map[string]interface{}); !ok {
			compose["networks"] = map[string]interface{}{}
		}
		networksMap := compose["networks"].(map[string]interface{})
		networksMap["default"] = map[string]interface{}{
			"external": true,
			"name":     "shared_network",
		}

		// Marshal back to YAML
		outBytes, err := yaml.Marshal(compose)
		if err != nil {
			log.Printf("Error marshaling YAML for %s: %s", file, err)
			continue
		}

		err = os.WriteFile(file, outBytes, 0644)
		if err != nil {
			log.Printf("Error writing %s: %s", file, err)
			continue
		}

		log.Printf("Updated compose file %s with shared_network and cleared host ports", file)
	}

	return nil
}

func createNetworkOverride(repoPath string, services map[string][]string) error {
	var sb strings.Builder
	sb.WriteString("services:\n")

	for service, ports := range services {
		sb.WriteString(fmt.Sprintf("  %s:\n", service))
		// Silencer: remove host ports
		sb.WriteString("    ports: []\n")
		// Attach to shared_network
		sb.WriteString("    networks:\n")
		sb.WriteString("      - shared_network\n")

		// Add Traefik labels for each internal port
		if len(ports) > 0 {
			sb.WriteString("    labels:\n")
			for i, port := range ports {
				// Example: route path /servicename or /servicename0, /servicename1 if multiple ports
				labelPath := service
				if i > 0 {
					labelPath = fmt.Sprintf("%s%d", service, i)
				}
				sb.WriteString(fmt.Sprintf("      - \"traefik.enable=true\"\n"))
				sb.WriteString(fmt.Sprintf("      - \"traefik.http.routers.%s.rule=PathPrefix(`/%s`)\"\n", labelPath, labelPath))
				sb.WriteString(fmt.Sprintf("      - \"traefik.http.services.%s.loadbalancer.server.port=%s\"\n", labelPath, port))
			}
		}
	}

	sb.WriteString("\nnetworks:\n")
	sb.WriteString("  shared_network:\n")
	sb.WriteString("    name: shared_network\n")
	sb.WriteString("    external: true\n")

	overridePath := filepath.Join(repoPath, "docker-compose.override.yml")
	return os.WriteFile(overridePath, []byte(sb.String()), 0644)
}

func insert(slice []string, index int, value string) []string {
	return append(slice[:index], append([]string{value}, slice[index:]...)...)
}

func dockerComposeHandler(repoName string, instructions_to_start string) error {
	// 1. Define the path inside your Go container
	// This must match the volume you mounted in your manager's compose file
	cleanName := strings.TrimPrefix(repoName, "/")
	repoPath := filepath.Join("/api/downloads", cleanName)

	// 2. Retrieve services and their internal ports from the compose file(s)
	services, err := retrieveServicesFromComposeFile(repoPath)
	if err != nil {
		return fmt.Errorf("failed to retrieve services from compose file: %w", err)
	}

	// Add the shared network to all compose files in the repo to ensure connectivity between services and Traefik
	if err := addSharedNetworkToComposeFiles(repoPath); err != nil {
		return fmt.Errorf("failed to add shared network to compose files: %w", err)
	}

	// Create a docker-compose.override.yml file with the necessary network configuration and Traefik labels for routing based on the services and their internal ports. This file will be used to ensure that all services are connected to the shared network and can be routed by Traefik without modifying the original compose files.

	if err := createNetworkOverride(repoPath, services); err != nil {
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
		// If one of the commands contains the "up" command, we must add the -d flag to run it in detached mode.
		if strings.Contains(cmdStr, "up") && !strings.Contains(cmdStr, "-d") {
			log.Printf("Adding -d flag to command: %s\n", cmdStr)
			parts := strings.Fields(cmdStr)
			lastFIndex := -1

			// Find the position of the last file listed after a -f
			for i := 0; i <= len(parts)-1; i++ {
				if parts[i] == "up" {
					lastFIndex = i
				}
			}

			if lastFIndex != -1 {
				cmdStr = strings.Join(insert(parts, lastFIndex+1, "-d"), " ")
			}
			log.Printf("Updated command with -d: %s\n", cmdStr)
		}
		// Ensure the --wait flag is included in the command to ensure that the command waits for the services to be healthy before proceeding.
		if strings.Contains(cmdStr, "up") && !strings.Contains(cmdStr, "--wait") {
			log.Printf("Adding --wait flag to command: %s\n", cmdStr)
			parts := strings.Fields(cmdStr)
			lastFIndex := -1

			log.Printf("Command parts: %v\n", parts)

			// Find the position of the last file listed after a -f
			for i := 0; i <= len(parts)-1; i++ {
				if parts[i] == "up" {
					lastFIndex = i
				}
			}

			if lastFIndex != -1 {
				cmdStr = strings.Join(insert(parts, lastFIndex+1, "--wait"), " ")
			}
			log.Printf("Updated command with --wait: %s\n", cmdStr)
		}
		if strings.Contains(cmdStr, "-f") {
			parts := strings.Fields(cmdStr)
			lastFIndex := -1

			// Find the position of the last file listed after a -f
			for i := 0; i <= len(parts)-1; i++ {
				if parts[i] == "-f" {
					lastFIndex = i + 1
				}
			}

			if lastFIndex != -1 {
				cmdStr = strings.Join(insert(insert(parts, lastFIndex+1, "docker-compose.override.yml"), lastFIndex+1, "-f"), " ")
			}
		}
		log.Printf("Executing command: %s\n", cmdStr)
		cmdParts := strings.Fields(cmdStr)
		// Print part by part
		for i, part := range cmdParts {
			log.Println("Part %d: %s\n", i, part)
		}
		cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
		cmd.Dir = repoPath
		cmd.Stderr = os.Stderr
		// Set the command's standard output to the Go application's standard output so we can see the output in the logs.
		// cmd.Stdout = os.Stdout
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
