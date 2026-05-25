import api from '../utils/api'

export const getAuditLogs = (page = 1, pageSize = 20) =>
  api.get('/admin/audit-logs', { params: { page, page_size: pageSize } })
