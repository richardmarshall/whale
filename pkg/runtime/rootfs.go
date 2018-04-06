package runtime

import (
	"os"
	"path"
	"path/filepath"
	"syscall"

	"github.com/richardmarshall/whale/pkg/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

// setupFilesystem prepares the filesystem for the container. The desired
// filesystem image is specified with the rootfs argument that is used to
// create COW overlay specific to the current container. With the overlay setup
// the new filesystem is populated with additional compenents such as the
// psudo-filesystems proc and sys as well as preparing the /dev directory.
func setupFilesystem(cntr *types.Container) error {
	log.Infof("Executing in rootfs: %s", cntr.Overlay.Mnt)
	log.Print("Remount root as private")
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return err
	}
	log.Printf("Setup device nodes")
	if err := setupDevices(cntr.Overlay.Mnt); err != nil {
		return err
	}
	log.Print("Setup dev symlinks")
	if err := setupSymlinks(cntr.Overlay.Mnt); err != nil {
		return err
	}
	log.Print("Mount volumes")
	if err := mountVolumes(cntr.Overlay.Mnt, cntr.Volumes); err != nil {
		return err
	}
	if err := pivotRoot(cntr.Overlay.Mnt); err != nil {
		return err
	}
	log.Print("Mount proc")
	if err := syscall.Mount("proc", "proc", "proc", 0, ""); err != nil {
		return err
	}
	log.Print("Mount sys")
	if err := syscall.Mount("sysfs", "sys", "sysfs", 0, ""); err != nil {
		return err
	}
	return nil
}

func mountVolumes(rootfs string, vs []types.Volume) error {
	for _, v := range vs {
		var flags uintptr = syscall.MS_BIND
		if !v.ReadWrite {
			flags |= syscall.MS_RDONLY
		}
		s, err := os.Stat(v.Source)
		if err != nil {
			return err
		}
		target := path.Join(rootfs, v.Target)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		if !s.IsDir() {
			f, err := os.OpenFile(target, os.O_RDONLY|os.O_CREATE, 0644)
			if err != nil {
				return err
			}
			f.Close()
		}
		log.Printf("Mounting %s at %s", v.Source, target)
		if err := syscall.Mount(v.Source, target, "bind", flags, ""); err != nil {
			return err
		}
	}
	return nil

}

// setupDevices creates a tempfs filesystem for /dev and creates the base set
// of device nodes that are commonly needed by applications.
func setupDevices(rootfs string) error {
	devs := []struct {
		name string
		mode os.FileMode
		dev  uint64
	}{
		{name: "null", mode: os.FileMode(0666) | os.ModeCharDevice, dev: unix.Mkdev(1, 3)},
		{name: "zero", mode: os.FileMode(0666) | os.ModeCharDevice, dev: unix.Mkdev(1, 5)},
		{name: "random", mode: os.FileMode(0666) | os.ModeCharDevice, dev: unix.Mkdev(1, 8)},
		{name: "urandom", mode: os.FileMode(0666) | os.ModeCharDevice, dev: unix.Mkdev(1, 9)},
		{name: "console", mode: os.FileMode(0666) | os.ModeCharDevice, dev: unix.Mkdev(136, 1)},
		{name: "tty", mode: os.FileMode(0666) | os.ModeCharDevice, dev: unix.Mkdev(5, 0)},
		{name: "full", mode: os.FileMode(0666) | os.ModeCharDevice, dev: unix.Mkdev(1, 7)},
	}
	log.Printf("Setup /dev tmpfs at %s", rootfs)
	if err := syscall.Mount("tmpfs", rootfs+"/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755"); err != nil {
		return err
	}
	log.Print("Creating device nodes")
	for _, dev := range devs {
		if err := syscall.Mknod(rootfs+"/dev/"+dev.name, 0666|syscall.S_IFCHR, int(dev.dev)); err != nil {
			log.WithError(err).Errorf("Unable to setup device: /dev/%s", dev.name)
			return err
		}
		log.Print("Created /dev/" + dev.name)
	}
	return nil
}

// setupSymlinks creates several symlinks that enable linking of the creating
// processes I/O file discriptors into the container.
func setupSymlinks(rootfs string) error {
	links := [][2]string{
		{"/proc/self/fd", "/dev/fd"},
		{"/proc/self/fd/0", "/dev/stdin"},
		{"/proc/self/fd/1", "/dev/stdout"},
		{"/proc/self/fd/2", "/dev/stderr"},
	}
	for _, link := range links {
		var (
			src = link[0]
			dst = rootfs + "/" + link[1]
		)
		if err := os.Symlink(src, dst); err != nil && !os.IsExist(err) {
			return err
		}
	}
	return nil
}

// pivotRoot updates the containers view of the top of the filesystem tree to
// be at the top of the overlay mount setup for the container image in use.
//
// References:
// http://man7.org/linux/man-pages/man2/pivot_root.2.html
func pivotRoot(newroot string) error {
	log.Print("Create oldroot target")
	oldroot := path.Join(newroot, "oldroot")
	if err := os.Mkdir(oldroot, 0755); err != nil {
		return err
	}
	log.Printf("Pivot root: %s", newroot)
	if err := syscall.PivotRoot(newroot, oldroot); err != nil {
		return err
	}
	log.Print("Chdir to root")
	if err := os.Chdir("/"); err != nil {
		return err
	}
	log.Print("Unmount old root")
	if err := syscall.Unmount("/oldroot", syscall.MNT_DETACH); err != nil {
		return err
	}
	log.Print("Remove oldroot dir")
	if err := os.Remove("/oldroot"); err != nil {
		return err
	}
	return nil
}
