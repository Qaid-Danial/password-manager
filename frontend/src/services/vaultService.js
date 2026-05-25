import api from '../utils/api'

export const getVault    = ()         => api.get(`/vault`)
export const getEntry    = (id)       => api.get(`/vault/${id}`)
export const createEntry = (data)     => api.post('/vault', data)
export const updateEntry = (id, data) => api.put(`/vault/${id}`, data)
export const deleteEntry = (id)       => api.delete(`/vault/${id}`)
