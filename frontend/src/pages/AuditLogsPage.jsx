import { useState, useEffect } from 'react'
import Navbar from '../components/Navbar'
import { getAuditLogs } from '../services/adminService'

const ACTION_BADGE = {
  login_failure:     'bg-red-100 text-red-700',
  delete_credential: 'bg-orange-100 text-orange-700',
  register:          'bg-green-100 text-green-700',
  login_success:     'bg-blue-100 text-blue-700',
  add_credential:    'bg-purple-100 text-purple-700',
  update_credential: 'bg-yellow-100 text-yellow-700',
  view_credential:   'bg-gray-100 text-gray-600',
}

const PAGE_SIZE = 20

export default function AuditLogsPage() {
  const [logs, setLogs] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    setLoading(true)
    getAuditLogs(page, PAGE_SIZE)
      .then((res) => {
        setLogs(res.data.logs || [])
        setTotal(res.data.total || 0)
      })
      .catch(() => setError('Failed to load audit logs.'))
      .finally(() => setLoading(false))
  }, [page])

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  return (
    <div className="min-h-screen bg-gray-100">
      <Navbar />

      <main className="max-w-6xl mx-auto px-4 py-8">
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-xl font-bold text-gray-800">Audit Logs</h2>
          <span className="text-sm text-gray-400">{total} total records</span>
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4 text-sm">
            {error}
          </div>
        )}

        <div className="bg-white rounded-lg border border-gray-200 shadow-sm overflow-x-auto">
          <table className="w-full text-sm min-w-[700px]">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <Th>Timestamp</Th>
                <Th>Action</Th>
                <Th>User ID</Th>
                <Th>IP Address</Th>
                <Th>User Agent</Th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {loading ? (
                <tr>
                  <td colSpan={5} className="px-4 py-10 text-center text-gray-400">
                    Loading…
                  </td>
                </tr>
              ) : logs.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-4 py-10 text-center text-gray-400">
                    No audit records yet.
                  </td>
                </tr>
              ) : (
                logs.map((log) => (
                  <tr key={log.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-gray-500 whitespace-nowrap">
                      {new Date(log.created_at).toLocaleString()}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${ACTION_BADGE[log.action] ?? 'bg-gray-100 text-gray-700'}`}>
                        {log.action}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-gray-400 font-mono text-xs">
                      {log.user_id ? `${log.user_id.slice(0, 8)}…` : '—'}
                    </td>
                    <td className="px-4 py-3 text-gray-500">{log.ip_address}</td>
                    <td className="px-4 py-3 text-gray-400 text-xs max-w-xs truncate">
                      {log.user_agent || '—'}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {totalPages > 1 && (
          <div className="flex items-center justify-between mt-4 text-sm">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              className="px-4 py-2 border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            <span className="text-gray-500">
              Page {page} of {totalPages}
            </span>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="px-4 py-2 border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        )}
      </main>
    </div>
  )
}

function Th({ children }) {
  return (
    <th className="px-4 py-3 text-left text-xs font-semibold text-gray-500 uppercase tracking-wide">
      {children}
    </th>
  )
}
