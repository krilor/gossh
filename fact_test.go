package gossh

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/krilor/gossh/rmt"
	"github.com/krilor/gossh/testing/docker"
	"golang.org/x/crypto/ssh"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestParseINI(t *testing.T) {

	var tests = []struct {
		in     string
		expect map[string]string
	}{
		{`distrib_id=ubuntu
distrib_release=18.04
distrib_codename=bionic

name="ubuntu"
#version="18.04.4 lts (bionic beaver)"
home_url="https://www.ubuntu.com/"`,
			map[string]string{
				"distrib_id":       "ubuntu",
				"distrib_release":  "18.04",
				"distrib_codename": "bionic",
				"name":             "ubuntu",
				"home_url":         "https://www.ubuntu.com/",
			},
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s", test.in), func(t *testing.T) {

			got := parseINI(test.in)

			if !reflect.DeepEqual(got, test.expect) {
				t.Errorf("notequal: got \"%s\" - expect \"%s\"", got, test.expect)
			}

		})
	}
}

func TestMajorVersion(t *testing.T) {

	var tests = []struct {
		in     string
		expect string
	}{
		{"18.04", "18"},
		{"7.7", "7"},
		{"6", "6"},
		{"", ""},
	}
	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {

			got := majorVersion(test.in)

			if got != test.expect {
				t.Errorf("notequal: got \"%s\" - expect \"%s\"", got, test.expect)
			}

		})
	}
}

func TestGather(t *testing.T) {

	var tests = []struct {
		img    docker.Image
		expect Facts
	}{
		{img: docker.Ubuntu("bionic"), expect: Facts{kv: map[Fact]string{
			OS:        "ubuntu",
			OSFamily:  "debian",
			OSVersion: "18",
		}}},
		{img: docker.Debian("buster"), expect: Facts{kv: map[Fact]string{
			OS:        "debian",
			OSFamily:  "debian",
			OSVersion: "10",
		}}},
		{img: docker.RedHat(7), expect: Facts{kv: map[Fact]string{
			OS:        "rhel",
			OSFamily:  "rhel",
			OSVersion: "7",
		}}},
		{img: docker.Oracle(7), expect: Facts{kv: map[Fact]string{
			OS:        "ol",
			OSFamily:  "rhel",
			OSVersion: "7",
		}}},
		{img: docker.Fedora(32), expect: Facts{kv: map[Fact]string{
			OS:        "fedora",
			OSFamily:  "rhel",
			OSVersion: "32",
		}}},
		{img: docker.CentOS(7), expect: Facts{kv: map[Fact]string{
			OS:        "centos",
			OSFamily:  "rhel",
			OSVersion: "7",
		}}},
	}

	for _, test := range tests {
		t.Run(test.img.Name(), func(t *testing.T) {
			c, err := docker.New(test.img)
			if err != nil {
				log.Fatalf("could not get throwaway container: %v", err)
			}
			defer c.Kill()

			h, err := rmt.New(fmt.Sprintf("localhost:%d", c.Port()), "gossh", "gosshpwd", ssh.InsecureIgnoreHostKey(), ssh.Password("gosshpwd"))

			if err != nil {
				log.Fatalf("could not connect to throwaway container %v", err)
			}

			f := Facts{}
			err = f.Gather(h)

			if err != nil {
				t.Errorf("could not run gather: %v", err)
			}

			if !reflect.DeepEqual(f.kv, test.expect.kv) {
				t.Errorf("notequal: got \"%v\" - expect \"%v\"", f.kv, test.expect.kv)
			}

		})
	}

}
