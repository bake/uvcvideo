package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

func logic() error {
	for _, mod := range []string{
		"kernel/drivers/media/common/videobuf2/videobuf2-common.ko",
		"kernel/drivers/media/common/videobuf2/videobuf2-v4l2.ko",
		"kernel/drivers/media/common/videobuf2/videobuf2-memops.ko",
		"kernel/drivers/media/common/videobuf2/videobuf2-vmalloc.ko",
		"kernel/drivers/media/common/uvc.ko",
		"kernel/drivers/media/usb/uvc/uvcvideo.ko",
	} {
		if err := loadModule(mod); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	dev := "video0"
	target, err := checkVideoInterface(dev)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Printf("Video device %s: %s\n", dev, target)
	}

	// gokrazy should not supervise this process even when manually started.
	os.Exit(125)
	return nil
}

func checkVideoInterface(device string) (string, error) {
	target, err := os.Readlink(fmt.Sprintf("/sys/class/video4linux/%s", device))
	if err != nil {
		return "", fmt.Errorf("Video interface %s not found", device)
	}
	return target, nil
}

func loadModule(mod string) error {
	fmt.Println(filepath.Join("/lib/modules", release, mod))
	f, err := os.Open(filepath.Join("/lib/modules", release, mod))
	if err != nil {
		return err
	}
	defer f.Close()
	if err := unix.FinitModule(int(f.Fd()), "", 0); err != nil {
		if err != unix.EEXIST &&
			err != unix.EBUSY &&
			err != unix.ENODEV &&
			err != unix.ENOENT {
			return fmt.Errorf("FinitModule(%v): %v", mod, err)
		}
	}
	modname := strings.TrimSuffix(filepath.Base(mod), ".ko")
	log.Printf("modprobe %v", modname)
	return nil
}

var release = func() string {
	var uts unix.Utsname
	if err := unix.Uname(&uts); err != nil {
		fmt.Fprintf(os.Stderr, "minitrd: %v\n", err)
		os.Exit(1)
	}
	return string(uts.Release[:bytes.IndexByte(uts.Release[:], 0)])
}()

func main() {
	if err := logic(); err != nil {
		log.Fatal(err)
	}
}
