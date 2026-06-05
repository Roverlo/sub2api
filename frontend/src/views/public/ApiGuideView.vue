<template>
  <div class="api-guide min-h-screen bg-[#f6f7fb] text-[#15171c]">
    <header class="sticky top-0 z-30 border-b border-white/10 bg-[#07090f]/95 text-white backdrop-blur">
      <div class="mx-auto flex max-w-7xl items-center justify-between gap-4 px-4 py-3 sm:px-6 lg:px-8">
        <RouterLink to="/home" class="flex min-w-0 items-center gap-3">
          <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center overflow-hidden rounded-md bg-white">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </span>
          <span class="truncate text-sm font-semibold">{{ siteName }}</span>
        </RouterLink>

        <nav class="flex flex-shrink-0 items-center gap-2">
          <RouterLink
            to="/login"
            class="hidden items-center gap-2 rounded-lg border border-white/15 px-3 py-2 text-sm font-medium text-white/80 transition hover:border-white/30 hover:text-white sm:inline-flex"
          >
            <Icon name="login" size="sm" />
            登录
          </RouterLink>
          <RouterLink
            :to="startPath"
            class="inline-flex items-center gap-2 rounded-lg bg-white px-3 py-2 text-sm font-semibold text-[#07090f] transition hover:bg-[#f1f4f8]"
          >
            <Icon name="userPlus" size="sm" />
            {{ isAuthenticated ? '进入控制台' : '注册账号' }}
          </RouterLink>
        </nav>
      </div>
    </header>

    <main>
      <section class="relative overflow-hidden bg-[#07090f] text-white">
        <div class="guide-grid absolute inset-0"></div>
        <div class="relative mx-auto max-w-7xl px-4 pb-8 pt-10 sm:px-6 sm:pb-10 sm:pt-14 lg:px-8 lg:pb-12 lg:pt-16">
          <div class="grid gap-8 lg:grid-cols-[0.78fr_1.22fr] lg:items-end">
          <div class="max-w-3xl">
            <p class="mb-4 inline-flex items-center gap-2 rounded-md border border-[#21d6b6]/30 bg-[#21d6b6]/10 px-3 py-1.5 text-sm font-medium text-[#8df5df]">
              <Icon name="bolt" size="sm" />
              小白向简版
            </p>
            <h1 class="max-w-3xl text-4xl font-bold leading-tight text-white sm:text-5xl lg:text-5xl">
              GPT 国内直连 API 注册与额度兑换说明
            </h1>
            <p class="mt-5 max-w-2xl text-base leading-8 text-white/72">
              注册账号、兑换额度、拿到 API 密钥后即可开始使用。通过国内服务器中转访问，无需翻墙，手机和电脑都可以用，也可以接入 ChatBox、NextChat 等常见客户端。
            </p>
            <div class="mt-7 flex flex-col gap-3 sm:flex-row">
              <RouterLink
                :to="startPath"
                class="inline-flex items-center justify-center gap-2 rounded-lg bg-[#21d6b6] px-5 py-3 text-sm font-semibold text-[#07100e] transition hover:bg-[#72f2db]"
              >
                <Icon name="arrowRight" size="sm" />
                {{ isAuthenticated ? '进入控制台' : '去注册并开始兑换' }}
              </RouterLink>
              <button
                type="button"
                class="inline-flex items-center justify-center gap-2 rounded-lg border border-white/18 px-5 py-3 text-sm font-semibold text-white transition hover:border-white/35 hover:bg-white/8"
                @click="copyRegisterUrl"
              >
                <Icon :name="copied ? 'checkCircle' : 'copy'" size="sm" />
                {{ copied ? '已复制注册网址' : '复制注册网址' }}
              </button>
            </div>

            <div class="mt-8 grid grid-cols-3 gap-2 lg:hidden">
              <div
                v-for="route in routeHighlights"
                :key="route.path"
                class="rounded-lg border border-white/12 bg-white/[0.04] p-3"
              >
                <Icon :name="route.icon" size="sm" class="text-[#8df5df]" />
                <div class="mt-2 text-xs font-semibold text-white">{{ route.title }}</div>
                <div class="mt-1 break-all font-mono text-[11px] text-white/48">{{ route.path }}</div>
              </div>
            </div>
          </div>

          <div class="hidden overflow-hidden rounded-lg border border-white/12 bg-[#0d1118]/88 shadow-2xl shadow-black/30 lg:block">
            <div class="flex flex-col gap-3 border-b border-white/10 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
              <div class="flex items-center gap-2 text-sm text-white/70">
                <span class="h-2.5 w-2.5 rounded-sm bg-[#ff6b6b]"></span>
                <span class="h-2.5 w-2.5 rounded-sm bg-[#ffd166]"></span>
                <span class="h-2.5 w-2.5 rounded-sm bg-[#21d6b6]"></span>
                <span class="ml-2 font-mono text-xs text-white/55">https://codex.260213.xyz</span>
              </div>
              <span class="inline-flex w-fit items-center gap-2 rounded-md border border-[#21d6b6]/30 bg-[#21d6b6]/10 px-2.5 py-1 text-xs font-medium text-[#8df5df]">
                <Icon name="checkCircle" size="xs" />
                按量计费 · 充多少用多少
              </span>
            </div>
            <div class="grid lg:grid-cols-[1.15fr_0.85fr]">
              <div class="border-b border-white/10 p-4 sm:p-6 lg:border-b-0 lg:border-r">
                <div class="grid gap-3 sm:grid-cols-3">
                  <div
                    v-for="metric in quickStats"
                    :key="metric.label"
                    class="rounded-lg border border-white/10 bg-white/[0.04] p-4"
                  >
                    <Icon :name="metric.icon" size="md" :class="metric.iconClass" />
                    <div class="mt-3 text-lg font-semibold text-white">{{ metric.value }}</div>
                    <div class="mt-1 text-xs leading-5 text-white/55">{{ metric.label }}</div>
                  </div>
                </div>

                <div class="mt-5 rounded-lg border border-white/10 bg-[#07090f] p-4 font-mono text-xs leading-6 text-white/72 sm:text-sm">
                  <div><span class="text-[#21d6b6]">$</span> openai --base-url https://codex.260213.xyz/v1</div>
                  <div><span class="text-[#21d6b6]">$</span> api-key sk-******</div>
                  <div class="text-[#ffd166]"># 模型与价格以网站实际显示为准</div>
                  <div class="text-[#8df5df]">200 OK · ready for ChatBox / NextChat / custom apps</div>
                </div>
              </div>

              <div class="grid divide-y divide-white/10">
                <div
                  v-for="route in routeHighlights"
                  :key="route.path"
                  class="flex items-start gap-4 p-4 sm:p-5"
                >
                  <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-md bg-white text-[#07090f]">
                    <Icon :name="route.icon" size="sm" />
                  </span>
                  <div class="min-w-0">
                    <div class="text-sm font-semibold text-white">{{ route.title }}</div>
                    <div class="mt-1 break-all font-mono text-xs text-[#8df5df]">{{ route.path }}</div>
                    <p class="mt-2 text-sm leading-6 text-white/60">{{ route.desc }}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
          </div>
        </div>
      </section>

      <section class="border-b border-[#dfe4ec] bg-white">
        <div class="mx-auto grid max-w-7xl gap-6 px-4 py-12 sm:px-6 lg:grid-cols-3 lg:px-8 lg:py-16">
          <article
            v-for="step in steps"
            :key="step.title"
            class="rounded-lg border border-[#dfe4ec] bg-white p-6 shadow-sm"
          >
            <div class="flex items-center justify-between gap-3">
              <span class="text-sm font-semibold text-[#6d7280]">{{ step.number }}</span>
              <span class="flex h-10 w-10 items-center justify-center rounded-md" :class="step.iconBoxClass">
                <Icon :name="step.icon" size="md" />
              </span>
            </div>
            <h2 class="mt-5 text-xl font-bold text-[#15171c]">{{ step.title }}</h2>
            <p class="mt-3 text-sm leading-7 text-[#5d6472]">{{ step.desc }}</p>
            <RouterLink
              :to="step.to"
              class="mt-5 inline-flex items-center gap-2 text-sm font-semibold text-[#0b806f] transition hover:text-[#075f53]"
            >
              {{ step.cta }}
              <Icon name="arrowRight" size="xs" />
            </RouterLink>
          </article>
        </div>
      </section>

      <section class="bg-[#f6f7fb]">
        <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8 lg:py-16">
          <div class="grid gap-8 lg:grid-cols-[0.85fr_1.15fr] lg:items-start">
            <div>
              <p class="text-sm font-semibold text-[#0b806f]">适合谁用</p>
              <h2 class="mt-3 text-3xl font-bold leading-tight text-[#15171c]">
                个人、开发者、日常办公用户都能轻量接入
              </h2>
              <p class="mt-4 text-sm leading-7 text-[#5d6472]">
                这是一项 API 中转服务。你不用自己维护上游账号和服务器，只需要一个站内账号、一份额度和一个 API 密钥，即可按实际调用消耗使用。
              </p>
            </div>

            <div class="grid gap-4 sm:grid-cols-2">
              <article
                v-for="feature in features"
                :key="feature.title"
                class="rounded-lg border border-[#dfe4ec] bg-white p-5 shadow-sm"
              >
                <Icon :name="feature.icon" size="md" :class="feature.iconClass" />
                <h3 class="mt-4 text-base font-bold text-[#15171c]">{{ feature.title }}</h3>
                <p class="mt-2 text-sm leading-6 text-[#5d6472]">{{ feature.desc }}</p>
              </article>
            </div>
          </div>
        </div>
      </section>

      <section class="border-y border-[#dfe4ec] bg-white">
        <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8 lg:py-16">
          <div class="grid gap-8 lg:grid-cols-[1fr_1fr]">
            <div>
              <p class="text-sm font-semibold text-[#0b806f]">接入客户端</p>
              <h2 class="mt-3 text-3xl font-bold text-[#15171c]">拿到密钥后，把它填到客户端或程序里</h2>
              <p class="mt-4 text-sm leading-7 text-[#5d6472]">
                推荐先用小额额度试跑。确认模型、速度和扣费符合预期后，再按自己的使用频率充值。
              </p>
            </div>

            <div class="grid gap-3">
              <div
                v-for="client in clients"
                :key="client.name"
                class="flex items-center justify-between gap-4 rounded-lg border border-[#dfe4ec] bg-[#fbfcff] px-4 py-3"
              >
                <div class="flex min-w-0 items-center gap-3">
                  <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-md" :class="client.iconBoxClass">
                    <Icon :name="client.icon" size="sm" />
                  </span>
                  <div class="min-w-0">
                    <div class="truncate text-sm font-bold text-[#15171c]">{{ client.name }}</div>
                    <div class="truncate text-xs text-[#6d7280]">{{ client.desc }}</div>
                  </div>
                </div>
                <span class="flex-shrink-0 rounded-md bg-[#edf8f5] px-2.5 py-1 text-xs font-semibold text-[#0b806f]">
                  可接入
                </span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="bg-[#07090f] text-white">
        <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8 lg:py-16">
          <div class="grid gap-6 lg:grid-cols-[0.8fr_1.2fr]">
            <div>
              <p class="text-sm font-semibold text-[#ffd166]">温馨提示</p>
              <h2 class="mt-3 text-3xl font-bold">密钥和兑换码只发给自己信任的人</h2>
              <p class="mt-4 text-sm leading-7 text-white/65">
                额度会根据实际调用消耗，具体扣费和模型价格以网站页面为准。如果不会注册、兑换或接入客户端，可以直接私信咨询。
              </p>
            </div>

            <div class="grid gap-4 sm:grid-cols-3">
              <article
                v-for="tip in tips"
                :key="tip.title"
                class="rounded-lg border border-white/12 bg-white/[0.04] p-5"
              >
                <Icon :name="tip.icon" size="md" :class="tip.iconClass" />
                <h3 class="mt-4 text-base font-bold">{{ tip.title }}</h3>
                <p class="mt-2 text-sm leading-6 text-white/62">{{ tip.desc }}</p>
              </article>
            </div>
          </div>

          <div class="mt-10 flex flex-col gap-3 border-t border-white/10 pt-8 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <div class="text-sm text-white/55">注册网址</div>
              <a
                :href="registerUrl"
                class="mt-1 inline-flex break-all font-mono text-sm font-semibold text-[#8df5df] hover:text-white"
              >
                {{ registerUrl }}
              </a>
            </div>
            <RouterLink
              :to="startPath"
              class="inline-flex items-center justify-center gap-2 rounded-lg bg-white px-5 py-3 text-sm font-semibold text-[#07090f] transition hover:bg-[#f1f4f8]"
            >
              <Icon name="arrowRight" size="sm" />
              {{ isAuthenticated ? '进入控制台' : '现在注册' }}
            </RouterLink>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore, useAuthStore } from '@/stores'

const appStore = useAppStore()
const authStore = useAuthStore()

const registerUrl = 'https://codex.260213.xyz/'
const copied = ref(false)

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const startPath = computed(() => (isAuthenticated.value ? (authStore.isAdmin ? '/admin/dashboard' : '/dashboard') : '/register'))

const quickStats = [
  {
    icon: 'globe',
    iconClass: 'text-[#8df5df]',
    value: '国内直连',
    label: '通过国内服务器中转访问'
  },
  {
    icon: 'dollar',
    iconClass: 'text-[#ffd166]',
    value: '按量计费',
    label: '充多少用多少，适合小额试用'
  },
  {
    icon: 'cpu',
    iconClass: 'text-[#9bb7ff]',
    value: '多端可用',
    label: '手机、电脑、常见客户端均可接入'
  }
] as const

const routeHighlights = [
  {
    icon: 'userPlus',
    title: '注册账号',
    path: '/register',
    desc: '按页面提示填写账号信息，完成后登录网站。'
  },
  {
    icon: 'gift',
    title: '兑换额度',
    path: '/redeem',
    desc: '填入提供给你的兑换码，确认后查看可用余额。'
  },
  {
    icon: 'key',
    title: '复制密钥',
    path: '/keys',
    desc: '创建或复制 API 密钥，再填到客户端或程序里。'
  }
] as const

const steps = [
  {
    number: '01',
    icon: 'userPlus',
    iconBoxClass: 'bg-[#e8f8f4] text-[#0b806f]',
    title: '打开网站并注册账号',
    desc: '登录 codex.260213.xyz，点击注册，按页面提示填写账号信息。注册完成后，使用账号登录网站。',
    to: '/register',
    cta: '去注册'
  },
  {
    number: '02',
    icon: 'gift',
    iconBoxClass: 'bg-[#fff4d8] text-[#9a6400]',
    title: '兑换额度',
    desc: '登录后找到“兑换”入口，把我们提供给你的兑换码填进去，然后点击确认兑换。',
    to: '/redeem',
    cta: '去兑换'
  },
  {
    number: '03',
    icon: 'key',
    iconBoxClass: 'bg-[#eaf0ff] text-[#395fcf]',
    title: '开始使用 API',
    desc: '在“API 密钥”页面复制你的密钥，之后按教程填到 ChatBox、NextChat 或自己的程序里使用。',
    to: '/keys',
    cta: '去拿密钥'
  }
] as const

const features = [
  {
    icon: 'server',
    iconClass: 'text-[#0b806f]',
    title: '不用自己搭建中转',
    desc: '面向个人和轻量开发场景，减少账号、网络和接入维护成本。'
  },
  {
    icon: 'chart',
    iconClass: 'text-[#395fcf]',
    title: '余额与消耗可查看',
    desc: '兑换成功后可在余额、用量等页面查看额度变化，方便控制成本。'
  },
  {
    icon: 'sparkles',
    iconClass: 'text-[#b15f00]',
    title: '模型以站内显示为准',
    desc: '支持 GPT-5.5 / GPT-5.4 等模型时，会以网站实际展示的渠道和价格为准。'
  },
  {
    icon: 'shield',
    iconClass: 'text-[#b42343]',
    title: '适合先试后用',
    desc: '按量计费，不需要一次投入很多，确认体验后再按需充值。'
  }
] as const

const clients = [
  {
    icon: 'chatBubble',
    iconBoxClass: 'bg-[#e8f8f4] text-[#0b806f]',
    name: 'ChatBox',
    desc: '填写 API Key 和接口地址即可测试对话'
  },
  {
    icon: 'terminal',
    iconBoxClass: 'bg-[#eaf0ff] text-[#395fcf]',
    name: 'NextChat',
    desc: '适合网页和移动端日常使用'
  },
  {
    icon: 'cpu',
    iconBoxClass: 'bg-[#fff4d8] text-[#9a6400]',
    name: '自己的程序',
    desc: '按 OpenAI 兼容接口方式接入'
  }
] as const

const tips = [
  {
    icon: 'lock',
    iconClass: 'text-[#8df5df]',
    title: '别泄露密钥',
    desc: '兑换码和 API 密钥请自己保存，不要发给陌生人。'
  },
  {
    icon: 'dollar',
    iconClass: 'text-[#ffd166]',
    title: '留意扣费',
    desc: '不同模型价格不同，实际扣费以网站页面为准。'
  },
  {
    icon: 'chat',
    iconClass: 'text-[#9bb7ff]',
    title: '不会就问',
    desc: '注册、兑换、客户端接入遇到问题，可以直接私信咨询。'
  }
] as const

async function copyRegisterUrl() {
  try {
    await navigator.clipboard.writeText(registerUrl)
    copied.value = true
    window.setTimeout(() => {
      copied.value = false
    }, 2200)
  } catch {
    window.prompt('复制注册网址', registerUrl)
  }
}

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.guide-grid {
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.06) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.06) 1px, transparent 1px),
    linear-gradient(135deg, rgba(33, 214, 182, 0.1), transparent 42%, rgba(255, 209, 102, 0.08));
  background-size:
    36px 36px,
    36px 36px,
    100% 100%;
}

.api-guide {
  letter-spacing: 0;
}
</style>
