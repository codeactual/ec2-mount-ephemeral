package ec2

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	cage_file "github.com/codeactual/ec2-mount-ephemeral/internal/cage/os/file"
)

// FindEphemeralDevices returns the absolute paths to all ephemeral devices.
//
// Based on this EBS-focused approach:
//   https://github.com/kubernetes/kubernetes/blob/f4472b1a92877ed4b1576e7e44496b0de7a8efe2/pkg/volume/aws_ebs/aws_util.go#L227
//   Apache 2.0: https://github.com/kubernetes/kubernetes/blob/f4472b1a92877ed4b1576e7e44496b0de7a8efe2/LICENSE
//
// Changes:
//   - Use to cage_file.Readdirnames to find multiple devices instead of only one.
func FindEphemeralDevices() (paths []string, err error) {
	const (
		devIndexPath  = "/dev/disk/by-id"
		symlinkPrefix = "nvme-Amazon_EC2_NVMe_Instance_Storage"
	)

	names, err := cage_file.Readdirnames(devIndexPath, -1)
	if err != nil {
		return []string{}, errors.Wrapf(err, "failed to collect contents of [%s]", devIndexPath)
	}

	for _, name := range names {
		if !strings.HasPrefix(name, symlinkPrefix) {
			continue
		}

		p := filepath.Join(devIndexPath, name)

		stat, err := os.Lstat(p)
		if err != nil {
			if os.IsNotExist(err) {
				return []string{}, nil
			}
			return []string{}, fmt.Errorf("error getting stat of %q: %v", p, err)
		}

		if stat.Mode()&os.ModeSymlink != os.ModeSymlink {
			return []string{}, nil
		}

		// Find the target, resolving to an absolute path
		// For example, /dev/disk/by-id/nvme-Amazon_Elastic_Block_Store_vol0fab1d5e3f72a5e23 -> ../../nvme2n1
		resolved, err := filepath.EvalSymlinks(p)
		if err != nil {
			return []string{}, fmt.Errorf("error reading target of symlink %q: %v", p, err)
		}

		if !strings.HasPrefix(resolved, "/dev") {
			return []string{}, fmt.Errorf("resolved symlink for %q was unexpected: %q", p, resolved)
		}

		paths = append(paths, resolved)
	}

	return paths, nil
}
