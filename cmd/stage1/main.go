// stage1 prepares the
package main

import (
	"os"
	"os/exec"
	"path"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/richardmarshall/whale/pkg/runtime"
	"github.com/richardmarshall/whale/pkg/types"
)

func main() {
	log.Info("stage1 starting")
	base := os.Getenv("WHALE_RUNTIME_DIR")
	if base == "" {
		log.Fatal("Unable to lookup runtime directory")
	}
	if len(os.Args) != 2 {
		log.Fatal("Incorrect number of arguments")
	}
	p := path.Join(base, "containers", os.Args[1], "config.json")
	c, err := types.LoadContainerConfig(p)
	if err != nil {
		log.WithError(err).Fatal("Unable to unmarshal container struct")
	}
	log.Infof("Loaded container: %s (%s)", c.ID, c.Name)
	if err := runtime.Stage1(c); err != nil {
		log.WithError(err).Fatal("Stage1 execution failed")
	}
	if err := types.SaveContainerConfig(c); err != nil {
		log.WithError(err).Fatal("Unable to marshal container struct")
	}
	log.Info("stage1 complete")
	stage2 := exec.Command(c.Stage2, c.ID)
	stage2.Stdin = os.Stdin
	stage2.Stdout = os.Stdout
	stage2.Stderr = os.Stderr
	stage2.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: c.Cloneflags,
	}
	if err := stage2.Run(); err != nil {
		log.WithError(err).Fatal("Stage 2 exitied with error")
	}
}
