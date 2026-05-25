import { Link, useNavigate } from 'react-router-dom'
import { logout, getCurrentUser } from '../services/authService'

export default function Navbar() {
  const navigate = useNavigate()
  const user = getCurrentUser()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <nav className="bg-gray-900 text-white px-6 py-4 flex items-center justify-between shadow-md">
      <Link to="/dashboard" className="text-lg font-bold tracking-tight">
        SecureVault
      </Link>

      <div className="flex items-center gap-5 text-sm">
        <Link to="/dashboard" className="text-gray-300 hover:text-white">
          Vault
        </Link>
        {user?.role === 'admin' && (
          <Link to="/admin/audit-logs" className="text-gray-300 hover:text-white">
            Audit Logs
          </Link>
        )}
        <span className="text-gray-500 text-xs border-l border-gray-700 pl-4">
          {user?.username}
        </span>
        <button
          onClick={handleLogout}
          className="bg-red-700 hover:bg-red-600 px-3 py-1.5 rounded text-xs font-medium"
        >
          Logout
        </button>
      </div>
    </nav>
  )
}
