package snapshot

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBucket_Add_HonorsInterval(t *testing.T) {
	type test struct {
		name   string
		bucket *bucket
		in     []Name
		out    []Name
	}
	end := Name{
		Timestamp:  time.Now().UTC(),
		FileSystem: "zsm_test",
	}
	tests := []test{
		{
			name: "remove interspersed minutes",
			bucket: &bucket{
				Interval: Hour,
				Size:     2,
			},
			in:  FakeNames(t, end, Minute, 61),
			out: FakeNames(t, end, Hour, 2),
		},
		{
			name: "remove interspersed hours",
			bucket: &bucket{
				Interval: Day,
				Size:     2,
			},
			in:  FakeNames(t, end, Hour, 25),
			out: FakeNames(t, end, Day, 2),
		},
		{
			name: "remove interspersed days",
			bucket: &bucket{
				Interval: Week,
				Size:     2,
			},
			in:  FakeNames(t, end, Day, 8),
			out: FakeNames(t, end, Week, 2),
		},
		{
			name: "remove interspersed weeks",
			bucket: &bucket{
				Interval: Month,
				Size:     2,
			},
			in:  FakeNames(t, end, Day, 32),
			out: FakeNames(t, end, Month, 2),
		},
		{
			name: "remove interspersed months",
			bucket: &bucket{
				Interval: Year,
				Size:     2,
			},
			in:  FakeNames(t, end, Month, 13),
			out: FakeNames(t, end, Year, 2),
		},
	}

	// Add a test for each interval which checks if this interval is known and
	// treated accordingly.
	for iv := Minute; iv < nIntervals; iv++ {
		size := rand.Intn(10) + 1
		in := FakeNames(t, end, iv, size*2)
		tt := test{
			name: fmt.Sprintf("interval: %s", iv),
			bucket: &bucket{
				Interval: iv,
				Size:     size,
			},
			in:  in,
			out: in[0:size],
		}
		tests = append(tests, tt)
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			for _, sn := range tt.in {
				tt.bucket.Add(sn)
			}
			assert.Equal(t, tt.out, tt.bucket.Elements)
		})
	}
}

func TestClean(t *testing.T) {
	tests := []struct {
		name     string
		cfg      BucketConfig
		testData func() ([]Name, []Name, []Name)
	}{
		{
			name: "keep all snapshots on empty config",
			testData: func() ([]Name, []Name, []Name) {
				end := Name{"zfs_test", time.Now().UTC()}
				in := FakeNames(t, end, Minute, 5)
				keep := in
				return in, keep, nil
			},
		},
		{
			name: "select according to bucket config",
			cfg:  BucketConfig{Minute: 3, Hour: 2},
			testData: func() ([]Name, []Name, []Name) {
				var (
					keep   []Name
					reject []Name
				)
				end := Name{"zfs_test", time.Now().UTC()}

				in := FakeNames(t, end, Minute, 5)
				keep = append(keep, in[2:5]...)
				reject = append(reject, in[0:2]...)

				oneHourAgo := Name{end.FileSystem, end.Timestamp.Add(-time.Hour)}
				in = append(in, oneHourAgo)
				keep = append(keep, oneHourAgo)

				return in, keep, reject
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			in, keep, reject := tt.testData()
			actualKeep, actualReject := clean(tt.cfg, in)
			assert.ElementsMatch(t, keep, actualKeep, "keeps don't match actual keeps")
			assert.ElementsMatch(t, reject, actualReject, "rejects don't match actual rejects")
			assert.NotContains(t, actualKeep, actualReject)
		})
	}
}

func TestClean_DoesNotModifyPassedNames(t *testing.T) {
	names := FakeNames(t, Name{FileSystem: "zfs_test", Timestamp: time.Now().UTC()}, Hour, 10)
	rand.Shuffle(len(names), func(i, j int) {
		names[i], names[j] = names[j], names[i]
	})

	passed := make([]Name, len(names))
	copy(passed, names)

	clean(BucketConfig{}, passed)

	assert.Equal(t, names, passed)
}

func TestClean_PanicsOnDivergentFileSystems(t *testing.T) {
	names := []Name{{FileSystem: "fs_1"}, {FileSystem: "fs_2"}}
	assert.Panics(t, func() {
		clean(BucketConfig{}, names)
	})
}
