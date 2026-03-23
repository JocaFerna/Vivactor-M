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