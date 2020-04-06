package gossh

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/krilor/gossh/testing/docker"
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

func TestGather(t *testing.T) {

	var tests = []struct {
		img    docker.Image
		expect Facts
	}{
		{img: docker.NewDebianImage("ubuntu", "bionic"), expect: Facts{kv: map[Fact]string{
			OS:        "ubuntu",
			OSFamily:  "debian",
			OSVersion: "18.04", // TODO this should probably be 18, and we should skip .04
		}}},
		{img: docker.NewRHELImage("centos", "7"), expect: Facts{kv: map[Fact]string{
			OS:        "centos",
			OSFamily:  "rhel fedora", // TODO - this doesn't look nice. Would like "rhel" only
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

			h, err := NewRemoteHost("localhost", c.Port(), "gossh", "gosshpwd")

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
