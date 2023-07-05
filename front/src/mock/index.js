/**
 * @description register mock data right here
 */
import Mock from 'mockjs'
// 设置拦截请求的响应时间 ajax 请求；
Mock.setup({
  timeout: '200-600'
})

// mock 一组角色数据；
const genRoles = () => ({
  status: 0,
  data: ['super', 'admin', 'nomal'],
  message: '成功'
})

const getSwiperInfo = () => ({
  status: 0,
  data: [
    {
      name: 'vue-next',
      itemSrc: 'https://p6-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/c588b8ab65a74d59aa379801136df4e0~tplv-k3u1fbpfcp-watermark.image',
      targetLink: 'https://github.com/vuejs/docs-next-zh-cn'
    },
    {
      name: 'vitejs',
      itemSrc: 'https://p6-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/a7351d2dcd7846158604ac8bd57222b5~tplv-k3u1fbpfcp-watermark.image',
      targetLink: 'https://github.com/vitejs'
    },
    {
      name: 'element-plus',
      itemSrc: 'https://p6-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/48a7fd198df44cca9c0dc10a8047bbef~tplv-k3u1fbpfcp-watermark.image',
      targetLink: 'https://github.com/element-plus/element-plus'
    },
    {
      name: 'tslang',
      itemSrc: 'https://p1-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/610fc57450884ceaae9578689663fe2f~tplv-k3u1fbpfcp-watermark.image',
      targetLink: 'https://github.com/Microsoft/TypeScript'
    }
  ],
  message: '成功'
})

Mock.mock('/api/auth/user/login', 'post', (option) => {
  const { email, password } = JSON.parse(option.body)
  console.log(email, password)
  if (email === 'mock123@outlook.com' && password === '1zsJJYx5/srrQaMycn5MYA==') {
    return {
      status: 0,
      data: {
        accessToken: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6IjIyODA1MjAxMjhAcXEuY29tIiwic3ViIjo5LCJpYXQiOjE2MjU4MzQ3MTksImV4cCI6MTYyODQyNjcxOX0.YQLVi-zw4XWQEd8Hy2YZGlFaqX8c7xyRPrYuxcFywFE'
      },
      success: true,
      message: '成功'
    }
  }
  if (email === 'admin@outlook.com' && password === '1zsJJYx5/srrQaMycn5MYA==') {
    return {
      status: 0,
      data: {
        accessToken: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6IjIyODA1MjAxMjhAcXEuY29tIiwic3ViIjo5LCJpYXQiOjE2MjU4MzQ3MTksImV4cCI6MTYyODQyNjcxOX0.YQLVi-zw4XWQEd8Hy2YZGlFaqX8c7xyRPrYuxcFywFE'
      },
      success: true,
      message: '成功'
    }
  }
  return {
    status: 0,
    data: null,
    message: '账户或者密码错误'
  }
})
// 登录成功，查询用户信息，包括角色
Mock.mock('/api/auth/user/userInfo', 'post', (option) => {
  const { email } = JSON.parse(option.body)

  if (email === 'mock123@outlook.com') {
    return {
      status: 0,
      data: {
        roleName: '超级管理员'
      },
      success: true,
      message: '成功'
    }
  }
  if (email === 'admin@outlook.com') {
    return {
      status: 0,
      data: {
        roleName: '管理员'
      },
      success: true,
      message: '成功'
    }
  }
  return {
    status: 0,
    data: null,
    message: '账户或者密码错误'
  }
})

Mock.mock('/api/auth/user/register', 'post', () => ({
  status: 0,
  data: {},
  success: true,
  message: '成功'
}))
//
Mock.mock('/api/auth/permission/routes', 'post', (option) => {
  const { roleName } = JSON.parse(option.body)
  if (roleName === '超级管理员') {
    return {
      status: 0,
      data: {
        authedRoutes: [
          '/dashboard',
          '/guide',
          '/dragable',
          '/calendar',
          '/copy',
          '/zip',
          '/role',
          '/menu',
          '/projectboard',
          '/excel',
          '/table',
          '/form',
          '/qrcode',
          '/editor',
          '/upload',
          '/cropper',
          '/personal'
        ]
      },
      success: true,
      message: '成功'
    }
  }
  if (roleName === '管理员') {
    return {
      status: 0,
      data: {
        authedRoutes: ['/dashboard', '/role', '/menu', '/personal']
      },
      success: true,
      message: '成功'
    }
  }
  return {
    status: 0,
    data: {
      authedRoutes: ['/dashboard', '/guide', '/dragable', '/calendar', '/copy', '/zip', '/excel', '/table', '/form', '/qrcode', '/editor', '/upload', '/cropper', '/personal']
    },
    success: true,
    message: '成功'
  }
})

Mock.mock('/api/auth/permission/permissions', 'post', () => ({
  status: 0,
  data: {
    permissions: ['test:permission-btn', 'test:permission-btn2', 'test:permission-btn3']
  },
  success: true,
  message: '成功'
}))

// /api/auth/user/reset-password
Mock.mock('/api/auth/user/reset-password', 'post', () => ({
  status: 0,
  data: {},
  success: true,
  message: '成功'
}))
//  /api/auth/email/forgot-password
Mock.mock('/api/auth/email/forgot-password', 'post', () => ({
  status: 0,
  data: {},
  success: true,
  message: '成功'
}))
// /api/auth/user/reset-password
Mock.mock('/api/auth/user/reset-password', 'post', () => ({
  status: 0,
  data: {},
  success: true,
  message: '成功'
}))

Mock.mock('/api/setting/basicInfo', 'post', (option) => {
  const { email, nickname, desc, mobile } = JSON.parse(option.body)
  return {
    status: 0,
    data: {
      email,
      nickname,
      desc,
      mobile
    },
    message: '更新成功'
  }
})

Mock.mock('/api/personal/tags', 'get', () => ({
  status: 0,
  data: {
    tags: ['积极阳光', '专注', '认真负责', '花痴']
  },
  message: '成功'
}))
Mock.mock('/api/data/world-population', 'get', () => ({
  status: 0,
  data: {
    dataSets: [
      { category: 'frontEnd', value: 13832220000, x: 'Vue-next' },
      { category: 'frontEnd', value: 13832210000, x: 'Vuex' },
      { category: 'frontEnd', value: 1383232300, x: 'vue-router' },
      { category: 'frontLib', value: 13832210000, x: 'ElementPlus' },
      { category: 'frontEnd', value: 1383232200, x: 'react' },
      { category: 'frontEnd', value: 13831322200, x: 'antd' },
      { category: 'frontEnd', value: 13831322200, x: 'antv' },
      { category: 'lowcode', value: 1383232400, x: 'lowcode' },
      { category: 'frontEnd', value: 1383232400, x: 'micro-frontend' },
      { category: 'frontEnd', value: 1383232400, x: 'flutter' },
      { category: 'frontEnd', value: 1383232300, x: '微信小程序' },
      { category: 'frontEnd', value: 1383232000, x: 'Taro' },
      { category: 'frontEnd', value: 1383231000, x: '抖音小程序' },
      { category: 'frontEnd', value: 1383236000, x: '快手小程序' },
      { category: 'frontEnd', value: 138322000, x: 'UniApp' },
      { category: 'frontEnd', value: 138322000, x: 'NodeJS' },
      { category: 'frontEnd', value: 138322000, x: 'Koa' },
      { category: 'frontEnd', value: 130922000, x: 'Vite' },
      { category: 'frontEnd', value: 130922009, x: 'VitePress' },
      { category: 'frontEnd', value: 130989000, x: 'TypeScript' },
      { category: 'frontEnd', value: 130989003, x: 'stylus' },
      { category: 'frontEnd', value: 130989003, x: 'less' },
      { category: 'frontEnd', value: 130989003, x: 'sass' },
      { category: 'frontEnd', value: 130989010, x: 'fidder' },
      { category: 'frontEnd', value: 130989015, x: 'G2' },
      { category: 'frontEnd', value: 130989010, x: 'mockjs' }
    ]
  },
  message: '更新成功'
}))
Mock.mock('/api/personal/tasks', 'get', () => ({
  status: 0,
  data: {
    tasks: [
      ['2021-05-19', [{ task: '读书看报' }]],

      ['2021-05-20', [{ task: '吃饭打屁' }]]
    ]
  },
  message: '更新成功'
}))
Mock.mock('/api/auth/roles', 'get', genRoles)
Mock.mock('/api/home/swiperInfo', 'get', getSwiperInfo)

// 项目看板数据
const getProjectInfo = {
  status: 0,
  message: '成功',
  data: [
    {
      projectId: '1',
      projectName: '后台管理系统',
      principal: '张三',
      timeConsuming: '20小时',
      status: '开发中',
      taskList: [
        {
          taskId: 1,
          taskName: '导航栏开发',
          developTime: '3工时',
          developMember: '李四',
          taskStatus: 'preparation' // preparation: 准备阶段，development : 开发中， completed: 开发完成， test：测试阶段，released： 待发布
        },
        {
          taskId: 2,
          taskName: '内容页开发',
          developTime: '8工时',
          developMember: '王五',
          taskStatus: 'development'
        },
        {
          taskId: 3,
          taskName: '侧边栏开发',
          developTime: '9工时',
          developMember: '赵六',
          taskStatus: 'completed'
        },
        {
          taskId: 4,
          taskName: '侧边栏开发',
          developTime: '9工时',
          developMember: '赵六',
          taskStatus: 'completed'
        },
        {
          taskId: 5,
          taskName: '侧边栏开发',
          developTime: '9工时',
          developMember: '赵六',
          taskStatus: 'test'
        },
        {
          taskId: 6,
          taskName: '侧边栏开发',
          developTime: '9工时',
          developMember: '赵六',
          taskStatus: 'released'
        }
      ]
    },
    {
      projectId: '2',
      projectName: '学生管理系统',
      principal: '老王',
      timeConsuming: '27小时',
      status: '开发中',
      taskList: [
        {
          taskId: 7,
          taskName: '导航栏开发',
          developTime: '10工时',
          developMember: '李四',
          taskStatus: 'preparation' // preparation: 准备阶段，development : 开发中， completed: 开发完成， test：测试阶段，released： 待发布
        },
        {
          taskId: 8,
          taskName: '内容页开发',
          developTime: '8工时',
          developMember: '王五',
          taskStatus: 'development'
        },
        {
          taskId: 9,
          taskName: '侧边栏开发',
          developTime: '9工时',
          developMember: '赵六',
          taskStatus: 'completed'
        },
        {
          taskId: 10,
          taskName: '侧边栏开发',
          developTime: '9工时',
          developMember: '赵六',
          taskStatus: 'test'
        }
      ]
    },
    {
      projectId: '3',
      projectName: '成绩管理系统',
      principal: '王五',
      timeConsuming: '40小时',
      status: '开发中',
      taskList: [
        {
          taskId: 11,
          taskName: '导航栏开发',
          developTime: '13工时',
          developMember: '李四',
          taskStatus: 'preparation' // preparation: 准备阶段，development : 开发中， completed: 开发完成， test：测试阶段，released： 待发布
        },
        {
          taskId: 12,
          taskName: '内容页开发',
          developTime: '18工时',
          developMember: '王五',
          taskStatus: 'development'
        },
        {
          taskId: 13,
          taskName: '侧边栏开发',
          developTime: '9工时',
          developMember: '赵六',
          taskStatus: 'completed'
        }
      ]
    }
  ]
}

// 获取项目详情
// eslint-disable-next-line no-useless-concat
Mock.mock(RegExp('/api/project/list' + '.*'), 'get', getProjectInfo)
