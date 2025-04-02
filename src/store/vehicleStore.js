// 车辆数据存储
const vehicleStore = {
  // 车辆数据
  vehicles: [
    {
      id: 15,
      plateNumber: '京A·12345',
      brand: '丰田',
      price: 300,
      status: 'available'
    },
    {
      id: 16,
      plateNumber: '京B·67890',
      brand: '本田',
      price: 280,
      status: 'rented'
    },
    {
      "id": 1,
      "plateNumber": "沪B·7K8V2",
      "brand": "本田",
      "price": 420,
      "status": "available"
    },
    {
      "id": 2,
      "plateNumber": "粤A·3D9XW",
      "brand": "大众",
      "price": 280,
      "status": "unavailable"
    },
    {
      "id": 3,
      "plateNumber": "京N·5H7Q3",
      "brand": "日产",
      "price": 360,
      "status": "available"
    },
    {
      "id": 4,
      "plateNumber": "浙C·2F4J8",
      "brand": "别克",
      "price": 390,
      "status": "available"
    },
    {
      "id": 5,
      "plateNumber": "苏E·9L3P6",
      "brand": "现代",
      "price": 310,
      "status": "unavailable"
    },
    {
      "id": 6,
      "plateNumber": "川A·1M5T9",
      "brand": "奥迪",
      "price": 480,
      "status": "available"
    },
    {
      "id": 7,
      "plateNumber": "津R·4S2K7",
      "brand": "雪佛兰",
      "price": 290,
      "status": "available"
    },
    {
      "id": 8,
      "plateNumber": "渝F·6Z9Y4",
      "brand": "比亚迪",
      "price": 330,
      "status": "unavailable"
    },
    {
      "id": 9,
      "plateNumber": "陕B·8N3V1",
      "brand": "特斯拉",
      "price": 450,
      "status": "available"
    },
    {
      "id": 10,
      "plateNumber": "鄂C·5Q7X2",
      "brand": "沃尔沃",
      "price": 410,
      "status": "available"
    },
    {
      "id": 11,
      "plateNumber": "湘D·3R9W6",
      "brand": "雷克萨斯",
      "price": 470,
      "status": "unavailable"
    },
    {
      "id": 12,
      "plateNumber": "闽E·0K4J5",
      "brand": "宝马",
      "price": 490,
      "status": "available"
    },
    {
      id: 13,
      plateNumber: '京C·11111',
      brand: '大众',
      price: 260,
      status: 'available'
    },
    {
      id: 14,
      plateNumber: '京D·22222',
      brand: '奔驰',
      price: 500,
      status: 'rented'
    }
  ],

  // 获取所有车辆
  getAllVehicles() {
    this.loadFromLocalStorage()
    return this.vehicles
  },

  // 添加车辆
  addVehicle(vehicle) {
    const newVehicle = {
      id: Date.now(),
      ...vehicle
    }
    this.vehicles.push(newVehicle)
    this.saveToLocalStorage()
    return { success: true, message: '添加成功', vehicle: newVehicle }
  },

  // 更新车辆
  updateVehicle(id, updatedVehicle) {
    const index = this.vehicles.findIndex(v => v.id === id)
    if (index !== -1) {
      this.vehicles[index] = {
        ...this.vehicles[index],
        ...updatedVehicle
      }
      this.saveToLocalStorage()
      return { success: true, message: '更新成功', vehicle: this.vehicles[index] }
    }
    return { success: false, message: '未找到车辆' }
  },

  // 删除车辆
  deleteVehicle(id) {
    const index = this.vehicles.findIndex(v => v.id === id)
    if (index !== -1) {
      const deletedVehicle = this.vehicles[index]
      this.vehicles.splice(index, 1)
      this.saveToLocalStorage()
      return { success: true, message: '删除成功', vehicle: deletedVehicle }
    }
    return { success: false, message: '未找到车辆' }
  },

  // 搜索车辆
  searchVehicles(criteria) {
    return this.vehicles.filter(vehicle => {
      const matchPlateNumber = !criteria.plateNumber ||
        vehicle.plateNumber.toLowerCase().includes(criteria.plateNumber.toLowerCase())
      const matchBrand = !criteria.brand ||
        vehicle.brand.toLowerCase().includes(criteria.brand.toLowerCase())
      const matchStatus = !criteria.status || vehicle.status === criteria.status
      return matchPlateNumber && matchBrand && matchStatus
    })
  },

  // 获取可租车辆数量
  getAvailableCount() {
    return this.vehicles.filter(v => v.status === 'available').length
  },

  // 获取已租车辆数量
  getRentedCount() {
    return this.vehicles.filter(v => v.status === 'rented').length
  },

  // 获取平均价格
  getAveragePrice() {
    if (this.vehicles.length === 0) return 0
    const total = this.vehicles.reduce((sum, v) => sum + v.price, 0)
    return total / this.vehicles.length
  },

  // 保存到本地存储
  saveToLocalStorage() {
    localStorage.setItem('vehicles', JSON.stringify(this.vehicles))
  },

  // 从本地存储加载
  loadFromLocalStorage() {
    const storedVehicles = localStorage.getItem('vehicles')
    if (storedVehicles) {
      this.vehicles = JSON.parse(storedVehicles)
    }
  },

  // 初始化
  init() {
    this.loadFromLocalStorage()
    if (this.vehicles.length === 0) {
      // 如果没有车辆数据，使用默认数据
      this.vehicles = [
        {
          id: 1,
          plateNumber: '京A12345',
          brand: '丰田',
          price: 300,
          status: 'available'
        },
        {
          id: 2,
          plateNumber: '京B67890',
          brand: '本田',
          price: 280,
          status: 'rented'
        },
        {
          id: 3,
          plateNumber: '京C11111',
          brand: '大众',
          price: 260,
          status: 'available'
        },
        {
          id: 4,
          plateNumber: '京D22222',
          brand: '奔驰',
          price: 500,
          status: 'rented'
        }
      ]
      this.saveToLocalStorage()
    }
  }
}

// 初始化车辆存储
vehicleStore.init()

export default vehicleStore 