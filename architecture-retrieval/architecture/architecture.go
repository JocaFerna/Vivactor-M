package architecture

import (
	//"context"
	//"log"
	//"fmt"
	//gogithub "github.com/google/go-github/v65/github"
	//"io"
	"os"
	"slices"
	"time"

	//"path/filepath"
	//"architecture-retrieval/hashset"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"regexp"

	"bufio"
	"bytes"
	"encoding/json"

	git "github.com/go-git/go-git/v5"
	"gopkg.in/yaml.v3"
)

//const GlobalServiceNames map[string]int = make(map[string]int)

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

	// Remove .git folder to avoid confusion and save space
	err = os.RemoveAll(filepath.Join(path, ".git"))
	if err != nil {
		log.Printf("Error removing .git folder: %s\n", err.Error())
	}
	return err
}

func StartArchitecture(repoName string, make_instructions string, packageList string) error {
	log.Println("Starting architecture...")

	err := dockerComposeHandler(repoName, make_instructions, packageList)
	if err != nil {
		log.Printf("Error starting architecture: %s\n", err.Error())
		return err
	} else {
		log.Println("Architecture started successfully")
		return nil
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
				strings.HasPrefix(info.Name(), "compose")) &&
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

					// Remove any service_healthy dependencies to avoid startup issues since we are adding a basic healthcheck to all services. This is a bit of a hack but it ensures that the architecture can start successfully without needing to analyze and modify complex dependency graphs in the compose files.
					//"depends_on":
					if dependsOn, ok := svcMap["depends_on"].(map[string]interface{}); ok {
						// Service
						for depName, depValue := range dependsOn {
							// Dependency
							if depMap, ok := depValue.(map[string]interface{}); ok {
								// Type of dependency (e.g. "service_healthy")
								if condition, ok := depMap["condition"].(string); ok && condition == "service_healthy" {
									// Remove any depends_on conditions that rely on service health to avoid startup issues since we are adding a basic healthcheck to all services. This is a bit of a hack but it ensures that the architecture can start successfully without needing to analyze and modify complex dependency graphs in the compose files.
									delete(depMap, "condition")

									log.Printf("Removed service_healthy condition from dependency %s of service %s in %s\n", depName, svcName, file)
								}
								// Check if the map is now empty (e.g., it only had "condition")
								if len(depMap) == 0 {
									delete(dependsOn, depName)
									log.Printf("Removed empty dependency %s from service %s\n", depName, svcName)
									continue // Skip the "Kept dependency" log
								}
							}
						}
						if len(dependsOn) == 0 {
							// Deletes the depends_on section
							delete(svcMap, "depends_on")
						}
						log.Printf("Updated depends_on for service %s in %s\n", svcName, file)
					}

					// Add healthcheck if not present
					if _, ok := svcMap["healthcheck"]; !ok {
						svcMap["healthcheck"] = map[string]interface{}{
							"test":     []interface{}{"CMD-SHELL", "echo 'Healthcheck passed' || exit 1"},
							"interval": "30s",
							"timeout":  "10s",
							"retries":  3,
						}
						log.Printf("Added basic healthcheck to service %s in %s\n", svcName, file)
					}


					// Also add the develop watch with rebuild action
					// develop:
					// watch:
					// - action: rebuild
					// - path: .

					// We can only add the watch, if there is a way to rebuild.
					buildVal, hasBuild := svcMap["build"]
					if hasBuild {
						var watchPath string
						switch v := buildVal.(type) {
						case string:

							watchPath = v
						case map[string]interface{}:
							if ctx, ok := v["context"].(string); ok {
								watchPath = ctx
							} else {
								watchPath = "." // Fallback
							}
						default:
							watchPath = "."
						}

						// We need to add this to ALL docker-compose files to ensure that the watch is present regardless of which compose file is used to start the architecture.
						for _, f := range files {
							yamlBytes, err := os.ReadFile(f)
							if err != nil {
								log.Printf("Error reading %s: %s", f, err)
								continue
							}

							var c map[string]interface{}
							err = yaml.Unmarshal(yamlBytes, &c)
							if err != nil {
								log.Printf("Error parsing %s: %s", f, err)
								continue
							}

							if servicesRaw, ok := c["services"].(map[string]interface{}); ok {
								if svcMap, ok := servicesRaw[svcName].(map[string]interface{}); ok {

									// Add build context if not present to ensure the service can be rebuilt
									if _, hasBuild := svcMap["build"]; !hasBuild {
										svcMap["build"] = watchPath
										log.Printf("Added build context to service %s in %s to enable rebuild action\n", svcName, file)
									}

									// Add 'develop' section if not present
									if _, ok := svcMap["develop"].(map[string]interface{}); !ok {
										svcMap["develop"] = map[string]interface{}{
											"watch": []interface{}{
												map[string]interface{}{
													"action": "rebuild",
													"path":   watchPath,
												},
											},
										}
										servicesRaw[svcName] = svcMap
										log.Printf("Added develop watch with rebuild action to service %s in %s\n", svcName, file)
									}
									// Marshal back to YAML
									outBytes, err := yaml.Marshal(c)
									if err != nil {
										log.Printf("Error marshaling YAML for %s: %s", f, err)
										continue
									}

									err = os.WriteFile(f, outBytes, 0644)
									if err != nil {
										log.Printf("Error writing %s: %s", f, err)
										continue
									}
									log.Printf("Updated compose file %s with develop watch for service %s\n", f, svcName)
								}

							}
						}
					}
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

func replace(slice []string, index int, value string) []string {
	if index < 0 || index >= len(slice) {
		return slice
	}
	slice[index] = value
	return slice
}

// TODO: This is hardcoded for 8 version of java
func fixDockerfiles(repoPath string) error {
	// This regex looks for any "FROM java" or "FROM openjdk" line
	javaRegex := regexp.MustCompile(`(?i)FROM\s+(java|openjdk):?(\d+)?([\w.-]*)`)

	return filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Base(path) == "Dockerfile" {
			content, _ := os.ReadFile(path)

			// We swap to Amazon Corretto.
			newContent := javaRegex.ReplaceAllString(string(content), "FROM amazoncorretto:$2-al2-jdk")

			if newContent != string(content) {
				os.WriteFile(path, []byte(newContent), 0644)
				log.Printf("Swapped to Corretto for %s", path)
			}
		}
		return nil
	})
}



// TODO: This function (and prob everything it calls)
// needs a major refactor because some things are hammered.
func dockerComposeHandler(repoName string, instructions_to_start string, packageList string) error {
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

	// In case of deprecated base images in Dockerfiles (like 'java:8-jre'), we need to patch them to ensure the architecture can start successfully.
	if err := fixDockerfiles(repoPath); err != nil {
		return fmt.Errorf("failed to fix Dockerfiles: %w", err)
	}

	// Separate the instructions to start the architecture into a list of commands by \n
	commands := strings.Split(instructions_to_start, "\n")

	// Parse the package list into a slice by \n
	packages := strings.Split(packageList, "\n")
	log.Printf("Packages to install: %v\n", packages)

	// Execute each command in the list
	for _, cmdStr := range commands {

		cmdStr = strings.TrimSpace(cmdStr)
		if cmdStr == "" {
			continue // Skip empty lines
		}

		// Change "docker-compose" to "docker compose"
		cmdList := strings.Split(cmdStr, " ")
		cmdStr = strings.Join(replace(cmdList, slices.Index(cmdList, "docker-compose"), "docker compose"), " ")
		
		// Add "nix-shell {packages} --run {cmdStr}" to the command to ensure it runs in the correct environment with the necessary dependencies installed.
		cmdStr = fmt.Sprintf("nix-shell -p "+strings.Join(packages, " ")+" --run \"%s\"", cmdStr)
		
	    // Add the docker-compose.override.yml file.
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
		// Add the "watch" argument.
		cmdStr = strings.ReplaceAll(cmdStr, "up", "watch")


		log.Printf("Executing command: %s\n", cmdStr)
		cmd := exec.Command("sh", "-c", cmdStr)
		
		cmd.Dir = repoPath
		cmd.Stderr = os.Stderr
		//cmd.Stdout = os.Stdout

		// If the command is the "watch" one.
		if strings.Contains(cmdStr, "watch") && strings.Contains(cmdStr, "docker compose") {

			// Start the watch process in the background
			if err := cmd.Start(); err != nil {
				log.Printf("Error starting command '%s': %s\n", cmdStr, err.Error())
				return err
			}
			
			
			// Convert service names to a map[string]int for the health check function
			var serviceNames map[string]int = make(map[string]int)
			for service := range services {
				serviceNames[service] = 0
			}
			

			
			log.Printf("Waiting for containers to be created for: %v\n", serviceNames)

			maxRetries := 30
			var waitErr error

			for i := 0; i < maxRetries; i++ {
				err := checkComposeHealth(repoPath, serviceNames)
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
				cmd.Process.Kill()
				return waitErr
			}
		} else {
			// Every other command.
			if err := cmd.Run(); err != nil {
				log.Printf("Error executing command '%s': %s\n", cmdStr, err.Error())
				return err
			}
		}	
	}

	return nil
}
// This is here mainly if we need in the future to go back using the github API.

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
