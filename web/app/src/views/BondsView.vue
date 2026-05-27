<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useLocaleStore } from '@/stores/locale'
import { api } from '@/api/client'
import { ElMessage } from 'element-plus'
import ConcordBadge from '@/components/bonds/ConcordBadge.vue'
import BondCard from '@/components/bonds/BondCard.vue'
import MatchLinkManager from '@/components/bonds/MatchLinkManager.vue'
import type { BondItem, MatchLink } from '@shared/types'

const auth = useAuthStore()
const { t } = useLocaleStore()

const bonds = ref<BondItem[]>([])
const links = ref<MatchLink[]>([])
const loading = ref(false)
const linksLoading = ref(false)
const error = ref('')
const expandedBond = ref<BondItem | null>(null)
const showInstantCompare = ref(false)
const compareUsername = ref('')
const comparing = ref(false)

async function loadBonds() {
  if (!auth.isLoggedIn) return
  loading.value = true
  try {
    const resp = await api<{ data: { items: BondItem[]; total: number } }>(`/profiles/${auth.userName}/bonds`)
    bonds.value = resp.data?.items || []
  } catch { bonds.value = [] }
  finally { loading.value = false }
}

async function loadLinks() {
  linksLoading.value = true
  try {
    const resp = await api<{ data: { items: MatchLink[]; total: number } }>('/match-links')
    links.value = resp.data?.items || []
  } catch { links.value = [] }
  finally { linksLoading.value = false }
}

onMounted(async () => {
  if (!auth.isLoggedIn) return
  await Promise.all([loadBonds(), loadLinks()])
})

async function createLink() {
  try {
    const resp = await api<{ data: MatchLink }>('/match-links', { method: 'POST', body: { type: 'assessment' } })
    links.value.unshift(resp.data)
    const url = `${window.location.origin}/m/${resp.data.token}`
    await navigator.clipboard?.writeText(url)
    ElMessage.success(t('Link created and copied', '链接已创建并复制'))
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to create link')
  }
}

async function deleteLink(id: number) {
  try {
    await api(`/match-links/${id}`, { method: 'DELETE' })
    links.value = links.value.filter(l => l.id !== id)
    ElMessage.success(t('Link deleted', '链接已删除'))
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : 'Failed to delete link')
  }
}

async function instantCompare() {
  if (!compareUsername.value.trim()) return
  comparing.value = true
  try {
    const resp = await api<{ data: BondItem['bond'] & { other_user: BondItem['other_user'] } }>('/bond', {
      method: 'POST',
      body: { with_name: compareUsername.value.trim() },
    })
    bonds.value.unshift({
      other_user: resp.data.other_user,
      bond: {
        self: resp.data.self || {},
        other: resp.data.other || {},
        delta_a: resp.data.delta_a || {},
        delta_b: resp.data.delta_b || {},
        concord: resp.data.concord || '',
      },
      source: 'instant',
      created_at: new Date().toISOString(),
    })
    showInstantCompare.value = false
    compareUsername.value = ''
    ElMessage.success(t('Bond computed', '关系已计算'))
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : 'Compare failed')
  } finally {
    comparing.value = false
  }
}

function toggleExpand(bond: BondItem) {
  expandedBond.value = expandedBond.value === bond ? null : bond
}
</script>

<template>
  <div class="bonds-page stagger-in">
    <div class="page-header">
      <div class="flex items-center justify-between">
        <div>
          <h1>{{ t('Bonds', '关系图谱') }}</h1>
          <p>{{ t('Relationship dynamics computed by the 25types Bond engine.', '由25types Bond引擎计算的关系动力学，双向视角。') }}</p>
        </div>
        <div class="flex gap-2">
          <el-button v-if="auth.isLoggedIn" round @click="showInstantCompare = true">
            {{ t('Instant Compare', '即时对比') }}
          </el-button>
          <el-button v-if="auth.isLoggedIn" type="primary" round @click="createLink">
            + {{ t('Create Link', '创建链接') }}
          </el-button>
        </div>
      </div>
    </div>

    <el-result
      v-if="!auth.isLoggedIn"
      icon="warning"
      :title="t('Please login', '请先登录')"
      :sub-title="t('Login to view your bonds.', '登录后查看你的关系分析。')"
    >
      <template #extra>
        <a href="/login"><el-button type="primary" size="large">{{ t('Login', '登录') }}</el-button></a>
      </template>
    </el-result>

    <el-skeleton v-else-if="loading" :rows="8" animated />

    <el-empty
      v-else-if="bonds.length === 0"
      description=""
    >
      <template #image>
        <div class="premium-empty">
          <div class="icon-wrap">🔗</div>
          <div class="text-sm text-gray-500 mt-2">{{ t('No bonds yet. Share a match link or compare instantly.', '暂无关系数据。分享匹配链接或即时对比。') }}</div>
        </div>
      </template>
      <template #default>
        <div class="flex gap-2">
          <el-button round @click="showInstantCompare = true">{{ t('Instant Compare', '即时对比') }}</el-button>
          <el-button type="primary" round @click="createLink">{{ t('Create Match Link', '创建匹配链接') }}</el-button>
        </div>
      </template>
    </el-empty>

    <div v-else>
      <!-- Summary stats -->
      <el-row :gutter="16" class="mb-4">
        <el-col :xs="8" class="mb-2">
          <div class="bond-stat">
            <div class="count" style="color:var(--wuxing-water)">{{ bonds.length }}</div>
            <div class="label">{{ t('Total Bonds', '全部关系') }}</div>
          </div>
        </el-col>
        <el-col :xs="8" class="mb-2">
          <div class="bond-stat">
            <div class="count" style="color:var(--wuxing-wood)">
              {{ bonds.filter(b => b.bond.concord === '顺' || b.bond.concord === 'harmonious').length }}
            </div>
            <div class="label">{{ t('Harmonious', '顺') }}</div>
          </div>
        </el-col>
        <el-col :xs="8" class="mb-2">
          <div class="bond-stat">
            <div class="count" style="color:var(--wuxing-metal)">{{ links.length }}</div>
            <div class="label">{{ t('Match Links', '匹配链接') }}</div>
          </div>
        </el-col>
      </el-row>

      <!-- Table -->
      <el-table :data="bonds" style="width:100%" @row-click="toggleExpand">
        <el-table-column :label="t('Partner', '对方')">
          <template #default="{ row }">
            <span class="font-medium">{{ row.other_user?.name || t('Anonymous', '匿名') }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('Date', '日期')">
          <template #default="{ row }">
            {{ new Date(row.created_at).toLocaleDateString() }}
          </template>
        </el-table-column>
        <el-table-column :label="t('Source', '来源')">
          <template #default="{ row }">
            <el-tag size="small" effect="plain" round :type="row.source === 'instant' ? '' : 'info'">
              {{ row.source === 'instant' ? t('Instant', '即时') : t('Match Link', '匹配链接') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('Concord', '关系')">
          <template #default="{ row }">
            <ConcordBadge :concord="row.bond.concord" />
          </template>
        </el-table-column>
      </el-table>

      <!-- Expanded detail -->
      <div v-if="expandedBond" class="mt-4">
        <el-card shadow="never" class="premium-card">
          <BondCard
            :self="expandedBond.bond.self"
            :other="expandedBond.bond.other"
            :delta-a="expandedBond.bond.delta_a"
            :delta-b="expandedBond.bond.delta_b"
            :concord="expandedBond.bond.concord"
          />
        </el-card>
      </div>

      <!-- Match links section -->
      <div class="mt-6">
        <MatchLinkManager
          :links="links"
          :loading="linksLoading"
          @create="createLink"
          @delete="deleteLink"
        />
      </div>
    </div>

    <!-- Instant compare dialog -->
    <el-dialog
      v-model="showInstantCompare"
      :title="t('Instant Compare', '即时对比')"
      width="400px"
      :close-on-click-modal="false"
    >
      <el-form @submit.prevent="instantCompare">
        <el-form-item :label="t('Username', '用户名')">
          <el-input
            v-model="compareUsername"
            :placeholder="t('Enter username to compare', '输入要对比的用户名')"
            @keyup.enter="instantCompare"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showInstantCompare = false">{{ t('Cancel', '取消') }}</el-button>
        <el-button type="primary" :loading="comparing" :disabled="!compareUsername.trim()" @click="instantCompare">
          {{ t('Compare', '对比') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>
