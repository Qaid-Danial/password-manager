import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import Navbar from '../components/Navbar'
import CredentialFields from '../components/CredentialFields'
import PasswordStrength from '../components/PasswordStrength'
import PasswordGenerator from '../components/PasswordGenerator'
import { getEntry, updateEntry } from '../services/vaultService'

export default function EditCredentialPage() {
  const { id } = useParams()
  const [form, setForm] = useState({ site_name: '', site_url: '', vault_username: '', password: '', notes: '' })
  const [showGenerator, setShowGenerator] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [fetching, setFetching] = useState(true)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    getEntry(id)
      .then((res) => {
        const e = res.data
        setForm({
          site_name:      e.site_name      ?? '',
          site_url:       e.site_url       ?? '',
          vault_username: e.vault_username ?? '',
          password:       e.password       ?? '',
          notes:          e.notes          ?? '',
        })
      })
      .catch(() => setError('Failed to load credential.'))
      .finally(() => setFetching(false))
  }, [id])

  const onChange = (e) => setForm((f) => ({ ...f, [e.target.name]: e.target.value }))

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await updateEntry(id, form)
      navigate('/dashboard')
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to update credential.')
    } finally {
      setLoading(false)
    }
  }

  if (fetching) {
    return (
      <div className="min-h-screen bg-gray-100">
        <Navbar />
        <p className="text-center text-gray-400 text-sm mt-20">Loading…</p>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-100">
      <Navbar />
      <main className="max-w-xl mx-auto px-4 py-8">
        <h2 className="text-xl font-bold text-gray-800 mb-6">Edit Credential</h2>

        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4 text-sm">
            {error}
          </div>
        )}

        <form
          onSubmit={handleSubmit}
          className="bg-white rounded-lg border border-gray-200 shadow-sm p-6 space-y-4"
        >
          <CredentialFields
            form={form}
            onChange={onChange}
            showPassword={showPassword}
            onTogglePassword={() => setShowPassword((v) => !v)}
          />

          <PasswordStrength password={form.password} />

          <div>
            <button
              type="button"
              onClick={() => setShowGenerator((v) => !v)}
              className="text-xs text-blue-600 hover:underline"
            >
              {showGenerator ? 'Hide' : 'Show'} password generator
            </button>
            {showGenerator && (
              <div className="mt-3">
                <PasswordGenerator
                  onUse={(p) => setForm((f) => ({ ...f, password: p }))}
                />
              </div>
            )}
          </div>

          <div className="flex gap-3 pt-2">
            <button
              type="button"
              onClick={() => navigate(-1)}
              className="flex-1 border border-gray-300 rounded py-2 text-sm hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex-1 bg-blue-600 hover:bg-blue-700 text-white rounded py-2 text-sm font-medium disabled:opacity-50"
            >
              {loading ? 'Saving…' : 'Update Credential'}
            </button>
          </div>
        </form>
      </main>
    </div>
  )
}
