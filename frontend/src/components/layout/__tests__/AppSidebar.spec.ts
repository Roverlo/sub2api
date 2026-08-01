import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const stylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../style.css')
const styleSource = readFileSync(stylePath, 'utf8')
const sourceRoot = resolve(dirname(fileURLToPath(import.meta.url)), '../../..')
const routerSource = readFileSync(resolve(sourceRoot, 'router/index.ts'), 'utf8')
const appSource = readFileSync(resolve(sourceRoot, 'App.vue'), 'utf8')
const headerSource = readFileSync(resolve(sourceRoot, 'components/layout/AppHeader.vue'), 'utf8')
const prefetchSource = readFileSync(resolve(sourceRoot, 'composables/useRoutePrefetch.ts'), 'utf8')

describe('redeem-only customer flow', () => {
  it('keeps redeem entry points and redirects retired subscription pages', () => {
    expect(componentSource).toContain("{ path: '/redeem'")
    expect(componentSource).toContain("{ path: '/admin/redeem'")
    expect(componentSource).not.toContain("{ path: '/subscriptions', label:")
    expect(componentSource).not.toContain("{ path: '/admin/subscriptions', label:")
    expect(routerSource).toMatch(/path: '\/subscriptions',[\s\S]*?redirect: '\/redeem'/)
    expect(routerSource).toMatch(/path: '\/admin\/subscriptions',[\s\S]*?redirect: '\/admin\/redeem'/)
    expect(appSource).not.toContain('subscriptionStore.fetchActiveSubscriptions')
    expect(appSource).not.toContain('subscriptionStore.startPolling')
    expect(appSource).toContain('subscriptionStore.clear()')
    expect(headerSource).not.toContain('SubscriptionProgressMini')
    expect(prefetchSource).not.toContain("'/admin/subscriptions':")
    expect(prefetchSource).toContain("'/admin/groups': ['/admin/redeem', '/admin/users']")
  })
})

describe('AppSidebar custom SVG styles', () => {
  it('does not override uploaded SVG fill or stroke colors', () => {
    expect(componentSource).toContain('.sidebar-svg-icon {')
    expect(componentSource).toContain('color: currentColor;')
    expect(componentSource).toContain('display: block;')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar scroll position persistence', () => {
  it('binds a template ref to the sidebar nav element', () => {
    expect(componentSource).toContain('ref="sidebarNavRef"')
    expect(componentSource).toContain('sidebar-nav')
  })

  it('declares sidebarNavRef in script setup', () => {
    expect(componentSource).toContain("const sidebarNavRef = ref<HTMLElement | null>(null)")
  })

  it('saves scroll position on beforeUnmount', () => {
    expect(componentSource).toContain('onBeforeUnmount')
    expect(componentSource).toContain('appStore.sidebarScrollTop')
    expect(componentSource).toContain('sidebarNavRef.value.scrollTop')
  })

  it('restores scroll position on mount', () => {
    expect(componentSource).toContain('onMounted')
    expect(componentSource).toContain('appStore.sidebarScrollTop')
    expect(componentSource).toContain('nextTick')
  })
})

describe('AppSidebar header styles', () => {
  it('does not clip the version badge dropdown', () => {
    const sidebarHeaderBlockMatch = styleSource.match(/\.sidebar-header\s*\{[\s\S]*?\n {2}\}/)
    const sidebarBrandBlockMatch = componentSource.match(/\.sidebar-brand\s*\{[\s\S]*?\n\}/)

    expect(sidebarHeaderBlockMatch).not.toBeNull()
    expect(sidebarBrandBlockMatch).not.toBeNull()
    expect(sidebarHeaderBlockMatch?.[0]).not.toContain('@apply overflow-hidden;')
    expect(sidebarBrandBlockMatch?.[0]).not.toContain('overflow: hidden;')
  })
})
