package readFile

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
	"os"
)

func ReadFile(graphData string, serviceName string) (string, error) {
	// Parse the graph data to find the service and its source code
	graph, err := graphparsing.ParseGraph(graphData)
	if err != nil {
		return "", err
	}

	sourceCodePath, err := graphparsing.GetServiceSourceCodePath(graph, serviceName)
	if err != nil {
		return "", err
	}
	if sourceCodePath == "Database" {
		return "This service is a Database and does not have source code to display.", nil
	}
	// Read the source code from the file path and return it as a string
	bytes, err := os.ReadFile(sourceCodePath)
	if err != nil {
		return "", err
	}

	// Convert the byte slice directly to a standard string
	return string(bytes), nil
}