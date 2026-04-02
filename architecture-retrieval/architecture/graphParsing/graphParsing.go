package graphparsing

import (
	"encoding/json"
	"fmt"
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