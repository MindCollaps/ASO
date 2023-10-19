package git

import (
	"context"
	"fmt"
	"github.com/google/go-github/v56/github"
)

// map of github clients
var GitHubClients map[string]*github.Client

func CheckUser(email string, username string, token string) string {
	gitClient := GetGithubClient(token)

	usr, res, error := gitClient.Users.Get(context.Background(), username)

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
	usr, res, error = gitClient.Users.Get(context.Background(), email)

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

func CheckRepoExists(owner string, token string, repo string) bool {
	c := context.Background()
	gitClient := GetGithubClient(token)
	_, _, err := gitClient.Repositories.Get(c, owner, repo)

	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func AddUserToRepo(username string, token string, repoName string) bool {
	//add user to repo
	c := context.Background()
	//make RepoAddColabo options
	opts := &github.RepositoryAddCollaboratorOptions{
		Permission: "pull",
	}

	gitClient := GetGithubClient(token)
	Repo, _, err := gitClient.Repositories.Get(c, username, repoName)

	if err != nil {
		return false
	}

	_, _, err = gitClient.Repositories.AddCollaborator(c, Repo.Owner.GetLogin(), Repo.GetName(), username, opts)

	if err != nil {
		return false
	}

	return true
}

func RemoveUserFromRepo(repoOwner string, usernameToBeAdded string, gitToken string, repoName string) bool {
	//add user to repo
	c := context.Background()

	gitClient := GetGithubClient(gitToken)
	Repo, _, err := gitClient.Repositories.Get(c, repoOwner, repoName)

	if err != nil {
		return false
	}

	_, err = gitClient.Repositories.RemoveCollaborator(c, repoOwner, Repo.GetName(), usernameToBeAdded)

	if err != nil {
		return false
	}

	return true
}

func CheckIfRepoExistsAndEditRights(owner string, username string, token string, repo string) bool {
	c := context.Background()
	gitClient := GetGithubClient(token)
	_, _, err := gitClient.Repositories.Get(c, "", repo)

	if err != nil {
		return false
	}

	isCol, _, err := gitClient.Repositories.IsCollaborator(c, owner, repo, username)

	return isCol
}

func GetGithubClient(token string) *github.Client {
	if GitHubClients == nil {
		GitHubClients = make(map[string]*github.Client)
	}

	if GitHubClients[token] == nil {
		GitHubClients[token] = github.NewClient(nil).WithAuthToken(token)
		if GitHubClients[token] == nil {
			return nil
		}
		return GitHubClients[token]
	}

	return GitHubClients[token]
}

func CheckNewToken(owner string, token string, tokenBefore string) bool {
	if GitHubClients == nil {
		GitHubClients = make(map[string]*github.Client)
	}

	if GitHubClients[tokenBefore] != nil {
		delete(GitHubClients, tokenBefore)
	}

	gitClient := github.NewClient(nil).WithAuthToken(token)

	if gitClient == nil {
		return false
	}

	_, _, err := gitClient.Users.Get(context.Background(), owner)

	if err != nil {
		return false
	}

	GitHubClients[token] = gitClient

	return true
}
