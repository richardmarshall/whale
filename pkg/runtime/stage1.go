package runtime

import (
	"fmt"
	"path"
	"syscall"

	"github.com/richardmarshall/whale/pkg/types"
	log "github.com/sirupsen/logrus"
)

var nsFlags = map[string]uintptr{
	"NS":  syscall.CLONE_NEWNS,
	"PID": syscall.CLONE_NEWPID,
	"UTS": syscall.CLONE_NEWUTS,
	"NET": syscall.CLONE_NEWNET,
	"IPC": syscall.CLONE_NEWIPC,
}

func cloneFlags(ns map[string]bool) (uintptr, error) {
	var flags uintptr
	for n, _ := range ns {
		f, ok := nsFlags[n]
		if !ok {
			return flags, fmt.Errorf("invalid namespace: %s", n)
		}
		flags |= f
	}
	return flags, nil
}

func Stage1(c *types.Container) error {
	flags, err := cloneFlags(c.Namespaces)
	if err != nil {
		return err
	}
	c.Cloneflags = flags
	imageDir := path.Join(c.RuntimeDir, "rootfs", c.Image)
	overlay, err := prepareOverlay(imageDir, c.Dir)
	if err != nil {
		log.Info("Error preparing overlay")
		return err
	}
	c.Overlay = overlay
	return nil
}
