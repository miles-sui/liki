<script setup lang="ts">
import { computed } from 'vue'
import type { BirthInfo } from '@shared/types'

const props = withDefaults(defineProps<{
  modelValue: Partial<BirthInfo>
  disabled?: boolean
  showGender?: boolean
  showDST?: boolean
}>(), {
  disabled: false,
  showGender: true,
  showDST: false,
})

const emit = defineEmits<{ (e: 'update:modelValue', v: Partial<BirthInfo>): void }>()

function set(key: keyof BirthInfo, val: unknown) {
  emit('update:modelValue', { ...props.modelValue, [key]: val })
}

const genderOptions = [
  { value: 'male', en: 'Male', zh: '男' },
  { value: 'female', en: 'Female', zh: '女' },
]
</script>

<template>
  <div class="birth-info-form">
    <el-row :gutter="12">
      <el-col :span="8">
        <el-form-item label="年">
          <el-input-number
            :model-value="modelValue.year"
            :min="1900" :max="2100" :disabled="disabled"
            controls-position="right" style="width:100%"
            @update:model-value="(v: number | undefined) => set('year', v)"
          />
        </el-form-item>
      </el-col>
      <el-col :span="4">
        <el-form-item label="月">
          <el-input-number
            :model-value="modelValue.month"
            :min="1" :max="12" :disabled="disabled"
            controls-position="right"
            @update:model-value="(v: number | undefined) => set('month', v)"
          />
        </el-form-item>
      </el-col>
      <el-col :span="4">
        <el-form-item label="日">
          <el-input-number
            :model-value="modelValue.day"
            :min="1" :max="31" :disabled="disabled"
            controls-position="right"
            @update:model-value="(v: number | undefined) => set('day', v)"
          />
        </el-form-item>
      </el-col>
      <el-col :span="4">
        <el-form-item label="时">
          <el-input-number
            :model-value="modelValue.hour"
            :min="0" :max="23" :disabled="disabled"
            controls-position="right"
            @update:model-value="(v: number | undefined) => set('hour', v)"
          />
        </el-form-item>
      </el-col>
      <el-col :span="4">
        <el-form-item label="分">
          <el-input-number
            :model-value="modelValue.minute"
            :min="0" :max="59" :disabled="disabled"
            controls-position="right"
            @update:model-value="(v: number | undefined) => set('minute', v)"
          />
        </el-form-item>
      </el-col>
    </el-row>
    <el-row :gutter="12">
      <el-col :span="6">
        <el-form-item label="经度">
          <el-input-number
            :model-value="modelValue.longitude"
            :min="-180" :max="180" :step="0.1" :disabled="disabled"
            controls-position="right" style="width:100%"
            @update:model-value="(v: number | undefined) => set('longitude', v)"
          />
        </el-form-item>
      </el-col>
      <el-col :span="6">
        <el-form-item label="时区(分)">
          <el-input-number
            :model-value="modelValue.timezone"
            :min="-720" :max="720" :step="30" :disabled="disabled"
            controls-position="right" style="width:100%"
            @update:model-value="(v: number | undefined) => set('timezone', v)"
          />
        </el-form-item>
      </el-col>
      <el-col v-if="showGender" :span="6">
        <el-form-item label="性别">
          <el-select
            :model-value="modelValue.gender"
            :disabled="disabled"
            @update:model-value="(v: string) => set('gender', v)"
          >
            <el-option
              v-for="g in genderOptions" :key="g.value"
              :label="g.en" :value="g.value"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col v-if="showDST" :span="6">
        <el-form-item label="夏令时">
          <el-switch
            :model-value="modelValue.is_dst"
            :disabled="disabled"
            @update:model-value="(v: boolean) => set('is_dst', v)"
          />
        </el-form-item>
      </el-col>
    </el-row>
  </div>
</template>
