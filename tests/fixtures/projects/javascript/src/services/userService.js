// Sample service file with proper patterns
// Tests: camelCase naming, absolute imports pattern

import { formatDate, deepClone } from '@/utils/helpers';
import { validateEmail } from '@/utils/validators';

const API_BASE_URL = '/api/v1';

/**
 * User Service - handles user operations
 */
class UserService {
  constructor(httpClient) {
    this.httpClient = httpClient;
    this.cache = new Map();
  }

  /**
   * Get user by ID
   * @param {string} userId - User ID
   * @returns {Promise<User>} User object
   */
  async getUserById(userId) {
    if (this.cache.has(userId)) {
      return this.cache.get(userId);
    }

    const response = await this.httpClient.get(`${API_BASE_URL}/users/${userId}`);
    const user = response.data;
    
    this.cache.set(userId, user);
    return user;
  }

  /**
   * Create a new user
   * @param {Object} userData - User data
   * @returns {Promise<User>} Created user
   */
  async createUser(userData) {
    if (!validateEmail(userData.email)) {
      throw new Error('Invalid email address');
    }

    const payload = deepClone(userData);
    payload.createdAt = formatDate(new Date());

    const response = await this.httpClient.post(`${API_BASE_URL}/users`, payload);
    return response.data;
  }

  /**
   * Update user
   * @param {string} userId - User ID
   * @param {Object} updates - Updates to apply
   * @returns {Promise<User>} Updated user
   */
  async updateUser(userId, updates) {
    const response = await this.httpClient.patch(
      `${API_BASE_URL}/users/${userId}`,
      updates
    );
    
    // Invalidate cache
    this.cache.delete(userId);
    
    return response.data;
  }

  /**
   * Delete user
   * @param {string} userId - User ID
   * @returns {Promise<void>}
   */
  async deleteUser(userId) {
    await this.httpClient.delete(`${API_BASE_URL}/users/${userId}`);
    this.cache.delete(userId);
  }
}

export default UserService;












