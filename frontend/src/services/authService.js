import api from '../utils/api'

export const register = (username, email, password) =>
  api.post('/auth/register', { username, email, password })

export const login = (email, password) =>
  api.post('/auth/login', { email, password })

export const logout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
}

export const getCurrentUser = () => {
  try {
    return JSON.parse(localStorage.getItem('user'))
  } catch {
    return null
  }
}

export const isAuthenticated = () => !!localStorage.getItem('token')
