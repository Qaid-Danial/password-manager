import { Navigate, Outlet } from 'react-router-dom'
import { isAuthenticated, getCurrentUser } from '../services/authService'

export default function ProtectedRoute({ adminOnly = false }) {
  if (!isAuthenticated()) {
    return <Navigate to="/login" replace />
  }

  if (adminOnly) {
    const user = getCurrentUser()
    if (!user || user.role !== 'admin') {
      return <Navigate to="/dashboard" replace />
    }
  }

  return <Outlet />
}
