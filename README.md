# zsm - ZFS Snapshot Manager

`zsm` intended as a lightweight and simple ZFS snapshot manager. For the
moment it is exclusively geared towards OpenZFS on Linux[1].

## Planed Features

The following is a list of planed features. Completed features are
documented below. The [CHANGELOG](CHANGELOG.md) provides detail about
the version in which a certain feature was made available.

### Sending snapshots

Calling `zsm send <user>@<host> <target-pool>` sends all snapshots
managed by `zsm` to `host` using ssh. The `user` used to log in to the
remote host must be able to write to `target-pool`. Additionally a `zsm`
executable of the same version must be on the users `PATH`.

`zsm send` does not perform any clean up on the remote or the local
host. The administrators are responsible to schedule regular calls to
`zsm clean` on both hosts on a regular basis. While this may lead to the
awkward situation, that already removed snapshots are re-transmitted to
a target host it keeps the implementation of `send` simple and straight
forward.

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
