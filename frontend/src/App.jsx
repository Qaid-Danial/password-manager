import { Routes, Route, Navigate } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import DashboardPage from './pages/DashboardPage'
import AddCredentialPage from './pages/AddCredentialPage'
import EditCredentialPage from './pages/EditCredentialPage'
import CredentialDetailPage from './pages/CredentialDetailPage'
import AuditLogsPage from './pages/AuditLogsPage'
import ProtectedRoute from './components/ProtectedRoute'

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />

      <Route element={<ProtectedRoute />}>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/vault/new" element={<AddCredentialPage />} />
        <Route path="/vault/:id" element={<CredentialDetailPage />} />
        <Route path="/vault/:id/edit" element={<EditCredentialPage />} />

        <Route element={<ProtectedRoute adminOnly />}>
          <Route path="/admin/audit-logs" element={<AuditLogsPage />} />
        </Route>
      </Route>

      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  )
}
