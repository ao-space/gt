import request from '@/utils/request'

const personalApi = {
  basicInfo: '/api/setting/basicInfo',
  personalTags: '/api/personal/tags'
}

class Service {
  /**
   * @description POST 设置基本信息
   */
  static postSetBasicInfo(data: any) {
    return request({
      url: personalApi.basicInfo,
      method: 'post',
      json: true,
      data
    }).then((res) => {
      if (res.status === 0) {
        return Promise.resolve(res)
      }
      return Promise.reject(res)
    })
  }

  /**
   * @description Get 获取tags
   */
  static getPersonTags() {
    return request({
      url: personalApi.personalTags,
      method: 'get',
      json: true
    }).then((res) => {
      if (res.status === 0) {
        return Promise.resolve(res)
      }
      return Promise.reject(res)
    })
  }
}
export default Service
