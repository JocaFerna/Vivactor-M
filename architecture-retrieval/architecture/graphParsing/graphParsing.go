package graphparsing

import (
	"encoding/json"
	"fmt"
	"strings"
)
type Graph struct{
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
	System SystemContext `json:"systemContext"`
}

type Node struct{
	Id string `json:"id"`
	Label string `json:"label"`
	Type string `json:"type"`
	Properties NodeProperties `json:"properties,omitempty"`
} 

type NodeProperties struct{
	Language string `json:"language"`
	OrderOfMagnitudeOfFiles string `json:"orderOfMagnitudeOfFiles"`
	Port string `json:"port,omitempty"`
}

type Edge struct{
	Source string `json:"source"`
	Target string `json:"target"`
	Endpoint string `json:"endpoint"`
	Properties EdgeProperties `json:"properties"`
}

type EdgeProperties struct {
    CallDefinitionInSource string `json:"callDefinitionInSource"`
	Method string `json:"method"`
}

type SystemContext struct {
	Name string `json:"name"`
	Description string `json:"description"`
	SelfManagedLibraries []SelfManagedLibraries `json:"selfManagedLibraries"`
	ServicesSeparatedByBusinessDomain bool `json:"servicesSeparatedByBusinessDomain"`
}

type SelfManagedLibraries struct{
	Name string `json:"name"`
	ServicesUsingLibrary []string `json:"servicesUsingLibrary"`
}


func ParseGraph (graph string) (Graph, error) {
	var result Graph
	err := json.Unmarshal([]byte(graph) ,&result)
	if err != nil{
		var none Graph
		return none, fmt.Errorf("Error parsing json: %s",err)
	} else{
		return result, nil
	}
}

func SerializeGraph(graph Graph) (string, error) {
	jsonBytes, err := json.Marshal(graph)
	if err != nil {
		return "", fmt.Errorf("Error serializing graph: %s", err)
	}
	return string(jsonBytes), nil
}

func GetNodeById(graph Graph, id string) (Node, error) {
	for _, node := range graph.Nodes {
		if node.Id == id {
			return node, nil
		}
	}
	var none Node
	return none, fmt.Errorf("Node with id %s not found", id)
}

func GetNodeByLabel(graph Graph, label string) (Node, error) {
	for _, node := range graph.Nodes {
		if node.Label == label {
			return node, nil
		}
	}
	var none Node
	return none, fmt.Errorf("Node with label %s not found", label)
}

func GetAdjacentNodes(graph Graph, nodeId string) ([]*Node, error) {
    var adjacentNodes []*Node
    for _, edge := range graph.Edges {
        if edge.Source == nodeId {
            // FIX: Find the node in the actual Nodes slice and take its address
            found := false
            for i := range graph.Nodes {
                if graph.Nodes[i].Id == edge.Target {
                    adjacentNodes = append(adjacentNodes, &graph.Nodes[i])
                    found = true
                    break
                }
            }
            if !found {
                return nil, fmt.Errorf("Target node %s not found", edge.Target)
            }
        }
    }
    return adjacentNodes, nil
}

func RemoveNodeAsAdjacent(adjList map[*Node][]*Node, nodeSource *Node, nodeToRemove *Node) map[*Node][]*Node {
	neighbors := adjList[nodeSource]
	for i, neighbor := range neighbors {
		if neighbor == nodeToRemove {
			adjList[nodeSource] = append(neighbors[:i], neighbors[i+1:]...)
			break
		}
	}
	return adjList
}

func GetOrderOfMagnitudeOfFiles(node Node) int {
	// Retrieve the exponent from the string
	var exponent int
	_, err := fmt.Sscanf(node.Properties.OrderOfMagnitudeOfFiles, "10^%d", &exponent)
	if err != nil {
		return 0 // Default to 0 if parsing fails
	}
	return exponent
}

func RemoveEdge(edges []Edge, edgeToRemove Edge) []Edge {
	for i, edge := range edges {
		if edge == edgeToRemove {
			return append(edges[:i], edges[i+1:]...)
		}
	}
	return edges
}

func MergeNodeIntoAnother(graph Graph, nodeToBeMerged Node, nodeToReceive Node) Graph {
	
	// Replace all edges that point to nodeToBeMerged with nodeToReceive
	for i, edge := range graph.Edges {
		if edge.Source == nodeToBeMerged.Id && edge.Target != nodeToReceive.Id {
			graph.Edges[i].Source = nodeToReceive.Id
		}
		if edge.Target == nodeToBeMerged.Id && edge.Source != nodeToReceive.Id {
			graph.Edges[i].Target = nodeToReceive.Id
			// If target is nodeToBeMerged, we need to update the call definition in source to reflect the new source node.
			graph.Edges[i].Properties.CallDefinitionInSource = strings.ReplaceAll(graph.Edges[i].Properties.CallDefinitionInSource, nodeToBeMerged.Label, nodeToReceive.Label)
		}
		// Also, we need to delete edges from nodeToBeMerged to nodeToReceive to avoid self-loop after merging. And vice-versa.
		if (edge.Source == nodeToBeMerged.Id && edge.Target == nodeToReceive.Id) || (edge.Source == nodeToReceive.Id && edge.Target == nodeToBeMerged.Id) {
			graph.Edges = RemoveEdge(graph.Edges, edge)
			i-- // Decrement i to account for the removed edge
		}
	}
	
	// Remove the merged node from the graph
	var updatedNodes []Node
	for _, node := range graph.Nodes {
		if node.Id != nodeToBeMerged.Id {
			updatedNodes = append(updatedNodes, node)
		}
	}
	graph.Nodes = updatedNodes

	// Remove the node from potential self-managed libraries in the system context
	for i, library := range graph.System.SelfManagedLibraries {
		var updatedServices []string
		for _, service := range library.ServicesUsingLibrary {
			if service != nodeToBeMerged.Label {
				updatedServices = append(updatedServices, service)
			}
		}
		graph.System.SelfManagedLibraries[i].ServicesUsingLibrary = updatedServices
	}
	
	
	return graph
}

func SanitizeName(name string) string {
	// Remove spaces and special characters from the name to create a valid identifier
	sanitized := strings.ReplaceAll(name, " ", "")
	sanitized = strings.ReplaceAll(sanitized, "-", "")
	sanitized = strings.ReplaceAll(sanitized, "_", "")
	return sanitized
}