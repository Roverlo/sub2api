package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	SchedulerFallbackSelectionLastUsed      = "last_used"
	SchedulerFallbackSelectionRandom        = "random"
	SchedulerFallbackSelectionRoundRobin    = "round_robin"
	SchedulerFallbackSelectionQuotaBalanced = "quota_balanced"
)

var localRoundRobinCounters sync.Map // key string -> *atomic.Uint64

func normalizeFallbackSelectionMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case SchedulerFallbackSelectionRandom:
		return SchedulerFallbackSelectionRandom
	case SchedulerFallbackSelectionRoundRobin, "round-robin", "roundrobin", "rr":
		return SchedulerFallbackSelectionRoundRobin
	case SchedulerFallbackSelectionQuotaBalanced, "quota-balanced", "quota_balance", "quotabalanced", "quota":
		return SchedulerFallbackSelectionQuotaBalanced
	default:
		return SchedulerFallbackSelectionLastUsed
	}
}

func isRoundRobinSelectionMode(mode string) bool {
	return normalizeFallbackSelectionMode(mode) == SchedulerFallbackSelectionRoundRobin
}

func isQuotaBalancedSelectionMode(mode string) bool {
	return normalizeFallbackSelectionMode(mode) == SchedulerFallbackSelectionQuotaBalanced
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

type quotaBalancedAccount struct {
	account *Account
	quota   quotaBalancedKey
}

type quotaBalancedAccountWithLoad struct {
	item  accountWithLoad
	quota quotaBalancedKey
}

type quotaBalancedOpenAICandidate struct {
	candidate openAIAccountCandidateScore
	quota     quotaBalancedKey
}

type quotaBalancedKey struct {
	has7d         bool
	bucket7d      int
	has5h         bool
	bucket5h      int
	hasGeneric    bool
	bucketGeneric int
}

func accountQuotaBalanceKey(account *Account, now time.Time) quotaBalancedKey {
	if account == nil {
		return quotaBalancedKey{}
	}

	key := quotaBalancedKey{}
	if account.IsOpenAI() {
		key.bucket7d, key.has7d = openAIQuotaUsageBucket(account.Extra, "7d", now)
		key.bucket5h, key.has5h = openAIQuotaUsageBucket(account.Extra, "5h", now)
	}
	key.bucketGeneric, key.hasGeneric = genericQuotaUsageBucket(account)
	return key
}

func openAIQuotaUsageBucket(extra map[string]any, window string, now time.Time) (int, bool) {
	usedPercent, ok := resolveAccountExtraNumber(extra, "codex_"+window+"_used_percent")
	if !ok {
		return 0, false
	}
	if openAIQuotaWindowReset(extra, window, now) {
		usedPercent = 0
	} else if openAICodexSnapshotStaleForPause(extra, now) {
		return 0, false
	}
	return quotaUtilizationBucket(usedPercent / 100), true
}

func genericQuotaUsageBucket(account *Account) (int, bool) {
	var maxUtilization float64
	hasQuota := false
	addUtilization := func(used, limit float64) {
		if limit <= 0 {
			return
		}
		if utilization := used / limit; utilization > maxUtilization {
			maxUtilization = utilization
		}
		hasQuota = true
	}
	addUtilization(account.GetQuotaUsed(), account.GetQuotaLimit())
	if account.GetQuotaDailyLimit() > 0 && account.IsDailyQuotaPeriodExpired() {
		addUtilization(0, account.GetQuotaDailyLimit())
	} else {
		addUtilization(account.GetQuotaDailyUsed(), account.GetQuotaDailyLimit())
	}
	if account.GetQuotaWeeklyLimit() > 0 && account.IsWeeklyQuotaPeriodExpired() {
		addUtilization(0, account.GetQuotaWeeklyLimit())
	} else {
		addUtilization(account.GetQuotaWeeklyUsed(), account.GetQuotaWeeklyLimit())
	}
	if !hasQuota {
		return 0, false
	}
	return quotaUtilizationBucket(maxUtilization), true
}

func quotaUtilizationBucket(utilization float64) int {
	return int(clamp01(utilization) * 100)
}

func compareQuotaBalancedKey(a, b quotaBalancedKey) int {
	if a.has7d != b.has7d {
		if a.has7d {
			return -1
		}
		return 1
	}
	if a.has7d && a.bucket7d != b.bucket7d {
		return compareInt(a.bucket7d, b.bucket7d)
	}
	if a.has5h != b.has5h {
		if a.has5h {
			return -1
		}
		return 1
	}
	if a.has5h && a.bucket5h != b.bucket5h {
		return compareInt(a.bucket5h, b.bucket5h)
	}
	if a.hasGeneric != b.hasGeneric {
		if a.hasGeneric {
			return -1
		}
		return 1
	}
	if a.hasGeneric && a.bucketGeneric != b.bucketGeneric {
		return compareInt(a.bucketGeneric, b.bucketGeneric)
	}
	return 0
}

func sameQuotaBalancedKey(a, b quotaBalancedKey) bool {
	return a.has7d == b.has7d &&
		a.bucket7d == b.bucket7d &&
		a.has5h == b.has5h &&
		a.bucket5h == b.bucket5h &&
		a.hasGeneric == b.hasGeneric &&
		a.bucketGeneric == b.bucketGeneric
}

func compareInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func sameQuotaBalancedAccountGroup(a, b quotaBalancedAccount, preferOAuth bool) bool {
	if a.account.Priority != b.account.Priority {
		return false
	}
	if preferOAuth && a.account.Type != b.account.Type {
		return false
	}
	return sameQuotaBalancedKey(a.quota, b.quota)
}

func sameQuotaBalancedAccountLoadGroup(a, b quotaBalancedAccountWithLoad, preferOAuth bool) bool {
	if a.item.account.Priority != b.item.account.Priority {
		return false
	}
	if a.item.loadInfo.LoadRate != b.item.loadInfo.LoadRate {
		return false
	}
	if preferOAuth && a.item.account.Type != b.item.account.Type {
		return false
	}
	return sameQuotaBalancedKey(a.quota, b.quota)
}

func rotateQuotaBalancedAccountsByOffset(items []quotaBalancedAccount, offset int) []quotaBalancedAccount {
	if len(items) <= 1 {
		return append([]quotaBalancedAccount(nil), items...)
	}
	offset %= len(items)
	if offset < 0 {
		offset += len(items)
	}
	out := make([]quotaBalancedAccount, 0, len(items))
	out = append(out, items[offset:]...)
	out = append(out, items[:offset]...)
	return out
}

func rotateQuotaBalancedAccountLoadsByOffset(items []quotaBalancedAccountWithLoad, offset int) []quotaBalancedAccountWithLoad {
	if len(items) <= 1 {
		return append([]quotaBalancedAccountWithLoad(nil), items...)
	}
	offset %= len(items)
	if offset < 0 {
		offset += len(items)
	}
	out := make([]quotaBalancedAccountWithLoad, 0, len(items))
	out = append(out, items[offset:]...)
	out = append(out, items[:offset]...)
	return out
}

func buildQuotaBalancedAccountOrder(ctx context.Context, counter roundRobinCounter, accounts []*Account, preferOAuth bool, key string) []*Account {
	if len(accounts) == 0 {
		return nil
	}
	now := time.Now()
	ordered := make([]quotaBalancedAccount, 0, len(accounts))
	for _, account := range accounts {
		ordered = append(ordered, quotaBalancedAccount{
			account: account,
			quota:   accountQuotaBalanceKey(account, now),
		})
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		a, b := ordered[i], ordered[j]
		if a.account.Priority != b.account.Priority {
			return a.account.Priority < b.account.Priority
		}
		if preferOAuth && a.account.Type != b.account.Type {
			return a.account.Type == AccountTypeOAuth
		}
		if cmp := compareQuotaBalancedKey(a.quota, b.quota); cmp != 0 {
			return cmp < 0
		}
		return a.account.ID < b.account.ID
	})

	out := make([]*Account, 0, len(ordered))
	for start := 0; start < len(ordered); {
		end := start + 1
		for end < len(ordered) && sameQuotaBalancedAccountGroup(ordered[start], ordered[end], preferOAuth) {
			end++
		}
		group := rotateQuotaBalancedAccountsByOffset(ordered[start:end], nextRoundRobinOffset(ctx, counter, key, end-start))
		for _, item := range group {
			out = append(out, item.account)
		}
		start = end
	}
	return out
}

func buildQuotaBalancedAccountLoadOrder(ctx context.Context, counter roundRobinCounter, items []accountWithLoad, preferOAuth bool, key string) []accountWithLoad {
	if len(items) == 0 {
		return nil
	}
	now := time.Now()
	ordered := make([]quotaBalancedAccountWithLoad, 0, len(items))
	for _, item := range items {
		ordered = append(ordered, quotaBalancedAccountWithLoad{
			item:  item,
			quota: accountQuotaBalanceKey(item.account, now),
		})
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		a, b := ordered[i], ordered[j]
		if a.item.account.Priority != b.item.account.Priority {
			return a.item.account.Priority < b.item.account.Priority
		}
		if a.item.loadInfo.LoadRate != b.item.loadInfo.LoadRate {
			return a.item.loadInfo.LoadRate < b.item.loadInfo.LoadRate
		}
		if preferOAuth && a.item.account.Type != b.item.account.Type {
			return a.item.account.Type == AccountTypeOAuth
		}
		if cmp := compareQuotaBalancedKey(a.quota, b.quota); cmp != 0 {
			return cmp < 0
		}
		return a.item.account.ID < b.item.account.ID
	})

	out := make([]accountWithLoad, 0, len(ordered))
	for start := 0; start < len(ordered); {
		end := start + 1
		for end < len(ordered) && sameQuotaBalancedAccountLoadGroup(ordered[start], ordered[end], preferOAuth) {
			end++
		}
		group := rotateQuotaBalancedAccountLoadsByOffset(ordered[start:end], nextRoundRobinOffset(ctx, counter, key, end-start))
		for _, item := range group {
			out = append(out, item.item)
		}
		start = end
	}
	return out
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

func buildQuotaBalancedOpenAIAccountOrder(ctx context.Context, counter roundRobinCounter, accounts []*Account, key string, requireCompact bool) []*Account {
	if !requireCompact {
		return buildQuotaBalancedAccountOrder(ctx, counter, accounts, false, key)
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
	out = append(out, buildQuotaBalancedAccountOrder(ctx, counter, supported, false, key+":compact_supported")...)
	out = append(out, buildQuotaBalancedAccountOrder(ctx, counter, unknown, false, key+":compact_unknown")...)
	return out
}

func buildQuotaBalancedOpenAIAccountLoadOrder(ctx context.Context, counter roundRobinCounter, items []accountWithLoad, key string, requireCompact bool, includeUnsupported bool) []accountWithLoad {
	if !requireCompact {
		return buildQuotaBalancedAccountLoadOrder(ctx, counter, items, false, key)
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
	out = append(out, buildQuotaBalancedAccountLoadOrder(ctx, counter, supported, false, key+":compact_supported")...)
	out = append(out, buildQuotaBalancedAccountLoadOrder(ctx, counter, unknown, false, key+":compact_unknown")...)
	if includeUnsupported {
		out = append(out, buildQuotaBalancedAccountLoadOrder(ctx, counter, unsupported, false, key+":compact_unsupported")...)
	}
	return out
}

func buildQuotaBalancedOpenAICandidateOrder(ctx context.Context, counter roundRobinCounter, candidates []openAIAccountCandidateScore, key string) []openAIAccountCandidateScore {
	if len(candidates) == 0 {
		return nil
	}
	now := time.Now()
	ordered := make([]quotaBalancedOpenAICandidate, 0, len(candidates))
	for _, candidate := range candidates {
		ordered = append(ordered, quotaBalancedOpenAICandidate{
			candidate: candidate,
			quota:     accountQuotaBalanceKey(candidate.account, now),
		})
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		a, b := ordered[i], ordered[j]
		if a.candidate.account.Priority != b.candidate.account.Priority {
			return a.candidate.account.Priority < b.candidate.account.Priority
		}
		if a.candidate.loadInfo.LoadRate != b.candidate.loadInfo.LoadRate {
			return a.candidate.loadInfo.LoadRate < b.candidate.loadInfo.LoadRate
		}
		if a.candidate.loadInfo.WaitingCount != b.candidate.loadInfo.WaitingCount {
			return a.candidate.loadInfo.WaitingCount < b.candidate.loadInfo.WaitingCount
		}
		if cmp := compareQuotaBalancedKey(a.quota, b.quota); cmp != 0 {
			return cmp < 0
		}
		if a.candidate.score != b.candidate.score {
			return a.candidate.score > b.candidate.score
		}
		return a.candidate.account.ID < b.candidate.account.ID
	})

	out := make([]openAIAccountCandidateScore, 0, len(ordered))
	for start := 0; start < len(ordered); {
		end := start + 1
		for end < len(ordered) &&
			ordered[end].candidate.account.Priority == ordered[start].candidate.account.Priority &&
			ordered[end].candidate.loadInfo.LoadRate == ordered[start].candidate.loadInfo.LoadRate &&
			ordered[end].candidate.loadInfo.WaitingCount == ordered[start].candidate.loadInfo.WaitingCount &&
			sameQuotaBalancedKey(ordered[end].quota, ordered[start].quota) {
			end++
		}
		group := ordered[start:end]
		if len(group) > 1 {
			group = rotateQuotaBalancedOpenAICandidatesByOffset(group, nextRoundRobinOffset(ctx, counter, key, len(group)))
		}
		for _, item := range group {
			out = append(out, item.candidate)
		}
		start = end
	}
	return out
}

func rotateQuotaBalancedOpenAICandidatesByOffset(items []quotaBalancedOpenAICandidate, offset int) []quotaBalancedOpenAICandidate {
	if len(items) <= 1 {
		return append([]quotaBalancedOpenAICandidate(nil), items...)
	}
	offset %= len(items)
	if offset < 0 {
		offset += len(items)
	}
	out := make([]quotaBalancedOpenAICandidate, 0, len(items))
	out = append(out, items[offset:]...)
	out = append(out, items[:offset]...)
	return out
}
