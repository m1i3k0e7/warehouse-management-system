import axios from 'axios';
import { API_BASE_URL } from '../utils/constants';
import { logger } from '../utils/logger';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    logger.error('API Error:', error.response?.data || error.message);
    return Promise.reject(error);
  }
);

export const inventoryApi = {
  getShelfStatus: (shelfId) => api.get(`/shelves/${shelfId}/status`),
  placeMaterial: (data) => api.post('/materials/place', data),
  removeMaterial: (data) => api.post('/materials/remove', data),
  moveMaterial: (data) => api.post('/materials/move', data),
  searchMaterials: (query, limit, offset) => api.get(`/materials/search?q=${query}&limit=${limit}&offset=${offset}`),
  // Add other inventory-related API calls here
};

export default api;
