// CLEAN FILE - No security issues
// Sentinel should report zero findings for this file

import { config } from './config';

/**
 * User service with proper security practices
 */
class UserService {
  constructor(apiClient) {
    this.apiClient = apiClient;
  }

  /**
   * Get user by ID using parameterized query
   * @param {string} userId - User ID
   * @returns {Promise<User>} User object
   */
  async getUserById(userId) {
    // API key from environment, not hardcoded
    const apiKey = process.env.API_KEY;
    
    const response = await this.apiClient.get('/users', {
      params: { id: userId },
      headers: {
        'Authorization': `Bearer ${apiKey}`
      }
    });
    
    return response.data;
  }

  /**
   * Create user with proper validation
   * @param {Object} userData - User data
   * @returns {Promise<User>} Created user
   */
  async createUser(userData) {
    // Validate input
    if (!this.validateUserData(userData)) {
      throw new Error('Invalid user data');
    }

    const response = await this.apiClient.post('/users', userData);
    return response.data;
  }

  /**
   * Validate user data
   * @param {Object} data - Data to validate
   * @returns {boolean} Validation result
   */
  validateUserData(data) {
    return data && 
           typeof data.email === 'string' && 
           typeof data.name === 'string';
  }
}

// Proper configuration loading
const CONFIG = {
  apiUrl: process.env.API_URL || 'https://api.example.com',
  timeout: parseInt(process.env.TIMEOUT, 10) || 5000
};

export default UserService;
export { CONFIG };

