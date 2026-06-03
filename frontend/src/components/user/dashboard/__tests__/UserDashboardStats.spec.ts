import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, string | number>) => {
      if (key === 'dashboard.platformCount') return `${params?.count ?? 0} platforms`
      return key
    },
  }),
}))

function buildStats(overrides: Partial<UserStatsType> = {}): UserStatsType {
  return {
    total_api_keys: 0,
    active_api_keys: 0,
    total_requests: 0,
    total_input_tokens: 0,
    total_output_tokens: 0,
    total_cache_creation_tokens: 0,
    total_cache_read_tokens: 0,
    total_tokens: 0,
    total_cost: 0,
    total_actual_cost: 0,
    today_requests: 0,
    today_input_tokens: 0,
    today_output_tokens: 0,
    today_cache_creation_tokens: 0,
    today_cache_read_tokens: 0,
    today_tokens: 0,
    today_cost: 0,
    today_actual_cost: 0,
    average_duration_ms: 0,
    rpm: 0,
    tpm: 0,
    by_platform: [],
    available_platforms: [],
    ...overrides,
  }
}

describe('UserDashboardStats', () => {
  it('shows only platforms configured by admin accounts', () => {
    const wrapper = mount(UserDashboardStats, {
      props: {
        stats: buildStats({
          total_actual_cost: 9,
          today_actual_cost: 4,
          available_platforms: ['openai'],
          by_platform: [
            {
              platform: 'openai',
              total_requests: 12,
              total_tokens: 3000,
              total_actual_cost: 3,
              today_requests: 5,
              today_tokens: 1000,
              today_actual_cost: 1,
            },
            {
              platform: 'anthropic',
              total_requests: 8,
              total_tokens: 2000,
              total_actual_cost: 6,
              today_requests: 3,
              today_tokens: 800,
              today_actual_cost: 3,
            },
          ],
        }),
        balance: 0,
        isSimple: false,
        platformQuotas: [
          {
            platform: 'openai',
            daily_limit_usd: 10,
            weekly_limit_usd: null,
            monthly_limit_usd: null,
            daily_usage_usd: 1,
            weekly_usage_usd: 0,
            monthly_usage_usd: 0,
          },
          {
            platform: 'gemini',
            daily_limit_usd: 10,
            weekly_limit_usd: null,
            monthly_limit_usd: null,
            daily_usage_usd: 0,
            weekly_usage_usd: 0,
            monthly_usage_usd: 0,
          },
        ],
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    expect(wrapper.text()).toContain('OpenAI')
    expect(wrapper.text()).toContain('1 platforms')
    expect(wrapper.text()).not.toContain('Claude')
    expect(wrapper.text()).not.toContain('Gemini')
    expect(wrapper.text()).not.toContain('Antigravity')
    expect(wrapper.text()).not.toContain('dashboard.platformOther')
  })
})
