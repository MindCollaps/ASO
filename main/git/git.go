package git

import (
	"context"
	"github.com/google/go-github/v56/github"
	"log"
)

// GitHubClients map of GitHub clients
var HubClients map[string]*github.Client

func CheckUser(username string, token string) bool {
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

	if statusUsrName {
		return true
	} else {
		return false
	}
}

func CheckRepoExists(owner string, token string, repo string) bool {
	c := context.Background()
	gitClient := GetGithubClient(token)
	_, _, err := gitClient.Repositories.Get(c, owner, repo)

	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func AddUserToRepo(username string, token string, repoName string, owner string) bool {
	//add user to repo
	c := context.Background()
	//make RepoAddColabo options
	opts := &github.RepositoryAddCollaboratorOptions{
		Permission: "pull",
	}

	gitClient := GetGithubClient(token)
	Repo, _, err := gitClient.Repositories.Get(c, owner, repoName)

	if err != nil {
		log.Println(err)
		return false
	}

	_, _, err = gitClient.Repositories.AddCollaborator(c, owner, Repo.GetName(), username, opts)

	if err != nil {
		return false
	}

	return true
}

func RemoveUserFromRepo(repoOwner string, usernameToBeRemoved string, gitToken string, repoName string) bool {
	c := context.Background()

	gitClient := GetGithubClient(gitToken)
	Repo, _, err := gitClient.Repositories.Get(c, repoOwner, repoName)

	if err != nil {
		return false
	}

	_, err = gitClient.Repositories.RemoveCollaborator(c, repoOwner, Repo.GetName(), usernameToBeRemoved)

	if err != nil {
		inv, _, err := gitClient.Repositories.ListInvitations(c, repoOwner, Repo.GetName(), nil)

		if err != nil {
			log.Println(err)
			return false
		}

		for _, invite := range inv {
			if invite.GetInvitee().GetLogin() == usernameToBeRemoved {
				gitClient.Repositories.DeleteInvitation(c, repoOwner, Repo.GetName(), invite.GetID())
				return true
			}
		}
	}

	return true
}

func CheckIfUserIsColabo(owner string, username string, token string, repo string) bool {
	c := context.Background()
	gitClient := GetGithubClient(token)
	_, _, err := gitClient.Repositories.Get(c, owner, repo)

	if err != nil {
		log.Println(err)
		return false
	}

	isCol, _, err := gitClient.Repositories.IsCollaborator(c, owner, repo, username)

	if err != nil {
		log.Println(err)
		return false
	}

	return isCol
}

func CheckIfUserIsPendingInvite(owner string, username string, token string, repo string) bool {
	c := context.Background()
	gitClient := GetGithubClient(token)
	_, _, err := gitClient.Repositories.Get(c, owner, repo)

	if err != nil {
		log.Println(err)
		return false
	}

	options := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}

	invites, _, err := gitClient.Repositories.ListInvitations(c, owner, repo, options)

	if err != nil {
		log.Println(err)
		return false
	}

	for _, invite := range invites {
		if invite.GetInvitee().GetLogin() == username {
			return true
		}
	}

	return false
}

func GetGithubClient(token string) *github.Client {
	if HubClients == nil {
		HubClients = make(map[string]*github.Client)
	}

	if HubClients[token] == nil {
		HubClients[token] = github.NewClient(nil).WithAuthToken(token)
		if HubClients[token] == nil {
			return nil
		}
		return HubClients[token]
	}

	return HubClients[token]
}

func CheckNewToken(owner string, token string, tokenBefore string) bool {
	if HubClients == nil {
		HubClients = make(map[string]*github.Client)
	}

	if HubClients[tokenBefore] != nil {
		delete(HubClients, tokenBefore)
	}

	gitClient := github.NewClient(nil).WithAuthToken(token)

	if gitClient == nil {
		return false
	}

	_, _, err := gitClient.Users.Get(context.Background(), owner)

	if err != nil {
		return false
	}

	HubClients[token] = gitClient

	return true
}

func GetColabosFromRepo(token string, owner string, repo string) []*github.User {
	c := context.Background()
	gitClient := GetGithubClient(token)
	_, _, err := gitClient.Repositories.Get(c, owner, repo)

	if err != nil {
		log.Println(err)
		return nil
	}

	var collaborators []*github.User

	pageCount := 0

	for {
		options := &github.ListCollaboratorsOptions{
			Affiliation: "all",
			ListOptions: github.ListOptions{
				Page:    pageCount,
				PerPage: 100,
			},
		}

		colabo, resposne, err := gitClient.Repositories.ListCollaborators(c, owner, repo, options)

		collaborators = append(collaborators, colabo...)

		if pageCount == resposne.LastPage || err != nil {
			break
		}

		pageCount++
	}

	if err != nil {
		log.Println(err)
		return nil
	}

	return collaborators
}
