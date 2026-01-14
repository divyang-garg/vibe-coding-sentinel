// Main Application Logic
let currentProjectId = null;

// Initialize application
document.addEventListener('DOMContentLoaded', () => {
    initializeApp();
});

async function initializeApp() {
    // Get project ID from URL or localStorage
    currentProjectId = getProjectId();
    
    if (!currentProjectId) {
        showToast('Project ID not found. Please provide project_id in URL or localStorage.', 'error');
        return;
    }

    // Setup navigation
    setupNavigation();

    // Initialize tabs
    const hash = window.location.hash || '#config';
    showTab(hash.substring(1));

    // Initialize tab-specific functionality
    if (hash === '#usage') {
        await initUsageDashboard();
    } else if (hash === '#optimization') {
        await initOptimizationDashboard();
    } else {
        await initConfigUI();
    }
}

// Get project ID from URL params or localStorage
function getProjectId() {
    // Try URL parameter first
    const urlParams = new URLSearchParams(window.location.search);
    const projectId = urlParams.get('project_id');
    if (projectId) {
        localStorage.setItem('sentinel_project_id', projectId);
        return projectId;
    }

    // Try localStorage
    return localStorage.getItem('sentinel_project_id');
}

// Setup navigation
function setupNavigation() {
    document.querySelectorAll('.nav-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const tab = e.target.getAttribute('data-tab');
            showTab(tab);
            window.location.hash = tab;
        });
    });

    // Handle hash changes
    window.addEventListener('hashchange', () => {
        const hash = window.location.hash.substring(1) || 'config';
        showTab(hash);
    });
}

// Show tab
function showTab(tabName) {
    // Update nav buttons
    document.querySelectorAll('.nav-btn').forEach(btn => {
        if (btn.getAttribute('data-tab') === tabName) {
            btn.classList.add('active');
        } else {
            btn.classList.remove('active');
        }
    });

    // Show/hide tab content
    document.querySelectorAll('.tab-content').forEach(content => {
        if (content.id === `${tabName}-section`) {
            content.classList.add('active');
            
            // Initialize tab if needed
            if (tabName === 'usage' && !trendsChart) {
                initUsageDashboard();
            } else if (tabName === 'config' && providers.length === 0) {
                initConfigUI();
            } else if (tabName === 'optimization') {
                initOptimizationDashboard();
            }
        } else {
            content.classList.remove('active');
        }
    });
}

// Utility functions
function formatNumber(num) {
    return new Intl.NumberFormat('en-US').format(num);
}

function formatCurrency(amount) {
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD',
        minimumFractionDigits: 2,
        maximumFractionDigits: 4,
    }).format(amount);
}

function formatDate(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
    });
}

function maskAPIKey(apiKey) {
    if (!apiKey || apiKey.length <= 4) {
        return '****';
    }
    return '****' + apiKey.slice(-4);
}

// Toast notification system
function showToast(message, type = 'info') {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;
    
    container.appendChild(toast);
    
    // Auto-remove after 5 seconds
    setTimeout(() => {
        toast.style.animation = 'slideOut 0.3s ease-out';
        setTimeout(() => {
            container.removeChild(toast);
        }, 300);
    }, 5000);
}

// Global error handler
window.addEventListener('error', (event) => {
    console.error('Global error:', event.error);
    showToast('An error occurred. Please check the console for details.', 'error');
});

// Handle unhandled promise rejections
window.addEventListener('unhandledrejection', (event) => {
    console.error('Unhandled promise rejection:', event.reason);
    showToast('An error occurred: ' + (event.reason?.message || 'Unknown error'), 'error');
});

