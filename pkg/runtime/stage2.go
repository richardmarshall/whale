/*
 */
package runtime

import (
	"syscall"

	"github.com/richardmarshall/whale/pkg/types"
	log "github.com/sirupsen/logrus"
)

// Stage2 is the runtime entrypoint for the second stage of the container startup process.
func Stage2(c *types.Container) error {
	if err := setupCPUCGroup(c.ID, c.CPUShares); err != nil {
		return err
	}
	if err := setupMemoryCGroup(c.ID, c.Memory); err != nil {
		return err
	}
	if c.Namespaces["NS"] {
		if err := setupOverlay(c.Overlay); err != nil {
			log.Info("Error setting up overlay")
			return err
		}
		if err := setupFilesystem(c); err != nil {
			return err
		}
	}
	if c.Namespaces["UTS"] {
		if err := syscall.Sethostname([]byte(c.Name)); err != nil {
			return err
		}
	}
	return nil
}
