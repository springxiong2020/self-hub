<template>
  <el-date-picker
    v-if="!lunar"
    :model-value="modelValue"
    @update:model-value="emit('update:modelValue', $event || null)"
    type="date"
    value-format="YYYY-MM-DD"
    placeholder="生日选填"
    :disabled-date="(date) => date.getTime() > Date.now()"
    clearable
  />
  <div v-else class="lunar-input">
    <el-select
      v-model="year"
      filterable
      placeholder="农历年"
      aria-label="农历出生年"
      @change="update"
    >
      <el-option
        v-for="value in years"
        :key="value"
        :label="`${value}年`"
        :value="value"
      />
    </el-select>
    <el-select
      v-model="month"
      placeholder="月"
      aria-label="农历出生月"
      @change="update"
    >
      <el-option
        v-for="value in 12"
        :key="value"
        :label="`${value}月`"
        :value="value"
      />
    </el-select>
    <el-select
      v-model="day"
      placeholder="日"
      aria-label="农历出生日"
      @change="update"
    >
      <el-option
        v-for="value in 30"
        :key="value"
        :label="`${value}日`"
        :value="value"
      />
    </el-select>
    <el-button link @click="clear">清空</el-button>
  </div>
</template>

<script setup>
  import { ref, watch } from 'vue'

  const props = defineProps({
    modelValue: { type: String, default: null },
    lunar: Boolean
  })
  const emit = defineEmits(['update:modelValue'])
  const year = ref()
  const month = ref()
  const day = ref()
  const years = Array.from(
    { length: new Date().getFullYear() - 1899 },
    (_, index) => new Date().getFullYear() - index
  )

  watch(
    () => props.modelValue,
    (value) => {
      const parts = value?.split('-').map(Number) || []
      ;[year.value, month.value, day.value] = parts
    },
    { immediate: true }
  )

  const clear = () => {
    year.value = month.value = day.value = undefined
    emit('update:modelValue', null)
  }
  const update = () => {
    if (year.value && month.value && day.value) {
      emit(
        'update:modelValue',
        `${year.value}-${String(month.value).padStart(2, '0')}-${String(day.value).padStart(2, '0')}`
      )
    }
  }
  defineExpose({
    clear,
    isIncomplete: () =>
      props.lunar &&
      !!(year.value || month.value || day.value) &&
      !(year.value && month.value && day.value)
  })
</script>

<style scoped>
  .lunar-input {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    width: 100%;
  }
  .lunar-input .el-select:first-child {
    width: 110px;
  }
  .lunar-input .el-select {
    width: 80px;
  }
</style>
