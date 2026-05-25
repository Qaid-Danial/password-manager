/**
 * Shared form fields used by both AddCredentialPage and EditCredentialPage.
 * Keeps form markup in one place without coupling the two pages together.
 */
export default function CredentialFields({ form, onChange, showPassword, onTogglePassword }) {
  return (
    <>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Site Name <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          name="site_name"
          value={form.site_name}
          onChange={onChange}
          required
          maxLength={255}
          placeholder="GitHub"
          className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-blue-500"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Site URL</label>
        <input
          type="text"
          name="site_url"
          value={form.site_url}
          onChange={onChange}
          maxLength={500}
          placeholder="https://github.com"
          className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-blue-500"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Username <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          name="vault_username"
          value={form.vault_username}
          onChange={onChange}
          required
          placeholder="alice@example.com"
          className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-blue-500"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Password <span className="text-red-500">*</span>
        </label>
        <div className="relative">
          <input
            type={showPassword ? 'text' : 'password'}
            name="password"
            value={form.password}
            onChange={onChange}
            required
            placeholder="••••••••"
            className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-blue-500 pr-16"
          />
          <button
            type="button"
            onClick={onTogglePassword}
            className="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-gray-500 hover:text-gray-800"
          >
            {showPassword ? 'Hide' : 'Show'}
          </button>
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Notes</label>
        <textarea
          name="notes"
          value={form.notes}
          onChange={onChange}
          rows={3}
          placeholder="Optional notes…"
          className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-blue-500 resize-none"
        />
      </div>
    </>
  )
}
