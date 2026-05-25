import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import Navbar from '../components/Navbar'
import { getVault, deleteEntry } from '../services/vaultService'

export default function DashboardPage() {
  const [entries, setEntries] = useState([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [deleteId, setDeleteId] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    getVault()
      .then((res) => setEntries(res.data.entries || []))
      .catch(() => setError('Failed to load vault entries.'))
      .finally(() => setLoading(false))
  }, [])

  const filtered = entries.filter(
    (e) =>
      e.site_name.toLowerCase().includes(search.toLowerCase()) ||
      e.vault_username.toLowerCase().includes(search.toLowerCase())
  )

  const confirmDelete = async () => {
    try {
      await deleteEntry(deleteId)
      setEntries((prev) => prev.filter((e) => e.id !== deleteId))
    } catch {
      setError('Failed to delete entry.')
    } finally {
      setDeleteId(null)
    }
  }

  return (
    <div className="min-h-screen bg-gray-100">
      <Navbar />

      <main className="max-w-5xl mx-auto px-4 py-8">
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-xl font-bold text-gray-800">
            Your Vault
            <span className="ml-2 text-sm font-normal text-gray-400">({entries.length})</span>
          </h2>
          <Link
            to="/vault/new"
            className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded text-sm font-medium"
          >
            + Add Credential
          </Link>
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4 text-sm">
            {error}
          </div>
        )}

        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search by site name or username…"
          className="w-full border border-gray-300 rounded px-4 py-2 text-sm mb-6 focus:outline-none focus:border-blue-500 bg-white"
        />

        {loading ? (
          <p className="text-gray-400 text-sm">Loading…</p>
        ) : filtered.length === 0 ? (
          <div className="text-center py-20 text-gray-400">
            <p className="text-5xl mb-3">&#128274;</p>
            <p className="text-sm">
              {search ? 'No results found.' : 'No credentials yet — add your first one.'}
            </p>
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {filtered.map((entry) => (
              <VaultCard
                key={entry.id}
                entry={entry}
                onDelete={() => setDeleteId(entry.id)}
              />
            ))}
          </div>
        )}
      </main>

      {deleteId && (
        <Modal
          title="Delete credential?"
          message="This action cannot be undone."
          onConfirm={confirmDelete}
          onCancel={() => setDeleteId(null)}
          confirmLabel="Delete"
          confirmClass="bg-red-600 hover:bg-red-700"
        />
      )}
    </div>
  )
}

function VaultCard({ entry, onDelete }) {
  const navigate = useNavigate()
  const [copied, setCopied] = useState(false)

  const copyPassword = async () => {
    await navigator.clipboard.writeText(entry.password)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4 shadow-sm hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between mb-2">
        <div className="min-w-0">
          <h3 className="font-semibold text-gray-800 truncate">{entry.site_name}</h3>
          {entry.site_url && (
            <a
              href={entry.site_url.startsWith('http') ? entry.site_url : `https://${entry.site_url}`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-xs text-blue-500 hover:underline truncate block"
            >
              {entry.site_url}
            </a>
          )}
        </div>
      </div>

      <p className="text-sm text-gray-600 truncate mb-1">{entry.vault_username}</p>
      <p className="text-sm text-gray-300 font-mono mb-3 tracking-widest">••••••••</p>

      <div className="flex gap-1.5 flex-wrap">
        <button
          onClick={copyPassword}
          className="text-xs px-2 py-1 bg-gray-100 hover:bg-gray-200 rounded"
        >
          {copied ? '✓ Copied' : 'Copy'}
        </button>
        <button
          onClick={() => navigate(`/vault/${entry.id}`)}
          className="text-xs px-2 py-1 bg-blue-50 text-blue-700 hover:bg-blue-100 rounded"
        >
          View
        </button>
        <button
          onClick={() => navigate(`/vault/${entry.id}/edit`)}
          className="text-xs px-2 py-1 bg-green-50 text-green-700 hover:bg-green-100 rounded"
        >
          Edit
        </button>
        <button
          onClick={onDelete}
          className="text-xs px-2 py-1 bg-red-50 text-red-700 hover:bg-red-100 rounded"
        >
          Delete
        </button>
      </div>
    </div>
  )
}

function Modal({ title, message, onConfirm, onCancel, confirmLabel, confirmClass }) {
  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50 px-4">
      <div className="bg-white rounded-lg p-6 shadow-xl w-full max-w-sm">
        <h3 className="text-base font-semibold text-gray-800 mb-1">{title}</h3>
        <p className="text-sm text-gray-500 mb-5">{message}</p>
        <div className="flex gap-2 justify-end">
          <button
            onClick={onCancel}
            className="px-4 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            className={`px-4 py-2 text-sm text-white rounded ${confirmClass}`}
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  )
}
