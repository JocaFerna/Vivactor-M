package readFile

import (
	graphparsing "architecture-retrieval/architecture/graphParsing"
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
	// Read the source code from the file path using I/O operations

	return sourceCode, nil
}