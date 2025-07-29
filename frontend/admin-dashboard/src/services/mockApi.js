import { logger } from '../utils/logger';

const mockStats = {
  totalShelves: 12,
  totalSlots: 1200,
  occupiedSlots: 750,
  emptySlots: 450,
  materials: 320,
  operationsToday: 128,
};

const mockShelfStatus = {
  shelfId: 'shelf-1',
  totalSlots: 100,
  emptySlots: 40,
  occupiedSlots: 60,
  slots: Array.from({ length: 100 }, (_, i) => ({
    ID: `shelf-1-${Math.floor(i / 10) + 1}-${(i % 10) + 1}`,
    Row: Math.floor(i / 10) + 1,
    Column: (i % 10) + 1,
    Status: Math.random() > 0.6 ? 'occupied' : 'empty',
    MaterialID: Math.random() > 0.6 ? `M${Math.floor(Math.random() * 1000)}` : null,
  })),
};

export const mockInventoryApi = {
  getDashboardStats: () => {
    logger.info('Using mock API for getDashboardStats');
    return Promise.resolve({ data: mockStats });
  },
  getShelfStatus: (shelfId) => {
    logger.info(`Using mock API for getShelfStatus for shelf: ${shelfId}`);
    return Promise.resolve({ data: { ...mockShelfStatus, shelfId } });
  },
  // Mock other API calls as needed
  placeMaterial: (data) => {
    logger.info('Mocking placeMaterial:', data);
    return Promise.resolve({ data: { message: 'Material placed successfully (mocked)' } });
  },
  removeMaterial: (data) => {
    logger.info('Mocking removeMaterial:', data);
    return Promise.resolve({ data: { message: 'Material removed successfully (mocked)' } });
  },
  moveMaterial: (data) => {
    logger.info('Mocking moveMaterial:', data);
    return Promise.resolve({ data: { message: 'Material moved successfully (mocked)' } });
  },
  searchMaterials: (query, limit, offset) => {
    logger.info('Mocking searchMaterials:', { query, limit, offset });
    return Promise.resolve({ data: [] });
  },
};
