package nonAPIGateway

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
	"architecture-retrieval/refactor/utils"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func MitigateNonAPIGateway(graphString string, selectedNodes []string) (string, error){
	// Parse the graph from JSON
	graph, err := graphparsing.ParseGraph(graphString)
	if err != nil {
		return "", err
	}

	// Check if the selected nodes are in the graph
	nodeMap := make(map[int]graphparsing.Node)
	index := 0
	for _, nodeLabel := range selectedNodes {
		// Remove any leading/trailing \" from the node label
		nodeLabel = strings.TrimPrefix(nodeLabel, "\"")
		nodeLabel = strings.TrimSuffix(nodeLabel, "\"")

		node, err := graphparsing.GetNodeById(graph, nodeLabel)
		if err != nil {
			return "", fmt.Errorf("Error: Node with ID %s not found in the graph", nodeLabel)
		}
		nodeMap[index] = node
		index++
	}

	// Get base path of the project from the graph.
	basePath := utils.GetBasePathOfGraph(graph)

	// Create a new API Gateway node
	apiGatewayNode := graphparsing.Node{
		Id: strconv.Itoa(len(graph.Nodes) + 1), // Generate a new unique ID
		Label: "inserted-api-gateway",
		Type: "APIGateway",
		Properties: graphparsing.NodeProperties{
			Language: "python",
			OrderOfMagnitudeOfFiles: "10^2",
			Port: "8080",
		},
	}
	graph.Nodes = append(graph.Nodes, apiGatewayNode)
	err = utils.GenerateAPIGatewayFromNode(apiGatewayNode, basePath)
	if err != nil {
		return "", fmt.Errorf("Error creating API Gateway node: %v", err)
	}

	// Eliminate any edges within the selected nodes and create new edges from the API Gateway to the selected nodes
	for i, _ := range nodeMap {
		node1 := nodeMap[i]
		for j := i + 1; j < len(nodeMap); j++ {
			node2 := nodeMap[j]
			// Eliminate edge between node1 and node2 if it exists
			for k, edge := range graph.Edges {
				if (edge.Source == node1.Id && edge.Target == node2.Id) || (edge.Source == node2.Id && edge.Target == node1.Id) {
					graph.Edges = append(graph.Edges[:k], graph.Edges[k+1:]...)
					if edge.Source == node1.Id {
						err := utils.RemoveCallToService(node1,node2,basePath)
						if err != nil {
							return "", fmt.Errorf("Error removing call to service: %v", err)
						}
					} else {
						err := utils.RemoveCallToService(node2,node1,basePath)
						if err != nil {
							return "", fmt.Errorf("Error removing call to service: %v", err)
						}
						
					}
					break
				}
			}
		}
			// Create new edges from the API Gateway to the selected nodes
			newEdgeRequest := graphparsing.Edge{
				Source: node1.Id,
				Target: apiGatewayNode.Id,
				Endpoint: fmt.Sprintf("/api/v1/request/%s", graphparsing.SanitizeName(node1.Label)),
				Properties: graphparsing.EdgeProperties{
					CallDefinitionInSource: fmt.Sprintf("http://inserted-api-gateway:8080/api/v1/request/%s", graphparsing.SanitizeName(node1.Label)),
					Method: "GET",
				},
			}
			// Get node1 port
			node1Port, err := utils.GetPortFromNode(node1)
			if err != nil {
				return "", fmt.Errorf("Error getting port from node: %v", err)
			}

			newEdgeResponse := graphparsing.Edge{
				Source: apiGatewayNode.Id,
				Target: node1.Id,
				Endpoint: fmt.Sprintf("/api/v1/response/%s", graphparsing.SanitizeName(node1.Label)),
				Properties: graphparsing.EdgeProperties{
					CallDefinitionInSource: fmt.Sprintf("http://%s:%s/api/v1/response/%s",graphparsing.SanitizeName(node1.Label), strconv.Itoa(node1Port), graphparsing.SanitizeName(node1.Label)),
					Method: "POST",
				},
			}
			graph.Edges = append(graph.Edges, newEdgeRequest)
			err = utils.HandleCallFromNode(node1, apiGatewayNode, newEdgeRequest, basePath)
			if err != nil {
				return "", fmt.Errorf("Error handling call from node: %v", err)
			}
			graph.Edges = append(graph.Edges, newEdgeResponse)
			err = utils.HandleCallFromNode(apiGatewayNode, node1, newEdgeResponse, basePath)
			if err != nil {
				return ""	, fmt.Errorf("Error handling call to node: %v", err)
			}
	}
	// Before Starting the new API Gateway service, we wait for a 40 seconds to ensure all changes are caught by watch
	fmt.Println("Waiting for 40 seconds before starting the new API Gateway service to ensure all changes are caught by watch...")
	time.Sleep(40 * time.Second)
	// Start the new API Gateway service
	err = utils.StartServiceFromNode(apiGatewayNode, graph, basePath)
	if err != nil {
		return "", fmt.Errorf("Error starting API Gateway service: %v", err)
	}
	// Serialize the graph back to JSON
	graphSerialized, err := graphparsing.SerializeGraph(graph)
	if err != nil {
		return "", err
	}
	return graphSerialized, nil
}