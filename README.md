# ec2-mount-ephemeral [![GoDoc](https://godoc.org/github.com/codeactual/ec2-mount-ephemeral?status.svg)](https://godoc.org/github.com/codeactual/ec2-mount-ephemeral) [![Go Report Card](https://goreportcard.com/badge/github.com/codeactual/ec2-mount-ephemeral)](https://goreportcard.com/report/github.com/codeactual/ec2-mount-ephemeral) [![Build Status](https://travis-ci.org/codeactual/ec2-mount-ephemeral.png)](https://travis-ci.org/codeactual/ec2-mount-ephemeral)

ec2-mount-ephemeral prepares an instance-storage device for immediate use.

It currently supports only EC2 types with one NVMe instance-storage device.

# Usage

> To install: `go get -v github.com/codeactual/ec2-mount-ephemeral/cmd/ec2-mount-ephemeral`

## Examples

> Usage:

```bash
ec2-mount-ephemeral --help
```

> Display the mount plan (dry run):

```bash
ec2-mount-ephemeral --mount-path /instance
```

> Same as above (live run):

```bash
ec2-mount-ephemeral --mount-path /instance --force
```

> Same as above but customize the filesystem:

```bash
ec2-mount-ephemeral --fs-type non_default_ext4 --mount-path /instance --force
```

# License

[Mozilla Public License Version 2.0](https://www.mozilla.org/en-US/MPL/2.0/) ([About](https://www.mozilla.org/en-US/MPL/), [FAQ](https://www.mozilla.org/en-US/MPL/2.0/FAQ/))

*(Exported from a private monorepo with [transplant](https://github.com/codeactual/transplant).)*
