import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
});

// Add token to requests automatically
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export const authAPI = {
  login: (credentials) => api.post('/login', credentials),
  register: (businessData) => api.post('/register', businessData),
};

export const servicesAPI = {
  create: (serviceData) => api.post('/services', serviceData),
  list: () => api.get('/services'),
  delete: (id) => api.delete(`/services/${id}`),
  getPublic: (businessId) => api.get(`/public/services?business_id=${businessId}`), // Add this line
};

export const slotsAPI = {
  generate: (slotData) => api.post('/slots/generate', slotData),
  list: () => api.get('/slots'),
  getPublic: (businessId, serviceId, date) => 
    api.get(`/public/slots?business_id=${businessId}&service_id=${serviceId}${date ? `&date=${date}` : ''}`),
};

export default api;