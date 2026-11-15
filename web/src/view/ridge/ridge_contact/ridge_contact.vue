<template>
  <div>
    <ContactStats ref="contactStatsRef" />
    <UpcomingBirthdays ref="upcomingBirthdaysRef" @detail="getDetails" />

    <!-- 搜索框 -->
    <div class="gva-search-box">
      <el-form
        ref="elSearchFormRef"
        :inline="true"
        :model="searchInfo"
        class="demo-form-inline"
        @keyup.enter="onSubmit"
      >
        <el-form-item label="姓名" prop="remarkName">
          <el-input v-model="searchInfo.remarkName" placeholder="姓名" />
        </el-form-item>

        <el-form-item label="性别" prop="gender">
          <el-select v-model="searchInfo.gender">
            <el-option label="未知" :value="0" />
            <el-option label="男" :value="1" />
            <el-option label="女" :value="2" />
          </el-select>
        </el-form-item>

        <el-form-item label="分类" prop="categoryId">
          <el-select v-model="searchInfo.categoryId">
            <el-option
              v-for="c in catList"
              :label="c.label"
              :value="c.value"
              :key="c.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="标签" prop="tagId">
          <el-select v-model="searchInfo.tagId">
            <el-option
              v-for="c in tagList"
              :label="c.label"
              :value="c.value"
              :key="c.id"
            />
          </el-select>
        </el-form-item>

        <template v-if="showAllQuery">
          <el-form-item label="好友状态" prop="status">
            <el-select v-model="searchInfo.status">
              <el-option
                v-for="c in statusList"
                :label="c.label"
                :value="parseInt(c.value)"
                :key="c.id"
              />
            </el-select>
          </el-form-item>

          <el-form-item label="分析豁免" prop="isExempt">
            <el-select v-model="searchInfo.isExempt">
              <el-option label="否" :value="0" />
              <el-option label="是" :value="1" />
            </el-select>
          </el-form-item>

          <el-form-item label="操作建议" prop="suggestedAction">
            <el-select v-model="searchInfo.suggestedAction">
              <el-option
                v-for="c in opList"
                :label="c.label"
                :value="parseInt(c.value)"
                :key="c.id"
              />
            </el-select>
          </el-form-item>
        </template>

        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit"
            >查询</el-button
          >
          <el-button icon="refresh" @click="onReset">重置</el-button>

          <el-button
            link
            type="primary"
            icon="arrow-down"
            @click="showAllQuery = true"
            v-if="!showAllQuery"
            >展开</el-button
          >
          <el-button
            link
            type="primary"
            icon="arrow-up"
            @click="showAllQuery = false"
            v-else
            >收起</el-button
          >
        </el-form-item>
      </el-form>
    </div>

    <!-- 表格 -->
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" icon="plus" @click="openDialog()"
          >新增</el-button
        >
        <el-button
          icon="delete"
          :disabled="!multipleSelection.length"
          @click="onDelete"
          type="danger"
          >删除</el-button
        >
      </div>
      <el-table
        ref="multipleTable"
        tooltip-effect="dark"
        :data="tableData"
        row-key="id"
        @selection-change="handleSelectionChange"
      >
        <el-table-column align="center" type="selection" width="55" />

        <el-table-column align="center" label="ID" prop="id" width="55" />

        <el-table-column align="center" label="姓名" prop="remarkName" />

        <el-table-column align="center" label="性别" prop="gender">
          <template #default="{ row }">{{ genderMap[row.gender] }} </template>
        </el-table-column>

        <el-table-column align="center" label="分类" prop="categoryId">
          <template #default="{ row }">{{ catMap[row.categoryId] }} </template>
        </el-table-column>

        <el-table-column align="center" label="标签" prop="tagId">
          <template #default="{ row }">{{ tagMap[row.tagId] }} </template>
        </el-table-column>

        <el-table-column
          align="center"
          label="生日"
          prop="birthday"
          min-width="120"
          show-overflow-tooltip
        >
          <template #default="{ row }">{{
            formatContactBirthday(row)
          }}</template>
        </el-table-column>

        <el-table-column align="center" label="排序值" prop="sortOrder" />

        <el-table-column align="center" label="添加时间" prop="friendedAt">
          <template #default="{ row }"
            >{{ formatDate(row.friendedAt, 'yyyy-MM-dd') }}
          </template>
        </el-table-column>

        <el-table-column align="center" label="好友状态" prop="status">
          <template #default="{ row }">{{ statusMap[row.status] }} </template>
        </el-table-column>

        <el-table-column align="center" label="分析豁免" prop="isExempt">
          <template #default="{ row }">
            <el-tag :type="tagTypeMap[row.isExempt]">
              {{ formatBoolean(row.isExempt) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column align="center" label="操作建议" prop="suggestedAction">
          <template #default="{ row }"
            ><el-tag :type="opMap[row.suggestedAction]?.extend">
              {{ opMap[row.suggestedAction]?.label }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column align="center" label="操作" fixed="right" width="210">
          <template #default="{ row }">
            <el-button
              type="info"
              icon="InfoFilled"
              link
              @click="getDetails(row)"
            >
              详情
            </el-button>
            <el-button
              type="warning"
              link
              icon="edit"
              @click="updateRidgeContactFunc(row)"
            >
              编辑
            </el-button>
            <el-button type="danger" link icon="delete" @click="deleteRow(row)"
              >删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="gva-pagination">
        <el-pagination
          layout="total, sizes, prev, pager, next, jumper"
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 30, 50, 100]"
          :total="total"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <!-- 新增 / 编辑 -->
    <el-drawer
      destroy-on-close
      :size="appStore.drawerSize"
      v-model="dialogFormVisible"
      :show-close="false"
      :before-close="closeDialog"
    >
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ type === 'create' ? '新增' : '编辑' }}</span>
          <div>
            <el-button :loading="btnLoading" type="primary" @click="enterDialog"
              >确 定</el-button
            >
            <el-button @click="closeDialog">取 消</el-button>
          </div>
        </div>
      </template>

      <el-form
        :model="formData"
        label-position="top"
        ref="elFormRef"
        :rules="rule"
        label-width="80px"
      >
        <el-form-item label="姓名" prop="remarkName">
          <el-input
            v-model="formData.remarkName"
            :clearable="true"
            placeholder="请输入姓名"
          />
        </el-form-item>
        <el-form-item label="备注" prop="notes">
          <el-input
            v-model="formData.notes"
            :clearable="true"
            placeholder="请输入备注"
          />
        </el-form-item>
        <el-form-item label="性别" prop="gender">
          <el-radio-group v-model="formData.gender">
            <el-radio :value="0">未知</el-radio>
            <el-radio :value="1">男</el-radio>
            <el-radio :value="2">女</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="生日历法">
          <el-radio-group
            v-model="formData.birthdayIsLunar"
            @change="changeBirthdayCalendar"
          >
            <el-radio :value="false">公历</el-radio>
            <el-radio :value="true">农历</el-radio>
          </el-radio-group>
          <el-checkbox
            v-if="formData.birthdayIsLunar"
            v-model="formData.birthdayIsLeapMonth"
            class="birthday-leap"
            >闰月生日</el-checkbox
          >
        </el-form-item>
        <el-form-item label="生日（选填）" prop="birthday">
          <ContactBirthdayInput
            ref="birthdayInputRef"
            v-model="formData.birthday"
            :lunar="formData.birthdayIsLunar"
          />
        </el-form-item>
        <el-form-item label="排序值" prop="sortOrder">
          <el-input-number
            v-model.number="formData.sortOrder"
            :min="0"
            :precision="0"
            placeholder="排序值"
          />
        </el-form-item>
        <el-form-item label="分析豁免" prop="isExempt">
          <el-switch
            v-model="formData.isExemptBool"
            @change="handleSwitchChange"
          />
        </el-form-item>
        <el-form-item label="好友状态" prop="status">
          <el-select v-model="formData.status" class="input-select">
            <el-option
              v-for="c in statusList"
              :label="c.label"
              :value="parseInt(c.value)"
              :key="c.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="分类" prop="categoryId">
          <el-select v-model="formData.categoryId" class="input-select">
            <el-option
              v-for="c in catList"
              :label="c.label"
              :value="parseInt(c.value)"
              :key="c.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="标签" prop="tagId">
          <el-select v-model="formData.tagId" class="input-select">
            <el-option
              v-for="c in tagList"
              :label="c.label"
              :value="parseInt(c.value)"
              :key="c.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="添加时间" prop="friendedAt">
          <el-date-picker
            v-model="formData.friendedAt"
            type="date"
            placeholder="选择日期"
            :clearable="false"
          />
        </el-form-item>
        <el-form-item label="上次聊天" prop="lastChatAt">
          <el-date-picker
            v-model="formData.lastChatAt"
            type="date"
            placeholder="选择日期"
            :clearable="true"
          />
        </el-form-item>
        <el-form-item label="上次互动" prop="lastInteractionAt">
          <el-date-picker
            v-model="formData.lastInteractionAt"
            type="date"
            placeholder="选择日期"
            :clearable="true"
          />
        </el-form-item>
        <el-form-item label="置仅聊天" prop="chatOnlySince">
          <el-date-picker
            v-model="formData.chatOnlySince"
            type="date"
            placeholder="选择日期"
            :clearable="true"
          />
        </el-form-item>
      </el-form>
    </el-drawer>

    <!-- 详情 -->
    <el-drawer
      destroy-on-close
      :size="appStore.drawerSize"
      v-model="detailShow"
      :show-close="true"
      :before-close="closeDetailShow"
      title="详情"
    >
      <el-descriptions :column="1" border>
        <el-descriptions-item label="ID">
          {{ detailForm.id }}
        </el-descriptions-item>
        <el-descriptions-item label="姓名">
          {{ detailForm.remarkName }}
        </el-descriptions-item>
        <el-descriptions-item label="性别">
          {{ genderMap[detailForm.gender] }}
        </el-descriptions-item>
        <el-descriptions-item label="生日">{{
          formatContactBirthday(detailForm)
        }}</el-descriptions-item>
        <el-descriptions-item label="分类">
          {{ catMap[detailForm.categoryId] }}
        </el-descriptions-item>
        <el-descriptions-item label="标签">
          {{ tagMap[detailForm.tagId] }}
        </el-descriptions-item>
        <el-descriptions-item label="备注">
          {{ detailForm.notes }}
        </el-descriptions-item>
        <el-descriptions-item label="排序值">
          {{ detailForm.sortOrder }}
        </el-descriptions-item>
        <el-descriptions-item label="好友添加时间">
          {{ formatDate(detailForm.friendedAt, 'yyyy-MM-dd') || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="上次聊天时间">
          {{ formatDate(detailForm.lastChatAt, 'yyyy-MM-dd') || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="上次互动时间">
          {{ formatDate(detailForm.lastInteractionAt, 'yyyy-MM-dd') || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="置仅聊天时间">
          {{ formatDate(detailForm.chatOnlySince, 'yyyy-MM-dd') || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="当前好友状态">
          {{ statusMap[detailForm.status] }}
        </el-descriptions-item>
        <el-descriptions-item label="是否分析豁免">
          {{ formatBoolean(detailForm.isExempt) }}
        </el-descriptions-item>
        <el-descriptions-item label="好友操作建议">
          {{ opMap[detailForm.suggestedAction]?.label }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ formatDate(detailForm.createdAt) }}
        </el-descriptions-item>
        <el-descriptions-item label="更新时间">
          {{ formatDate(detailForm.updatedAt) }}
        </el-descriptions-item>
      </el-descriptions>
    </el-drawer>
  </div>
</template>

<script setup>
  import {
    createRidgeContact,
    deleteRidgeContact,
    deleteRidgeContactByIds,
    updateRidgeContact,
    findRidgeContact,
    getRidgeContactList
  } from '@/api/ridge/ridge_contact'
  import { formatBoolean, getDictFunc, formatDate } from '@/utils/format'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { ref, reactive, onMounted } from 'vue'
  import { useAppStore } from '@/pinia'
  import ContactStats from './components/ContactStats.vue'
  import ContactBirthdayInput from './components/ContactBirthdayInput.vue'
  import UpcomingBirthdays from './components/UpcomingBirthdays.vue'
  import { formatContactBirthday } from '@/utils/ridgeBirthday'

  const contactStatsRef = ref()
  const upcomingBirthdaysRef = ref()
  const birthdayInputRef = ref()

  defineOptions({
    name: 'RidgeContact'
  })

  // 提交按钮loading
  const btnLoading = ref(false)
  const appStore = useAppStore()

  // 控制更多查询条件显示/隐藏状态
  const showAllQuery = ref(false)

  // 自动化生成的字典（可能为空）以及字段
  const formData = ref({
    id: undefined,
    remarkName: '',
    birthday: null,
    birthdayIsLunar: false,
    birthdayIsLeapMonth: false,
    gender: 1,
    categoryId: undefined,
    tagId: undefined,
    notes: '',
    sortOrder: 0,
    friendedAt: undefined,
    lastChatAt: undefined,
    lastInteractionAt: undefined,
    chatOnlySince: undefined,
    status: 0,
    isExempt: 0,
    isExemptBool: false
  })

  const changeBirthdayCalendar = () => {
    birthdayInputRef.value?.clear()
    formData.value.birthday = null
    formData.value.birthdayIsLeapMonth = false
  }

  // 验证规则
  const rule = reactive({
    birthday: [
      {
        validator: (_rule, _value, callback) => {
          if (birthdayInputRef.value?.isIncomplete())
            callback(new Error('请选择完整的农历生日，或清空'))
          else callback()
        },
        trigger: 'change'
      }
    ],
    remarkName: [
      {
        required: true,
        message: '',
        trigger: ['input', 'blur']
      },
      {
        whitespace: true,
        message: '不能只输入空格',
        trigger: ['input', 'blur']
      }
    ],
    gender: [
      {
        required: true,
        message: '',
        trigger: ['input', 'blur']
      }
    ],
    categoryId: [
      {
        required: true,
        message: '',
        trigger: ['input', 'blur']
      }
    ],
    tagId: [
      {
        required: true,
        message: '',
        trigger: ['input', 'blur']
      }
    ],
    notes: [
      {
        required: true,
        message: '',
        trigger: ['input', 'blur']
      },
      {
        whitespace: true,
        message: '不能只输入空格',
        trigger: ['input', 'blur']
      }
    ],
    sortOrder: [
      {
        required: true,
        message: '',
        trigger: ['input', 'blur']
      }
    ],
    friendedAt: [
      {
        required: true,
        message: '',
        trigger: ['input', 'blur']
      }
    ],
    status: [
      {
        required: true,
        message: '',
        trigger: ['input', 'blur']
      }
    ],
    isExempt: [
      {
        required: true,
        message: '',
        trigger: ['input', 'blur']
      }
    ]
  })

  const elFormRef = ref()
  const elSearchFormRef = ref()

  // =========== 表格控制部分 ===========
  const page = ref(1)
  const total = ref(0)
  const pageSize = ref(10)
  const tableData = ref([])
  const searchInfo = ref({})
  // 重置
  const onReset = () => {
    searchInfo.value = {}
    getTableData()
  }

  // 搜索
  const onSubmit = () => {
    elSearchFormRef.value?.validate(async (valid) => {
      if (!valid) return
      page.value = 1
      getTableData()
    })
  }

  // 分页
  const handleSizeChange = (val) => {
    pageSize.value = val
    getTableData()
  }

  // 修改页面容量
  const handleCurrentChange = (val) => {
    page.value = val
    getTableData()
  }

  // 查询
  const getTableData = async () => {
    const table = await getRidgeContactList({
      page: page.value,
      pageSize: pageSize.value,
      ...searchInfo.value
    })
    if (table.code === 0) {
      tableData.value = table.data.list
      total.value = table.data.total
      page.value = table.data.page
      pageSize.value = table.data.pageSize
    }
  }

  getTableData()

  // ============== 表格控制部分结束 ===============

  // 获取需要的字典 可能为空 按需保留
  const setOptions = async () => {}

  // 获取需要的字典 可能为空 按需保留
  setOptions()

  // 多选数据
  const multipleSelection = ref([])
  // 多选
  const handleSelectionChange = (val) => {
    multipleSelection.value = val
  }

  // 删除行
  const deleteRow = (row) => {
    ElMessageBox.confirm('确定要删除吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(() => {
      deleteRidgeContactFunc(row)
    })
  }

  // 多选删除
  const onDelete = async () => {
    ElMessageBox.confirm('确定要删除吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      const ids = []
      if (multipleSelection.value.length === 0) {
        ElMessage({
          type: 'warning',
          message: '请选择要删除的数据'
        })
        return
      }
      multipleSelection.value &&
        multipleSelection.value.map((item) => {
          ids.push(item.id)
        })
      const res = await deleteRidgeContactByIds({ ids })
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: '删除成功'
        })
        if (tableData.value.length === ids.length && page.value > 1) {
          page.value--
        }
        getTableData()
        contactStatsRef.value?.refresh()
        upcomingBirthdaysRef.value?.refresh()
      }
    })
  }

  // 行为控制标记（弹窗内部需要增还是改）
  const type = ref('')

  // 更新行
  const updateRidgeContactFunc = async (row) => {
    const res = await findRidgeContact({ id: row.id })
    type.value = 'update'
    if (res.code === 0) {
      formData.value = res.data
      formData.value.isExemptBool = formData.value.isExempt === 1
      dialogFormVisible.value = true
    }
  }

  // 删除行
  const deleteRidgeContactFunc = async (row) => {
    const res = await deleteRidgeContact({ id: row.id })
    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: '删除成功'
      })
      if (tableData.value.length === 1 && page.value > 1) {
        page.value--
      }
      getTableData()
      contactStatsRef.value?.refresh()
      upcomingBirthdaysRef.value?.refresh()
    }
  }

  // 弹窗控制标记
  const dialogFormVisible = ref(false)

  // 打开弹窗
  const openDialog = () => {
    type.value = 'create'
    dialogFormVisible.value = true
  }

  // 关闭弹窗
  const closeDialog = () => {
    dialogFormVisible.value = false
    formData.value = {
      id: undefined,
      remarkName: '',
      birthday: null,
      birthdayIsLunar: false,
      birthdayIsLeapMonth: false,
      gender: 1,
      categoryId: undefined,
      tagId: undefined,
      notes: '',
      sortOrder: 0,
      friendedAt: undefined,
      lastChatAt: undefined,
      lastInteractionAt: undefined,
      chatOnlySince: undefined,
      status: 0,
      isExempt: 0,
      isExemptBool: false
    }
  }
  // 弹窗确定
  const enterDialog = async () => {
    btnLoading.value = true
    elFormRef.value?.validate(async (valid) => {
      if (!valid) return (btnLoading.value = false)
      let res
      switch (type.value) {
        case 'create':
          res = await createRidgeContact(formData.value)
          break
        case 'update':
          res = await updateRidgeContact(formData.value)
          break
        default:
          res = await createRidgeContact(formData.value)
          break
      }
      btnLoading.value = false
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: '创建/更改成功'
        })
        closeDialog()
        getTableData()
        contactStatsRef.value?.refresh()
        upcomingBirthdaysRef.value?.refresh()
      }
    })
  }

  const detailForm = ref({})

  // 查看详情控制标记
  const detailShow = ref(false)

  // 打开详情弹窗
  const openDetailShow = () => {
    detailShow.value = true
  }

  // 打开详情
  const getDetails = async (row) => {
    // 打开弹窗
    const res = await findRidgeContact({ id: row.id })
    if (res.code === 0) {
      detailForm.value = res.data
      openDetailShow()
    }
  }

  // 关闭详情弹窗
  const closeDetailShow = () => {
    detailShow.value = false
    detailForm.value = {}
  }

  /********************************************************************************************************************/
  // 是否
  const tagTypeMap = {
    0: 'danger',
    1: 'success'
  }
  // 性别
  const genderMap = {
    0: '未知',
    1: '男',
    2: '女'
  }

  const handleSwitchChange = (value) => {
    formData.value.isExempt = value ? 1 : 0
  }

  // 分类
  const catList = ref([])
  const catMap = ref({})
  const fetchCategoryList = async () => {
    const dataList = await getDictFunc('ridge_contact_category')

    const list = []
    const map = {}
    dataList.forEach(({ value, label }) => {
      list.push({ value, label })
      map[value] = label
    })

    catList.value = list
    catMap.value = map
  }

  // 标签
  const tagList = ref([])
  const tagMap = ref({})
  const fetchTagList = async () => {
    const dataList = await getDictFunc('ridge_contact_tag')

    const list = []
    const map = {}
    dataList.forEach(({ value, label }) => {
      list.push({ value, label })
      map[value] = label
    })

    tagList.value = list
    tagMap.value = map
  }

  // 状态
  const statusList = ref([])
  const statusMap = ref({})
  const fetchStatusList = async () => {
    const dataList = await getDictFunc('ridge_contact_status')

    const list = []
    const map = {}
    dataList.forEach(({ value, label }) => {
      list.push({ value, label })
      map[value] = label
    })

    statusList.value = list
    statusMap.value = map
  }

  // 操作
  const opList = ref([])
  const opMap = ref({})
  const fetchOperationList = async () => {
    const dataList = await getDictFunc('ridge_contact_operation')

    const list = []
    const map = {}
    dataList.forEach(({ value, label, extend }) => {
      list.push({ value, label })
      map[value] = { label, extend }
    })

    opList.value = list
    opMap.value = map
  }

  onMounted(async () => {
    await Promise.all([
      fetchCategoryList(),
      fetchTagList(),
      fetchStatusList(),
      fetchOperationList()
    ])
  })
</script>

<style scoped>
  .birthday-leap {
    margin-left: 16px;
  }
  .input-select {
    width: 210px;
  }
</style>
