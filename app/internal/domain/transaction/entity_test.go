package transaction

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTransition_SubmittedToMinedToConfirmed(t *testing.T) {
	now := time.Now()
	tx := NewSubmitted("req-1", "0xabc", 10, "0xfrom", "0xcontract", now)

	block := uint64(10)
	require.NoError(t, tx.TransitionTo(StatusMined, &block, "", now.Add(time.Second)))
	require.NoError(t, tx.TransitionTo(StatusConfirmed, &block, "", now.Add(2*time.Second)))
}

func TestTransition_ConfirmedToSubmittedFails(t *testing.T) {
	now := time.Now()
	tx := NewSubmitted("req-1", "0xabc", 10, "0xfrom", "0xcontract", now)
	block := uint64(10)
	_ = tx.TransitionTo(StatusMined, &block, "", now.Add(time.Second))
	_ = tx.TransitionTo(StatusConfirmed, &block, "", now.Add(2*time.Second))

	err := tx.TransitionTo(StatusSubmitted, nil, "", now.Add(3*time.Second))
	require.Error(t, err)
}
