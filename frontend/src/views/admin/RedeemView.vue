<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="space-y-3">
          <section
            class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800"
          >
            <div class="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
              <div
                class="grid flex-1 grid-cols-1 gap-3 md:grid-cols-[minmax(180px,220px)_minmax(260px,360px)]"
              >
                <Select
                  v-model="filters.status"
                  :options="filterStatusOptions"
                  @change="handleFilterChange"
                />
                <div class="grid grid-cols-[1fr_auto_1fr] items-center gap-2">
                  <input
                    v-model="filters.value_min"
                    type="number"
                    min="0"
                    step="0.01"
                    :aria-label="t('admin.redeem.valueMin')"
                    :placeholder="t('admin.redeem.valueMin')"
                    class="input"
                    @input="handleValueFilterInput"
                  />
                  <span class="text-sm text-gray-400">-</span>
                  <input
                    v-model="filters.value_max"
                    type="number"
                    min="0"
                    step="0.01"
                    :aria-label="t('admin.redeem.valueMax')"
                    :placeholder="t('admin.redeem.valueMax')"
                    class="input"
                    @input="handleValueFilterInput"
                  />
                </div>
              </div>

              <div class="flex shrink-0 flex-wrap items-center gap-2 xl:justify-end">
                <button
                  v-if="hasActiveFilters"
                  type="button"
                  class="btn btn-secondary btn-sm"
                  @click="resetFilters"
                >
                  <Icon name="x" size="sm" />
                  {{ t('common.reset') }}
                </button>
                <button type="button" class="btn btn-primary" @click="showGenerateDialog = true">
                  <Icon name="plus" size="sm" />
                  {{ t('admin.redeem.generateCodes') }}
                </button>
              </div>
            </div>
          </section>

          <section
            class="rounded-lg border border-gray-200 bg-white px-4 py-3 shadow-sm dark:border-dark-700 dark:bg-dark-800"
          >
            <div class="flex flex-col gap-3 xl:flex-row xl:items-center">
              <div
                class="grid min-w-0 flex-1 grid-cols-[repeat(auto-fit,minmax(120px,1fr))] gap-2"
              >
                <button
                  v-for="bucket in visibleValueBuckets"
                  :key="bucket.key"
                  type="button"
                  data-test="value-bucket-filter"
                  :aria-pressed="isValueBucketSelected(bucket.key)"
                  :class="[
                    'flex min-w-0 items-center justify-between gap-3 rounded-lg border px-3 py-2 text-left transition-colors',
                    isValueBucketSelected(bucket.key)
                      ? 'border-green-500 bg-green-50 text-green-800 shadow-sm dark:border-green-400 dark:bg-green-900/20 dark:text-green-200'
                      : 'border-gray-200 bg-gray-50 text-gray-700 hover:border-green-200 hover:bg-white hover:text-green-700 dark:border-dark-700 dark:bg-dark-900/60 dark:text-gray-300 dark:hover:border-green-700 dark:hover:bg-dark-800'
                  ]"
                  @click="toggleValueBucketFilter(bucket.key)"
                >
                  <span class="min-w-0">
                    <span class="block text-sm font-semibold leading-5">
                      {{ formatRedeemValue(bucket.value, bucket.type) }}
                    </span>
                    <span class="block truncate text-xs text-gray-500 dark:text-dark-400">
                      {{ t('admin.redeem.types.' + bucket.type) }}
                    </span>
                  </span>
                  <span
                    :class="[
                      'rounded-md px-2 py-1 text-xs font-semibold shadow-sm',
                      isValueBucketSelected(bucket.key)
                        ? 'bg-white text-green-700 dark:bg-dark-800 dark:text-green-200'
                        : 'bg-white text-gray-600 dark:bg-dark-800 dark:text-gray-300'
                    ]"
                  >
                    {{ bucket.count }}
                  </span>
                </button>
                <span
                  v-if="visibleValueBuckets.length === 0"
                  class="inline-flex min-h-10 items-center text-sm text-gray-400"
                >
                  {{ loadingValueBuckets ? t('common.loading') : t('empty.noData') }}
                </span>
              </div>
              <button
                v-if="selectedValueBucketKeys.size > 0"
                type="button"
                class="inline-flex shrink-0 items-center justify-center gap-1.5 rounded-full border border-green-200 bg-green-50 px-3 py-1.5 text-sm font-medium text-green-700 transition-colors hover:border-green-300 hover:bg-white dark:border-green-800 dark:bg-green-900/20 dark:text-green-200 dark:hover:border-green-700 dark:hover:bg-dark-800"
                @click="resetValueBucketFilters"
              >
                <Icon name="x" size="sm" />
                {{ t('common.reset') }}
              </button>
            </div>
          </section>

          <section
            v-if="selectedCount > 0"
            class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-primary-200 bg-primary-50 px-4 py-3 dark:border-primary-800 dark:bg-primary-900/20"
          >
            <div class="flex flex-wrap items-center gap-2 text-sm text-primary-900 dark:text-primary-100">
              <span class="rounded-md bg-white px-2 py-1 font-semibold shadow-sm dark:bg-dark-800">
                {{ t('admin.redeem.selectedCount', { count: selectedCount }) }}
              </span>
              <span v-if="selectedValueSummary" class="text-primary-700 dark:text-primary-200">
                {{ selectedValueSummary }}
              </span>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <button
                type="button"
                data-test="copy-selected-codes"
                :class="[
                  'btn btn-sm',
                  copiedSelectedCodes ? 'btn-success' : 'btn-secondary'
                ]"
                @click="copySelectedCodes"
              >
                <Icon
                  :name="copiedSelectedCodes ? 'check' : 'copy'"
                  size="sm"
                  :stroke-width="2"
                />
                {{
                  copiedSelectedCodes
                    ? t('admin.redeem.copied')
                    : t('admin.redeem.copySelectedCodes')
                }}
              </button>
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                @click="clearSelectedCodes"
              >
                <Icon name="x" size="sm" />
                {{ t('admin.redeem.clearSelection') }}
              </button>
            </div>
          </section>
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="codes"
          :loading="loading"
          :server-side-sort="true"
          default-sort-key="id"
          default-sort-order="desc"
          @sort="handleSort"
        >
          <template #header-select>
            <input
              data-test="select-all-codes"
              type="checkbox"
              :aria-label="t('common.selectAll')"
              class="h-4 w-4 cursor-pointer rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              :checked="allVisibleSelected"
              @click.stop
              @change="toggleSelectAllVisible($event)"
            />
          </template>

          <template #cell-select="{ row }">
            <input
              data-test="select-code"
              type="checkbox"
              :aria-label="`${t('admin.redeem.columns.code')}: ${row.code}`"
              class="h-4 w-4 cursor-pointer rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              :checked="selectedCodeIds.has(row.id)"
              @click.stop
              @change="toggleSelectRow(row.id, $event)"
            />
          </template>

          <template #cell-code="{ value }">
            <div class="flex items-center space-x-2">
              <code class="font-mono text-sm text-gray-900 dark:text-gray-100">{{ value }}</code>
              <button
                type="button"
                @click="copyToClipboard(value)"
                :class="[
                  'flex items-center transition-colors',
                  copiedCode === value
                    ? 'text-green-500'
                    : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'
                ]"
                :title="copiedCode === value ? t('admin.redeem.copied') : t('keys.copyToClipboard')"
                :aria-label="
                  copiedCode === value ? t('admin.redeem.copied') : t('keys.copyToClipboard')
                "
              >
                <Icon v-if="copiedCode !== value" name="copy" size="sm" :stroke-width="2" />
                <Icon v-else name="check" size="sm" :stroke-width="2" />
              </button>
            </div>
          </template>

          <template #cell-type="{ value }">
            <span
              :class="[
                'badge',
                value === 'balance'
                  ? 'badge-success'
                  : value === 'subscription'
                    ? 'badge-warning'
                    : 'badge-primary'
              ]"
            >
              {{ t('admin.redeem.types.' + value) }}
            </span>
          </template>

          <template #cell-value="{ value, row }">
            <span class="text-sm font-medium text-gray-900 dark:text-white">
              <template v-if="row.type === 'balance'">${{ value.toFixed(2) }}</template>
              <template v-else-if="row.type === 'subscription'">
                {{ row.validity_days || 30 }} {{ t('admin.redeem.days') }}
                <span v-if="row.group" class="ml-1 text-xs text-gray-500 dark:text-gray-400"
                  >({{ row.group.name }})</span
                >
              </template>
              <template v-else>{{ value }}</template>
            </span>
          </template>

          <template #cell-status="{ value }">
            <span
              :class="[
                'badge',
                value === 'unused'
                  ? 'badge-success'
                  : value === 'used'
                    ? 'badge-gray'
                    : 'badge-danger'
              ]"
            >
              {{ t('admin.redeem.status.' + value) }}
            </span>
          </template>

          <template #cell-used_at="{ row }">
            <div class="min-w-40">
              <p class="truncate text-sm text-gray-700 dark:text-gray-200">
                {{
                  row.user?.email ||
                  (row.used_by ? t('admin.redeem.userPrefix', { id: row.used_by }) : '-')
                }}
              </p>
              <p class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">
                {{ row.used_at ? formatDateTime(row.used_at) : '-' }}
              </p>
            </div>
          </template>

          <template #cell-expires_at="{ value, row }">
            <span
              :class="[
                'text-sm',
                row.status === 'expired'
                  ? 'text-red-600 dark:text-red-400'
                  : 'text-gray-500 dark:text-dark-400'
              ]"
            >
              {{ value ? formatDateTime(value) : t('admin.redeem.neverExpires') }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center justify-end">
              <button
                v-if="row.status === 'unused'"
                type="button"
                @click="handleDelete(row)"
                class="inline-flex h-9 w-9 items-center justify-center rounded-lg text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 focus:outline-none focus:ring-2 focus:ring-red-500/40 dark:hover:bg-red-900/20 dark:hover:text-red-400"
                :title="t('common.delete')"
                :aria-label="t('common.delete')"
              >
                <Icon name="trash" size="sm" :stroke-width="2" />
              </button>
              <span v-else class="text-gray-400 dark:text-dark-500">-</span>
            </div>
          </template>

          <template #empty>
            <div class="flex flex-col items-center">
              <div
                class="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-lg bg-primary-50 text-primary-600 dark:bg-primary-900/20 dark:text-primary-300"
              >
                <Icon name="gift" size="lg" :stroke-width="1.75" />
              </div>
              <p class="text-lg font-semibold text-gray-900 dark:text-gray-100">
                {{ t('admin.redeem.noCodes') }}
              </p>
              <p class="mt-1 max-w-md text-sm text-gray-500 dark:text-dark-400">
                {{ t('admin.redeem.noCodesDescription') }}
              </p>
              <button type="button" class="btn btn-primary btn-sm mt-4" @click="showGenerateDialog = true">
                <Icon name="plus" size="sm" />
                {{ t('admin.redeem.generateCodes') }}
              </button>
            </div>
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <div
          class="flex flex-col gap-3 rounded-lg border border-gray-200 bg-white px-4 py-3 shadow-sm dark:border-dark-700 dark:bg-dark-800 lg:flex-row lg:items-center lg:justify-between"
        >
          <Pagination
            v-if="pagination.total > 0"
            :page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.page_size"
            @update:page="handlePageChange"
            @update:pageSize="handlePageSizeChange"
          />
          <div v-else class="text-sm text-gray-500 dark:text-dark-400">
            {{ loading ? t('common.loading') : t('empty.noData') }}
          </div>

          <button
            v-if="filters.status === 'unused'"
            type="button"
            class="btn btn-danger btn-sm self-start lg:self-auto"
            @click="showDeleteUnusedDialog = true"
          >
            <Icon name="trash" size="sm" />
            {{ t('admin.redeem.deleteAllUnused') }}
          </button>
        </div>
      </template>
    </TablePageLayout>

    <!-- Delete Confirmation Dialog -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.redeem.deleteCode')"
      :message="t('admin.redeem.deleteCodeConfirm')"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />

    <!-- Delete Unused Codes Dialog -->
    <ConfirmDialog
      :show="showDeleteUnusedDialog"
      :title="t('admin.redeem.deleteAllUnused')"
      :message="t('admin.redeem.deleteAllUnusedConfirm')"
      :confirm-text="t('admin.redeem.deleteAll')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmDeleteUnused"
      @cancel="showDeleteUnusedDialog = false"
    />

    <!-- Generate Codes Dialog -->
    <Teleport to="body">
      <div v-if="showGenerateDialog" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="fixed inset-0 bg-black/50" @click="showGenerateDialog = false"></div>
        <div
          class="relative z-10 w-full max-w-md rounded-xl bg-white p-6 shadow-xl dark:bg-dark-800"
        >
          <h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('admin.redeem.generateCodesTitle') }}
          </h2>
          <form @submit.prevent="handleGenerateCodes" class="space-y-4">
            <div>
              <label class="input-label">{{ t('admin.redeem.codeType') }}</label>
              <Select v-model="generateForm.type" :options="typeOptions" />
            </div>
            <!-- 余额/并发类型：显示数值输入 -->
            <div v-if="generateForm.type !== 'subscription' && generateForm.type !== 'invitation'">
              <label class="input-label">
                {{
                  generateForm.type === 'balance'
                    ? t('admin.redeem.amount')
                    : t('admin.redeem.columns.value')
                }}
              </label>
              <input
                v-model.number="generateForm.value"
                type="number"
                :step="generateForm.type === 'balance' ? '0.01' : '1'"
                :min="generateForm.type === 'balance' ? '0.01' : '1'"
                required
                class="input"
              />
              <div class="mt-2 flex flex-wrap gap-2">
                <button
                  v-for="preset in currentValuePresets"
                  :key="preset"
                  type="button"
                  :class="[
                    'rounded-md border px-2.5 py-1.5 text-xs font-medium transition-colors',
                    Number(generateForm.value) === preset
                      ? 'border-primary-500 bg-primary-50 text-primary-700 dark:border-primary-400 dark:bg-primary-900/20 dark:text-primary-300'
                      : 'border-gray-200 text-gray-600 hover:border-primary-200 hover:text-primary-700 dark:border-dark-600 dark:text-gray-300'
                  ]"
                  @click="generateForm.value = preset"
                >
                  {{ formatRedeemValue(preset, generateForm.type) }}
                </button>
              </div>
            </div>
            <!-- 邀请码类型：显示提示信息 -->
            <div v-if="generateForm.type === 'invitation'" class="rounded-lg bg-blue-50 p-3 dark:bg-blue-900/20">
              <p class="text-sm text-blue-700 dark:text-blue-300">
                {{ t('admin.redeem.invitationHint') }}
              </p>
            </div>
            <!-- 订阅类型：显示分组选择和有效天数 -->
            <template v-if="generateForm.type === 'subscription'">
              <div>
                <label class="input-label">{{ t('admin.redeem.selectGroup') }}</label>
                <Select
                  v-model="generateForm.group_id"
                  :options="subscriptionGroupOptions"
                  :placeholder="t('admin.redeem.selectGroupPlaceholder')"
                >
                  <template #selected="{ option }">
                    <GroupBadge
                      v-if="option"
                      :name="(option as unknown as GroupOption).label"
                      :platform="(option as unknown as GroupOption).platform"
                      :subscription-type="(option as unknown as GroupOption).subscriptionType"
                      :rate-multiplier="(option as unknown as GroupOption).rate"
                    />
                    <span v-else class="text-gray-400">{{
                      t('admin.redeem.selectGroupPlaceholder')
                    }}</span>
                  </template>
                  <template #option="{ option, selected }">
                    <GroupOptionItem
                      :name="(option as unknown as GroupOption).label"
                      :platform="(option as unknown as GroupOption).platform"
                      :subscription-type="(option as unknown as GroupOption).subscriptionType"
                      :rate-multiplier="(option as unknown as GroupOption).rate"
                      :description="(option as unknown as GroupOption).description"
                      :selected="selected"
                    />
                  </template>
                </Select>
              </div>
              <div>
                <label class="input-label">{{ t('admin.redeem.validityDays') }}</label>
                <input
                  v-model.number="generateForm.validity_days"
                  type="number"
                  min="1"
                  max="365"
                  required
                  class="input"
                />
              </div>
            </template>
            <div>
              <label class="input-label">{{ t('admin.redeem.codeExpiry') }}</label>
              <div class="grid grid-cols-2 gap-2 sm:grid-cols-5">
                <button
                  v-for="option in redeemCodeExpiryOptions"
                  :key="option.value"
                  type="button"
                  @click="generateForm.expiry_option = option.value"
                  :class="[
                    'rounded-lg border px-3 py-2 text-sm transition-colors',
                    generateForm.expiry_option === option.value
                      ? 'border-primary-500 bg-primary-50 text-primary-700 dark:border-primary-400 dark:bg-primary-900/20 dark:text-primary-300'
                      : 'border-gray-200 text-gray-700 hover:bg-gray-50 dark:border-dark-600 dark:text-gray-300 dark:hover:bg-dark-700'
                  ]"
                >
                  {{ option.label }}
                </button>
              </div>
              <input
                v-if="generateForm.expiry_option === 'custom'"
                v-model.number="generateForm.custom_expiry_days"
                type="number"
                min="1"
                max="3650"
                required
                class="input mt-2"
                :placeholder="t('admin.redeem.customExpiryDays')"
              />
            </div>
            <div>
              <label class="input-label">{{ t('admin.redeem.count') }}</label>
              <input
                v-model.number="generateForm.count"
                type="number"
                min="1"
                max="100"
                required
                class="input"
              />
              <div class="mt-2 flex flex-wrap gap-2">
                <button
                  v-for="preset in countPresets"
                  :key="preset"
                  type="button"
                  :class="[
                    'rounded-md border px-2.5 py-1.5 text-xs font-medium transition-colors',
                    Number(generateForm.count) === preset
                      ? 'border-primary-500 bg-primary-50 text-primary-700 dark:border-primary-400 dark:bg-primary-900/20 dark:text-primary-300'
                      : 'border-gray-200 text-gray-600 hover:border-primary-200 hover:text-primary-700 dark:border-dark-600 dark:text-gray-300'
                  ]"
                  @click="generateForm.count = preset"
                >
                  {{ preset }}
                </button>
              </div>
              <p
                v-if="generateForm.type === 'balance'"
                class="mt-2 text-xs text-gray-500 dark:text-gray-400"
              >
                {{ t('admin.redeem.generateTotal', { total: formatCurrency(generateForm.value * generateForm.count) }) }}
              </p>
            </div>
            <div class="flex justify-end gap-3 pt-2">
              <button type="button" @click="showGenerateDialog = false" class="btn btn-secondary">
                {{ t('common.cancel') }}
              </button>
              <button type="submit" :disabled="generating" class="btn btn-primary">
                {{ generating ? t('admin.redeem.generating') : t('admin.redeem.generate') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- Generated Codes Result Dialog -->
    <Teleport to="body">
      <div v-if="showResultDialog" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="closeResultDialog"></div>
        <div class="relative z-10 w-full max-w-lg rounded-xl bg-white shadow-xl dark:bg-dark-800">
          <!-- Header -->
          <div
            class="flex items-center justify-between border-b border-gray-200 px-5 py-4 dark:border-dark-600"
          >
            <div class="flex items-center gap-3">
              <div
                class="flex h-10 w-10 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/30"
              >
                <svg
                  class="h-5 w-5 text-green-600 dark:text-green-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M5 13l4 4L19 7"
                  />
                </svg>
              </div>
              <div>
                <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                  {{ t('admin.redeem.generatedSuccessfully') }}
                </h2>
                <p class="text-sm text-gray-500 dark:text-gray-400">
                  {{ t('admin.redeem.codesCreated', { count: generatedCodes.length }) }}
                </p>
              </div>
            </div>
            <button
              @click="closeResultDialog"
              class="rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-gray-300"
            >
              <Icon name="x" size="md" :stroke-width="2" />
            </button>
          </div>
          <!-- Content -->
          <div class="p-5">
            <div class="relative">
              <textarea
                readonly
                :value="generatedCodesText"
                :style="{ height: textareaHeight }"
                class="w-full resize-none rounded-lg border border-gray-200 bg-gray-50 p-3 font-mono text-sm text-gray-800 focus:outline-none dark:border-dark-600 dark:bg-dark-700 dark:text-gray-200"
              ></textarea>
            </div>
          </div>
          <!-- Footer -->
          <div
            class="flex justify-end gap-2 rounded-b-xl border-t border-gray-200 bg-gray-50 px-5 py-4 dark:border-dark-600 dark:bg-dark-700/50"
          >
            <button
              @click="copyGeneratedCodes"
              :class="[
                'btn flex items-center gap-2 transition-all',
                copiedAll ? 'btn-success' : 'btn-secondary'
              ]"
            >
              <Icon v-if="!copiedAll" name="copy" size="sm" :stroke-width="2" />
              <svg v-else class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M5 13l4 4L19 7"
                />
              </svg>
              {{ copiedAll ? t('admin.redeem.copied') : t('admin.redeem.copyAll') }}
            </button>
            <button @click="downloadGeneratedCodes" class="btn btn-primary flex items-center gap-2">
              <Icon name="download" size="sm" :stroke-width="2" />
              {{ t('admin.redeem.download') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import { useTableSelection } from '@/composables/useTableSelection'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { adminAPI } from '@/api/admin'
import { formatDateTime } from '@/utils/format'
import type {
  RedeemCode,
  RedeemCodeType,
  Group,
  GroupPlatform,
  SubscriptionType
} from '@/types'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select from '@/components/common/Select.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import GroupOptionItem from '@/components/common/GroupOptionItem.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard: clipboardCopy } = useClipboard()

interface GroupOption {
  value: number
  label: string
  description: string | null
  platform: GroupPlatform
  subscriptionType: SubscriptionType
  rate: number
}

const showGenerateDialog = ref(false)
const showResultDialog = ref(false)
const generatedCodes = ref<RedeemCode[]>([])
const subscriptionGroups = ref<Group[]>([])

// 订阅类型分组选项
const subscriptionGroupOptions = computed(() => {
  return subscriptionGroups.value
    .filter((g) => g.subscription_type === 'subscription')
    .map((g) => ({
      value: g.id,
      label: g.name,
      description: g.description,
      platform: g.platform,
      subscriptionType: g.subscription_type,
      rate: g.rate_multiplier
    }))
})

const generatedCodesText = computed(() => {
  return generatedCodes.value.map((code) => code.code).join('\n')
})

const textareaHeight = computed(() => {
  const lineCount = generatedCodes.value.length
  const lineHeight = 24 // approximate line height in px
  const padding = 24 // top + bottom padding
  const minHeight = 60
  const maxHeight = 240
  const calculatedHeight = Math.min(
    Math.max(lineCount * lineHeight + padding, minHeight),
    maxHeight
  )
  return `${calculatedHeight}px`
})

const copiedAll = ref(false)

const closeResultDialog = () => {
  showResultDialog.value = false
  generatedCodes.value = []
  copiedAll.value = false
}

const copyGeneratedCodes = async () => {
  const success = await clipboardCopy(generatedCodesText.value, t('admin.redeem.copied'))
  if (success) {
    copiedAll.value = true
    setTimeout(() => {
      copiedAll.value = false
    }, 2000)
  }
}

const downloadGeneratedCodes = () => {
  const blob = new Blob([generatedCodesText.value], { type: 'text/plain' })
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `redeem-codes-${new Date().toISOString().split('T')[0]}.txt`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

const columns = computed<Column[]>(() => [
  { key: 'select', label: '', class: 'w-12' },
  { key: 'code', label: t('admin.redeem.columns.code'), class: 'min-w-[260px]' },
  { key: 'type', label: t('admin.redeem.columns.type'), sortable: true },
  { key: 'value', label: t('admin.redeem.columns.value'), sortable: true },
  { key: 'status', label: t('admin.redeem.columns.status'), sortable: true },
  {
    key: 'used_at',
    label: `${t('admin.redeem.columns.usedBy')} / ${t('admin.redeem.columns.usedAt')}`,
    sortable: true,
    class: 'min-w-[220px]'
  },
  {
    key: 'expires_at',
    label: t('admin.redeem.columns.expiresAt'),
    sortable: true,
    class: 'min-w-[180px]'
  },
  { key: 'actions', label: t('admin.redeem.columns.actions'), class: 'w-16' }
])

const typeOptions = computed(() => [
  { value: 'balance', label: t('admin.redeem.balance') },
  { value: 'concurrency', label: t('admin.redeem.concurrency') },
  { value: 'subscription', label: t('admin.redeem.subscription') },
  { value: 'invitation', label: t('admin.redeem.invitation') }
])

const filterStatusOptions = computed(() => [
  { value: '', label: t('admin.redeem.allStatus') },
  { value: 'unused', label: t('admin.redeem.unused') },
  { value: 'used', label: t('admin.redeem.used') },
  { value: 'expired', label: t('admin.redeem.status.expired') },
  { value: 'disabled', label: t('admin.redeem.status.disabled') }
])

const balanceValuePresets = [5, 10, 20, 50, 100, 200]
const concurrencyValuePresets = [1, 2, 5, 10, 20]
const countPresets = [1, 5, 10, 20, 50, 100]

const codes = ref<RedeemCode[]>([])
type ValueBucket = { key: string; value: number; type: RedeemCodeType; count: number }

const valueBuckets = ref<ValueBucket[]>([])
const loading = ref(false)
const generating = ref(false)
const loadingValueBuckets = ref(false)
const selectedValueBucketKeys = ref(new Set<string>())
const filters = reactive({
  type: '',
  status: '',
  value_min: '',
  value_max: ''
})
const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
  pages: 0
})
const sortState = reactive({
  sort_by: 'id',
  sort_order: 'desc' as 'asc' | 'desc'
})

let abortController: AbortController | null = null

const showDeleteDialog = ref(false)
const showDeleteUnusedDialog = ref(false)
const deletingCode = ref<RedeemCode | null>(null)
const copiedCode = ref<string | null>(null)
const copiedSelectedCodes = ref(false)

const {
  selectedSet: selectedCodeIds,
  selectedCount,
  allVisibleSelected,
  select,
  deselect,
  clear: clearSelectedCodes,
  toggleVisible
} = useTableSelection<RedeemCode>({
  rows: codes,
  getId: (code) => code.id
})

type RedeemCodeExpiryOption = 'never' | '1' | '3' | '7' | 'custom'

const redeemCodeExpiryOptions = computed<{ value: RedeemCodeExpiryOption; label: string }[]>(() => [
  { value: 'never', label: t('admin.redeem.neverExpires') },
  { value: '1', label: t('admin.redeem.expiryPresetDays', { days: 1 }) },
  { value: '3', label: t('admin.redeem.expiryPresetDays', { days: 3 }) },
  { value: '7', label: t('admin.redeem.expiryPresetDays', { days: 7 }) },
  { value: 'custom', label: t('admin.redeem.customExpiry') }
])

const generateForm = reactive({
  type: 'balance' as RedeemCodeType,
  value: 10,
  count: 1,
  group_id: null as number | null,
  validity_days: 30,
  expiry_option: 'never' as RedeemCodeExpiryOption,
  custom_expiry_days: 7
})

// 面值筛选和批量生成辅助函数
const parseOptionalNumber = (value: string) => {
  const trimmed = String(value ?? '').trim()
  if (!trimmed) return undefined
  const parsed = Number(trimmed)
  return Number.isFinite(parsed) ? parsed : undefined
}

const formatCurrency = (value: number) => `$${Number(value || 0).toFixed(2)}`

const formatRedeemValue = (value: number, type: RedeemCodeType) => {
  if (type === 'balance') return formatCurrency(value)
  if (type === 'subscription') return `${value || 0} ${t('admin.redeem.days')}`
  return String(value)
}

const getValueBucketKey = (value: number, type: RedeemCodeType) => `${type}:${value}`

const currentValuePresets = computed(() => {
  if (generateForm.type === 'balance') return balanceValuePresets
  if (generateForm.type === 'concurrency') return concurrencyValuePresets
  return []
})

const hasActiveFilters = computed(
  () =>
    Boolean(filters.type) ||
    Boolean(filters.status) ||
    Boolean(filters.value_min.trim()) ||
    Boolean(filters.value_max.trim()) ||
    selectedValueBucketKeys.value.size > 0
)

const visibleValueBuckets = computed(() => {
  return [...valueBuckets.value].sort((a, b) => b.count - a.count || a.value - b.value)
})

const selectedValueBucketQuery = computed(() =>
  Array.from(selectedValueBucketKeys.value).sort().join(',')
)

const selectedValueSummary = computed(() => {
  const selectedCodes = codes.value.filter((code) => selectedCodeIds.value.has(code.id))
  if (selectedCodes.length === 0) return ''

  const bucketMap = new Map<string, { value: number; type: RedeemCodeType; count: number }>()
  for (const code of selectedCodes) {
    if (code.type !== 'balance' && code.type !== 'concurrency') continue
    const key = `${code.type}:${code.value}`
    const existing = bucketMap.get(key)
    if (existing) {
      existing.count += 1
      continue
    }
    bucketMap.set(key, { value: code.value, type: code.type, count: 1 })
  }

  return Array.from(bucketMap.values())
    .sort((a, b) => b.count - a.count || a.value - b.value)
    .slice(0, 4)
    .map((bucket) => `${formatRedeemValue(bucket.value, bucket.type)} x ${bucket.count}`)
    .join(' / ')
})

const selectedCodesText = computed(() =>
  codes.value
    .filter((code) => selectedCodeIds.value.has(code.id))
    .map((code) => code.code)
    .join('\n')
)

const isValueBucketSelected = (key: string) => selectedValueBucketKeys.value.has(key)

const handleFilterChange = () => {
  pagination.page = 1
  loadCodes()
  loadValueBuckets()
}

const handleValueFilterInput = () => {
  filters.type = ''
  selectedValueBucketKeys.value = new Set()
  clearTimeout(valueFilterTimeout)
  valueFilterTimeout = setTimeout(() => {
    pagination.page = 1
    loadCodes()
    loadValueBuckets()
  }, 300)
}

const toggleValueBucketFilter = (key: string) => {
  const next = new Set(selectedValueBucketKeys.value)
  if (next.has(key)) {
    next.delete(key)
  } else {
    next.add(key)
  }
  selectedValueBucketKeys.value = next
  pagination.page = 1
  loadCodes()
}

const resetValueBucketFilters = () => {
  if (selectedValueBucketKeys.value.size === 0) return
  selectedValueBucketKeys.value = new Set()
  pagination.page = 1
  loadCodes()
}

const resetFilters = () => {
  filters.type = ''
  filters.status = ''
  filters.value_min = ''
  filters.value_max = ''
  selectedValueBucketKeys.value = new Set()
  pagination.page = 1
  loadCodes()
  loadValueBuckets()
}

// 监听类型变化，邀请码类型时自动设置 value 为 0
watch(
  () => generateForm.type,
  (newType) => {
    if (newType === 'invitation') {
      generateForm.value = 0
    } else if (newType === 'concurrency' && !concurrencyValuePresets.includes(generateForm.value)) {
      generateForm.value = concurrencyValuePresets[0]
    } else if (newType === 'balance' && generateForm.value === 0) {
      generateForm.value = balanceValuePresets[2]
    } else if (generateForm.value === 0) {
      generateForm.value = 10
    }
  }
)

const buildRedeemQueryFilters = () => {
  const valueMin = parseOptionalNumber(filters.value_min)
  const valueMax = parseOptionalNumber(filters.value_max)
  if (valueMin !== undefined && valueMax !== undefined && valueMin > valueMax) {
    appStore.showError(t('admin.redeem.valueRangeInvalid'))
    return null
  }

  return {
    type: (filters.type || undefined) as RedeemCodeType | undefined,
    status: (filters.status || undefined) as 'used' | 'expired' | 'unused' | 'disabled' | undefined,
    value_min: valueMin,
    value_max: valueMax,
    value_buckets: selectedValueBucketQuery.value || undefined,
    sort_by: sortState.sort_by,
    sort_order: sortState.sort_order
  }
}

const buildValueBucketQueryFilters = () => {
  const valueMin = parseOptionalNumber(filters.value_min)
  const valueMax = parseOptionalNumber(filters.value_max)
  if (valueMin !== undefined && valueMax !== undefined && valueMin > valueMax) {
    return null
  }

  return {
    type: (filters.type || undefined) as RedeemCodeType | undefined,
    status: (filters.status || undefined) as 'used' | 'expired' | 'unused' | 'disabled' | undefined,
    value_min: valueMin,
    value_max: valueMax,
    sort_by: 'value',
    sort_order: 'asc' as const
  }
}

const loadCodes = async () => {
  const queryFilters = buildRedeemQueryFilters()
  if (!queryFilters) return

  if (abortController) {
    abortController.abort()
  }
  const currentController = new AbortController()
  abortController = currentController
  loading.value = true
  try {
    const response = await adminAPI.redeem.list(
      pagination.page,
      pagination.page_size,
      queryFilters,
      {
        signal: currentController.signal
      }
    )
    if (currentController.signal.aborted) {
      return
    }
    codes.value = response.items
    pagination.total = response.total
    pagination.pages = response.pages
  } catch (error: any) {
    if (
      currentController.signal.aborted ||
      error?.name === 'AbortError' ||
      error?.code === 'ERR_CANCELED'
    ) {
      return
    }
    appStore.showError(t('admin.redeem.failedToLoad'))
    console.error('Error loading redeem codes:', error)
  } finally {
    if (abortController === currentController && !currentController.signal.aborted) {
      loading.value = false
      abortController = null
    }
  }
}

const loadValueBuckets = async () => {
  const queryFilters = buildValueBucketQueryFilters()
  if (!queryFilters) return

  loadingValueBuckets.value = true
  try {
    const response = await adminAPI.redeem.listValueBuckets(queryFilters)
    valueBuckets.value = response.map((bucket) => ({
      ...bucket,
      key: getValueBucketKey(bucket.value, bucket.type)
    }))
    const availableKeys = new Set(valueBuckets.value.map((bucket) => bucket.key))
    const nextSelected = new Set(
      Array.from(selectedValueBucketKeys.value).filter((key) => availableKeys.has(key))
    )
    if (nextSelected.size !== selectedValueBucketKeys.value.size) {
      selectedValueBucketKeys.value = nextSelected
    }
  } catch (error) {
    console.error('Error loading redeem value buckets:', error)
  } finally {
    loadingValueBuckets.value = false
  }
}

let valueFilterTimeout: ReturnType<typeof setTimeout>

const handlePageChange = (page: number) => {
  pagination.page = page
  loadCodes()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  loadCodes()
}

const handleSort = (key: string, order: 'asc' | 'desc') => {
  sortState.sort_by = key
  sortState.sort_order = order
  pagination.page = 1
  loadCodes()
}

const toggleSelectRow = (id: number, event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.checked) {
    select(id)
    return
  }
  deselect(id)
}

const toggleSelectAllVisible = (event: Event) => {
  const target = event.target as HTMLInputElement
  toggleVisible(target.checked)
}

const getRedeemCodeExpiresInDays = () => {
  if (generateForm.expiry_option === 'never') {
    return undefined
  }
  if (generateForm.expiry_option === 'custom') {
    if (
      !Number.isFinite(generateForm.custom_expiry_days) ||
      generateForm.custom_expiry_days < 1
    ) {
      return null
    }
    return Math.floor(generateForm.custom_expiry_days)
  }
  return Number(generateForm.expiry_option)
}

const handleGenerateCodes = async () => {
  // 订阅类型必须选择分组
  if (generateForm.type === 'subscription' && !generateForm.group_id) {
    appStore.showError(t('admin.redeem.groupRequired'))
    return
  }

  const expiresInDays = getRedeemCodeExpiresInDays()
  if (expiresInDays === null) {
    appStore.showError(t('admin.redeem.expiryDaysRequired'))
    return
  }

  generating.value = true
  try {
    const result = await adminAPI.redeem.generate(
      generateForm.count,
      generateForm.type,
      generateForm.value,
      generateForm.type === 'subscription' ? generateForm.group_id : undefined,
      generateForm.type === 'subscription' ? generateForm.validity_days : undefined,
      expiresInDays
    )
    showGenerateDialog.value = false
    generatedCodes.value = result
    showResultDialog.value = true
    // 重置表单
    generateForm.group_id = null
    generateForm.validity_days = 30
    generateForm.expiry_option = 'never'
    generateForm.custom_expiry_days = 7
    loadCodes()
    loadValueBuckets()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.redeem.failedToGenerate'))
    console.error('Error generating codes:', error)
  } finally {
    generating.value = false
  }
}

const copyToClipboard = async (text: string) => {
  const success = await clipboardCopy(text, t('admin.redeem.copied'))
  if (success) {
    copiedCode.value = text
    setTimeout(() => {
      copiedCode.value = null
    }, 2000)
  }
}

const copySelectedCodes = async () => {
  if (!selectedCodesText.value) {
    appStore.showError(t('admin.redeem.noSelectedCodesToCopy'))
    return
  }

  const success = await clipboardCopy(
    selectedCodesText.value,
    t('admin.redeem.selectedCodesCopied')
  )
  if (success) {
    copiedSelectedCodes.value = true
    setTimeout(() => {
      copiedSelectedCodes.value = false
    }, 2000)
  }
}

const handleDelete = (code: RedeemCode) => {
  deletingCode.value = code
  showDeleteDialog.value = true
}

const confirmDelete = async () => {
  if (!deletingCode.value) return

  try {
    await adminAPI.redeem.delete(deletingCode.value.id)
    appStore.showSuccess(t('admin.redeem.codeDeleted'))
    showDeleteDialog.value = false
    deletingCode.value = null
    loadCodes()
    loadValueBuckets()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.redeem.failedToDelete'))
    console.error('Error deleting code:', error)
  }
}

const confirmDeleteUnused = async () => {
  try {
    // Get all unused codes and delete them
    const unusedCodesResponse = await adminAPI.redeem.list(1, 1000, { status: 'unused' })
    const unusedCodeIds = unusedCodesResponse.items.map((code) => code.id)

    if (unusedCodeIds.length === 0) {
      appStore.showInfo(t('admin.redeem.noUnusedCodes'))
      showDeleteUnusedDialog.value = false
      return
    }

    const result = await adminAPI.redeem.batchDelete(unusedCodeIds)
    appStore.showSuccess(t('admin.redeem.codesDeleted', { count: result.deleted }))
    showDeleteUnusedDialog.value = false
    loadCodes()
    loadValueBuckets()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.redeem.failedToDeleteUnused'))
    console.error('Error deleting unused codes:', error)
  }
}

// 加载订阅类型分组
const loadSubscriptionGroups = async () => {
  try {
    const groups = await adminAPI.groups.getAll()
    subscriptionGroups.value = groups
  } catch (error) {
    console.error('Error loading subscription groups:', error)
  }
}

onMounted(() => {
  loadCodes()
  loadValueBuckets()
  loadSubscriptionGroups()
})

onUnmounted(() => {
  clearTimeout(valueFilterTimeout)
  abortController?.abort()
})
</script>
