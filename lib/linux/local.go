package linux

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"github.com/djeebus/github-users-sync/lib/github"
	"github.com/pkg/errors"
	"github.com/willdonnelly/passwd"
)

const (
	MinUID = 5000
	MaxUID = 6000
)

type UserRepo struct {
}

func New() *UserRepo {
	return new(UserRepo)
}

type User struct {
	Login string
	UID   int
	GID   int
}

func (r *UserRepo) GetAll() (map[string]*User, error) {
	localUsers, err := passwd.Parse()
	if err != nil {
		return nil, err
	}

	users := make(map[string]*User)
	for username, localUser := range localUsers {
		uid, err := strconv.Atoi(localUser.Uid)
		if err != nil {
			fmt.Printf("failed to convert uid '%s': %v\n", localUser.Uid, err)
			continue
		}

		if uid > MaxUID {
			continue
		}

		if uid < MinUID {
			continue
		}

		user := new(User)
		user.Login = username

		if user.UID, err = strconv.Atoi(localUser.Uid); err != nil {
			return nil, errors.Wrapf(err, "failed to convert uid: %s", localUser.Uid)
		}

		if user.GID, err = strconv.Atoi(localUser.Gid); err != nil {
			return nil, errors.Wrapf(err, "failed to convert gid: %s", localUser.Gid)
		}

		users[username] = user
	}

	return users, nil
}

func (r *UserRepo) New(remoteUser *github.User) (*User, error) {
	login := remoteUser.Username
	login = strings.TrimSpace(login)
	login = strings.ToLower(login)

	// create the user
	cmd := exec.Command(
		"adduser",
		"--firstuid", strconv.Itoa(MinUID),
		"--lastuid", strconv.Itoa(MaxUID),
		"--shell", "/bin/bash",
		"--disabled-password",
		"--gecos", "Github",
		login,
	)
	if err := runCmd(cmd); err != nil {
		return nil, err
	}

	// add the user to the sudo group
	cmd = exec.Command(
		"usermod",
		"--groups", "sudo",
		"--append",
		login,
	)
	if err := runCmd(cmd); err != nil {
		return nil, err
	}

	localUser, err := user.Lookup(login)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find created user %s", login)
	}

	result := new(User)
	if result.UID, err = strconv.Atoi(localUser.Uid); err != nil {
		return nil, errors.Wrapf(err, "failed to parse uid %s", localUser.Uid)
	}
	if result.GID, err = strconv.Atoi(localUser.Gid); err != nil {
		return nil, errors.Wrapf(err, "failed to parse gid %s", localUser.Gid)
	}

	return result, nil
}

func runCmd(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (r *UserRepo) EnableUser(username string) error {
	cmd := exec.Command("usermod", "--unlock", username)
	return runCmd(cmd)
}

func (r *UserRepo) DisableUser(username string) error {
	cmd := exec.Command("usermod", "--lock", username)
	return runCmd(cmd)
}
