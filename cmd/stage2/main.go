// stage2 prepares the
package main

import (
	"os"
	"os/exec"
	"path"

	log "github.com/sirupsen/logrus"

	"github.com/richardmarshall/whale/pkg/runtime"
	"github.com/richardmarshall/whale/pkg/types"
)

func main() {
	log.Info("stage2 starting")
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
	if err := runtime.Stage2(c); err != nil {
		log.WithError(err).Fatal("Stage2 execution failed")
	}
	log.Info("stage2 complete")
	cmd := exec.Command(c.Cmd[0], c.Cmd[1:]...)
	if c.Interactive {
		log.Info("Running in interactive mode")
		cmd.Stdin = os.Stdin
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = c.Env
	log.Infof("Executing %v", c.Cmd)
	if err := cmd.Run(); err != nil {
		log.WithError(err).Fatal("Entrypoint exited with error")
	}
}
