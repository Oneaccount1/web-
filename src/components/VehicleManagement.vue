<template>
  <div class="vehicle-container">
    <el-card class="main-card">
      <div class="header">
        <div class="title-section">
          <el-icon class="title-icon"><Van /></el-icon>
          <h2>车辆管理系统</h2>
        </div>
        <div class="header-actions">
          <span class="welcome-text" v-if="currentUser">
            欢迎，{{ currentUser.username }}
            <el-button type="text" @click="handleLogout">退出登录</el-button>
          </span>
          <el-button type="primary" size="large" @click="dialogVisible = true">
            <el-icon><Plus /></el-icon>添加车辆
          </el-button>
        </div>
      </div>

      <!-- 统计卡片 -->
      <div class="statistics">
        <el-row :gutter="20">
          <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
            <el-card shadow="hover" class="stat-card">
              <template #header>
                <div class="stat-header">
                  <el-icon class="stat-icon"><Van /></el-icon>
                  <span>总车辆数</span>
                </div>
              </template>
              <div class="stat-value">{{ vehicles.length }}</div>
            </el-card>
          </el-col>
          <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
            <el-card shadow="hover" class="stat-card">
              <template #header>
                <div class="stat-header">
                  <el-icon class="stat-icon available"><Check /></el-icon>
                  <span>可租车辆</span>
                </div>
              </template>
              <div class="stat-value">{{ availableCount }}</div>
            </el-card>
          </el-col>
          <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
            <el-card shadow="hover" class="stat-card">
              <template #header>
                <div class="stat-header">
                  <el-icon class="stat-icon rented"><Timer /></el-icon>
                  <span>已租车辆</span>
                </div>
              </template>
              <div class="stat-value">{{ rentedCount }}</div>
            </el-card>
          </el-col>
          <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
            <el-card shadow="hover" class="stat-card">
              <template #header>
                <div class="stat-header">
                  <el-icon class="stat-icon"><Money /></el-icon>
                  <span>平均日租金</span>
                </div>
              </template>
              <div class="stat-value">¥{{ averagePrice.toFixed(2) }}</div>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <!-- 搜索栏 -->
      <el-card class="search-card">
        <template #header>
          <div>搜索条件</div>
        </template>
        <div class="search-bar">
          <el-input
            v-model="search.plateNumber"
            placeholder="搜索车牌号"
            class="search-input"
            clearable
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-input
            v-model="search.brand"
            placeholder="搜索品牌"
            class="search-input"
            clearable
          >
            <template #prefix>
              <el-icon><Box /></el-icon>
            </template>
          </el-input>
          <el-select 
            v-model="search.status" 
            placeholder="租赁状态" 
            class="search-select"
            clearable
          >
            <el-option label="全部" value=""></el-option>
            <el-option label="可租" value="available"></el-option>
            <el-option label="已租" value="rented"></el-option>
          </el-select>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>搜索
          </el-button>
          <el-button @click="resetSearch">
            <el-icon><Refresh /></el-icon>重置
          </el-button>
        </div>
      </el-card>

      <!-- 车辆列表 -->
      <el-card class="table-card">
        <template #header>
          <div>车辆列表</div>
        </template>
        <el-table 
          :data="paginatedVehicles" 
          style="width: 100%"
          border
          stripe
          v-loading="loading"
          :empty-text="'暂无数据'"
        >
          <el-table-column prop="plateNumber" label="车牌号" min-width="120" sortable>
            <template #default="scope">
              <el-tag size="small" effect="plain">{{ scope.row.plateNumber }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="brand" label="品牌" min-width="120" sortable></el-table-column>
          <el-table-column prop="price" label="价格" min-width="120" sortable>
            <template #default="scope">
              <span class="price">¥{{ scope.row.price }}/天</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="租赁状态" min-width="120">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'available' ? 'success' : 'danger'" effect="dark">
                {{ scope.row.status === 'available' ? '可租' : '已租' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" fixed="right" min-width="200">
            <template #default="scope">
              <el-button-group>
                <el-button
                  size="small"
                  type="primary"
                  @click="handleEdit(scope.row)"
                >
                  <el-icon><Edit /></el-icon>
                  编辑
                </el-button>
                <el-button
                  size="small"
                  type="danger"
                  @click="handleDelete(scope.row)"
                >
                  <el-icon><Delete /></el-icon>
                  删除
                </el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>

        <!-- 分页器 -->
        <div class="pagination-container">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 30, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            :total="filteredVehicles.length"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
            background
          />
        </div>
      </el-card>
    </el-card>

    <!-- 添加/编辑车辆对话框 -->
    <el-dialog
      :title="editingVehicle ? '编辑车辆' : '添加车辆'"
      v-model="dialogVisible"
      width="500px"
      destroy-on-close
      center
    >
      <el-form :model="vehicleForm" :rules="rules" ref="vehicleFormRef" label-width="100px" label-position="left">
        <el-form-item label="车牌号" prop="plateNumber">
          <el-input v-model="vehicleForm.plateNumber" placeholder="请输入车牌号"></el-input>
        </el-form-item>
        <el-form-item label="品牌" prop="brand">
          <el-input v-model="vehicleForm.brand" placeholder="请输入品牌"></el-input>
        </el-form-item>
        <el-form-item label="价格" prop="price">
          <el-input-number 
            v-model="vehicleForm.price" 
            :min="0" 
            :precision="2"
            :step="10"
            style="width: 100%"
          ></el-input-number>
        </el-form-item>
        <el-form-item label="租赁状态" prop="status">
          <el-select v-model="vehicleForm.status" placeholder="请选择租赁状态" style="width: 100%">
            <el-option label="可租" value="available"></el-option>
            <el-option label="已租" value="rented"></el-option>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Van,
  Plus,
  Check,
  Timer,
  Money,
  Search,
  Box,
  Refresh,
  Edit,
  Delete
} from '@element-plus/icons-vue'
import userStore from '../store/userStore'
import vehicleStore from '../store/vehicleStore'
import { useRouter } from 'vue-router'

export default {
  name: 'VehicleManagement',
  components: {
    Van,
    Plus,
    Check,
    Timer,
    Money,
    Search,
    Box,
    Refresh,
    Edit,
    Delete
  },
  setup() {
    const router = useRouter()
    return { router }
  },
  data() {
    return {
      currentUser: null,
      loading: false,
      vehicles: [],
      search: {
        plateNumber: '',
        brand: '',
        status: ''
      },
      currentPage: 1,
      pageSize: 10,
      dialogVisible: false,
      editingVehicle: null,
      vehicleForm: {
        plateNumber: '',
        brand: '',
        price: 0,
        status: 'available'
      },
      rules: {
        plateNumber: [
          { required: true, message: '请输入车牌号', trigger: 'blur' }
        ],
        brand: [
          { required: true, message: '请输入品牌', trigger: 'blur' }
        ],
        price: [
          { required: true, message: '请输入价格', trigger: 'blur' }
        ],
        status: [
          { required: true, message: '请选择租赁状态', trigger: 'change' }
        ]
      }
    }
  },
  created() {
    // 获取当前登录用户
    const storedUser = localStorage.getItem('currentUser')
    if (storedUser) {
      this.currentUser = JSON.parse(storedUser)
    } else {
      // 如果没有登录，跳转到登录页
      this.router.push('/login')
    }
    
    // 加载车辆数据
    this.loadVehicles()
  },
  computed: {
    filteredVehicles() {
      return vehicleStore.searchVehicles(this.search)
    },
    paginatedVehicles() {
      const start = (this.currentPage - 1) * this.pageSize
      const end = start + this.pageSize
      return this.filteredVehicles.slice(start, end)
    },
    availableCount() {
      return vehicleStore.getAvailableCount()
    },
    rentedCount() {
      return vehicleStore.getRentedCount()
    },
    averagePrice() {
      return vehicleStore.getAveragePrice()
    }
  },
  methods: {
    loadVehicles() {
      this.loading = true
      setTimeout(() => {
        this.vehicles = vehicleStore.getAllVehicles()
        this.loading = false
      }, 500)
    },
    handleLogout() {
      ElMessageBox.confirm('确定要退出登录吗?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        const result = userStore.logout()
        if (result.success) {
          ElMessage.success(result.message)
          this.router.push('/login')
        }
      }).catch(() => {
        // 取消退出
      })
    },
    handleSizeChange(val) {
      this.pageSize = val
      this.currentPage = 1
    },
    handleCurrentChange(val) {
      this.currentPage = val
    },
    handleSearch() {
      this.currentPage = 1
      this.loading = true
      setTimeout(() => {
        this.loading = false
      }, 500)
    },
    resetSearch() {
      this.search = {
        plateNumber: '',
        brand: '',
        status: ''
      }
      this.currentPage = 1
      this.handleSearch()
    },
    handleEdit(vehicle) {
      this.editingVehicle = vehicle
      this.vehicleForm = { ...vehicle }
      this.dialogVisible = true
    },
    handleDelete(vehicle) {
      ElMessageBox.confirm('确认删除该车辆?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        const result = vehicleStore.deleteVehicle(vehicle.id)
        if (result.success) {
          ElMessage.success(result.message)
          this.loadVehicles()
        } else {
          ElMessage.error(result.message)
        }
      }).catch(() => {
        ElMessage.info('已取消删除')
      })
    },
    handleSubmit() {
      this.$refs.vehicleFormRef.validate((valid) => {
        if (valid) {
          let result
          if (this.editingVehicle) {
            result = vehicleStore.updateVehicle(this.editingVehicle.id, this.vehicleForm)
          } else {
            result = vehicleStore.addVehicle(this.vehicleForm)
          }
          
          if (result.success) {
            this.dialogVisible = false
            ElMessage.success(result.message)
            this.loadVehicles()
            this.resetForm()
          } else {
            ElMessage.error(result.message)
          }
        }
      })
    },
    resetForm() {
      this.editingVehicle = null
      this.vehicleForm = {
        plateNumber: '',
        brand: '',
        price: 0,
        status: 'available'
      }
      if (this.$refs.vehicleFormRef) {
        this.$refs.vehicleFormRef.resetFields()
      }
    }
  }
}
</script>

<style scoped>
.vehicle-container {
  padding: 20px;
  background-color: #f0f2f5;
  min-height: 100vh;
  box-sizing: border-box;
  border-radius: 10px;
}

.main-card {
  margin-bottom: 20px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  border-radius: 10px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #ebeef5;
}

.title-section {
  display: flex;
  align-items: center;
  gap: 12px;
}

.title-icon {
  font-size: 28px;
  color: #409EFF;
}

h2 {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: #303133;
}

.statistics {
  margin-bottom: 24px;
  border-radius: 10px;
}

.stat-card {
  transition: all 0.3s;
  height: 100%;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
}

.stat-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 4px 16px 0 rgba(0, 0, 0, 0.1);
}

.stat-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  font-weight: 500;
}

.stat-icon {
  font-size: 20px;
}

.stat-icon.available {
  color: #67C23A;
}

.stat-icon.rented {
  color: #F56C6C;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
  text-align: center;
  padding: 20px 0;
}

.search-card {
  margin-bottom: 24px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
}

.search-bar {
  display: flex;
  gap: 16px;
  align-items: center;
  flex-wrap: wrap;
  padding: 8px 0;
}

.search-input {
  width: 220px;
}

.search-select {
  width: 220px;
}

.table-card {
  margin-bottom: 24px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
}

.price {
  color: #F56C6C;
  font-weight: bold;
}

.pagination-container {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
  border-top: 1px solid #ebeef5;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

:deep(.el-card__body) {
  padding: 20px;
}

:deep(.el-card__header) {
  padding: 16px 20px;
  font-weight: 500;
}

:deep(.el-table .cell) {
  padding: 10px 12px;
}

:deep(.el-button-group .el-button) {
  padding: 8px 15px;
}

:deep(.el-tag) {
  font-weight: 500;
}

:deep(.el-table thead th) {
  background-color: #f5f7fa !important;
  color: #606266;
  font-weight: 600;
  height: 50px;
}

:deep(.el-pagination) {
  font-weight: normal;
}

/* 添加响应式设计 */
@media (max-width: 1200px) {
  .statistics .el-row {
    margin-left: -10px !important;
    margin-right: -10px !important;
  }
  
  .statistics .el-col {
    padding-left: 10px !important;
    padding-right: 10px !important;
  }
}

@media (max-width: 991px) {
  .search-input, .search-select {
    width: 180px;
  }
}

@media (max-width: 768px) {
  .search-input, .search-select {
    width: 100%;
  }
  
  .search-bar {
    flex-direction: column;
    align-items: stretch;
  }
  
  .search-bar .el-button {
    margin-top: 10px;
  }
  
  .statistics .el-col {
    width: 50%;
    margin-bottom: 16px;
  }
  
  .pagination-container {
    justify-content: center;
  }
}

.welcome-text {
  margin-right: 20px;
  font-size: 14px;
  color: #606266;
}

.header-actions {
  display: flex;
  align-items: center;
}
</style> 