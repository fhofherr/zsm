package snapshot

import (
	"fmt"
	"sort"
	"time"
)

//go:generate stringer -type=Interval

// Interval represents the interval between two consecutive snapshots.
type Interval int

func (i Interval) exceeded(a, b time.Time) bool {
	if b.Before(a) {
		a, b = b, a
	}
	switch i {
	case Minute:
		return b.Sub(a) >= time.Minute
	case Hour:
		return b.Sub(a) >= time.Hour
	case Day:
		return b.Sub(a) >= 24*time.Hour
	case Week:
		return b.Sub(a) >= 7*24*time.Hour
	case Month:
		x := a.AddDate(0, 1, 0)
		return b.Equal(x) || b.After(x)
	case Year:
		x := a.AddDate(1, 0, 0)
		return b.Equal(x) || b.After(x)
	default:
		msg := fmt.Sprintf("programming error: unknown interval: %s", i)
		panic(msg)
	}
}

// Intervals in which snapshots can be kept.
const (
	Minute Interval = iota
	Hour
	Day
	Week
	Month
	Year
	nIntervals // this must always be last
)

// BucketConfig configures the buckets for retaining snapshots.
//
// A value >0 for each of the elements of BucketConfig signals that each
// corresponding bucket should be filled with that many snapshots, each apart
// by at least the amount signaled by the bucket name.
//
// Example:
//
// If the element Minute is set to 5 the Minute bucket is to be filled with 5
// snapshots at least a minute apart.
type BucketConfig [nIntervals]int

func (b BucketConfig) createBuckets() []*bucket {
	var buckets []*bucket

	for i := Minute; i < nIntervals; i++ {
		if b[i] == 0 {
			continue
		}
		buckets = append(buckets, &bucket{
			Interval: i,
			Size:     b[i],
		})
	}
	return buckets
}

type bucket struct {
	Interval Interval
	Size     int
	Elements []Name
}

// Add tries to add the snapshot with Name sn to the bucket, if it fits.
// A snapshot is considered to fit if the bucket is either empty, or
// sn.Timestamp is more than b.Interval away from the last entry in the bucket
// and the bucket is not full yet.
//
// Add will never remove entries from the bucket once they are added. In order
// to distribute multiple snapshots across multiple buckets and be sure, that
// each bucket is filled with the correct values, the snapshots **must** be
// sorted according to their timestamps first.
func (b *bucket) Add(sn Name) bool {
	if len(b.Elements) == b.Size {
		return false
	}
	if b.Elements == nil {
		b.Elements = make([]Name, 0, b.Size)
	}
	prev := len(b.Elements) - 1
	if prev > -1 && !b.Interval.exceeded(b.Elements[prev].Timestamp, sn.Timestamp) {
		return false
	}
	b.Elements = append(b.Elements, sn)
	return true
}

// clean determines which snapshots should be removed from names according to
// cfg. It returns a slice for kept and a slice for removed snapshot names.
//
// All snapshot names must belong to the same file system. If this is not the
// case clean panics.
func clean(cfg BucketConfig, names []Name) ([]Name, []Name) {
	var (
		keep   []Name
		reject []Name
	)

	if len(names) == 0 {
		return keep, reject
	}

	// Create a defensive copy of names, as we sort it in-place
	names = append(names[:0:0], names...)
	// Sort our defensive copy of names to begin with the most recent snapshot
	// name and end with the oldest.
	sort.Slice(names, func(i, j int) bool {
		return names[i].Timestamp.After(names[j].Timestamp)
	})

	// Remember the first file system. We check for all other snapshot names
	// if they belong to the same file system. If not we panic, as this is a
	// programming error. The caller should have taken care to separate
	// snapshots according to their file systems.
	fs := names[0].FileSystem

	buckets := cfg.createBuckets()
	for _, name := range names {
		if name.FileSystem != fs {
			msg := fmt.Sprintf("programming error: name has file system %s; expected %s", name.FileSystem, fs)
			panic(msg)
		}

		// In case we don't have any buckets we don't want to add name to
		// rejects.
		ok := len(buckets) == 0 || false
		for _, b := range buckets {
			if b.Add(name) {
				ok = true
			}
		}
		if ok {
			keep = append(keep, name)
		} else {
			reject = append(reject, name)
		}
	}
	return keep, reject
}
