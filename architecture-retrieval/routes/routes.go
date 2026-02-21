package routes

import (
	"fmt"
	"log"
	"context"
	"time"
	"net/http"
	"architecture-retrieval/architecture"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"os"
	gogithub "github.com/google/go-github/v65/github"
	"path/filepath"
	"strings"
)

type Handler = func(writer http.ResponseWriter, request *http.Request)

// Used to register all routes in the application. This is called in main.go
func Register() {
	log.Println("Registering routes")
	routes := map[string]Handler{
		"/":      home,
		"/cloneRepository":  cloneHandler,
		"/startArchitecture" : startHandler,
	}

	for route, handler := range routes {
		http.HandleFunc(route, handler)
	}
}



func home(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "{\"message\": \"Hello World\"}")
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
	last_appearance_of_separator := strings.LastIndex(url,"/")

	repo_name := request.URL.Query().Get("url")[last_appearance_of_separator:]
	log.Printf("Starting architecture for repository: %s\n", repo_name)

	architecture.StartArchitecture(repo_name,command)

	// Return 200 OK
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("{\"message\": \"Starting architecture...\"}"))
	return
	
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