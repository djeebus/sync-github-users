package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/djeebus/github-users-sync/lib/github"
	"github.com/djeebus/github-users-sync/lib/linux"
	"github.com/pkg/errors"
)

func main() {
	var (
		err error

		ctx   = context.Background()
	)

	config, err := getConfig()
	if err != nil {
		panic(err)
	}

	remote := github.New(ctx, config.org, config.team, config.token)
	local := linux.New()

	fmt.Printf("[github] getting all users in %s/%s ...\n", config.org, config.team)
	remoteUsers, err := remote.GetAll(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("[local] getting all local users ... \n")
	localUsers, err := local.GetAll()
	if err != nil {
		panic(err)
	}

	// add new local users
	for username, remoteUser := range remoteUsers {
		localUser, ok := localUsers[username]
		if ok {
			fmt.Printf("[%s] enabling user ...\n", username)
			if err = local.EnableUser(username); err != nil {
				fmt.Printf("failed to enable %s: %v\n", username, err)
				continue
			}
		} else {
			fmt.Printf("[%s] creating user ...\n", username)
			if localUser, err = local.New(remoteUser); err != nil {
				fmt.Printf("failed to create local user %s: %v\n", username, err)
				continue
			}
		}

		fmt.Printf("[%s] downloading authorized keys ...\n", username)
		if err = writeAuthorizedKeys(remote, remoteUser, localUser); err != nil {
			fmt.Printf("failed to write auth keys for %s: %v\n", username, err)
			continue
		}
	}

	// remove local users that have been removed from remote
	for username := range localUsers {
		_, ok := remoteUsers[username]
		if !ok {
			fmt.Printf("[%s] disabling user ...\n", username)
			if err = local.DisableUser(username); err != nil {
				fmt.Printf("failed to disable %s: %v\n", username, err)
			}
			continue
		}

		fmt.Printf("[%s] enabling user ...\n", username)
		if err = local.EnableUser(username); err != nil {
			fmt.Printf("failed to enable %s: %v\n", username, err)
		}
	}
}

func writeAuthorizedKeys(remote *github.UserRepo, githubUser *github.User, localUser *linux.User) error {
	data, err := remote.GetAuthorizedKeys(githubUser.Username)
	if err != nil {
		return errors.Wrapf(err, "failed to get authorized keys for %s", githubUser.Username)
	}

	sshConfigPath := fmt.Sprintf("/home/%s/.ssh/", localUser.Login)
	if err = os.MkdirAll(sshConfigPath, 0o700); err != nil {
		return errors.Wrapf(err, "failed to create '%s'", sshConfigPath)
	}

	authKeysPath := filepath.Join("/", "home", localUser.Login, ".ssh", "authorized_keys")
	if err = ioutil.WriteFile(authKeysPath, data, 0o600); err != nil {
		return errors.Wrapf(err, "failed to write auth keys data")
	}

	for _, path := range []string{sshConfigPath, authKeysPath} {
		if err = os.Chown(path, localUser.UID, localUser.GID); err != nil {
			return errors.Wrapf(err, "failed to chown %s", path)
		}
	}

	return nil
}
