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

func TestBuildQuotaBalancedAccountOrder_PrefersLowerQuotaUsageWithinPriority(t *testing.T) {
	resetLocalRoundRobinCountersForTest()
	ctx := context.Background()
	key := accountLoadRoundRobinKey("quota_unit", nil, PlatformOpenAI, "gpt-5.2")
	accounts := []*Account{
		{
			ID:       21,
			Platform: PlatformOpenAI,
			Priority: 0,
			Extra: map[string]any{
				"codex_7d_used_percent": 80.0,
				"codex_5h_used_percent": 1.0,
			},
		},
		{
			ID:       22,
			Platform: PlatformOpenAI,
			Priority: 0,
			Extra: map[string]any{
				"codex_7d_used_percent": 20.0,
				"codex_5h_used_percent": 20.0,
			},
		},
		{
			ID:       23,
			Platform: PlatformOpenAI,
			Priority: 1,
			Extra: map[string]any{
				"codex_7d_used_percent": 1.0,
				"codex_5h_used_percent": 1.0,
			},
		},
	}

	ordered := buildQuotaBalancedAccountOrder(ctx, nil, accounts, false, key)

	require.Equal(t, []int64{22, 21, 23}, accountIDsForTest(ordered))
}

func TestBuildQuotaBalancedAccountOrder_UsesFiveHourAfterSevenDayTie(t *testing.T) {
	resetLocalRoundRobinCountersForTest()
	ctx := context.Background()
	key := accountLoadRoundRobinKey("quota_7d_then_5h", nil, PlatformOpenAI, "gpt-5.2")
	accounts := []*Account{
		{ID: 25, Platform: PlatformOpenAI, Priority: 0, Extra: map[string]any{"codex_7d_used_percent": 20.0, "codex_5h_used_percent": 90.0}},
		{ID: 26, Platform: PlatformOpenAI, Priority: 0, Extra: map[string]any{"codex_7d_used_percent": 20.0, "codex_5h_used_percent": 10.0}},
	}

	ordered := buildQuotaBalancedAccountOrder(ctx, nil, accounts, false, key)

	require.Equal(t, []int64{26, 25}, accountIDsForTest(ordered))
}

func TestBuildQuotaBalancedAccountOrder_RotatesEqualQuotaBucket(t *testing.T) {
	resetLocalRoundRobinCountersForTest()
	ctx := context.Background()
	key := accountLoadRoundRobinKey("quota_equal", nil, PlatformOpenAI, "gpt-5.2")
	accounts := []*Account{
		{ID: 31, Platform: PlatformOpenAI, Priority: 0, Extra: map[string]any{"codex_7d_used_percent": 20.1, "codex_5h_used_percent": 10.1}},
		{ID: 32, Platform: PlatformOpenAI, Priority: 0, Extra: map[string]any{"codex_7d_used_percent": 20.9, "codex_5h_used_percent": 10.9}},
	}

	first := buildQuotaBalancedAccountOrder(ctx, nil, accounts, false, key)
	second := buildQuotaBalancedAccountOrder(ctx, nil, accounts, false, key)

	require.Equal(t, []int64{31, 32}, accountIDsForTest(first))
	require.Equal(t, []int64{32, 31}, accountIDsForTest(second))
}

func TestBuildQuotaBalancedAccountLoadOrder_RespectsPriorityAndLoad(t *testing.T) {
	resetLocalRoundRobinCountersForTest()
	ctx := context.Background()
	key := accountLoadRoundRobinKey("quota_load", nil, PlatformOpenAI, "gpt-5.2")
	items := []accountWithLoad{
		{account: &Account{ID: 41, Platform: PlatformOpenAI, Priority: 0, Extra: map[string]any{"codex_5h_used_percent": 80.0}}, loadInfo: &AccountLoadInfo{LoadRate: 10}},
		{account: &Account{ID: 42, Platform: PlatformOpenAI, Priority: 0, Extra: map[string]any{"codex_5h_used_percent": 1.0}}, loadInfo: &AccountLoadInfo{LoadRate: 90}},
		{account: &Account{ID: 43, Platform: PlatformOpenAI, Priority: 1, Extra: map[string]any{"codex_5h_used_percent": 1.0}}, loadInfo: &AccountLoadInfo{LoadRate: 0}},
	}

	ordered := buildQuotaBalancedAccountLoadOrder(ctx, nil, items, false, key)

	require.Equal(t, []int64{41, 42, 43}, accountLoadIDsForTest(ordered))
}

func TestNormalizeFallbackSelectionMode(t *testing.T) {
	require.Equal(t, SchedulerFallbackSelectionLastUsed, normalizeFallbackSelectionMode(""))
	require.Equal(t, SchedulerFallbackSelectionLastUsed, normalizeFallbackSelectionMode("bad"))
	require.Equal(t, SchedulerFallbackSelectionRandom, normalizeFallbackSelectionMode(" random "))
	require.Equal(t, SchedulerFallbackSelectionRoundRobin, normalizeFallbackSelectionMode("round-robin"))
	require.Equal(t, SchedulerFallbackSelectionRoundRobin, normalizeFallbackSelectionMode("rr"))
	require.Equal(t, SchedulerFallbackSelectionQuotaBalanced, normalizeFallbackSelectionMode("quota-balanced"))
	require.Equal(t, SchedulerFallbackSelectionQuotaBalanced, normalizeFallbackSelectionMode("quota"))
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
