package rmt

// Package rmt contains functionality for remote targets
// A remote target is a target that is connected to using SSH

type Remote struct {
}

// Put puts the contents of a Reader on a path on the remote machine
//
// Inspiration:
// https://github.com/laher/scp-go/blob/master/scp/toremote.go
// https://gist.github.com/jedy/3357393
//
// SCP notes:
// https://web.archive.org/web/20170215184048/https://blogs.oracle.com/janp/entry/how_the_scp_protocol_works
// https://en.wikipedia.org/wiki/Secure_copy#cite_note-Pechanec-2
func (r *remote) put(content io.Reader, size int64, path string, mode uint32) error {

	// consider using github.com/pkg/sftp

	session, err := r.client.NewSession()
	if err != nil {
		return errors.Wrap(err, "failed to create scp session")
	}
	defer session.Close()

	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()

		// header message has the format C<mode> <size> <filename>
		fmt.Fprintf(w, "C%04o %d %s\n", mode, size, filepath.Base(path))

		io.Copy(w, content)

		// transfer end with \x00
		fmt.Fprint(w, "\x00")
	}()

	if b, err := session.CombinedOutput(fmt.Sprintf("/usr/bin/scp -tr %s", path)); err != nil {
		return errors.Wrapf(err, "unable to copy content: %s", string(b))
	}

	return nil
}

type remote struct {
	client   *ssh.Client
	addr     string
	port     int
	usr      string
	sudopass string
}

func (r *remote) user() string {
	return r.usr
}

func (r *remote) String() string {
	return fmt.Sprintf("%s@%s:%d", r.usr, r.addr, r.port)
}

func newRemote(addr string, port int, user string, sudopass string) (*remote, error) {
	r := remote{
		addr:     addr,
		port:     port,
		usr:      user,
		sudopass: sudopass,
	}

	var err error

	a, err := getAgentAuths()
	auths := []ssh.AuthMethod{
		ssh.Password(sudopass),
	}
	if err == nil {
		auths = append(auths, a)
	}

	cc := ssh.ClientConfig{
		User:            user,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO
	}

	r.client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", addr, port), &cc)
	if err != nil {
		return &r, errors.Wrapf(err, "unable to establish connection to %s:%d", addr, port)
	}

	return &r, nil

}

// run runs cmd on remote
func (r *remote) run(cmd string, stdin string, sudo bool, user string) (Response, error) {

	session, err := r.client.NewSession()
	resp := Response{}

	if err != nil {
		return resp, errors.Wrap(err, "unable to create new session")
	}
	defer session.Close()

	o := bytes.Buffer{}
	e := bytes.Buffer{}

	session.Stdout = &o
	session.Stderr = &e

	// TODO - consider using session.Shell - http://networkbit.ch/golang-ssh-client/#multiple_commands
	if sudo {
		session.Stdin = strings.NewReader(r.sudopass + "\n" + stdin + "\n")
		if user == "" || user == "-" {
			user = "root"
		}
		sudocmd := fmt.Sprintf(`sudo -k -S -u %s bash -c "%s"`, user, cmd)
		err = session.Run(sudocmd)

	} else {
		session.Stdin = strings.NewReader(stdin + "\n")
		err = session.Run(cmd)
	}

	resp.Stdout = scrubStd(o.String())
	resp.Stderr = scrubStd(e.String())

	if err != nil {

		switch t := err.(type) {
		case *ssh.ExitError:
			resp.ExitStatus = t.Waitmsg.ExitStatus()
		case *ssh.ExitMissingError:
			resp.ExitStatus = -1
		default:
			return resp, errors.Wrap(err, "run of command failed")
		}

	} else {
		resp.ExitStatus = 0
	}

	return resp, nil
}

}


// getAgentAuths is a helper function to get SSH keys from an ssh agent
func getAgentAuths() (ssh.AuthMethod, error) {

	socket := os.Getenv("SSH_AUTH_SOCK")
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open SSH_AUTH_SOCK")
	}

	agentClient := agent.NewClient(conn)

	return ssh.PublicKeysCallback(agentClient.Signers), nil
}


// NewRemoteHost returns a Host based on address, port and user
// It will connect to the SSH agent to get any ssh keys
func NewRemoteHost(addr string, port int, user string, sudopass string) (*Host, error) {
	r, err := newRemote(addr, port, user, sudopass)

	if err != nil {
		return &Host{}, err
	}

	return &Host{r, false, newTrace()}, nil
}
