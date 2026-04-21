package routes

import (
	"architecture-retrieval/architecture"
	"architecture-retrieval/architecture/emulation"

	"architecture-retrieval/refactor/nonAPIVersioned"
	hardCodedEnpointsRefactor "architecture-retrieval/refactor/hardCodedEndpoints"
	sharedPersistencyRefactor "architecture-retrieval/refactor/sharedPersistency"
	sharedLibrariesRefactor "architecture-retrieval/refactor/sharedLibraries"
	wrongCutsRefactor "architecture-retrieval/refactor/wrongCuts"
	microserviceGreedyRefactor "architecture-retrieval/refactor/microserviceGreedy"


	"architecture-retrieval/smells/apiNonVersioned"
	"architecture-retrieval/smells/cyclicDependency"
	"architecture-retrieval/smells/esbUsage"
	"architecture-retrieval/smells/hardCodedEndpoints"
	"architecture-retrieval/smells/inapropriateServiceIntimacity"
	"architecture-retrieval/smells/microserviceGreedy"
	"architecture-retrieval/smells/sharedLibraries"
	"architecture-retrieval/smells/wrongCuts"
	"architecture-retrieval/smells/sharedPersistency"
	"architecture-retrieval/smells/tooManyStandards"
	"architecture-retrieval/smells/noAPIGateway"

	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	gogithub "github.com/google/go-github/v65/github"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type Handler = func(writer http.ResponseWriter, request *http.Request)

// Used to register all routes in the application. This is called in main.go
func Register() {
	log.Println("Registering routes")
	routes := map[string]Handler{
		"/":      home,
		
		// Architecture Handling
		"/cloneRepository":  cloneHandler,
		"/startArchitecture" : startHandler,
		"/emulateArchitecture" : emulateHandler,

		// Smells Detection
		"/smells/apiNonVersioned": apiNonVersionedSmellHandler,
		"/smells/cyclicDependency": cyclicDependencyHandler,
		"/smells/esbUsage": esbUsageHandler,
		"/smells/hardcodedEndpoints": hardcodedEndpointsHandler,
		"/smells/innapropriateServiceIntimacity": inapropriateServiceIntimacityHandler,
		
		// All Reports of Smells Detection
		"/smells/report": smellsHandler,

		// Refactor Handling
		"/refactor/mitigateNonAPIVersionedSmells": nonAPIVersionedHandler,
		"/refactor/mitigateHardcodedEndpointsSmells": hardcodedEndpointsRefactorHandler,
		"/refactor/mitigateSharedPersistencySmells": sharedPersistencyRefactorHandler,
		"/refactor/mitigateSharedLibrariesSmells": sharedLibrariesRefactorHandler,
		"/refactor/mitigateWrongCutsSmells": wrongCutsRefactorHandler,
		"/refactor/mitigateMicroserviceGreedySmells": microserviceGreedyRefactorHandler,
	}

	for route, handler := range routes {
		http.HandleFunc(route, handler)
	}
}



func home(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "{\"message\": \"Hello World\"}")
}

// Refactor of Microservice Greedy -> Handling of the route
func microserviceGreedyRefactorHandler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received mitigate microservice greedy smells request")
	graph := request.URL.Query().Get("graph")
	greedyMicroservices := request.URL.Query().Get("greedyMicroservices")
	// Remove [ and ] from the greedyMicroservices string
	greedyMicroservices = strings.TrimPrefix(greedyMicroservices, "[")
	greedyMicroservices = strings.TrimSuffix(greedyMicroservices, "]")
	greedyMicroservicesList := strings.Split(greedyMicroservices, ",")
	fmt.Printf("Greedy Microservices: %v\n", greedyMicroservicesList)

	graphRefactored, err := microserviceGreedyRefactor.RefactorMicroserviceGreedy(graph, greedyMicroservicesList)
	if err != nil {
		log.Printf("Error mitigating microservice greedy smells: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error mitigating microservice greedy smells\"}"))
		return
	} else {
		// Return 200 OK
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("{\"message\": \"Mitigating microservice greedy smells...\", \"graph\": " + graphRefactored + "}"))
		return
	}
}

// Refactor of Wrong Cuts -> Handling of the route
func wrongCutsRefactorHandler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received mitigate wrong cuts smells request")
	graph := request.URL.Query().Get("graph")
	graphRefactored, err := wrongCutsRefactor.MitigateWrongCuts(graph)
	// For now, let's not return. We are just testing.
	if err != nil {
		log.Printf("Error mitigating wrong cuts smells: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error mitigating wrong cuts smells\"}"))
		return
	} else {
		// Return 200 OK
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("{\"message\": \"Mitigating wrong cuts smells...\", \"graph\": " + graphRefactored + "}"))
		return
	}
}


// Refactor of Shared Libraries -> Handling of the route
func sharedLibrariesRefactorHandler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received mitigate shared libraries smells request")
	graph := request.URL.Query().Get("graph")
	sharedLibrariesSmells := request.URL.Query().Get("sharedLibrariesSmells")
	// Remove [ and ] from the sharedLibrariesSmells string
	sharedLibrariesSmells = strings.TrimPrefix(sharedLibrariesSmells, "[")
	sharedLibrariesSmells = strings.TrimSuffix(sharedLibrariesSmells, "]")
	sharedLibrariesSmellsList := strings.Split(sharedLibrariesSmells, ",")
	fmt.Printf("Shared Libraries Smells: %v\n", sharedLibrariesSmellsList)

	graphRefactored, err := sharedLibrariesRefactor.MitigateSharedLibraries(graph, sharedLibrariesSmellsList)
	if err != nil {
		log.Printf("Error mitigating shared libraries smells: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error mitigating shared libraries smells\"}"))
		return
	} else {
		// Return 200 OK
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("{\"message\": \"Mitigating shared libraries smells...\", \"graph\": " + graphRefactored + "}"))
		return
	}
}

// Refactor of Shared Persistency -> Handling of the route
func sharedPersistencyRefactorHandler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received mitigate shared persistency smells request")

	graph := request.URL.Query().Get("graph")
	sharedPersistencySmells := request.URL.Query().Get("sharedPersistencySmells")
	// Remove [ and ] from the sharedPersistencySmells string
	sharedPersistencySmells = strings.TrimPrefix(sharedPersistencySmells, "[")
	sharedPersistencySmells = strings.TrimSuffix(sharedPersistencySmells, "]")
	sharedPersistencySmellsList := strings.Split(sharedPersistencySmells, ",")
	// Convert the strings in the list to the format of [["service1","service2"],["service3","service4"]]
	sharedPersistencyNestedList := [][]string{}
	
	for _, smell := range sharedPersistencySmellsList {
		smell = strings.TrimPrefix(smell, "[")
		smell = strings.TrimSuffix(smell, "]")
		smellList := strings.Split(smell, ",")
		sharedPersistencyNestedList = append(sharedPersistencyNestedList, smellList)
	}
	fmt.Printf("Shared Persistency Smells: %v\n", sharedPersistencySmellsList)

	graphRefactored, err := sharedPersistencyRefactor.MitigateSharedPersistencySmellsGraph(graph, sharedPersistencyNestedList)
	if err != nil {
		log.Printf("Error mitigating shared persistency smells: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error mitigating shared persistency smells\"}"))
		return
	} else {
		// Return 200 OK
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("{\"message\": \"Mitigating shared persistency smells...\", \"graph\": " + graphRefactored + "}"))
		return
	}
}

// Refactor of Hardcoded Endpoints -> Handling of the route
func hardcodedEndpointsRefactorHandler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received mitigate hardcoded endpoints smells request")
	
	graph := request.URL.Query().Get("graph")
	endpoints := request.URL.Query().Get("endpoints")
	// Remove [ and ] from the endpoints string
	endpoints = strings.TrimPrefix(endpoints, "[")
	endpoints = strings.TrimSuffix(endpoints, "]")
	hardcodedEndpoints := strings.Split(endpoints, ",")
	fmt.Printf("Hardcoded Endpoints: %v\n", hardcodedEndpoints)


	graphRefactored, err := hardCodedEnpointsRefactor.RefactorGraphWithHardcodedEndpoints(graph, hardcodedEndpoints)
	
	fmt.Printf("Refactored Graph: %s\n", graphRefactored)
	if err != nil {
		fmt.Printf("Error mitigating hardcoded endpoints smells: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error mitigating hardcoded endpoints smells\"}"))
		return
	} else {
		// Return 200 OK
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("{\"message\": \"Mitigating hardcoded endpoints smells...\", \"graph\": " + graphRefactored + "}"))
		return
	}
}

// All Smells Report -> Handling of the route
func smellsHandler(writer http.ResponseWriter, request *http.Request){	log.Println("Received Smells Report request")
	// Get the url properly
	graph := request.URL.Query().Get("graph")
	log.Printf("Generating Smells Report")

	// Call all the smells detection functions and gather their results
	apiNonVersionedSmells, err := apiNonVersioned.DetectApiNonVersioned(graph)
	if err != nil {
		log.Printf("Error detecting API Non-Versioned smells: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error detecting API Non-Versioned smells\"}"))
		return
	}

	// Cyclic Dependency
	cyclicDependencySmells, err := cyclicDependency.DetectCyclicDependency(graph)
	if err != nil {
		log.Printf("Error detecting Cyclic Dependency smells: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error detecting Cyclic Dependency smells\"}"))
		return
	}

	// ESB Usage
	esbUsageSmell, err := esbUsage.DetectESBUsage(graph)
	if err != nil {
		log.Printf("Error detecting ESB Usage smells: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error detecting ESB Usage smells\"}"))
		return
	}

	// Hardcoded Endpoints
	hardcodedEndpointsSmells, err := hardCodedEndpoints.DetectHardCodedEndpoints(graph)
	if err != nil {
		log.Printf("Error detecting Hardcoded Endpoints smells: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error detecting Hardcoded Endpoints smells\"}"))
		return
	}

	// Inapropriate Service Intimacity
	innapropriateServiceIntimacitySmells, err := inapropriateServiceIntimacity.GetInnapropriateServiceIntimacity(graph)
	if err != nil {
		log.Printf("Error detecting Innapropriate Service Intimacity smells: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error detecting Innapropriate Service Intimacity smells\"}"))
		return
	}

	// Microservice Greedy
	microserviceGreedySmells, err := microserviceGreedy.GetMicroserviceGreedy(graph)
	if err != nil {
		log.Printf("Error detecting Microservice Greedy smells: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error detecting Microservice Greedy smells\"}"))
		return
	}

	// Shared Libraries
	sharedLibrariesSmells, err := sharedLibraries.GetSharedLibraries(graph)
	if err != nil {
		log.Printf("Error detecting Shared Libraries smells: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error detecting Shared Libraries smells\"}"))
		return
	}

	// Wrong Cuts
	wrongCutsSmells, err := wrongCuts.GetWrongCuts(graph)
	wrongCutsNA := false
	// This case may be due to high demand on Gemini API, so we can just say that is N/A.
	if err != nil {
		wrongCutsNA = true
	}

	// Shared Persistency
	sharedPersistencySmells, err := sharedPersistency.GetSharedPersistency(graph)
	if err != nil {
		log.Printf("Error detecting Shared Persistency smells: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error detecting Shared Persistency smells\"}"))
		return
	}
	
	// Too Many Standards
	tooManyStandardsSmells, err := tooManyStandards.GetTooManyStandards(graph)
	if err != nil {
		log.Printf("Error detecting Too Many Standards smells: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error detecting Too Many Standards smells\"}"))
		return
	}

	// No API Gateway
	noAPIGatewaySmells, err := noAPIGateway.GetNoAPIGateway(graph)
	if err != nil {
		log.Printf("Error detecting No API Gateway smells: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error detecting No API Gateway smells\"}"))
		return
	}

	// Return 200 OK with all the smells in the response body
	writer.WriteHeader(http.StatusOK)
	response := "{\"message\": \"Smells Report generated successfully\", \"smells\": {"
	response += "\"apiNonVersioned\": ["
	for i, smell := range apiNonVersionedSmells {
		response += fmt.Sprintf("\"%s\"", smell)
		if i < len(apiNonVersionedSmells)-1 {
			response += ","
		}
	}
	response += "],"
	response += "\"cyclicDependency\": ["
	for i, smell := range cyclicDependencySmells {
		response += "["
		for j, node := range smell {
			response += fmt.Sprintf("\"%s\"", node)
			if j < len(smell)-1 {
				response += ","
			}
		}
		response += "]"
		if i < len(cyclicDependencySmells)-1 {
			response += ","
		}
	}
	response += "],"
	response += "\"esbUsage\": " + fmt.Sprintf("%t", esbUsageSmell) + ","
	response += "\"hardcodedEndpoints\": ["
	for i, smell := range hardcodedEndpointsSmells {
		response += fmt.Sprintf("\"%s\"", smell)
		if i < len(hardcodedEndpointsSmells)-1 {
			response += ","
		}
	}
	response += "],"
	response += "\"innapropriateServiceIntimacity\": ["
	for i, smell := range innapropriateServiceIntimacitySmells {
		response += fmt.Sprintf("\"%s\"", smell)
		if i < len(innapropriateServiceIntimacitySmells)-1 {
			response += ","
		}
	}
	response += "],"
	response += "\"microserviceGreedy\": ["
	for i, smell := range microserviceGreedySmells {
		response += fmt.Sprintf("\"%s\"", smell)
		if i < len(microserviceGreedySmells)-1 {
			response += ","
		}
	}
	response += "],"
	response += "\"sharedLibraries\": ["
	for i, smell := range sharedLibrariesSmells {
		response += fmt.Sprintf("\"%s\"", smell)
		if i < len(sharedLibrariesSmells)-1 {
			response += ","
		}
	}
	response += "],"
	response += "\"wrongCuts\": "
	if wrongCutsNA {
		response += "\"N/A\""
	} else {
		response += fmt.Sprintf("%t", wrongCutsSmells)
	}
	response += ","
	response += "\"sharedPersistency\": ["
	for i, smell := range sharedPersistencySmells {
		response += "["
		for j, service := range smell {
			response += fmt.Sprintf("\"%s\"", service)
			if j < len(smell)-1 {
				response += ","
			}
		}
		response += "]"
		if i < len(sharedPersistencySmells)-1 {
			response += ","
		}
	}
	response += "],"
	response += "\"tooManyStandards\": "
	response += fmt.Sprintf("%v", tooManyStandardsSmells)
	response += ","
	response += "\"noAPIGateway\": "
	response += fmt.Sprintf("%t", noAPIGatewaySmells)

	response += "}}"
	writer.Write([]byte(response))
	return
	
}
// Hardcoded Endpoints -> Handling of the route
func hardcodedEndpointsHandler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received Hardcoded Endpoints detection request")
	// Get the url properly
	graph := request.URL.Query().Get("graph")
	log.Printf("Detecting Hardcoded Endpoints smells")
	
	
	hardcodedEndpointsSmells, err := hardCodedEndpoints.DetectHardCodedEndpoints(graph)
	if err != nil {
		log.Printf("Error detecting Hardcoded Endpoints smells: %s\n", err.Error())
		
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error detecting Hardcoded Endpoints smells\"}"))
		return
	} else {
		// Return 200 OK
		
		writer.WriteHeader(http.StatusOK)
		// Add hardcodedEndpointsSmells to the response body
		response := "{\"message\": \"Hardcoded Endpoints smells detected successfully\", \"smells\": {\"hardcodedEndpoints\": ["
		for i, smell := range hardcodedEndpointsSmells {
			response += fmt.Sprintf("\"%s\"", smell)
			if i < len(hardcodedEndpointsSmells)-1 {
				response += ","
			}
		}
		response += "]}}"
		writer.Write([]byte(response))
		return
	}
}

// Innapropriate Service Intimacity -> Handling of the route
func inapropriateServiceIntimacityHandler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received Innapropriate Service Intimacity detection request")
	// Get the url properly
	graph := request.URL.Query().Get("graph")
	log.Printf("Detecting Innapropriate Service Intimacity smells")
	
	
	innapropriateServiceIntimacitySmells, err := inapropriateServiceIntimacity.GetInnapropriateServiceIntimacity(graph)
	if err != nil {
		log.Printf("Error detecting Innapropriate Service Intimacity smells: %s\n", err.Error())
		
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error detecting Innapropriate Service Intimacity smells\"}"))
		return
	} else {
		// Return 200 OK
		
		writer.WriteHeader(http.StatusOK)
		// Add innapropriateServiceIntimacitySmells to the response body
		response := "{\"message\": \"Innapropriate Service Intimacity smells detected successfully\", \"smells\": {\"innapropriateServiceIntimacity\": ["
		for i, smell := range innapropriateServiceIntimacitySmells {
			response += fmt.Sprintf("\"%s\"", smell)
			if i < len(innapropriateServiceIntimacitySmells)-1 {
				response += ","
			}
		}
		response += "]}}"
		writer.Write([]byte(response))
		return
	}
}


// ESB Usage -> Handling of the route
func esbUsageHandler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received ESB Usage detection request")
	// Get the url properly
	graph := request.URL.Query().Get("graph")
	log.Printf("Detecting ESB Usage smells")
	
	
	esbUsageSmell, err := esbUsage.DetectESBUsage(graph)
	if err != nil {
		log.Printf("Error detecting ESB Usage smells: %s\n", err.Error())
		
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error detecting ESB Usage smells\"}"))
		return
	} else {
		// Return 200 OK
		
		writer.WriteHeader(http.StatusOK)
		response := "{\"message\": \"ESB Usage smells detected successfully\", \"smells\": {\"esbUsage\": " + fmt.Sprintf("%t", esbUsageSmell) + "}}"
		writer.Write([]byte(response))
		return
	}
}


// Cyclic Dependency -> Handling of the route
func cyclicDependencyHandler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received Cyclic Dependency detection request")
	// Get the url properly
	graph := request.URL.Query().Get("graph")
	log.Printf("Detecting Cyclic Dependencies")
	
	
	cycles, err := cyclicDependency.DetectCyclicDependency(graph)
	if err != nil {
		log.Printf("Error detecting Cyclic Dependencies: %s\n", err.Error())
		
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error detecting Cyclic Dependencies\"}"))
		return
	} else {
		// Return 200 OK
		
		writer.WriteHeader(http.StatusOK)
		// Add cycles to the response body
		response := "{\"message\": \"Cyclic Dependencies detected successfully\", \"smells\": {\"cycles\": ["
		for i, cycle := range cycles {
			response += "["
			for j, node := range cycle {
				response += fmt.Sprintf("\"%s\"", node)
				if j < len(cycle)-1 {
					response += ","
				}
			}
			response += "]"
			if i < len(cycles)-1 {
				response += ","
			}
		}
		response += "]}}"
		writer.Write([]byte(response))
		return
	}
}

// Emulate the architecture -> Handling of the route
func emulateHandler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received Architecture Handling")
	// Get the url properly
	graph := request.URL.Query().Get("graph")
	
	// Call Architecture Emulation Block
	err := emulation.EmulateArchitecture(graph)
	if err != nil {
		log.Printf("Error starting to emulate the desired architecture: %s\n", err.Error())
		
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error starting to emulate the desired architecture\"}"))
		return
	} else {
		// Return 200 OK
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("{\"message\": \"Architecture Emulated Successfully!\"}"))
		return
	}
}
func apiNonVersionedSmellHandler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received API Non-Versioned detection request")
	// Get the url properly
	graph := request.URL.Query().Get("graph")
	log.Printf("Detecting API Non-Versioned smells")
	
	
	nonAPIVersionedSmells, err := apiNonVersioned.DetectApiNonVersioned(graph)
	if err != nil {
		log.Printf("Error detecting API Non-Versioned smells: %s\n", err.Error())
		
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error detecting API Non-Versioned smells\"}"))
		return
	} else {
		// Return 200 OK
		
		writer.WriteHeader(http.StatusOK)
		// Add nonAPIVersionedSmells to the response body
		response := "{\"message\": \"API Non-Versioned smells detected successfully\", \"smells\": {\"nonAPIVersionedEndpoints\": ["
		for i, smell := range nonAPIVersionedSmells {
			response += fmt.Sprintf("\"%s\"", smell)
			if i < len(nonAPIVersionedSmells)-1 {
				response += ","
			}
		}
		response += "]}}"
		writer.Write([]byte(response))
		return
	}
}
func nonAPIVersionedHandler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received mitigate non-API versioned smells request")
	
	graph := request.URL.Query().Get("graph")
	apis := request.URL.Query().Get("apis")
	// Remove [ and ] from the apis string
	apis = strings.TrimPrefix(apis, "[")
	apis = strings.TrimSuffix(apis, "]")
	nonAPIVersionedSmells := strings.Split(apis, ",")
	fmt.Printf("Non-API Versioned Smells: %v\n", nonAPIVersionedSmells)


	graphRefactored, err := nonAPIVersioned.MitigateNonAPIVersionedSmellsGraph(graph, nonAPIVersionedSmells)
	fmt.Printf("Refactored Graph: %s\n", graphRefactored)
	if err != nil {
		fmt.Printf("Error mitigating non-API versioned smells: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error mitigating non-API versioned smells\"}"))
		return
	} else {
		// Return 200 OK
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("{\"message\": \"Mitigating shared library smells...\", \"graph\": " + graphRefactored + "}"))
		return
	}
}

func cloneHandler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received clone request")
	// Get the url properly
	url := request.URL.Query().Get("url")
	last_appearance_of_separator := strings.LastIndex(url,"/")
	
	// Create the path to save the file
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(cwd,"downloads",request.URL.Query().Get("url")[last_appearance_of_separator:])

	err = os.MkdirAll(path,os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	

	err = architecture.CloneRepository(url,path,request.URL.Query().Get("url")[last_appearance_of_separator:])

	if err != nil{
		log.Printf("Error cloning repository: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Println("Done cloning to " + path)

	// Return 200 OK
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("{\"message\": \"Repository cloned successfully\"}"))
	return

}
func startHandler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received start architecture request")

	// Get the url properly
	url := request.URL.Query().Get("url")
	command := request.URL.Query().Get("command")
	packageList := request.URL.Query().Get("packages")
	last_appearance_of_separator := strings.LastIndex(url,"/")

	repo_name := request.URL.Query().Get("url")[last_appearance_of_separator:]
	log.Printf("Starting architecture for repository: %s\n", repo_name)

	err := architecture.StartArchitecture(repo_name,command,packageList)
	if err != nil {
		log.Printf("Error starting architecture: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("{\"message\": \"Error starting architecture\"}"))
		return
	} else {
		// Return 200 OK
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("{\"message\": \"Starting architecture...\"}"))
		return
	}

	
	
}

/**func callback(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received callback request")
	code := request.URL.Query().Get("code")

	conf := getOAuthConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// conf.Exchange(ctx, code)
	log.Printf("Exchanging code for token: %s\n", code)
	log.Printf("OAuth config: %+v\n", conf)
	log.Printf("Context: %+v\n", ctx)

	token, err := conf.Exchange(ctx, code)
	log.Printf("Exchanged code for token: %s\n", token)
	if err != nil {
		log.Printf("Error exchanging code for token: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Println("Received token: %s\n", token.AccessToken)
	repos, err := getCurrentUserRepos(token.AccessToken)
	if err != nil {
		log.Printf("Error fetching user repositories: %s\n", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("Found %d repositories\n", len(repos))
	for _, repo := range repos {
		writer.Write([]byte(repo.GetFullName() + "\n"))
	}
	writer.WriteHeader(http.StatusOK)
	architecture.ProcessRepositories(token.AccessToken, repos)
}*/

/*func redirect(writer http.ResponseWriter, request *http.Request) {
	http.Redirect(writer, request, getRedirectURL(), http.StatusTemporaryRedirect)
}*/

func getOAuthConfig() *oauth2.Config {
	value_client := os.Getenv("GITHUB_CLIENT_ID")
	value_secret := os.Getenv("GITHUB_SECRET")
	if len(value_client) == 0 || len(value_secret) == 0 {
		log.Fatal("GITHUB_CLIENT_ID and GITHUB_SECRET must be set in environment variables")
	}
	return &oauth2.Config{
		ClientID:     value_client,
		ClientSecret: value_secret,
		Scopes:       []string{"user:email", "repo:public_repo"},
		Endpoint:     github.Endpoint,
		RedirectURL:  "http://localhost:8000/callback",
	}
}

func getRedirectURL() string {
	conf := getOAuthConfig()
	return conf.AuthCodeURL("state",)
}

func getCurrentUserRepos(accessToken string) ([]*gogithub.Repository, error) {
	client := gogithub.NewClient(nil).WithAuthToken(accessToken)

	opt := &gogithub.RepositoryListByAuthenticatedUserOptions{
		Affiliation: "owner",
	}
	opt.PerPage = 50
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	repos, _, err := client.Repositories.ListByAuthenticatedUser(ctx, opt)

	return repos, err
}