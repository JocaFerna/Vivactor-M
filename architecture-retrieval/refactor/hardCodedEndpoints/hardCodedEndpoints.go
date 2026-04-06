package hardCodedEndpoints

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
	"architecture-retrieval/refactor/utils"
	"fmt"
	"log"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// This function, given a graph and the hardcoded endpoints, refactors them.
// Using a service discovery approach, using Traefik for that.
func RefactorGraphWithHardcodedEndpoints(graph string, hardcodedEndpoints []string) (string, error) {
	
	// Parse the graph
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return "", err
	}

	log.Println("Hardcoded endpoints to refactor:", hardcodedEndpoints)

	// Get path of repo
	basePath := utils.GetBasePathOfGraph(graphStruct)

	// Retrieve the docker compose file.
	dockerComposeFile,err := utils.ReadFileContent(basePath+"/docker-compose.yml")
	if err != nil {
		return "", err
	}

	// Add Traefik configuration to the docker compose file.
	if !strings.Contains(dockerComposeFile, "traefik") && len(hardcodedEndpoints) > 0 {
		dockerComposeFile, err = InjectTraefikService(dockerComposeFile)
		if err != nil {
			return "", fmt.Errorf("error injecting Traefik service: %v", err)
		}
		err = utils.WriteFileContent(basePath+"/docker-compose.yml", dockerComposeFile)
		if err != nil {
			return "", fmt.Errorf("error writing updated docker-compose file: %v", err)
		}
	}


	// For each edge, check if the endpoint is in the hardcodedEndpoints list. If it is, refactor it.
	for i, edge := range graphStruct.Edges {
		for _, hardcodedEndpoint := range hardcodedEndpoints {
			hardcodedEndpoint = strings.TrimPrefix(hardcodedEndpoint,"\"")
			hardcodedEndpoint = strings.TrimSuffix(hardcodedEndpoint,"\"")
			log.Printf("Checking edge with call definition in source: %s against hardcoded endpoint: %s", edge.Properties.CallDefinitionInSource, hardcodedEndpoint)
			if edge.Properties.CallDefinitionInSource == hardcodedEndpoint {

				log.Printf("Found hardcoded endpoint in edge: %s", hardcodedEndpoint)

				//sourceNode, err := graphparsing.GetNodeById(graphStruct, edge.Source)
				if err != nil {
					return "", fmt.Errorf("error getting source node: %v", err)
				}
				targetNode, err := graphparsing.GetNodeById(graphStruct, edge.Target)
				if err != nil {
					return "", fmt.Errorf("error getting target node: %v", err)
				}
				dockerComposeFile, err = addTraefikConfigurationToService(targetNode,dockerComposeFile)
				if err != nil {
					return "", fmt.Errorf("error adding Traefik configuration to service: %v", err)
				}
				err = utils.WriteFileContent(basePath+"/docker-compose.yml", dockerComposeFile)
				if err != nil {
					return "", fmt.Errorf("error writing updated docker-compose file: %v", err)
				}

				// Update the call definition in the graph to point to Traefik instead of the hardcoded endpoint.
				graphStruct.Edges[i].Properties.CallDefinitionInSource = fmt.Sprintf("http://%s%s", targetNode.Label, edge.Endpoint)
				
				// Also, update the call definition in the source code files.
				filePathSource, err := utils.GetFilePathFromNode(graphStruct, edge.Source)
				if err != nil {
					return "", fmt.Errorf("error getting file path from node %s: %v", edge.Source, err)
				}
				fileSource, err := utils.ReadFileContent(filePathSource)
				if err != nil {
					return "", fmt.Errorf("error reading file %s: %v", filePathSource, err)
				}
				// Maybe the hardcoded endpoint is not sanitized, so we must sanitize it, except the characters
				// "/:?" that are common in endpoints.
				sanitizedHardcodedEndpoint := utils.SanitizeURL(hardcodedEndpoint)
				log.Printf("Sanitized hardcoded endpoint: %s", sanitizedHardcodedEndpoint)
				updatedFileSource := strings.ReplaceAll(fileSource, sanitizedHardcodedEndpoint, utils.SanitizeURL(graphStruct.Edges[i].Properties.CallDefinitionInSource))
				err = utils.WriteFileContent(filePathSource, updatedFileSource)
				if err != nil {
					return "", fmt.Errorf("error writing file %s: %v", filePathSource, err)
				}
				log.Println("File content after refactoring:\n", updatedFileSource)

				break
			}
		}
	}

	// Give a small delay to ensure all restart processes are completed before we return the updated graph.
	time.Sleep(20 * time.Second)
	// Restart the architecture after the refactor, to update docker compose.
	err = utils.RestartDockerCompose(basePath)
	if err != nil {
		return "", fmt.Errorf("error restarting docker compose: %v", err)
	}
	// Return the updated graph.
	graphString, err := graphparsing.SerializeGraph(graphStruct)
	if err != nil {
		return "", fmt.Errorf("error serializing graph: %v", err)
	}
	return graphString, nil
}



func InjectTraefikService(yamlContent string) (string, error) {
    var root map[string]interface{}
    if err := yaml.Unmarshal([]byte(yamlContent), &root); err != nil {
        return "", err
    }

    services := getOrCreateMap(root, "services")

    traefikService := map[string]interface{}{
        "image": "traefik:v2.5",
        "command": []string{
            "--providers.docker=true",
            "--providers.docker.exposedbydefault=false",
            "--entrypoints.web.address=:80",
            "--entrypoints.legacy.address=:8080",
        },
        "ports": []string{"80:80", "8080:8080"},
        "volumes": []string{"/var/run/docker.sock:/var/run/docker.sock:ro"},
        // SINGLE network at compose time — shared_network connected post-up
        "networks": map[string]interface{}{
            "traefik_internal": map[string]interface{}{
                "aliases": []string{},
            },
        },
    }

    services["traefik"] = traefikService

    networks := getOrCreateMap(root, "networks")
    if _, exists := networks["traefik_internal"]; !exists {
        networks["traefik_internal"] = map[string]interface{}{
            "name": "traefik_internal", "external": true,
        }
    }

    output, err := yaml.Marshal(root)
    return string(output), err
}

func addTraefikConfigurationToService(node graphparsing.Node, yamlContent string) (string, error) {
	var root map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &root); err != nil {
		return yamlContent, err
	}

	services := getOrCreateMap(root, "services")
	serviceName := utils.SanitizeName(node.Label)

	targetService, exists := services[serviceName].(map[string]interface{})
	if !exists {
		return yamlContent, nil
	}

	// Move service exclusively onto traefik_internal.
	// It must NOT also be on shared_network — that triggers the multi-endpoint error.
	replaceNetworks(targetService, "traefik_internal")

	// Configure Traefik labels
	labels := getOrCreateMap(targetService, "labels")
	labels["traefik.enable"] = "true"
	labels["traefik.docker.network"] = "traefik_internal"

	labels[fmt.Sprintf("traefik.http.routers.%s-new.rule", serviceName)] = fmt.Sprintf("Host(`%s`)", serviceName)
	labels[fmt.Sprintf("traefik.http.routers.%s-new.entrypoints", serviceName)] = "web"
	labels[fmt.Sprintf("traefik.http.routers.%s-old.rule", serviceName)] = fmt.Sprintf("Host(`%s`)", serviceName)
	labels[fmt.Sprintf("traefik.http.routers.%s-old.entrypoints", serviceName)] = "legacy"

	port := "8080"
	if node.Properties.Port != "" {
		port = node.Properties.Port
	}
	labels[fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", serviceName)] = port

	// Add a DNS alias on traefik_internal so Traefik can resolve this service by name
	addAliasToTraefik(services, serviceName)

	updatedYaml, err := yaml.Marshal(root)
	return string(updatedYaml), err
}

// addAliasToTraefik adds the service name as a DNS alias under
// services.traefik.networks.traefik_internal.aliases
func addAliasToTraefik(services map[string]interface{}, serviceName string) {
    traefik, ok := services["traefik"].(map[string]interface{})
    if !ok {
        return
    }

    networks := getOrCreateMap(traefik, "networks")

    // Only add alias to traefik_internal in the compose file.
    // shared_network is connected post-up via `docker network connect --alias`
    // in RestartDockerCompose to avoid the multi-endpoint daemon error.
    internalNet := getOrCreateMap(networks, "traefik_internal")
    addUniqueAlias(internalNet, serviceName)
}

// replaceNetworks replaces a service's entire network config with a single
// named network in map format. This avoids any dual-network assignment.
func replaceNetworks(service map[string]interface{}, netName string) {
	service["networks"] = map[string]interface{}{
		netName: map[string]interface{}{},
	}
}

// assignNetwork adds netName to a service's networks without removing existing ones.
// Use this for non-Traefik services that just need shared_network.
func assignNetwork(service map[string]interface{}, netName string) {
	networksInterface, exists := service["networks"]

	if !exists {
		service["networks"] = map[string]interface{}{
			netName: map[string]interface{}{},
		}
		return
	}

	networksMap := make(map[string]interface{})

	switch v := networksInterface.(type) {
	case []interface{}:
		for _, n := range v {
			networksMap[fmt.Sprint(n)] = map[string]interface{}{}
		}
	case []string:
		for _, n := range v {
			networksMap[n] = map[string]interface{}{}
		}
	case map[string]interface{}:
		networksMap = v
	}

	if _, ok := networksMap[netName]; !ok {
		networksMap[netName] = map[string]interface{}{}
	}

	service["networks"] = networksMap
}

func getOrCreateMap(parent map[string]interface{}, key string) map[string]interface{} {
	if val, ok := parent[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return m
		}
	}
	m := make(map[string]interface{})
	parent[key] = m
	return m
}


func addUniqueAlias(netConfig map[string]interface{}, alias string) {
    rawAliases, exists := netConfig["aliases"]
    var aliases []string

    if exists {
        switch v := rawAliases.(type) {
        case []interface{}:
            for _, s := range v {
                aliases = append(aliases, fmt.Sprint(s))
            }
        case []string:
            aliases = v
        }
    }

    for _, a := range aliases {
        if a == alias {
            return
        }
    }
    netConfig["aliases"] = append(aliases, alias)
}