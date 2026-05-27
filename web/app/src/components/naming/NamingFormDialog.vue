<script setup lang="ts">
import { ref } from 'vue'
import { useLocaleStore } from '@/stores/locale'
import BirthInfoForm from '@/components/shared/BirthInfoForm.vue'
import type { BirthInfo } from '@shared/types'

const props = defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void
  (e: 'submit', data: { surname: string; birth_info: Partial<BirthInfo> }): void
}>()

const { t } = useLocaleStore()

const surname = ref('')
const birthInfo = ref<Partial<BirthInfo>>({
  year: 2000, month: 1, day: 1, hour: 12, minute: 0,
  longitude: 120, timezone: 120, is_dst: false, gender: 'male',
})
const submitting = ref(false)

async function handleSubmit() {
  if (!surname.value.trim()) return
  submitting.value = true
  emit('submit', { surname: surname.value.trim(), birth_info: birthInfo.value })
}

function onClose() {
  emit('update:modelValue', false)
}
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    :title="t('New Naming', '新增起名')"
    width="560px"
    :close-on-click-modal="false"
    @update:model-value="onClose"
  >
    <el-form label-position="top">
      <el-form-item :label="t('Surname', '姓氏')" required>
        <el-input v-model="surname" maxlength="2" :placeholder="t('e.g. 李', '如：李')" style="width:120px" />
      </el-form-item>
      <el-divider />
      <p class="text-sm text-gray-500 mb-3">{{ t('Birth Info', '出生信息') }}</p>
      <BirthInfoForm v-model="birthInfo" />
    </el-form>
    <template #footer>
      <el-button @click="onClose">{{ t('Cancel', '取消') }}</el-button>
      <el-button type="primary" :loading="submitting" :disabled="!surname.trim()" @click="handleSubmit">
        {{ t('Analyze', '分析') }}
      </el-button>
    </template>
  </el-dialog>
</template>
