package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	SchedulerFallbackSelectionLastUsed   = "last_used"
	SchedulerFallbackSelectionRandom     = "random"
	SchedulerFallbackSelectionRoundRobin = "round_robin"
)

var localRoundRobinCounters sync.Map // key string -> *atomic.Uint64

func normalizeFallbackSelectionMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case SchedulerFallbackSelectionRandom:
		return SchedulerFallbackSelectionRandom
	case SchedulerFallbackSelectionRoundRobin, "round-robin", "roundrobin", "rr":
		return SchedulerFallbackSelectionRoundRobin
	default:
		return SchedulerFallbackSelectionLastUsed
	}
}

func isRoundRobinSelectionMode(mode string) bool {
	return normalizeFallbackSelectionMode(mode) == SchedulerFallbackSelectionRoundRobin
}

func nextLocalRoundRobin(key string) uint64 {
	if key == "" {
		key = "default"
	}
	value, _ := localRoundRobinCounters.LoadOrStore(key, &atomic.Uint64{})
	counter, _ := value.(*atomic.Uint64)
	if counter == nil {
		counter = &atomic.Uint64{}
		localRoundRobinCounters.Store(key, counter)
	}
	return counter.Add(1)
}

func resetLocalRoundRobinCountersForTest() {
	localRoundRobinCounters.Range(func(key, _ any) bool {
		localRoundRobinCounters.Delete(key)
		return true
	})
}

func accountRoundRobinKey(scope string, groupID *int64, platform string, requestedModel string, requireCompact bool, requiredCapability OpenAIEndpointCapability) string {
	compact := "0"
	if requireCompact {
		compact = "1"
	}
	return fmt.Sprintf("%s:g%d:p%s:m%s:c%s:e%s",
		strings.TrimSpace(scope),
		derefGroupID(groupID),
		strings.TrimSpace(platform),
		strings.TrimSpace(requestedModel),
		compact,
		strings.TrimSpace(string(requiredCapability)),
	)
}

func accountLoadRoundRobinKey(scope string, groupID *int64, platform string, requestedModel string) string {
	return accountRoundRobinKey(scope, groupID, platform, requestedModel, false, "")
}

type roundRobinCounter interface {
	nextRoundRobin(ctx context.Context, key string) uint64
}

func nextRoundRobinOffset(ctx context.Context, counter roundRobinCounter, key string, length int) int {
	if length <= 1 {
		return 0
	}
	var seq uint64
	if counter != nil {
		seq = counter.nextRoundRobin(ctx, key)
	} else {
		seq = nextLocalRoundRobin(key)
	}
	if seq == 0 {
		return 0
	}
	return int((seq - 1) % uint64(length))
}

func (s *SchedulerSnapshotService) nextRoundRobin(ctx context.Context, key string) uint64 {
	if s != nil && s.cache != nil {
		if seq, err := s.cache.NextRoundRobin(ctx, key); err == nil && seq > 0 {
			return seq
		}
	}
	return nextLocalRoundRobin(key)
}

func (s *OpenAIGatewayService) nextRoundRobin(ctx context.Context, key string) uint64 {
	if s != nil && s.schedulerSnapshot != nil {
		return s.schedulerSnapshot.nextRoundRobin(ctx, "openai:"+key)
	}
	return nextLocalRoundRobin("openai:" + key)
}

func (s *GatewayService) nextRoundRobin(ctx context.Context, key string) uint64 {
	if s != nil && s.schedulerSnapshot != nil {
		return s.schedulerSnapshot.nextRoundRobin(ctx, "gateway:"+key)
	}
	return nextLocalRoundRobin("gateway:" + key)
}

func rotateAccountWithLoadByOffset(items []accountWithLoad, offset int) []accountWithLoad {
	if len(items) <= 1 {
		return append([]accountWithLoad(nil), items...)
	}
	offset %= len(items)
	if offset < 0 {
		offset += len(items)
	}
	out := make([]accountWithLoad, 0, len(items))
	out = append(out, items[offset:]...)
	out = append(out, items[:offset]...)
	return out
}

func rotateAccountCandidatesByOffset(items []openAIAccountCandidateScore, offset int) []openAIAccountCandidateScore {
	if len(items) <= 1 {
		return append([]openAIAccountCandidateScore(nil), items...)
	}
	offset %= len(items)
	if offset < 0 {
		offset += len(items)
	}
	out := make([]openAIAccountCandidateScore, 0, len(items))
	out = append(out, items[offset:]...)
	out = append(out, items[:offset]...)
	return out
}

func rotateAccountsByOffset(items []*Account, offset int) []*Account {
	if len(items) <= 1 {
		return append([]*Account(nil), items...)
	}
	offset %= len(items)
	if offset < 0 {
		offset += len(items)
	}
	out := make([]*Account, 0, len(items))
	out = append(out, items[offset:]...)
	out = append(out, items[:offset]...)
	return out
}

func sortAccountLoadByPriorityLoadAndID(items []accountWithLoad, preferOAuth bool) {
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if a.account.Priority != b.account.Priority {
			return a.account.Priority < b.account.Priority
		}
		if a.loadInfo.LoadRate != b.loadInfo.LoadRate {
			return a.loadInfo.LoadRate < b.loadInfo.LoadRate
		}
		if preferOAuth && a.account.Type != b.account.Type {
			return a.account.Type == AccountTypeOAuth
		}
		return a.account.ID < b.account.ID
	})
}

func buildRoundRobinAccountLoadOrder(ctx context.Context, counter roundRobinCounter, items []accountWithLoad, preferOAuth bool, key string) []accountWithLoad {
	if len(items) == 0 {
		return nil
	}
	ordered := append([]accountWithLoad(nil), items...)
	sortAccountLoadByPriorityLoadAndID(ordered, preferOAuth)
	out := make([]accountWithLoad, 0, len(ordered))
	for start := 0; start < len(ordered); {
		end := start + 1
		for end < len(ordered) &&
			ordered[end].account.Priority == ordered[start].account.Priority &&
			ordered[end].loadInfo.LoadRate == ordered[start].loadInfo.LoadRate &&
			(!preferOAuth || ordered[end].account.Type == ordered[start].account.Type) {
			end++
		}
		group := ordered[start:end]
		out = append(out, rotateAccountWithLoadByOffset(group, nextRoundRobinOffset(ctx, counter, key, len(group)))...)
		start = end
	}
	return out
}

func buildRoundRobinAccountOrder(ctx context.Context, counter roundRobinCounter, accounts []*Account, preferOAuth bool, key string) []*Account {
	if len(accounts) == 0 {
		return nil
	}
	ordered := append([]*Account(nil), accounts...)
	sortAccountsByPriorityOnly(ordered, preferOAuth)
	out := make([]*Account, 0, len(ordered))
	for start := 0; start < len(ordered); {
		end := start + 1
		for end < len(ordered) &&
			ordered[end].Priority == ordered[start].Priority &&
			(!preferOAuth || ordered[end].Type == ordered[start].Type) {
			end++
		}
		group := ordered[start:end]
		out = append(out, rotateAccountsByOffset(group, nextRoundRobinOffset(ctx, counter, key, len(group)))...)
		start = end
	}
	return out
}

func buildRoundRobinOpenAIAccountOrder(ctx context.Context, counter roundRobinCounter, accounts []*Account, key string, requireCompact bool) []*Account {
	if !requireCompact {
		return buildRoundRobinAccountOrder(ctx, counter, accounts, false, key)
	}
	supported := make([]*Account, 0, len(accounts))
	unknown := make([]*Account, 0, len(accounts))
	for _, account := range accounts {
		switch openAICompactSupportTier(account) {
		case 2:
			supported = append(supported, account)
		case 1:
			unknown = append(unknown, account)
		}
	}
	out := make([]*Account, 0, len(supported)+len(unknown))
	out = append(out, buildRoundRobinAccountOrder(ctx, counter, supported, false, key+":compact_supported")...)
	out = append(out, buildRoundRobinAccountOrder(ctx, counter, unknown, false, key+":compact_unknown")...)
	return out
}

func buildRoundRobinOpenAIAccountLoadOrder(ctx context.Context, counter roundRobinCounter, items []accountWithLoad, key string, requireCompact bool, includeUnsupported bool) []accountWithLoad {
	if !requireCompact {
		return buildRoundRobinAccountLoadOrder(ctx, counter, items, false, key)
	}
	supported := make([]accountWithLoad, 0, len(items))
	unknown := make([]accountWithLoad, 0, len(items))
	unsupported := make([]accountWithLoad, 0, len(items))
	for _, item := range items {
		switch openAICompactSupportTier(item.account) {
		case 2:
			supported = append(supported, item)
		case 1:
			unknown = append(unknown, item)
		default:
			if includeUnsupported {
				unsupported = append(unsupported, item)
			}
		}
	}
	out := make([]accountWithLoad, 0, len(items))
	out = append(out, buildRoundRobinAccountLoadOrder(ctx, counter, supported, false, key+":compact_supported")...)
	out = append(out, buildRoundRobinAccountLoadOrder(ctx, counter, unknown, false, key+":compact_unknown")...)
	if includeUnsupported {
		out = append(out, buildRoundRobinAccountLoadOrder(ctx, counter, unsupported, false, key+":compact_unsupported")...)
	}
	return out
}
