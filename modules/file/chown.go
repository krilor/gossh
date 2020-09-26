package file

import (
	"errors"
	"fmt"
	"strings"
)

// Package sh contains (ba)sh related methods and types

// ChownCmd returns a chown/chgrp command based on the input
// If path,
func ChownCmd(path, username, groupname string) (string, error) {

	path = strings.TrimSpace(path)
	username = strings.TrimSpace(username)
	groupname = strings.TrimSpace(groupname)

	if path == "" {
		return "", errors.New("path is empty")
	}

	if username == "" && groupname == "" {
		return "", errors.New("both user and group name cannot be empty")
	}

	if username == "" {
		return fmt.Sprintf("chgrp %s %s", groupname, path), nil
	}

	if groupname == "" {
		return fmt.Sprintf("chown %s %s", username, path), nil
	}

	return fmt.Sprintf("chown %s:%s %s", username, groupname, path), nil

}

// Chown changes the ownership of the file
func (l Local) Chown(path, username, groupname string) error {

	cmd, err := sh.ChownCmd(path, username, groupname)
	if err != nil {
		return err
	}

	resp, err := l.Run(cmd, nil)

	if resp.ExitStatus != 0 {
		err = errors.New(resp.Err())
	}

	return err

}
