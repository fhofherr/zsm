# zsm - ZFS Snapshot Manager

`zsm` intended as a lightweight and simple ZFS snapshot manager. For the
moment it is exclusively geared towards OpenZFS on Linux[1].

## Planed Features

The following is a list of planed features. Completed features are
documented below. The [CHANGELOG](CHANGELOG.md) provides detail about
the version in which a certain feature was made available.

### Create snapshots

When called with `zsm create` snapshots of all datasets except for those
listed in a blacklist should be created. The snapshots start with the
same name as the dataset and are suffixed with `@<TIMESTAMP>` where
`<TIMESTAMP>` is an [RFC3339](https://tools.ietf.org/html/rfc3339)
timestamp. The `<TIMESTAMP>` is always UTC regardles of the system time.

Additionally `zsm` supports the invocation `zsm create <dataset>`. This
creates a snapshot of the the dataset `<dataset>`.

### Clean snapshots

If called with `zsm clean` `zsm` removes all but the last *h* hourly,
*d* daily, *w* weekly, and *m* monthly snapshots. *h*, *d*, *w* and *m*
are configurable.

Example: *h = 48*, *d = 7*, *w = 4*, *m = 12*

In this case `zsm` keeps the last 48 hourly snapshots, one for each of
the last 48 hours. If the total number of snapshots is less than 48,
`zsm` does not delete any snapshots. If some or all hours contain more
than one snapshot, `zsm` keeps the youngest snapshot of the hour. If
there are gaps between the hourly snapshots older snapshots are kept to
maintain a total of 48.

Additionally keeps a snapshot for each of the last 7 days. If there are
more snapshots per day, `zsm` retains the youngest snapshot of that day.
If there are gaps between the days, older snapshots are kept to maintain
a total of 7. Since `zsm` also keeps the last 48 hourly snapshots, the
two youngest daily snapshots are identical to the youngest hourly
snapshot of the respective day.

Then `zsm` keeps a snapshot for each of the last 4 weeks. Again gaps are
filled with snapshots of older weeks to maintain the total of 4. And as
with days, `zsm` prefers the youngest snapshot for each week if there
are multiple candidates. And again the youngest weekly snapshots are
identical to weekly or daily snapshots, for overlapping periods of time.

The same is repeated for the snapshots of the last 12 months.

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
