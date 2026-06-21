const BASE = 'http://127.0.0.1:17320'

let token = ''
// 尝试从同目录读取 token（file:// 加载时有效），失败则不启用 token
fetch('token.txt').then(r => r.text()).then(t => { token = t.trim() }).catch(() => {})

async function request(url, options = {}) {
  if (token) {
    options.headers = options.headers || {}
    options.headers['X-Token'] = token
  }
  const res = await fetch(BASE + url, options)
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  const body = await res.json()
  if (body.code !== 0) throw new Error(body.msg || 'error')
  return body.data !== undefined ? body.data : { ok: true, msg: body.msg }
}

export const api = {
  status() { return request('/api/status') },
  readCookie() { return request('/api/cookie') },
  upload(body) {
    const opts = { method: 'POST', headers: { 'Content-Type': 'application/json' } }
    if (body) opts.body = JSON.stringify(body)
    return request('/api/cookie', opts)
  },
  getConfig() { return request('/api/config') },
  setConfig(config) {
    return request('/api/config', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(config)
    })
  },
  testQL(forceCfg) {
    const opts = { method: 'POST', headers: { 'Content-Type': 'application/json' } }
    if (forceCfg) opts.body = JSON.stringify(forceCfg)
    return request('/api/test', opts)
  },
  clearLog() { return request('/api/log', { method: 'DELETE' }) }
}

export function logStream(onLine) {
  const url = BASE + '/api/log/stream' + (token ? '?token=' + encodeURIComponent(token) : '')
  const es = new EventSource(url)
  es.onmessage = (e) => { if (e.data) onLine(e.data) }
  es.onerror = () => { es.close(); setTimeout(() => { logStream(onLine) }, 2000) }
  return es
}
