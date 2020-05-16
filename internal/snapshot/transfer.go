package snapshot

import (
	"fmt"
	"io"
	"sort"
)

// Lister defines the ListSnapshots method.
type Lister interface {
	ListSnapshots() ([]Name, error)
}

// Receiver defines the ReceiveSnapshot method.
type Receiver interface {
	ReceiveSnapshot(string, Name, io.Reader) error
}

// Sender defines the SendSnapshot method.
type Sender interface {
	SendSnapshot(Name, io.Writer, ...SendOption) error
}

// ListerReceiver defines a type that can list all snapshots known to it and
// can receive additional snapshots.
type ListerReceiver interface {
	Lister
	Receiver
}

// ListerSender defines a type that can list all snapshots known to it and
// can send snapshots.
type ListerSender interface {
	Lister
	Sender
}

// Transfer transfers all snapshots not already known on dst from src to dst.
func Transfer(targetFS string, dst ListerReceiver, src ListerSender) error {
	local, err := src.ListSnapshots()
	if err != nil {
		return fmt.Errorf("transfer: list src snapshots: %w", err)
	}
	remote, err := dst.ListSnapshots()
	if err != nil {
		return fmt.Errorf("transfer: list dst snapshots: %w", err)
	}
	if len(local) == 0 && len(remote) == 0 {
		return nil
	}

	localGrouped := groupByFS(local)
	remoteGrouped := groupByFS(remote)
	for fs, localNames := range localGrouped {
		localNames := localNames
		sort.Slice(localNames, func(i, j int) bool {
			return localNames[i].Timestamp.Before(localNames[j].Timestamp)
		})
		remoteNames, ok := remoteGrouped[fs]
		if !ok {
			// The destination has no snapshots for fs. Just transfer
			// everything we have.
			if err := transfer(targetFS, dst, src, localNames[len(localNames)-1]); err != nil {
				return fmt.Errorf("transfer: %w", err)
			}
			continue
		}
		// We assume that all snapshots on destination are sent by this source.
		if len(remoteNames) == len(localNames) {
			// If remote and local have the same number of snapshots, we assume
			// that remote is up-to date. We continue with the next file system.
			continue
		}
		if len(remoteNames) > len(localNames) {
			// Abort if the remote has more than snapshots local. This
			// indicates a problem, since all snapshots on remote for the file
			// system should come from this host.
			return fmt.Errorf("transfer: dst has more snapshots: %d > %d", len(remoteNames), len(localNames))
		}
		// There are fewer snapshots on the destination than there are locally.
		// We need to determine the difference and send the missing snapshots.
		sort.Slice(remoteNames, func(i, j int) bool {
			return remoteNames[i].Timestamp.Before(remoteNames[j].Timestamp)
		})
		ref := localNames[len(remoteNames)]
		if err := transfer(targetFS, dst, src, localNames[len(localNames)-1], Reference(ref)); err != nil {
			return fmt.Errorf("transfer: %w", err)
		}
	}
	return nil
}

func groupByFS(names []Name) map[string][]Name {
	grp := make(map[string][]Name)
	for _, n := range names {
		if grp[n.FileSystem] == nil {
			grp[n.FileSystem] = make([]Name, 0, 10)
		}
		grp[n.FileSystem] = append(grp[n.FileSystem], n)
	}
	return grp
}

func transfer(targetFS string, dst Receiver, src Sender, n Name, opts ...SendOption) error {
	r, w := io.Pipe()

	sendErr := send(src, n, w, opts...)
	recvErr := receive(dst, targetFS, n, r)

	if err := <-sendErr; err != nil {
		return err
	}
	if err := <-recvErr; err != nil {
		return err
	}

	return nil
}

func send(src Sender, n Name, w io.WriteCloser, opts ...SendOption) <-chan error {
	errC := make(chan error, 1)
	go func() {
		defer close(errC)
		defer w.Close() // Close w before closing errC => signals EOF to reader

		if err := src.SendSnapshot(n, w, opts...); err != nil {
			errC <- err
		}
	}()
	return errC
}

func receive(dst Receiver, targetFS string, n Name, r io.Reader) <-chan error {
	errC := make(chan error, 1)
	go func() {
		defer close(errC)

		if err := dst.ReceiveSnapshot(targetFS, n, r); err != nil {
			errC <- err
		}
	}()
	return errC
}
