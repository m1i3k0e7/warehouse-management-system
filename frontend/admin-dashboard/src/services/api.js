import axios from 'axios';
import { API_BASE_URL } from '../utils/constants';
import { logger } from '../utils/logger';
import { mockInventoryApi } from './mockApi';

const useMockApi = process.env.REACT_APP_USE_MOCK_API === 'true';

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

const realInventoryApi = {
  getDashboardStats: () => api.get('/stats/dashboard'), // Assuming this endpoint exists
  getShelfStatus: (shelfId) => api.get(`/shelves/${shelfId}/status`),
  placeMaterial: (data) => api.post('/materials/place', data),
  removeMaterial: (data) => api.post('/materials/remove', data),
  moveMaterial: (data) => api.post('/materials/move', data),
  searchMaterials: (query, limit, offset) => api.get(`/materials/search?q=${query}&limit=${limit}&offset=${offset}`),
};

export const inventoryApi = useMockApi ? mockInventoryApi : realInventoryApi;

export default api;
