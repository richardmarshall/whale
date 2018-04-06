package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/richardmarshall/whale/pkg/runtime"
	"github.com/richardmarshall/whale/pkg/types"
)

var rootCmd = &cobra.Command{
	Use:   "whale",
	Short: "Whale is a simple container runtime",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run starts a container",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			args = []string{"/bin/sh"}
		}
		rundir, _ := cmd.Flags().GetString("rundir")
		name, _ := cmd.Flags().GetString("name")
		image, _ := cmd.Flags().GetString("image")
		user, _ := cmd.Flags().GetString("user")
		interactive, _ := cmd.Flags().GetBool("interactive")
		env, _ := cmd.Flags().GetStringSlice("env")
		volumes, _ := cmd.Flags().GetStringSlice("volume")
		// namespace flags
		pid, _ := cmd.Flags().GetString("pid")
		uts, _ := cmd.Flags().GetString("uts")
		net, _ := cmd.Flags().GetString("net")
		ipc, _ := cmd.Flags().GetString("ipc")
		mount, _ := cmd.Flags().GetString("mount")
		stage1, _ := cmd.Flags().GetString("stage1")
		stage2, _ := cmd.Flags().GetString("stage2")

		c, err := types.NewContainer(name)
		if err != nil {
			return err
		}
		c.Interactive = interactive
		c.User = user
		c.Env = env
		c.Image = image
		c.Cmd = args
		c.Stage1 = stage1
		c.Stage2 = stage2
		c.RuntimeDir = rundir
		if mount != "host" {
			c.Namespaces["NS"] = true
		}
		if pid != "host" {
			c.Namespaces["PID"] = true
		}
		if uts != "host" {
			c.Namespaces["UTS"] = true
		}
		if net != "host" {
			c.Namespaces["NET"] = true
		}
		if ipc != "host" {
			c.Namespaces["IPC"] = true
		}
		vols := c.Volumes
		for _, v := range volumes {
			readwrite := true
			chunks := strings.Split(v, ":")
			if len(chunks) != 2 && len(chunks) != 3 {
				return fmt.Errorf("invalid volume: %s", v)
			}
			if len(chunks) == 3 && chunks[2] == "ro" {
				readwrite = false
			}
			vols = append(vols, types.Volume{
				Source:    chunks[0],
				Target:    chunks[1],
				ReadWrite: readwrite,
			})
		}
		c.Volumes = vols
		d, err := runtime.CreateContainerDir(rundir, c.ID)
		if err != nil {
			return err
		}
		c.Dir = d
		if err := types.SaveContainerConfig(c); err != nil {
			return err
		}
		stg1 := exec.Command(c.Stage1, c.ID)
		envv := os.Environ()
		envv = append(envv, "WHALE_RUNTIME_DIR="+rundir)
		stg1.Env = envv
		stg1.Stdin = os.Stdin
		stg1.Stdout = os.Stdout
		stg1.Stderr = os.Stderr
		return stg1.Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().String("name", "", "container name")
	runCmd.Flags().String("image", "debian", "Distribution rootfs to use")
	runCmd.Flags().StringP("net", "n", "private", "Run with private or host networking")
	runCmd.Flags().StringP("uts", "s", "private", "Run with private or host UTS")
	runCmd.Flags().StringP("pid", "p", "private", "Run with private of host PID")
	runCmd.Flags().String("mount", "private", "Run with private or host mount")
	runCmd.Flags().StringP("user", "u", "root:root", "Run as user:group")
	runCmd.Flags().StringSliceP("env", "e", nil, "Environment variables")
	runCmd.Flags().StringSliceP("volume", "v", nil, "Volumes to mount in the container")
	runCmd.Flags().BoolP("interactive", "i", true, "Attach stdin to container")
	runCmd.Flags().IntP("memory", "m", 0, "Memory limit in bytes.")
	runCmd.Flags().IntP("cpu", "c", 0, "CPU shares.")
	runCmd.Flags().String("rundir", "/var/run/whale", "runtime directory")
	runCmd.Flags().String("stage1", "./bin/stage1", "")
	runCmd.Flags().String("stage2", "./bin/stage2", "")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
