package runtime

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/richardmarshall/whale/pkg/types"
)

// CreateContainerDir uses the provided container ID to create the runtime
// state directory for the container.
func CreateContainerDir(base, cid string) (string, error) {
	dir := path.Join(base, "containers", cid)
	if err := os.Mkdir(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func prepareOverlay(image, cdir string) (types.Overlay, error) {
	o := types.Overlay{
		Lower: image,
		Upper: path.Join(cdir, "rw"),
		Work:  path.Join(cdir, "work"),
		Mnt:   path.Join(cdir, "rootfs"),
	}
	fs, err := ioutil.ReadFile("/proc/filesystems")
	if err != nil {
		return o, err
	}
	if strings.Contains(string(fs), "overlay") {
		log.Info("Kernel supports overlayfs")
		o.Type = "overlay"
	} else if strings.Contains(string(fs), "aufs") {
		log.Info("Kernel supports AUFS")
		o.Type = "aufs"
	} else {
		return o, fmt.Errorf("no supported overlay filesystem available")
	}
	for _, dir := range []string{"rw", "work", "rootfs"} {
		d := path.Join(cdir, dir)
		if err := os.MkdirAll(d, 0755); err != nil {
			log.WithError(err).Errorf("Unable to create dir %s", d)
			return o, err
		}
		log.Infof("Created dir: %s", d)
	}
	return o, nil
}

// setupOverlay prepares a COW filesystem for the container with the lower
// layer being the requested distribution root filesystem. The overlayfs
// filesystem is used with fallback to AUFS if overlayfs is not available
// in the kernel.
//
// References:
// https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/plain/Documentation/filesystems/overlayfs.txt
// http://aufs.sourceforge.net/
func setupOverlay(o types.Overlay) error {
	log.Info("Mounting overlay")
	var err error
	switch o.Type {
	case "overlay":
		err = mountOverlayFS(o)
	case "aufs":
		err = mountAUFS(o)
	default:
		return fmt.Errorf("unsupported overlay type: %s", o.Type)
	}
	return err
}

func mountOverlayFS(o types.Overlay) error {
	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", o.Lower, o.Upper, o.Work)
	log.Infof("Mounting %s with opts=%s", o.Mnt, opts)
	return syscall.Mount("overlay", o.Mnt, "overlay", syscall.MS_NODEV, opts)
}

func mountAUFS(o types.Overlay) error {
	opts := fmt.Sprintf("br=%s=rw:%s=ro", o.Upper, o.Lower)
	log.Infof("Mounting %s with opts=%s", o.Mnt, opts)
	return syscall.Mount("", o.Mnt, "aufs", syscall.MS_NODEV, opts)
}

func unmountOverlay(o types.Overlay) error {
	return nil
}
