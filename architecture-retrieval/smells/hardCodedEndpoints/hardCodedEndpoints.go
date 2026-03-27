package hardCodedEndpoints

import (
	// Graph Parsing
	graphparsing "architecture-retrieval/architecture/graphParsing"
	"regexp"
)

// Given a graph, detects if there are hard-coded endpoints. If there are, it returns a list of all the hard-coded endpoints in the graph. Each hard-coded endpoint is represented as a string.
func DetectHardCodedEndpoints(graph string) ([]string, error) {

	// Parse the graph
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return nil, err
	}

	var hardCodedEndpoints []string
	for _, edge := range graphStruct.Edges {
		call := edge.Properties.CallDefinitionInSource
		if isHardCodedEndpoint(call) {
			hardCodedEndpoints = append(hardCodedEndpoints, call)
		}
	}
	return hardCodedEndpoints, nil
}

// Checks if the call has either
// - Hard-Coded IP address (e.g.: no DNS, localhost, etc.)
// - Hard-Coded Port (e.g.: 80800, 3000, etc.)
func isHardCodedEndpoint(endpoint string) bool {
	// Check for hard-coded IP addresses
	if containsIP(endpoint) {
		return true
	}

	// Check for hard-coded ports
	if containsPort(endpoint) {
		return true
	}

	return false
}

// Checks if the endpoint contains an IP address (e.g., 127.0.0.1)
func containsIP(endpoint string) bool {
	// Simple regex to match IP addresses (both IPv4 and IPv6)
	ipRegex := `(\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b)|(\b[0-9a-fA-F:]+:+[0-9a-fA-F:]+\b)`
	matched, _ := regexp.MatchString(ipRegex, endpoint)
	return matched
}

// Checks if the endpoint contains a hard-coded port (e.g., :8080)
func containsPort(endpoint string) bool{	
	// Simple regex to match ports (e.g., :8080)
	portRegex := `:\d{1,5}\b`
	matched, _ := regexp.MatchString(portRegex, endpoint)
	return matched
}