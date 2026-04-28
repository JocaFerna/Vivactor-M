package innapropriateServiceIntimacity

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
	"architecture-retrieval/refactor/utils"
	"fmt"
	"strings"
)

func MitigateInnapropriateServiceIntimacity(graphString string, smells []string) (string, error){
	// Parse the graph from JSON
	graph, err := graphparsing.ParseGraph(graphString)
	if err != nil {
		return "", err
	}
	
	// Get base path of the project from the graph.
	basePath := utils.GetBasePathOfGraph(graph)

	for _, smell := range smells {
		// Example of smell: "Service: service1, Database: db1, Owner: service2"
		// Trim the "
		smell = strings.TrimPrefix(smell,"\"")
		smell = strings.TrimSuffix(smell,"\"")
		
		parts := strings.Split(smell, ",")
		service_to_fix := strings.TrimSpace(strings.Split(parts[0], ":")[1])
		owner := strings.TrimSpace(strings.Split(parts[2], ":")[1])

		// Get the nodes of the service, database and owner
		serviceNode, err := graphparsing.GetNodeByLabel(graph, service_to_fix)
		if err != nil {
			fmt.Errorf("Error: Service node with label %s not found in the graph", service_to_fix)
			continue
		}
		ownerNode, err := graphparsing.GetNodeByLabel(graph, owner)
		if err != nil {
			return "", fmt.Errorf("Error: Owner node with label %s not found in the graph", owner)
		}
	
		// Move the calls to the database from the service to fix to the owner service.
		ownerNode, graph, err = utils.MoveNodeCalls(graph,serviceNode,ownerNode,basePath)
		if err != nil {
			return "", fmt.Errorf("Error moving calls from service %s to owner service %s: %v", service_to_fix, owner, err)
		}

		// Remove the service to fix if it has no more calls.
		err = utils.RemoveService(serviceNode,graph,basePath)
		if err != nil {
			return "", fmt.Errorf("Error removing service %s: %v", service_to_fix, err)
		}
		// Remove service node from the graph
		graph = graphparsing.RemoveNode(graph, serviceNode)
	}

	// Convert the graph back to JSON
	modifiedGraphString, err := graphparsing.SerializeGraph(graph)
	if err != nil {
		return "", fmt.Errorf("Error converting modified graph to JSON: %v", err)
	}

	return modifiedGraphString, nil
}	