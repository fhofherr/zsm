# zsm - ZFS Snapshot Manager

`zsm` intended as a lightweight and simple ZFS snapshot manager. For the
moment it is exclusively geared towards OpenZFS on Linux[1].

## Development

### ZFS in Vagrant

Since ZFS requires a kernel module to be loaded the tests cannot be
executed in a Docker container on a host machine without ZFS support.
Therefore this repository includes a `Vagrantfile` with OpenZFS
installed.

Some providers don't support synced folders out of the box (e.b. the
libvirt-provider). In this cases it is necessary to call

    vagrant rsync-auto

after starting the box. This sync is only uni-directional from the host
to the guest machine.

## License

Copyright © 2020 Ferdinand Hofherr

Distributed under the MIT License.

[1]: https://zfsonlinux.org/
