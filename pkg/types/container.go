package types

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
)

// containerID generates a UUID like string that can uniquely identify a
// running container.
func containerID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

type Container struct {
	Name        string          `json:"name"`
	ID          string          `json:"id"`
	Image       string          `json:"image"`
	RuntimeDir  string          `json:"runtime_dir"`
	Dir         string          `json:"dir"`
	Overlay     Overlay         `json:"overlay"`
	CPUShares   int             `json:"cpu_shares"`
	Memory      int             `json:"memory"`
	Volumes     []Volume        `json:"volumes"`
	Namespaces  map[string]bool `json:"namespaces"`
	Capabilites []string        `json:"capabilites"`
	Interactive bool            `json:"interactive"`
	User        string          `json:"user"`
	Env         []string        `json:"env"`
	Cmd         []string        `json:"cmd"`
	Cloneflags  uintptr         `json:"cloneflags"`
	Stage1      string          `json:"stage1"`
	Stage2      string          `json:"stage2"`
}

type Overlay struct {
	Type  string `json:"type"`
	Upper string `json:"upper"`
	Lower string `json:"lower"`
	Work  string `json:"work"`
	Mnt   string `json:"mnt"`
}

type Volume struct {
	Source    string `json:"source"`
	Target    string `json:"target"`
	ReadWrite bool   `json:"read_write"`
}

func NewContainer(name string) (*Container, error) {
	cid, err := containerID()
	if name == "" {
		name = cid
	}
	return &Container{
		Name:       name,
		ID:         cid,
		Namespaces: map[string]bool{},
		Volumes: []Volume{
			{
				Source: "/etc/resolv.conf",
				Target: "/etc/resolv.conf",
			},
		},
	}, err
}

func SaveContainerConfig(c *Container) error {
	b, err := json.MarshalIndent(*c, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(c.Dir, "config.json"), b, 0644)
}

func LoadContainerConfig(path string) (*Container, error) {
	c := Container{}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &c)
	return &c, err
}
