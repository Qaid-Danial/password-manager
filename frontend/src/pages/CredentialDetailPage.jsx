import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import Navbar from '../components/Navbar'
import { getEntry, deleteEntry } from '../services/vaultService'

export default function CredentialDetailPage() {
  const { id } = useParams()
  const [entry, setEntry] = useState(null)
  const [showPassword, setShowPassword] = useState(false)
  const [copied, setCopied] = useState(null)   // 'username' | 'password' | null
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [deleteConfirm, setDeleteConfirm] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    getEntry(id)
      .then((res) => setEntry(res.data))
      .catch(() => setError('Failed to load credential.'))
      .finally(() => setLoading(false))
  }, [id])

  const copy = async (text, field) => {
    await navigator.clipboard.writeText(text)
    setCopied(field)
    setTimeout(() => setCopied(null), 2000)
  }

  const handleDelete = async () => {
    try {
      await deleteEntry(id)
      navigate('/dashboard')
    } catch {
      setError('Failed to delete credential.')
      setDeleteConfirm(false)
    }
  }

  if (loading) return <Screen><p className="text-gray-400 text-sm mt-20 text-center">Loading…</p></Screen>
  if (error)   return <Screen><p className="text-red-500 text-sm mt-20 text-center">{error}</p></Screen>

  return (
    <Screen>
      <main className="max-w-xl mx-auto px-4 py-8">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-bold text-gray-800">{entry.site_name}</h2>
          <div className="flex gap-2">
            <button
              onClick={() => navigate(`/vault/${id}/edit`)}
              className="text-sm px-3 py-1.5 bg-green-600 hover:bg-green-700 text-white rounded"
            >
              Edit
            </button>
            <button
              onClick={() => setDeleteConfirm(true)}
              className="text-sm px-3 py-1.5 bg-red-600 hover:bg-red-700 text-white rounded"
            >
              Delete
            </button>
          </div>
        </div>

        <div className="bg-white rounded-lg border border-gray-200 shadow-sm divide-y divide-gray-100">
          {entry.site_url && (
            <Row label="Website">
              <a
                href={entry.site_url.startsWith('http') ? entry.site_url : `https://${entry.site_url}`}
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm text-blue-600 hover:underline break-all"
              >
                {entry.site_url}
              </a>
            </Row>
          )}

          <Row label="Username">
            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-800">{entry.vault_username}</span>
              <CopyBtn onClick={() => copy(entry.vault_username, 'username')} copied={copied === 'username'} />
            </div>
          </Row>

          <Row label="Password">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="text-sm font-mono text-gray-800">
                {showPassword ? entry.password : '••••••••••••'}
              </span>
              <button
                onClick={() => setShowPassword((v) => !v)}
                className="text-xs px-2 py-0.5 bg-gray-100 hover:bg-gray-200 rounded"
              >
                {showPassword ? 'Hide' : 'Show'}
              </button>
              <CopyBtn onClick={() => copy(entry.password, 'password')} copied={copied === 'password'} />
            </div>
          </Row>

          {entry.notes && (
            <Row label="Notes">
              <p className="text-sm text-gray-700 whitespace-pre-wrap">{entry.notes}</p>
            </Row>
          )}

          <Row label="Added">
            <span className="text-sm text-gray-400">
              {entry.created_at ? new Date(entry.created_at).toLocaleString() : '—'}
            </span>
          </Row>
        </div>

        <button
          onClick={() => navigate('/dashboard')}
          className="mt-5 text-sm text-gray-400 hover:text-gray-600"
        >
          ← Back to vault
        </button>
      </main>

      {deleteConfirm && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50 px-4">
          <div className="bg-white rounded-lg p-6 shadow-xl w-full max-w-sm">
            <h3 className="text-base font-semibold text-gray-800 mb-1">Delete credential?</h3>
            <p className="text-sm text-gray-500 mb-5">This action cannot be undone.</p>
            <div className="flex gap-2 justify-end">
              <button
                onClick={() => setDeleteConfirm(false)}
                className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                onClick={handleDelete}
                className="px-4 py-2 text-sm bg-red-600 hover:bg-red-700 text-white rounded"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      )}
    </Screen>
  )
}

function Screen({ children }) {
  return (
    <div className="min-h-screen bg-gray-100">
      <Navbar />
      {children}
    </div>
  )
}

function Row({ label, children }) {
  return (
    <div className="px-5 py-4 flex items-start gap-4">
      <span className="text-xs font-medium text-gray-400 uppercase tracking-wide w-24 flex-shrink-0 pt-0.5">
        {label}
      </span>
      <div className="flex-1 min-w-0">{children}</div>
    </div>
  )
}

function CopyBtn({ onClick, copied }) {
  return (
    <button
      onClick={onClick}
      className="text-xs px-2 py-0.5 bg-gray-100 hover:bg-gray-200 rounded"
    >
      {copied ? '✓' : 'Copy'}
    </button>
  )
}
