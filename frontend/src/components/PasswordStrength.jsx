import { calculateStrength } from '../utils/passwordGenerator'

export default function PasswordStrength({ password }) {
  if (!password) return null

  const { score, label, color } = calculateStrength(password)

  return (
    <div className="mt-1.5">
      <div className="flex gap-1">
        {Array.from({ length: 5 }).map((_, i) => (
          <div
            key={i}
            className={`h-1.5 flex-1 rounded-full ${i < score ? color : 'bg-gray-200'}`}
          />
        ))}
      </div>
      <p className="text-xs text-gray-500 mt-1">{label}</p>
    </div>
  )
}
