// Configuration UI Management
let currentEditingId = null;
let providers = [];
let models = {};

// Initialize configuration UI
async function initConfigUI() {
    await loadProviders();
    await loadConfigs();
    setupEventHandlers();
}

// Load providers from API
async function loadProviders() {
    try {
        const response = await getProviders();
        providers = response.providers || [];
        
        const providerSelect = document.getElementById('provider');
        providerSelect.innerHTML = '<option value="">Select a provider</option>';
        
        providers.forEach(provider => {
            const option = document.createElement('option');
            option.value = provider.name;
            option.textContent = provider.display_name;
            providerSelect.appendChild(option);
        });
    } catch (error) {
        showToast('Failed to load providers: ' + error.message, 'error');
    }
}

// Load models for a provider
async function loadModels(provider) {
    if (!provider) {
        const modelSelect = document.getElementById('model');
        modelSelect.innerHTML = '<option value="">Select a provider first</option>';
        modelSelect.disabled = true;
        return;
    }

    try {
        const response = await getModels(provider);
        models[provider] = response.models || [];
        
        const modelSelect = document.getElementById('model');
        modelSelect.innerHTML = '<option value="">Select a model</option>';
        modelSelect.disabled = false;
        
        models[provider].forEach(model => {
            const option = document.createElement('option');
            option.value = model.name;
            option.textContent = `${model.display_name} ($${model.price_per_1k}/1K tokens)`;
            modelSelect.appendChild(option);
        });
    } catch (error) {
        showToast('Failed to load models: ' + error.message, 'error');
    }
}

// Load configurations list
async function loadConfigs() {
    const projectId = getProjectId();
    if (!projectId) {
        document.getElementById('config-list').innerHTML = '<p class="error-message">Project ID not found</p>';
        return;
    }

    try {
        const response = await listConfigs(projectId);
        const configs = response.configs || [];
        displayConfigList(configs);
    } catch (error) {
        document.getElementById('config-list').innerHTML = `<p class="error-message">Failed to load configurations: ${error.message}</p>`;
    }
}

// Display configuration list
function displayConfigList(configs) {
    const container = document.getElementById('config-list');
    
    if (configs.length === 0) {
        container.innerHTML = '<p>No configurations found. Create one using the form above.</p>';
        return;
    }

    const table = document.createElement('table');
    table.className = 'config-table';
    
    const thead = document.createElement('thead');
    thead.innerHTML = `
        <tr>
            <th>Provider</th>
            <th>Model</th>
            <th>Status</th>
            <th>Actions</th>
        </tr>
    `;
    
    const tbody = document.createElement('tbody');
    configs.forEach(config => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${config.provider}</td>
            <td>${config.model}</td>
            <td><span class="badge active">Active</span></td>
            <td class="action-buttons">
                <button class="action-btn edit" onclick="editConfig('${config.id || ''}')">Edit</button>
                <button class="action-btn delete" onclick="deleteConfigConfirm('${config.id || ''}')">Delete</button>
            </td>
        `;
        tbody.appendChild(row);
    });
    
    table.appendChild(thead);
    table.appendChild(tbody);
    container.innerHTML = '';
    container.appendChild(table);
}

// Setup event handlers
function setupEventHandlers() {
    // Provider change -> load models
    document.getElementById('provider').addEventListener('change', (e) => {
        loadModels(e.target.value);
    });

    // Form submission
    document.getElementById('config-form').addEventListener('submit', handleFormSubmit);

    // Test connection button
    document.getElementById('test-connection').addEventListener('click', handleTestConnection);

    // Cancel edit button
    document.getElementById('cancel-edit').addEventListener('click', cancelEdit);
}

// Handle form submission
async function handleFormSubmit(e) {
    e.preventDefault();
    
    if (!validateForm()) {
        return;
    }

    const formData = getFormData();
    const projectId = getProjectId();
    
    if (!projectId) {
        showToast('Project ID not found', 'error');
        return;
    }

    try {
        if (currentEditingId) {
            await updateConfig(currentEditingId, formData);
            showToast('Configuration updated successfully', 'success');
        } else {
            await createConfig(formData);
            showToast('Configuration created successfully', 'success');
        }
        
        resetForm();
        await loadConfigs();
    } catch (error) {
        showToast('Failed to save configuration: ' + error.message, 'error');
    }
}

// Validate form
function validateForm() {
    let isValid = true;
    
    // Clear previous errors
    document.querySelectorAll('.error-message').forEach(el => {
        el.textContent = '';
    });

    const provider = document.getElementById('provider').value;
    const model = document.getElementById('model').value;
    const apiKey = document.getElementById('api-key').value;

    if (!provider) {
        document.getElementById('provider-error').textContent = 'Provider is required';
        isValid = false;
    }

    if (!model) {
        document.getElementById('model-error').textContent = 'Model is required';
        isValid = false;
    }

    if (!apiKey) {
        document.getElementById('api-key-error').textContent = 'API key is required';
        isValid = false;
    } else if (apiKey.length < 10) {
        document.getElementById('api-key-error').textContent = 'API key is too short';
        isValid = false;
    }

    const cacheTTL = parseInt(document.getElementById('cache-ttl').value);
    if (isNaN(cacheTTL) || cacheTTL < 1 || cacheTTL > 8760) {
        document.getElementById('cache-ttl').style.borderColor = '#e74c3c';
        isValid = false;
    } else {
        document.getElementById('cache-ttl').style.borderColor = '';
    }

    return isValid;
}

// Get form data
function getFormData() {
    return {
        provider: document.getElementById('provider').value,
        api_key: document.getElementById('api-key').value,
        model: document.getElementById('model').value,
        key_type: 'user-provided',
        cost_optimization: {
            use_cache: document.getElementById('use-cache').checked,
            cache_ttl_hours: parseInt(document.getElementById('cache-ttl').value) || 24,
            progressive_depth: document.getElementById('progressive-depth').checked,
            max_cost_per_request: parseFloat(document.getElementById('max-cost').value) || undefined,
        }
    };
}

// Handle test connection
async function handleTestConnection() {
    if (!validateForm()) {
        showToast('Please fill in all required fields', 'error');
        return;
    }

    const formData = getFormData();
    const testBtn = document.getElementById('test-connection');
    testBtn.disabled = true;
    testBtn.textContent = 'Testing...';

    try {
        await validateConfig({
            provider: formData.provider,
            api_key: formData.api_key,
            model: formData.model,
        });
        showToast('Connection test successful!', 'success');
    } catch (error) {
        showToast('Connection test failed: ' + error.message, 'error');
    } finally {
        testBtn.disabled = false;
        testBtn.textContent = 'Test Connection';
    }
}

// Edit configuration
async function editConfig(configId) {
    const projectId = getProjectId();
    if (!projectId) {
        showToast('Project ID not found', 'error');
        return;
    }

    try {
        const response = await getConfig(configId);
        const config = response;
        
        // Populate form
        document.getElementById('provider').value = config.provider;
        await loadModels(config.provider);
        setTimeout(() => {
            document.getElementById('model').value = config.model;
        }, 100);
        
        // Note: API key is masked, so we can't populate it
        document.getElementById('api-key').value = '';
        document.getElementById('api-key').placeholder = 'Enter new API key or leave blank to keep existing';
        
        document.getElementById('use-cache').checked = config.cost_optimization?.use_cache ?? true;
        document.getElementById('cache-ttl').value = config.cost_optimization?.cache_ttl_hours ?? 24;
        document.getElementById('progressive-depth').checked = config.cost_optimization?.progressive_depth ?? true;
        document.getElementById('max-cost').value = config.cost_optimization?.max_cost_per_request || '';

        currentEditingId = configId;
        document.getElementById('cancel-edit').style.display = 'inline-block';
        
        // Scroll to form
        document.getElementById('config-form').scrollIntoView({ behavior: 'smooth' });
    } catch (error) {
        showToast('Failed to load configuration: ' + error.message, 'error');
    }
}

// Cancel edit
function cancelEdit() {
    resetForm();
}

// Reset form
function resetForm() {
    document.getElementById('config-form').reset();
    document.getElementById('provider').value = '';
    document.getElementById('model').innerHTML = '<option value="">Select a provider first</option>';
    document.getElementById('model').disabled = true;
    document.getElementById('use-cache').checked = true;
    document.getElementById('cache-ttl').value = 24;
    document.getElementById('progressive-depth').checked = true;
    document.getElementById('max-cost').value = '';
    document.getElementById('api-key').placeholder = '';
    currentEditingId = null;
    document.getElementById('cancel-edit').style.display = 'none';
    
    // Clear errors
    document.querySelectorAll('.error-message').forEach(el => {
        el.textContent = '';
    });
}

// Delete configuration with confirmation
async function deleteConfigConfirm(configId) {
    if (!confirm('Are you sure you want to delete this configuration?')) {
        return;
    }

    try {
        await deleteConfig(configId);
        showToast('Configuration deleted successfully', 'success');
        await loadConfigs();
    } catch (error) {
        showToast('Failed to delete configuration: ' + error.message, 'error');
    }
}

// Toggle password visibility
document.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('.toggle-password').forEach(btn => {
        btn.addEventListener('click', () => {
            const targetId = btn.getAttribute('data-target');
            const input = document.getElementById(targetId);
            if (input.type === 'password') {
                input.type = 'text';
                btn.textContent = 'Hide';
            } else {
                input.type = 'password';
                btn.textContent = 'Show';
            }
        });
    });
});









