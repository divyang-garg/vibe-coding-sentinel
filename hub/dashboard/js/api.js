// API Client for Sentinel Hub
const API_BASE_URL = window.location.origin + '/api/v1';

// Helper function for API calls
async function apiCall(method, endpoint, data = null) {
    const url = API_BASE_URL + endpoint;
    const options = {
        method: method,
        headers: {
            'Content-Type': 'application/json',
        },
    };

    // Add API key from localStorage if available
    const apiKey = localStorage.getItem('sentinel_api_key');
    if (apiKey) {
        options.headers['X-API-Key'] = apiKey;
    }

    if (data) {
        options.body = JSON.stringify(data);
    }

    try {
        const response = await fetch(url, options);
        const responseData = await response.json();

        if (!response.ok) {
            throw new Error(responseData.error || `HTTP ${response.status}: ${response.statusText}`);
        }

        return responseData;
    } catch (error) {
        console.error('API call failed:', error);
        throw error;
    }
}

// Configuration API functions
async function createConfig(config) {
    return apiCall('POST', '/llm/config', config);
}

async function getConfig(id) {
    return apiCall('GET', `/llm/config/${id}`);
}

async function updateConfig(id, config) {
    return apiCall('PUT', `/llm/config/${id}`, config);
}

async function deleteConfig(id) {
    return apiCall('DELETE', `/llm/config/${id}`);
}

async function listConfigs(projectId) {
    return apiCall('GET', `/llm/config/project/${projectId}`);
}

async function getProviders() {
    return apiCall('GET', '/llm/providers');
}

async function getModels(provider) {
    return apiCall('GET', `/llm/models/${provider}`);
}

async function validateConfig(config) {
    return apiCall('POST', '/llm/config/validate', config);
}

// Usage API functions
async function getUsageReport(projectId, startDate, endDate) {
    const params = new URLSearchParams({
        project_id: projectId,
        start_date: startDate,
        end_date: endDate,
    });
    return apiCall('GET', `/llm/usage/report?${params}`);
}

async function getUsageStats(projectId, period) {
    const params = new URLSearchParams({
        project_id: projectId,
        period: period,
    });
    return apiCall('GET', `/llm/usage/stats?${params}`);
}

async function getCostBreakdown(projectId, period) {
    const params = new URLSearchParams({
        project_id: projectId,
        period: period,
    });
    return apiCall('GET', `/llm/usage/cost-breakdown?${params}`);
}

async function getUsageTrends(projectId, period, groupBy = 'day') {
    const params = new URLSearchParams({
        project_id: projectId,
        period: period,
        group_by: groupBy,
    });
    return apiCall('GET', `/llm/usage/trends?${params}`);
}

// Phase 14D: Cost Optimization Metrics API functions
async function getCacheMetrics(projectId) {
    const params = new URLSearchParams({
        project_id: projectId,
    });
    return apiCall('GET', `/metrics/cache?${params}`);
}

async function getCostMetrics(projectId, period = 'monthly') {
    const params = new URLSearchParams({
        project_id: projectId,
        period: period,
    });
    return apiCall('GET', `/metrics/cost?${params}`);
}

