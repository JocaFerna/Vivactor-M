package sharedPersistency
import (
	"fmt"
	"log"
	"strings"
	"time"
	"strconv"

	graphparsing "architecture-retrieval/architecture/graphParsing"
	"architecture-retrieval/refactor/utils"
)


// The approach for this mitigation has three proposed solutions
// 1 - use independent db for each service.
// 2 - use a shared database with proper isolation of tables.
// 3 - use a shared database with proper isolation of schemas.
// We will implement the first solution, due to the nature of this
// emulation structure.
func MitigateSharedPersistencySmellsGraph(graph string, sharedPersistencySmells [][]string) (string, error) {
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return "", err
	}
	refactoredSmells := []string{}
	
	basePath := utils.GetBasePathOfGraph(graphStruct)


	for _,smell := range sharedPersistencySmells {
		nodeList := []graphparsing.Node{}
		for _, service := range smell {	
			// Get service by name
			serviceName := strings.TrimPrefix(service, "\"")
			serviceName = strings.TrimSuffix(serviceName, "\"")
			node, err := graphparsing.GetNodeByLabel(graphStruct, serviceName)
			if err != nil {
				log.Printf("Error getting node by label %s: %v", serviceName, err)
				continue
			}
			// Add to the list of nodes
			nodeList = append(nodeList, node)
			// Check if the service has its own database.
			hasOwnDB, ownDB, err := hasOwnDatabase(graphStruct, node)
			if err != nil {
				log.Printf("Error checking if service %s has its own database: %v", serviceName, err)
				continue
			}
			if hasOwnDB {
				log.Printf("Service %s already has its own database %s, only removing the shared dependency...", serviceName, ownDB.Label)
			} else {
				log.Printf("Service %s does not have its own database, creating one and removing the shared dependency...", serviceName)
				// Create a new database node for the service.
				currentHighestId := 0
				for _, node := range graphStruct.Nodes {
					id, err := strconv.Atoi(node.Id)
					if err != nil {
						log.Printf("Error converting node ID to integer: %v", err)
						continue
					}
					if id > currentHighestId {
						currentHighestId = id
					}
				}
				newDBNode := graphparsing.Node{
					Id:    fmt.Sprintf("%d", currentHighestId+1),
					Type:  "DatabaseNode",
					Label: fmt.Sprintf("%s_db", node.Label),
				}
				graphStruct.Nodes = append(graphStruct.Nodes, newDBNode)

				// Add the new database to the architecture.
				node, graphStruct, err = utils.CreateDatabaseFromNode(graphStruct, newDBNode, basePath, node)
				if err != nil {
					log.Printf("Error adding database node to service %s: %v", serviceName, err)
					continue
				}
			}
		}
		// Get the shared database node for the services in the smell.
		sharedDBNode, err := getSharedDatabaseByNodes(graphStruct, nodeList)
		if err != nil {
			log.Printf("Error getting shared database for services in smell %v: %v", smell, err)
			continue
		}
		// Remove the edges between the shared database and the services in the smell.
		for _, edge := range graphStruct.Edges {
			if edge.Target == sharedDBNode.Id {
				for _, serviceNode := range nodeList {
					if edge.Source == serviceNode.Id {
						// Remove the edge
						graphStruct.Edges = graphparsing.RemoveEdge(graphStruct.Edges, edge)
						// Also, remove the connection in the service's file.
						err = utils.RemoveCallToDBNode(&serviceNode, &sharedDBNode, basePath)
						if err != nil {
							log.Printf("Error removing call to shared database for service %s: %v", serviceNode.Label, err)
						}
						break
					}
				}
			}
		}
		refactoredSmells = append(refactoredSmells, fmt.Sprintf("Shared persistency smell between services %v mitigated by creating independent databases and removing shared dependencies.", smell))
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
func getSharedDatabaseByNodes(graphStruct graphparsing.Graph, nodes []graphparsing.Node) (graphparsing.Node, error) {
	for _, node := range graphStruct.Nodes {
		checker := make(map[string]bool)

		if node.Type == "DatabaseNode" {
			// Check if this database edge's source nodes matches the
			// services in the smell.
			for _, edge := range graphStruct.Edges {
				if edge.Target == node.Id {
					checker[edge.Source] = true
				}
			}
			// If all services in the smell are in the checker, then we found the shared database.
			allServicesMatch := true
			for _, serviceNode := range nodes {
				if _, exists := checker[serviceNode.Id]; !exists {
					allServicesMatch = false
					break
				}
			}
			if allServicesMatch {
				return node, nil
			}
		}
	}
	return graphparsing.Node{}, fmt.Errorf("Shared database not found")
}

func hasOwnDatabase(graphStruct graphparsing.Graph, node graphparsing.Node) (bool, graphparsing.Node, error) {
	hasOwnDB := false
	var ownDB graphparsing.Node
	for _, edge := range graphStruct.Edges {
		if edge.Source == node.Id {
			targetNode, err := graphparsing.GetNodeById(graphStruct, edge.Target)
			if err != nil {
				continue
			}
			if targetNode.Type == "DatabaseNode" {
				// Check if the database is shared with other services.
				// For now, we are using as criteria if the database name contains
				// the service name, but we can use other criteria in the future.
				if strings.Contains(targetNode.Label, node.Label) || strings.Contains(targetNode.Label, utils.SanitizeName(node.Label)) {
					hasOwnDB = true
					ownDB = targetNode
					break
				}
			}
		}
	}
	return hasOwnDB, ownDB, nil
}

