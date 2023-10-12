package git

import (
	"context"
	"github.com/google/go-github/v56/github"
	"os"
)

var GitHubClient *github.Client
var Repo *github.Repository

func InitGit() bool {
	c := context.Background()
	//get token from .env
	token := os.Getenv("GITHUB_TOKEN")
	GitHubClient = github.NewClient(nil).WithAuthToken(token)

	ownerName := os.Getenv("GITHUB_REPO_OWNER")
	repoName := os.Getenv("GITHUB_REPO_NAME")
	rep, res, err := GitHubClient.Repositories.Get(c, ownerName, repoName)

	Repo = rep

	if err != nil {
		return false
	}

	if res.StatusCode != 200 {
		return false
	}

	if Repo == nil {
		return false
	}

	return true
}

func CheckUser(email string, username string) string {
	usr, res, error := GitHubClient.Users.Get(context.Background(), username)

	statusUsrName := true
	if error != nil {
		statusUsrName = false
	}

	if res.StatusCode != 200 {
		statusUsrName = false
	}

	if usr == nil {
		statusUsrName = false
	}

	//try again with email
	usr, res, error = GitHubClient.Users.Get(context.Background(), email)

	statusEmail := true

	if error != nil {
		statusEmail = false
	}

	if res.StatusCode != 200 {
		statusEmail = false
	}

	if usr == nil {
		statusEmail = false
	}

	if statusUsrName && statusEmail {
		return username
	} else if statusUsrName {
		return username
	} else if statusEmail {
		return email
	} else {
		return ""
	}
}

func AddUserToRepo(username string) bool {
	//add user to repo
	c := context.Background()
	//make RepoAddColabo options
	opts := &github.RepositoryAddCollaboratorOptions{
		Permission: "pull",
	}

	_, _, err := GitHubClient.Repositories.AddCollaborator(c, Repo.Owner.GetLogin(), Repo.GetName(), username, opts)

	if err != nil {
		return false
	}

	return true
}
