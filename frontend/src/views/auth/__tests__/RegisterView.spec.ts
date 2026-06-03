import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import RegisterView from '@/views/auth/RegisterView.vue'

const {
  pushMock,
  registerMock,
  showSuccessMock,
  showErrorMock,
  showWarningMock,
  getPublicSettingsMock,
} = vi.hoisted(() => ({
  pushMock: vi.fn(),
  registerMock: vi.fn(),
  showSuccessMock: vi.fn(),
  showErrorMock: vi.fn(),
  showWarningMock: vi.fn(),
  getPublicSettingsMock: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ query: { promo: 'PROMO-URL' } }),
}))

vi.mock('vue-i18n', () => ({
  createI18n: () => ({
    global: {
      t: (key: string) => key,
    },
  }),
  useI18n: () => ({
    t: (key: string, params?: Record<string, string | number>) => {
      if (key === 'auth.accountCreatedSuccess') return `created:${params?.siteName ?? ''}`
      return key
    },
    locale: { value: 'en' },
  }),
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => ({
    register: (...args: any[]) => registerMock(...args),
  }),
  useAppStore: () => ({
    showSuccess: (...args: any[]) => showSuccessMock(...args),
    showError: (...args: any[]) => showErrorMock(...args),
    showWarning: (...args: any[]) => showWarningMock(...args),
  }),
}))

vi.mock('@/api/auth', () => {
  return {
    getPublicSettings: (...args: any[]) => getPublicSettingsMock(...args),
    isWeChatWebOAuthEnabled: () => false,
    validateInvitationCode: vi.fn(),
  }
})

vi.mock('@/utils/oauthAffiliate', () => {
  return {
    resolveAffiliateReferralCode: () => '',
    loadAffiliateReferralCode: () => '',
    clearAffiliateReferralCode: vi.fn(),
  }
})

describe('RegisterView', () => {
  beforeEach(() => {
    pushMock.mockReset()
    registerMock.mockReset()
    showSuccessMock.mockReset()
    showErrorMock.mockReset()
    showWarningMock.mockReset()
    getPublicSettingsMock.mockReset()
    sessionStorage.clear()
    localStorage.clear()

    getPublicSettingsMock.mockResolvedValue({
      registration_enabled: true,
      email_verify_enabled: false,
      promo_code_enabled: true,
      invitation_code_enabled: false,
      turnstile_enabled: false,
      turnstile_site_key: '',
      site_name: 'codex',
      linuxdo_oauth_enabled: false,
      oidc_oauth_enabled: false,
      oidc_oauth_provider_name: 'OIDC',
      github_oauth_enabled: false,
      google_oauth_enabled: false,
      registration_email_suffix_whitelist: [],
      login_agreement_enabled: false,
      login_agreement_documents: [],
    })
    registerMock.mockResolvedValue({})
  })

  it('does not render or submit promo code during registration', async () => {
    const wrapper = mount(RegisterView, {
      global: {
        stubs: {
          AuthLayout: { template: '<div><slot /><slot name="footer" /></div>' },
          Icon: true,
          TurnstileWidget: true,
          LoginAgreementPrompt: true,
          EmailOAuthButtons: true,
          LinuxDoOAuthSection: true,
          WechatOAuthSection: true,
          OidcOAuthSection: true,
          RouterLink: { template: '<a><slot /></a>' },
          transition: false,
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('#promo_code').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('auth.promoCodeLabel')

    await wrapper.get('#email').setValue('new@example.com')
    await wrapper.get('#password').setValue('secret-123')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(registerMock).toHaveBeenCalledWith({
      email: 'new@example.com',
      password: 'secret-123',
      turnstile_token: undefined,
      invitation_code: undefined,
    })
  })
})
