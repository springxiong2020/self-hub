<template>
  <section
    v-loading="loading"
    class="birthday-panel"
    aria-label="即将到来的好友生日"
  >
    <header>
      <div>
        <h3>即将到来的生日</h3>
        <span class="subtitle">提前准备，记得祝福</span>
      </div>
      <el-radio-group v-model="days" size="small" aria-label="生日提醒范围">
        <el-radio-button :value="7">未来 7 天</el-radio-button>
        <el-radio-button :value="30">未来 30 天</el-radio-button>
        <el-radio-button :value="90">未来 90 天</el-radio-button>
      </el-radio-group>
    </header>
    <div v-if="failed" class="birthday-error" role="status">
      生日提醒加载失败
      <el-button link type="primary" @click="refresh">重新加载</el-button>
    </div>
    <el-timeline v-else-if="groups.length" class="birthday-timeline">
      <el-timeline-item
        v-for="group in groups"
        :key="group.date"
        :timestamp="group.date"
        placement="top"
        color="#00ccc6"
        :hollow="group.daysUntil > 0"
      >
        <div class="birthday-group">
          <el-tag
            :type="
              group.daysUntil === 0
                ? 'danger'
                : group.daysUntil <= 7
                  ? 'warning'
                  : 'info'
            "
            size="small"
          >
            {{
              group.daysUntil === 0 ? '今天生日' : `还有 ${group.daysUntil} 天`
            }}
          </el-tag>
          <div class="birthday-people">
            <div
              v-for="person in group.people"
              :key="person.id"
              class="birthday-person"
              :class="
                person.milestoneLevel
                  ? `birthday-person--${person.milestoneLevel}`
                  : ''
              "
            >
              <div class="birthday-person-heading">
                <el-button
                  link
                  type="primary"
                  @click="emit('detail', person)"
                  >{{ person.name }}</el-button
                >
                <span v-if="Number.isInteger(person.age)" class="birthday-age"
                  >{{ person.age }} 岁生日</span
                >
                <span v-if="person.milestoneLabel" class="birthday-milestone">{{
                  person.milestoneLabel
                }}</span>
              </div>
              <span class="birthday-original">{{
                formatContactBirthday(person)
              }}</span>
            </div>
          </div>
        </div>
      </el-timeline-item>
    </el-timeline>
    <el-empty
      v-else
      :description="`未来 ${days} 天暂无已登记的好友生日（含今天）`"
      :image-size="58"
    />
    <p v-if="groups.length" class="birthday-note">
      按公历日期排序，包含今天；农历闰月无对应闰月时按普通月，三十遇小月取廿九；2
      月 29 日遇平年取 2 月 28 日。生日年龄按所选历法计算，不加虚岁。
    </p>
  </section>
</template>

<script setup>
  import { computed, ref, watch } from 'vue'
  import { getUpcomingContactBirthdays } from '@/api/ridge/ridge_contact'
  import { formatContactBirthday } from '@/utils/ridgeBirthday'

  const emit = defineEmits(['detail'])
  const days = ref(30)
  const loading = ref(false)
  const failed = ref(false)
  const birthdays = ref([])
  let requestId = 0
  const groups = computed(() => {
    const grouped = new Map()
    birthdays.value.forEach((person) => {
      if (!grouped.has(person.nextBirthday))
        grouped.set(person.nextBirthday, {
          date: person.nextBirthday,
          daysUntil: person.daysUntil,
          people: []
        })
      grouped.get(person.nextBirthday).people.push(person)
    })
    return [...grouped.values()]
  })
  const refresh = async () => {
    const current = ++requestId
    loading.value = true
    failed.value = false
    try {
      const res = await getUpcomingContactBirthdays({ days: days.value })
      if (current !== requestId) return
      if (res.code !== 0 || !Array.isArray(res.data))
        throw new Error('获取生日提醒失败')
      birthdays.value = res.data
    } catch {
      if (current === requestId) failed.value = true
    } finally {
      if (current === requestId) loading.value = false
    }
  }
  watch(days, refresh, { immediate: true })
  defineExpose({ refresh })
</script>

<style scoped>
  .birthday-panel {
    margin-bottom: 16px;
    padding: 18px 20px 12px;
    border-radius: 4px;
    background: var(--el-bg-color);
  }
  header {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 18px;
  }
  header > div {
    display: flex;
    align-items: baseline;
    flex-wrap: wrap;
    gap: 12px;
  }
  h3 {
    margin: 0;
    font-size: 15px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }
  .subtitle,
  .birthday-original,
  .birthday-note {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
  .birthday-timeline {
    max-height: 260px;
    overflow-y: auto;
    margin: 0;
    padding: 4px 12px 0 6px;
  }
  .birthday-group {
    display: flex;
    align-items: flex-start;
    flex-wrap: wrap;
    gap: 12px;
  }
  .birthday-people {
    display: flex;
    flex-wrap: wrap;
    gap: 8px 24px;
  }
  .birthday-person {
    --occasion-color: var(--el-text-color-regular);
    --occasion-background: transparent;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
    max-width: 100%;
    padding: 6px 10px;
    border-radius: 4px;
    background: var(--occasion-background);
  }
  .birthday-person-heading {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
  }
  .birthday-age {
    color: var(--el-text-color-regular);
    font-size: 13px;
  }
  .birthday-milestone {
    padding: 2px 7px;
    border: 1px solid currentColor;
    border-radius: 3px;
    color: var(--occasion-color);
    font-size: 12px;
    font-weight: 600;
    line-height: 1.5;
  }
  .birthday-person--adult {
    --occasion-color: #007772;
    --occasion-background: rgba(0, 204, 198, 0.09);
  }
  .birthday-person--anniversary {
    --occasion-color: var(--el-text-color-secondary);
  }
  .birthday-person--commemorative {
    --occasion-color: #8a641b;
  }
  .birthday-person--longevity {
    --occasion-color: #805000;
    --occasion-background: rgba(213, 158, 36, 0.12);
  }
  .birthday-person--grand {
    --occasion-color: #a52a2a;
    --occasion-background: rgba(196, 57, 52, 0.09);
  }
  :global(.dark) .birthday-person--adult {
    --occasion-color: #5be0dc;
  }
  :global(.dark) .birthday-person--commemorative,
  :global(.dark) .birthday-person--longevity {
    --occasion-color: #f3c969;
  }
  :global(.dark) .birthday-person--grand {
    --occasion-color: #ff9e96;
  }
  .birthday-note {
    margin: 8px 0 0;
    line-height: 1.6;
  }
  .birthday-error {
    padding: 20px;
    text-align: center;
    color: var(--el-text-color-secondary);
  }
  :deep(.el-empty) {
    padding: 4px 0 12px;
  }
</style>
