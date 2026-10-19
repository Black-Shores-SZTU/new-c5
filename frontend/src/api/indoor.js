import axios from 'axios'
import { allLocations, buildings, buildRoute, floors } from '../mock/indoorData'

// ============================================================
// 后端未联调：以下函数当前返回本地 mock。
// TODO: 联调时取消 axios.get 注释，并移除本地 mock 返回。
// ============================================================

export async function fetchFloors() {
  // TODO: return (await axios.get('/api/indoor/floors')).data
  return Promise.resolve(floors)
}

export async function fetchBuildings() {
  // TODO: return (await axios.get('/api/indoor/buildings')).data
  return Promise.resolve(buildings)
}

export async function fetchLocations() {
  // TODO: return (await axios.get('/api/indoor/locations')).data
  return Promise.resolve(allLocations)
}

export async function fetchRoute(fromId, toId, elevatorOnly = false) {
  // TODO: return (await axios.get('/api/indoor/route', { params: { fromId, toId, elevatorOnly } })).data
  const from = allLocations.find((location) => location.id === fromId)
  const to = allLocations.find((location) => location.id === toId)
  return Promise.resolve(buildRoute(from, to, elevatorOnly))
}

// 预留 axios 实例，后续统一配置 baseURL / 超时 / 拦截器
export const indoorRequest = axios.create({
  baseURL: '/api/indoor',
  timeout: 10000
})
