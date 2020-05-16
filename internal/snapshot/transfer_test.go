package snapshot_test

import (
	"errors"
	"testing"
	"time"

	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransfer(t *testing.T) {
	type testCase struct {
		name        string
		local       []snapshot.Name
		remote      []snapshot.Name
		mock        func(t *testing.T, tt *testCase)
		expectedErr error

		// Set during test execution
		targetFS string
		dst      *snapshot.MockManager
		src      *snapshot.MockManager
	}
	now := time.Now().UTC()
	tests := []testCase{
		{
			name: "list snapshots on src fails",
			mock: func(_ *testing.T, tt *testCase) {
				err := errors.New("list failed")
				tt.src.On("ListSnapshots").Return(tt.local, err)
			},
			expectedErr: errors.New("transfer: list src snapshots: list failed"),
		},
		{
			name: "list snapshots on dst fails",
			mock: func(_ *testing.T, tt *testCase) {
				tt.src.On("ListSnapshots").Return(tt.local, nil)

				err := errors.New("list failed")
				tt.dst.On("ListSnapshots").Return(tt.remote, err)
			},
			expectedErr: errors.New("transfer: list dst snapshots: list failed"),
		},
		{
			name: "no snapshots at all",
			mock: func(_ *testing.T, tt *testCase) {
				tt.src.On("ListSnapshots").Return(tt.local, nil)
				tt.dst.On("ListSnapshots").Return(tt.remote, nil)
			},
		},
		{
			name:   "dst more snapshots than src",
			local:  snapshot.FakeNames(t, snapshot.Name{FileSystem: "zsm_test", Timestamp: now}, snapshot.Hour, 1),
			remote: snapshot.FakeNames(t, snapshot.Name{FileSystem: "zsm_test", Timestamp: now}, snapshot.Hour, 2),
			mock: func(_ *testing.T, tt *testCase) {
				tt.src.On("ListSnapshots").Return(tt.local, nil)
				tt.dst.On("ListSnapshots").Return(tt.remote, nil)
			},
			expectedErr: errors.New("transfer: dst has more snapshots: 2 > 1"),
		},
		{
			name: "no snapshots on dst",
			local: snapshot.FakeNames(
				t, snapshot.Name{FileSystem: "zsm_test", Timestamp: now}, snapshot.Hour, 5,
			),
			mock: func(t *testing.T, tt *testCase) {
				// Shuffle the local snapshot to ensure we are not relying on
				// any specific order of the snapshots in the actual test.
				localShuffled := snapshot.ShuffleNamesC(tt.local)
				tt.src.On("ListSnapshots").Return(localShuffled, nil)
				tt.src.On("SendSnapshot", tt.local[4], mock.AnythingOfType("*io.PipeWriter")).Return(nil)

				tt.dst.On("ListSnapshots").Return(tt.remote, nil)
				tt.dst.On("ReceiveSnapshot", tt.targetFS, tt.local[4], mock.AnythingOfType("*io.PipeReader")).
					Return(nil)
			},
		},
		{
			name:  "initial snapshot transfer send snapshot fails",
			local: []snapshot.Name{{FileSystem: "zsm_test", Timestamp: now}},
			mock: func(t *testing.T, tt *testCase) {
				err := errors.New("send failed")
				tt.src.On("ListSnapshots").Return(tt.local, nil)
				tt.src.On("SendSnapshot", tt.local[0], mock.AnythingOfType("*io.PipeWriter")).Return(err)

				tt.dst.On("ListSnapshots").Return(tt.remote, nil)
				tt.dst.On("ReceiveSnapshot", tt.targetFS, tt.local[0], mock.AnythingOfType("*io.PipeReader")).
					Return(nil)
			},
			expectedErr: errors.New("transfer: send failed"),
		},
		{
			name:  "initial snapshot transfer receive snapshot fails",
			local: []snapshot.Name{{FileSystem: "zsm_test", Timestamp: now}},
			mock: func(t *testing.T, tt *testCase) {
				tt.src.On("ListSnapshots").Return(tt.local, nil)
				tt.src.On("SendSnapshot", tt.local[0], mock.AnythingOfType("*io.PipeWriter")).Return(nil)

				err := errors.New("receive failed")
				tt.dst.On("ListSnapshots").Return(tt.remote, nil)
				tt.dst.On("ReceiveSnapshot", tt.targetFS, tt.local[0], mock.AnythingOfType("*io.PipeReader")).
					Return(err)
			},
			expectedErr: errors.New("transfer: receive failed"),
		},
		{
			name: "src has snapshots for multiple file systems",
			local: []snapshot.Name{
				{FileSystem: "zsm_test", Timestamp: now},
				{FileSystem: "zsm_test/fs1", Timestamp: now},
			},
			mock: func(t *testing.T, tt *testCase) {
				localShuffled := snapshot.ShuffleNamesC(tt.local)

				tt.src.On("ListSnapshots").Return(localShuffled, nil)
				tt.src.On("SendSnapshot", tt.local[0], mock.AnythingOfType("*io.PipeWriter")).Return(nil)
				tt.src.On("SendSnapshot", tt.local[1], mock.AnythingOfType("*io.PipeWriter")).Return(nil)

				tt.dst.On("ListSnapshots").Return(tt.remote, nil)
				tt.dst.On("ReceiveSnapshot", tt.targetFS, tt.local[0], mock.AnythingOfType("*io.PipeReader")).
					Return(nil)
				tt.dst.On("ReceiveSnapshot", tt.targetFS, tt.local[1], mock.AnythingOfType("*io.PipeReader")).
					Return(nil)
			},
		},
		{
			name:   "dst has all snapshots of src",
			local:  snapshot.FakeNames(t, snapshot.Name{FileSystem: "zsm_test", Timestamp: now}, snapshot.Day, 5),
			remote: snapshot.FakeNames(t, snapshot.Name{FileSystem: "zsm_test", Timestamp: now}, snapshot.Day, 5),
			mock: func(t *testing.T, tt *testCase) {
				localShuffled := snapshot.ShuffleNamesC(tt.local)
				tt.src.On("ListSnapshots").Return(localShuffled, nil)

				remoteShuffled := snapshot.ShuffleNamesC(tt.remote)
				tt.dst.On("ListSnapshots").Return(remoteShuffled, nil)
			},
		},
		{
			name: "src has more snapshots than dst",
			local: snapshot.FakeNames(
				t, snapshot.Name{FileSystem: "zsm_test", Timestamp: now}, snapshot.Hour, 10,
			),
			remote: snapshot.FakeNames(
				t, snapshot.Name{FileSystem: "zsm_test", Timestamp: now.Add(-5 * time.Hour)},
				snapshot.Hour, 5,
			),
			mock: func(t *testing.T, tt *testCase) {
				localShuffled := snapshot.ShuffleNamesC(tt.local)
				tt.src.On("ListSnapshots").Return(localShuffled, nil)
				tt.src.On("SendSnapshot",
					tt.local[9], mock.AnythingOfType("*io.PipeWriter"), mock.AnythingOfType("snapshot.SendOption"),
				).Return(nil)
				tt.src.ExpectSendOptions(snapshot.Reference(tt.local[5]))

				remoteShuffled := snapshot.ShuffleNamesC(tt.remote)
				tt.dst.On("ListSnapshots").Return(remoteShuffled, nil)
				tt.dst.On("ReceiveSnapshot", tt.targetFS, tt.local[9], mock.AnythingOfType("*io.PipeReader")).
					Return(nil)
			},
		},
		{
			name:  "incremental transfer send fails",
			local: snapshot.FakeNames(t, snapshot.Name{FileSystem: "zsm_test", Timestamp: now}, snapshot.Hour, 3),
			remote: snapshot.FakeNames(
				t, snapshot.Name{FileSystem: "zsm_test", Timestamp: now.Add(-3 * time.Hour)}, snapshot.Hour, 1,
			),
			mock: func(t *testing.T, tt *testCase) {
				err := errors.New("send failed")
				tt.src.On("ListSnapshots").Return(tt.local, nil)
				tt.src.On("SendSnapshot",
					tt.local[2], mock.AnythingOfType("*io.PipeWriter"), mock.AnythingOfType("snapshot.SendOption"),
				).Return(err)
				tt.src.ExpectSendOptions(snapshot.Reference(tt.local[1]))

				tt.dst.On("ListSnapshots").Return(tt.remote, nil)
				tt.dst.On("ReceiveSnapshot", tt.targetFS, tt.local[2], mock.AnythingOfType("*io.PipeReader")).
					Return(nil)
			},
			expectedErr: errors.New("transfer: send failed"),
		},
		{
			name:  "incremental transfer receive fails",
			local: snapshot.FakeNames(t, snapshot.Name{FileSystem: "zsm_test", Timestamp: now}, snapshot.Hour, 3),
			remote: snapshot.FakeNames(
				t, snapshot.Name{FileSystem: "zsm_test", Timestamp: now.Add(-3 * time.Hour)}, snapshot.Hour, 1,
			),
			mock: func(t *testing.T, tt *testCase) {
				tt.src.On("ListSnapshots").Return(tt.local, nil)
				tt.src.On("SendSnapshot",
					tt.local[2], mock.AnythingOfType("*io.PipeWriter"), mock.AnythingOfType("snapshot.SendOption"),
				).Return(nil)
				tt.src.ExpectSendOptions(snapshot.Reference(tt.local[1]))

				err := errors.New("receive failed")
				tt.dst.On("ListSnapshots").Return(tt.remote, nil)
				tt.dst.On("ReceiveSnapshot", tt.targetFS, tt.local[2], mock.AnythingOfType("*io.PipeReader")).
					Return(err)
			},
			expectedErr: errors.New("transfer: receive failed"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.dst = &snapshot.MockManager{}
			tt.dst.Test(t)
			tt.src = &snapshot.MockManager{}
			tt.src.Test(t)
			tt.targetFS = "target_fs"

			tt.mock(t, &tt)

			err := snapshot.Transfer(tt.targetFS, tt.dst, tt.src)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			tt.dst.AssertExpectations(t)
			tt.src.AssertExpectations(t)
			tt.src.AssertSendOptions(t)
		})
	}
}
