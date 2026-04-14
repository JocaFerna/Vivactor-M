package sharedLibraries

import (
	"fmt"
	"strings"
	"time"

	graphparsing "architecture-retrieval/architecture/graphParsing"
	"architecture-retrieval/refactor/utils"
)

func MitigateSharedLibraries(graph string, sharedLibrariesSmells []string) (string,error){
	// Parse the graph
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return "", err
	}

	basePath := utils.GetBasePathOfGraph(graphStruct)
	
	graphLibraries := graphStruct.System.SelfManagedLibraries

	for _, libraryName := range sharedLibrariesSmells {
		// Trim \" from the library name if it exists.
		libraryName = strings.TrimPrefix(libraryName,"\"")
		libraryName = strings.TrimSuffix(libraryName,"\"")
		// Match library name by library in graphStruct
		for _, library := range graphLibraries {
			if library.Name == libraryName {
				fmt.Printf("Mitigating shared library smell for library %s...\n", library.Name)
				// First,let's create a new service for the library,
				// to allow it to be deployed and developed independently
				// by the connected microservices.
				newServiceName := fmt.Sprintf("%s-service", library.Name)
				newNode := graphparsing.Node{
					Id: fmt.Sprintf("%d", len(graphStruct.Nodes)+1),
					Label: newServiceName,
					Type: "BasicNode", //TODO: We are using basicNode, but we can add a new type in the future.
					Properties: graphparsing.NodeProperties{
						Language: "java",
						OrderOfMagnitudeOfFiles: "10^1",
					},
				}
				graphStruct.Nodes = append(graphStruct.Nodes, newNode)

				// Create the new service.
				err := utils.CreateBasicFileFromNode(newNode, basePath)
				if err != nil {
					return "", fmt.Errorf("error creating service for library %s: %v", library.Name, err)
				}

				// Now, for each service that is using the library, we must 
				// create a edge between the new service and the service
				// from the service that is using the library and the library.
				for _, serviceName := range library.ServicesUsingLibrary {
					// Get the node of the service by its name.
					serviceNode, err := graphparsing.GetNodeByLabel(graphStruct, serviceName)
					if err != nil {
						return "", fmt.Errorf("error getting node for service %s: %v", serviceName, err)
					}
					// Create an endpoint for the new service, we can use the library name as the endpoint.
					endpoint := fmt.Sprintf("/%s", utils.SanitizeName(libraryName))

					// Create the call Definition
					callDefinition := fmt.Sprintf("http://%s:8080%s", newServiceName, endpoint)

					// Create an edge between the new service and the service that is using the library.
					newEdge := graphparsing.Edge{
						Source: serviceNode.Id,
						Target: newNode.Id,
						Endpoint: endpoint,
						Properties: graphparsing.EdgeProperties{
							CallDefinitionInSource: callDefinition,
							Method: "GET",
						},
					}
					graphStruct.Edges = append(graphStruct.Edges, newEdge)
					

					// Add the new edge to the architecture.
					err = utils.HandleCallFromNode(serviceNode, newNode, newEdge, basePath)
					if err != nil {
						return "", fmt.Errorf("error handling call from node %s: %v", serviceNode.Label, err)
					}
				}
			}
		}
		// Remove the library from the graphStruct, since it is now a service.
		newLibraries := []graphparsing.SelfManagedLibraries{}
		for _, library := range graphStruct.System.SelfManagedLibraries {
			if library.Name != libraryName {
				newLibraries = append(newLibraries, library)
			}
		}
		graphStruct.System.SelfManagedLibraries = newLibraries
	}
	// Restart the docker compose to apply the changes.
	err = utils.RestartDockerComposeWithoutTraefik(basePath)
	if err != nil {
		return "", fmt.Errorf("error restarting docker compose: %v", err)
	}
	
	// Wait for the services to restart and the graph to be updated.
	time.Sleep(10 * time.Second)
	stringGraph, err := graphparsing.SerializeGraph(graphStruct)
	if err != nil {
		return "", fmt.Errorf("error serializing graph: %v", err)
	}
	return stringGraph, nil
}
