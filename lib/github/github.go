package github

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

type UserRepo struct {
	org, team, token string
	httpClient       *http.Client
	githubClient     *github.Client
}

func New(ctx context.Context, org, role, token string) *UserRepo {
	t := oauth2.Token{AccessToken: token}
	ts := oauth2.StaticTokenSource(&t)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &UserRepo{org, role, token, http.DefaultClient, client}
}

type User struct {
	Username string
	FullName string
}

func (r *UserRepo) GetAll(ctx context.Context) (map[string]*User, error) {
	githubUsers, _, err := r.githubClient.Teams.ListTeamMembersBySlug(ctx, r.org, r.team, nil)
	if err != nil {
		return nil, err
	}

	users := make(map[string]*User)
	for _, githubUser := range githubUsers {
		user := new(User)
		if githubUser.Login != nil {
			user.Username = *githubUser.Login
		}
		if githubUser.Name != nil {
			user.FullName = *githubUser.Name
		}

		users[*githubUser.Login] = user
	}

	return users, nil
}

func (r *UserRepo) GetAuthorizedKeys(username string) ([]byte, error) {
	response, err := r.httpClient.Get( fmt.Sprintf("https://github.com/%s.keys", username))
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code: %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
