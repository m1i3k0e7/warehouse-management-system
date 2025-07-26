const axios = require('axios');
const config = require('../config');
const logger = require('../utils/logger');

class InventoryAPIService {
  constructor() {
    this.inventoryServiceUrl = config.services.inventory.url;
  }

  async placeMaterial(data) {
    try {
      const response = await axios.post(`${this.inventoryServiceUrl}/materials/place`, data);
      return response.data;
    } catch (error) {
      logger.error('Error placing material via Inventory API:', error.message);
      throw new Error(`Failed to place material: ${error.response?.data?.error || error.message}`);
    }
  }

  async removeMaterial(data) {
    try {
      const response = await axios.post(`${this.inventoryServiceUrl}/materials/remove`, data);
      return response.data;
    } catch (error) {
      logger.error('Error removing material via Inventory API:', error.message);
      throw new Error(`Failed to remove material: ${error.response?.data?.error || error.message}`);
    }
  }

  async moveMaterial(data) {
    try {
      const response = await axios.post(`${this.inventoryServiceUrl}/materials/move`, data);
      return response.data;
    } catch (error) {
      logger.error('Error moving material via Inventory API:', error.message);
      throw new Error(`Failed to move material: ${error.response?.data?.error || error.message}`);
    }
  }

  async getShelfStatus(shelfId) {
    try {
      const response = await axios.get(`${this.inventoryServiceUrl}/shelves/${shelfId}/status`);
      return response.data;
    } catch (error) {
      logger.error(`Error getting shelf status for ${shelfId} via Inventory API:`, error.message);
      throw new Error(`Failed to get shelf status: ${error.response?.data?.error || error.message}`);
    }
  }

  // You can add more methods here for other inventory-related API calls
}

module.exports = InventoryAPIService;
