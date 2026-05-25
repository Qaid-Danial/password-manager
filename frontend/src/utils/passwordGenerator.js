const UPPERCASE = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
const LOWERCASE = 'abcdefghijklmnopqrstuvwxyz'
const NUMBERS   = '0123456789'
const SYMBOLS   = '!@#$%^&*()_+-=[]{}|;:,.<>?'

/**
 * Generates a cryptographically random password.
 * Uses crypto.getRandomValues — never Math.random — so the output
 * is suitable for use as an actual credential.
 */
export function generatePassword({
  length = 16,
  uppercase = true,
  lowercase = true,
  numbers = true,
  symbols = true,
} = {}) {
  let charset = ''
  if (uppercase) charset += UPPERCASE
  if (lowercase) charset += LOWERCASE
  if (numbers)   charset += NUMBERS
  if (symbols)   charset += SYMBOLS
  if (!charset)  charset = LOWERCASE

  const arr = new Uint32Array(length)
  crypto.getRandomValues(arr)
  return Array.from(arr, (n) => charset[n % charset.length]).join('')
}

/**
 * Scores a password from 0–5 based on length, case, digits, and symbols.
 * Returns { score, label, color } for use with the strength bar component.
 */
export function calculateStrength(password) {
  if (!password) return { score: 0, label: '', color: 'bg-gray-200' }

  let score = 0
  if (password.length >= 8)  score++
  if (password.length >= 12) score++
  if (/[A-Z]/.test(password)) score++
  if (/[0-9]/.test(password)) score++
  if (/[^A-Za-z0-9]/.test(password)) score++

  const levels = [
    { score: 0, label: 'Very Weak',   color: 'bg-red-500' },
    { score: 1, label: 'Weak',        color: 'bg-orange-400' },
    { score: 2, label: 'Fair',        color: 'bg-yellow-400' },
    { score: 3, label: 'Good',        color: 'bg-blue-400' },
    { score: 4, label: 'Strong',      color: 'bg-green-500' },
    { score: 5, label: 'Very Strong', color: 'bg-green-700' },
  ]

  return levels[score] ?? levels[levels.length - 1]
}
