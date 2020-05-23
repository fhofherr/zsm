# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep
a Changelog](https://keepachangelog.com/en/1.0.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.1.0-alpha.1]

### Added

* `zsm create` command which allows to create snapshots of a single
  specified or all availalbe file systems.
* `zsm clean` command which allows to remove obsolete snapshots.
* `zsm list` command which allows to list all snapshots managed by zsm.
* `zsm receive` command which allows to read snapshot data from `stdin`
  and store it in a `target_fs`. Care must be taken that `target_fs` is
  excluded when calling `zsm create`.
* `zsm version` command which prints the current zsm version.

[Unreleased]: https://github.com/fhofherr/zsm/compare/v0.1.0-alpha.1...HEAD
[v0.1.0-alpha.1]: https://github.com/fhofherr/zsm/compare/v0.0.0...v0.1.0-alpha.1
