// Copyright (C) 2019 The ec2-mount-ephemeral Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Command ec2-mount-ephemeral prepares an instance-storage device for immediate use.
//
// It currently supports only EC2 types with one NVMe instance-storage device.
//
// Usage:
//
//   ec2-mount-ephemeral --help
//
// Display the mount plan (dry run):
//
//   ec2-mount-ephemeral --mount-path /instance
//
// Same as above (live run):
//
//   ec2-mount-ephemeral --mount-path /instance --force
//
// Same as above but customize the filesystem:
//
//   ec2-mount-ephemeral --fs-type <type> --mount-path /instance --force
//
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tp_ec2 "github.com/codeactual/ec2-mount-ephemeral/internal/third_party/github.com/aws/ec2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/codeactual/ec2-mount-ephemeral/internal/cage/cli/handler"
	handler_cobra "github.com/codeactual/ec2-mount-ephemeral/internal/cage/cli/handler/cobra"
	log_zap "github.com/codeactual/ec2-mount-ephemeral/internal/cage/cli/handler/mixin/log/zap"
	cage_exec "github.com/codeactual/ec2-mount-ephemeral/internal/cage/os/exec"
	cage_reflect "github.com/codeactual/ec2-mount-ephemeral/internal/cage/reflect"
)

const (
	defaultMountOpt = "defaults"
	defaultFsType   = "ext4"
	defaultTimeout  = 60
)

func main() {
	err := handler_cobra.NewHandler(&Handler{
		Session: &handler.DefaultSession{},
	}).Execute()
	if err != nil {
		panic(errors.WithStack(err))
	}
}

// Handler defines the sub-command flags and logic.
type Handler struct {
	handler.Session

	MountPath string `yaml:"Mount the ephemeral disk at this path"`
	Force     bool   `yaml:"Disable the default dry-run mode"`
	FsType    string `yaml:"Filesystem type"`
	MountOpt  string `yaml:"'mount' option list"`
	Timeout   uint   `usage:"Number of seconds to wait for all devices to be mounted before cancellation"`

	Log *log_zap.Mixin
}

// Init defines the command, its environment variable prefix, etc.
//
// It implements cli/handler/cobra.Handler.
func (h *Handler) Init() handler_cobra.Init {
	h.Log = &log_zap.Mixin{}

	return handler_cobra.Init{
		Cmd: &cobra.Command{
			Use:   "ec2-mount-ephemeral",
			Short: "Mount a single expected ephemeral disk",
		},
		EnvPrefix: "EC2_MOUNT_EPHEMERAL",
		Mixins: []handler.Mixin{
			h.Log,
		},
	}
}

// BindFlags binds the flags to Handler fields.
//
// It implements cli/handler/cobra.Handler.
func (h *Handler) BindFlags(cmd *cobra.Command) []string {
	cmd.Flags().BoolVarP(&h.Force, "force", "", false, cage_reflect.GetFieldTag(*h, "Force", "usage"))
	cmd.Flags().StringVarP(&h.FsType, "fs-type", "", defaultFsType, cage_reflect.GetFieldTag(*h, "FsType", "usage"))
	cmd.Flags().StringVarP(&h.MountOpt, "mount-opt", "", defaultMountOpt, cage_reflect.GetFieldTag(*h, "MountOpt", "usage"))
	cmd.Flags().StringVarP(&h.MountPath, "mount-path", "", "", cage_reflect.GetFieldTag(*h, "MountPath", "usage"))
	cmd.Flags().UintVarP(&h.Timeout, "timeout", "", defaultTimeout, cage_reflect.GetFieldTag(*h, "Timeout", "usage"))
	return []string{"mount-path"}
}

// Run performs the sub-command logic.
//
// It implements cli/handler/cobra.Handler.
func (h *Handler) Run(ctx context.Context, input handler.Input) {
	dryrun := !h.Force

	paths, err := tp_ec2.FindEphemeralDevices()
	h.Log.ExitOnErr(1, err)

	devicesLen := len(paths)
	switch devicesLen {
	case 0:
		panic(errors.New("No ephemeral devices found."))
	case 1:
		fmt.Printf("found 1 ephemeral device at: [%s] mount [%s]\n", paths[0], h.MountPath)
	default:
		panic(errors.Errorf("Found [%d] ephemeral devices, expected only 1.", devicesLen))
	}

	var mkfsArgs [][]string
	var fsckArgs [][]string
	var mountArgs [][]string

	ctx, cancel := context.WithTimeout(ctx, time.Duration(h.Timeout)*time.Second)
	defer cancel()

	for _, devicePath := range paths {
		mkfsArgs = append(mkfsArgs, []string{"mkfs", "-V", "-t", h.FsType, devicePath})

		// -M: error if already mounted; -y: attempt to repair issues; -V: verbose
		fsckArgs = append(fsckArgs, []string{"fsck", "-M", "-y", "-V", devicePath})

		mountArgs = append(mountArgs, []string{
			"mount",
			"-o", h.MountOpt,
			"-t", h.FsType,
			devicePath,
			h.MountPath,
		})
	}

	for _, args := range mkfsArgs {
		if dryrun {
			fmt.Println(strings.Join(args, " "))
			continue
		}
		mustExecCmd(ctx, args)
	}

	for _, args := range fsckArgs {
		if dryrun {
			fmt.Println(strings.Join(args, " "))
			continue
		}
		mustExecCmd(ctx, args)
	}

	for _, args := range mountArgs {
		if dryrun {
			fmt.Println(strings.Join(args, " "))
			continue
		}
		mustExecCmd(ctx, args)
	}

	if dryrun {
		fmt.Println("Dry run complete. Run with --force to execute the commands.")
	}
}

func mustExecCmd(ctx context.Context, args []string) {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	_, err := cage_exec.CommonExecutor{}.Standard(ctx, os.Stdout, os.Stderr, nil, cmd)
	if err != nil {
		panic(errors.WithStack(err))
	}
}

var _ handler_cobra.Handler = (*Handler)(nil)
