import { useState, useEffect } from 'react'
import { generatePassword } from '../utils/passwordGenerator'
import PasswordStrength from './PasswordStrength'

export default function PasswordGenerator({ onUse }) {
  const [password, setPassword] = useState('')
  const [length, setLength] = useState(16)
  const [opts, setOpts] = useState({ uppercase: true, lowercase: true, numbers: true, symbols: true })
  const [copied, setCopied] = useState(false)

  const generate = () => setPassword(generatePassword({ length, ...opts }))

  useEffect(() => { generate() }, [length, opts])

  const copy = async () => {
    await navigator.clipboard.writeText(password)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const toggle = (key) => setOpts((o) => ({ ...o, [key]: !o[key] }))

  return (
    <div className="border border-gray-200 rounded-lg p-4 bg-gray-50 space-y-3">
      <p className="text-xs font-semibold text-gray-600 uppercase tracking-wide">Password Generator</p>

      <div className="flex items-center gap-2">
        <code className="flex-1 bg-white border border-gray-200 rounded px-3 py-2 text-sm font-mono truncate select-all">
          {password}
        </code>
        <button
          type="button"
          onClick={copy}
          className="text-xs px-3 py-2 bg-gray-200 hover:bg-gray-300 rounded whitespace-nowrap"
        >
          {copied ? '✓ Copied' : 'Copy'}
        </button>
      </div>

      <PasswordStrength password={password} />

      <div>
        <label className="text-xs text-gray-500">Length: <span className="font-medium text-gray-700">{length}</span></label>
        <input
          type="range" min="8" max="64" value={length}
          onChange={(e) => setLength(Number(e.target.value))}
          className="w-full mt-1 accent-blue-600"
        />
      </div>

      <div className="flex flex-wrap gap-4 text-xs text-gray-600">
        {Object.entries(opts).map(([key, val]) => (
          <label key={key} className="flex items-center gap-1.5 cursor-pointer">
            <input
              type="checkbox"
              checked={val}
              onChange={() => toggle(key)}
              className="rounded"
            />
            {key.charAt(0).toUpperCase() + key.slice(1)}
          </label>
        ))}
      </div>

      <div className="flex gap-2 pt-1">
        <button
          type="button"
          onClick={generate}
          className="text-xs px-3 py-1.5 bg-gray-200 hover:bg-gray-300 rounded"
        >
          Regenerate
        </button>
        {onUse && (
          <button
            type="button"
            onClick={() => onUse(password)}
            className="text-xs px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded"
          >
            Use this password
          </button>
        )}
      </div>
    </div>
  )
}
