// 简单的用户数据存储
const userStore = {
  // 用户数据
  users: [
    {
      id: 1,
      username: 'admin',
      password: '123456'
    }
  ],
  
  // 当前登录用户
  currentUser: null,
  
  // 添加用户
  addUser(username, password) {
    // 检查用户名是否已存在
    const existingUser = this.users.find(user => user.username === username)
    if (existingUser) {
      return { success: false, message: '用户名已存在' }
    }
    
    // 添加新用户
    const newUser = {
      id: this.users.length + 1,
      username,
      password
    }
    this.users.push(newUser)
    
    // 保存到本地存储
    this.saveToLocalStorage()
    
    return { success: true, message: '注册成功' }
  },
  
  // 用户登录
  login(username, password) {
    // 从本地存储加载用户数据
    this.loadFromLocalStorage()
    
    // 查找用户
    const user = this.users.find(
      user => user.username === username && user.password === password
    )
    
    if (user) {
      this.currentUser = user
      localStorage.setItem('isLoggedIn', 'true')
      localStorage.setItem('currentUser', JSON.stringify(user))
      return { success: true, message: '登录成功', user }
    } else {
      return { success: false, message: '用户名或密码错误' }
    }
  },
  
  // 用户登出
  logout() {
    this.currentUser = null
    localStorage.removeItem('isLoggedIn')
    localStorage.removeItem('currentUser')
    return { success: true, message: '已登出' }
  },
  
  // 保存到本地存储
  saveToLocalStorage() {
    localStorage.setItem('users', JSON.stringify(this.users))
  },
  
  // 从本地存储加载
  loadFromLocalStorage() {
    const storedUsers = localStorage.getItem('users')
    if (storedUsers) {
      this.users = JSON.parse(storedUsers)
    }
    
    const storedUser = localStorage.getItem('currentUser')
    if (storedUser) {
      this.currentUser = JSON.parse(storedUser)
    }
  },
  
  // 初始化
  init() {
    this.loadFromLocalStorage()
    if (this.users.length === 0) {
      // 如果没有用户，添加默认管理员账户
      this.users.push({
        id: 1,
        username: 'admin',
        password: '123456'
      })
      this.saveToLocalStorage()
    }
  }
}

// 初始化用户存储
userStore.init()

export default userStore 