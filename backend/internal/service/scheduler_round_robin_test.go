//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildRoundRobinAccountOrder_RotatesWithinPriority(t *testing.T) {
	resetLocalRoundRobinCountersForTest()
	ctx := context.Background()
	groupID := int64(17)
	key := accountLoadRoundRobinKey("unit", &groupID, PlatformAnthropic, "claude-sonnet")
	accounts := []*Account{
		{ID: 3, Priority: 1},
		{ID: 1, Priority: 0},
		{ID: 2, Priority: 0},
		{ID: 4, Priority: 2},
	}

	first := buildRoundRobinAccountOrder(ctx, nil, accounts, false, key)
	second := buildRoundRobinAccountOrder(ctx, nil, accounts, false, key)
	third := buildRoundRobinAccountOrder(ctx, nil, accounts, false, key)

	require.Equal(t, []int64{1, 2, 3, 4}, accountIDsForTest(first))
	require.Equal(t, []int64{2, 1, 3, 4}, accountIDsForTest(second))
	require.Equal(t, []int64{1, 2, 3, 4}, accountIDsForTest(third))
}

func TestBuildRoundRobinAccountLoadOrder_RotatesWithinPriorityAndLoad(t *testing.T) {
	resetLocalRoundRobinCountersForTest()
	ctx := context.Background()
	groupID := int64(18)
	key := accountLoadRoundRobinKey("unit_load", &groupID, PlatformOpenAI, "gpt-5.1")
	accounts := []accountWithLoad{
		{account: &Account{ID: 11, Priority: 0}, loadInfo: &AccountLoadInfo{LoadRate: 20}},
		{account: &Account{ID: 12, Priority: 0}, loadInfo: &AccountLoadInfo{LoadRate: 20}},
		{account: &Account{ID: 13, Priority: 0}, loadInfo: &AccountLoadInfo{LoadRate: 60}},
	}

	first := buildRoundRobinAccountLoadOrder(ctx, nil, accounts, false, key)
	second := buildRoundRobinAccountLoadOrder(ctx, nil, accounts, false, key)

	require.Equal(t, []int64{11, 12, 13}, accountLoadIDsForTest(first))
	require.Equal(t, []int64{12, 11, 13}, accountLoadIDsForTest(second))
}

func TestNormalizeFallbackSelectionMode(t *testing.T) {
	require.Equal(t, SchedulerFallbackSelectionLastUsed, normalizeFallbackSelectionMode(""))
	require.Equal(t, SchedulerFallbackSelectionLastUsed, normalizeFallbackSelectionMode("bad"))
	require.Equal(t, SchedulerFallbackSelectionRandom, normalizeFallbackSelectionMode(" random "))
	require.Equal(t, SchedulerFallbackSelectionRoundRobin, normalizeFallbackSelectionMode("round-robin"))
	require.Equal(t, SchedulerFallbackSelectionRoundRobin, normalizeFallbackSelectionMode("rr"))
}

func accountIDsForTest(accounts []*Account) []int64 {
	out := make([]int64, 0, len(accounts))
	for _, account := range accounts {
		out = append(out, account.ID)
	}
	return out
}

func accountLoadIDsForTest(accounts []accountWithLoad) []int64 {
	out := make([]int64, 0, len(accounts))
	for _, item := range accounts {
		out = append(out, item.account.ID)
	}
	return out
}
