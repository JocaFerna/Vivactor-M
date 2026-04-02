package nonAPIVersioned

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	graphparsing "architecture-retrieval/architecture/graphParsing"
	"architecture-retrieval/refactor/utils"
)

func MitigateNonAPIVersionedSmellsGraph(graph string, nonAPIVersionedSmells []string) (string, error) {
	graphStruct, err := graphparsing.ParseGraph(graph)
	if err != nil {
		return "", err
	}
	refactoredSmells := []string{}
	
	basePath := utils.GetBasePathOfGraph(graphStruct)

	services := make(map[string]bool)

	timerforMitigation := time.Now()

	// First, let's update the graph.
	for _,smell := range nonAPIVersionedSmells {
		// Due to the nature we receive the smells, we need to remove the \"
		smell = strings.TrimPrefix(smell, "\"")
		smell = strings.TrimSuffix(smell, "\"")
		for i, edge := range graphStruct.Edges {
			if edge.Properties.CallDefinitionInSource == smell {


				// Update the source file.
				filePathSource, err := utils.GetFilePathFromNode(graphStruct, edge.Source)
				if err != nil {
					return "", fmt.Errorf("error getting file path from node %s: %v", edge.Source, err)
				}
				fileSource, err := utils.ReadFileContent(filePathSource)
				if err != nil {
					return "", fmt.Errorf("error reading file %s: %v", filePathSource, err)
				}
				
				updatedFileSource := strings.ReplaceAll(fileSource, graphStruct.Edges[i].Endpoint, "/v1"+graphStruct.Edges[i].Endpoint)
				err = utils.WriteFileContent(filePathSource, updatedFileSource)
				if err != nil {
					return "", fmt.Errorf("error writing file %s: %v", filePathSource, err)
				}
				
				serviceName, err := graphparsing.GetNodeById(graphStruct, edge.Source)
				if exists, _ := services[utils.SanitizeName(serviceName.Label)]; !exists {
					services[utils.SanitizeName(serviceName.Label)] = true
				}

				// Update the target file.
				filePathTarget, err := utils.GetFilePathFromNode(graphStruct, edge.Target)
				if err != nil {
					return "", fmt.Errorf("error getting file path from node %s: %v", edge.Target, err)
				}
				fileTarget, err := utils.ReadFileContent(filePathTarget)
				if err != nil {
					return "", fmt.Errorf("error reading file %s: %v", filePathTarget, err)
				}
				updatedFileTarget := strings.ReplaceAll(fileTarget, graphStruct.Edges[i].Endpoint, "/v1"+graphStruct.Edges[i].Endpoint)
				err = utils.WriteFileContent(filePathTarget, updatedFileTarget)
				if err != nil {
					return "", fmt.Errorf("error writing file %s: %v", filePathTarget, err)
				}

				serviceName, err = graphparsing.GetNodeById(graphStruct, edge.Target)
				if exists, _ := services[utils.SanitizeName(serviceName.Label)]; !exists {
					services[utils.SanitizeName(serviceName.Label)] = true
				}

				// Add the versioning. (Always v1 for now).
				graphStruct.Edges[i].Endpoint = "/v1" + graphStruct.Edges[i].Endpoint
				// Also, update the call definition in source, as it is used to detect the smell.
				// We will add it before the endpoint.
				graphStruct.Edges[i].Properties.CallDefinitionInSource = addVersioningToCallDefinition(graphStruct.Edges[i].Properties.CallDefinitionInSource)


				// Add to refactored smells, to be able to return it later.
				refactoredSmells = append(refactoredSmells, graphStruct.Edges[i].Properties.CallDefinitionInSource)

			}
		}
	}

	// Inserting a delay here to give the containers time to restart and become healthy after the changes. This is a bit of a hack, but it should work for now. We will replace this with a more robust solution later.
	log.Println("Waiting for services to restart and become healthy...")
	time.Sleep(20 * time.Second)

	// Before returning the graph, we need to check that everything is fine.
	// We must check the health of the containers.
	// This works by running `docker compose ps --format json` and parsing the output to verify that all expected services are running and healthy. We will retry this check several times with a delay in between, to give the containers time to start up and become healthy.
	maxRetries := 30
	var waitErr error

	expectedServices := make(map[string]int)
	for serviceName := range services {
		expectedServices[serviceName] = 1
	}


	for i := 0; i < maxRetries; i++ {
		err := utils.CheckComposeHealth(basePath, expectedServices, timerforMitigation)
		var nonHealthy = false

		if err == nil {
			// Save serviceNames into global value
			log.Println("Architecture is healthy!")
			waitErr = nil
			break
		}

		switch err.(type) {

		case *utils.ServiceUnhealthyError:
			nonHealthy = true
			log.Printf("Service unhealthy: %v", err)

		case *utils.ServiceNotStartedError:
			log.Printf("Service not started yet: %v", err)

		default:
			log.Printf("Unknown error: %v", err)
		}
		if !nonHealthy {
			log.Printf("Waiting for services to start... (attempt %d/%d)\n", i+1, maxRetries)
			// Wait for 10 seconds before retrying
			time.Sleep(10 * time.Second)
			waitErr = err
		} else {
			log.Printf("Services not healthy!")
			//IDEA: This could need a "break" here but I don't want to cause any other bugs now.
		}
		
	}

	if waitErr != nil {
		log.Printf("Wait failed after retries: %s. Returning error.\n", waitErr.Error())
		return "", waitErr
	}

	log.Println("Refactored smells: ", refactoredSmells)

    log.Println("Architecture is healthy")

	// Now, let's test if the calls are working as they should
	//FIXME: This is a bit messy.
	for i := 0; i < maxRetries; i++ {
		allCallsVerified := true
		// First, let's update the graph.
		for _,smell := range refactoredSmells {
			// Due to the nature we receive the smells, we need to remove the \"
			smell = strings.TrimPrefix(smell, "\"")
			smell = strings.TrimSuffix(smell, "\"")
			for i, edge := range graphStruct.Edges {
				if edge.Properties.CallDefinitionInSource == smell {

					serviceName, err := graphparsing.GetNodeById(graphStruct, edge.Source)
					if err != nil {
						return "", fmt.Errorf("error getting node by id %s: %v", edge.Source, err)
					}

					verified, err := VerifyCallsViaLogs(utils.SanitizeName(serviceName.Label), edge.Endpoint)
					if err != nil {
						log.Printf("Error verifying calls via logs: %v", err)
						verified = false
					}

					if !verified {
						log.Printf("Call not verified for service %s and endpoint %s. Retrying... (attempt %d/%d)\n", serviceName.Label, edge.Endpoint, i+1, maxRetries)
						allCallsVerified = false
						continue
					} else {
						log.Printf("Call verified for service %s and endpoint %s.\n", serviceName.Label, edge.Endpoint)
					}		
				}
			}
			
		}
		if allCallsVerified {
			log.Println("All calls verified!")
			break
		} else {
			log.Printf("Not all calls verified. Retrying... (attempt %d/%d)\n", i+1, maxRetries)
			// Wait for 10 seconds before retrying
			time.Sleep(10 * time.Second)
		}
	}


	// Return the updated graph.
	graphString, err := graphparsing.SerializeGraph(graphStruct)
	if err != nil {
		return "", fmt.Errorf("error serializing graph: %v", err)
	}

	return graphString, nil
}

// This function adds versioning to the call definition in source. It assumes that the call definition is in the format "package/path/to/service/endpoint" and it adds "/v1" after the third "/". For example, if the call definition is "github.com/myorg/myservice/endpoint", it will become "github.com/myorg/myservice/v1/endpoint".
func addVersioningToCallDefinition(callDefinition string) string {
	// Split data by third appearance of "/"
	parts := strings.SplitN(callDefinition, "/", 4)
	if len(parts) < 4 {
		return callDefinition
	}
	// Add versioning after the third "/"
	return parts[0] + "/" + parts[1] + "/" + parts[2] + "/v1/" + parts[3]
}

func VerifyCallsViaLogs(serviceName string, expectedEndpoint string) (bool, error) {
    // Command: docker logs --tail 20 <serviceName>
    cmd := exec.Command("docker", "logs", "--tail", "20", serviceName)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return false, err
    }

    logLines := string(output)

	// Let's take only the last 5 lines of the logs, as the call should be there if the service is healthy and the call is working. This is to avoid false positives from old logs.
	lastLines := strings.Split(logLines, "\n")
	if len(lastLines) > 5 {
		lastLines = lastLines[len(lastLines)-5:]
	}
	logLines = strings.Join(lastLines, "\n")
	
    
    // We are looking for a line that contains BOTH our new endpoint and a success indicator
    // E.g., "[ordersservice] Response from inventoryservice on /v1/endpoint: 200" 
    // and the endpoint must have /v1
	
	// Go through each line of logLines
	for _, line := range strings.Split(logLines, "\n") {
		if strings.Contains(line, expectedEndpoint) && strings.Contains(line, "200") {
			return true, nil
		}
	}
    return false, nil
}

