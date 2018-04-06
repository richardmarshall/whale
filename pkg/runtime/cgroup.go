package runtime

import (
	"io/ioutil"
	"os"
	"strconv"
)

// setupCPUCGroup creates a new cpu cgroup and applies any requested limits
func setupCPUCGroup(cid string, shares int) error {
	cdir := "/sys/fs/cgroup/cpu/whale/" + cid
	if err := os.MkdirAll(cdir, 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(cdir+"/tasks", []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		return err
	}
	if shares > 0 {
		if err := ioutil.WriteFile(cdir+"/cpu.shares", []byte(strconv.Itoa(shares)), 0644); err != nil {
			return err
		}
	}
	return nil
}

// setupMemoryCGroup creates a new memory cgroup and applies any requested limits
func setupMemoryCGroup(cid string, memory int) error {
	cdir := "/sys/fs/cgroup/memory/whale/" + cid
	if err := os.MkdirAll(cdir, 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(cdir+"/tasks", []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		return err
	}
	if memory > 0 {
		if err := ioutil.WriteFile(cdir+"/memory.limit_in_bytes", []byte(strconv.Itoa(memory)), 0644); err != nil {
			return err
		}
	}
	return nil
}
