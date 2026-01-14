// Phase 14D: Cost Optimization Dashboard Management
let cacheHitRateChart = null;
let costSavingsChart = null;

// Initialize optimization dashboard
async function initOptimizationDashboard() {
    setupOptimizationEventHandlers();
    await loadOptimizationData();
}

// Setup event handlers for optimization dashboard
function setupOptimizationEventHandlers() {
    document.getElementById('optimization-period-select').addEventListener('change', loadOptimizationData);
    document.getElementById('refresh-optimization').addEventListener('click', loadOptimizationData);
}

// Load optimization data
async function loadOptimizationData() {
    const projectId = getProjectId();
    if (!projectId) {
        showToast('Project ID not found', 'error');
        return;
    }

    const period = document.getElementById('optimization-period-select').value;
    
    try {
        // Load cache metrics
        const cacheMetrics = await getCacheMetrics(projectId);
        updateCacheMetrics(cacheMetrics);

        // Load cost metrics
        const costMetrics = await getCostMetrics(projectId, period);
        updateCostSavings(costMetrics);
    } catch (error) {
        console.error('Failed to load optimization data:', error);
        showToast('Failed to load optimization data: ' + (error.message || 'Unknown error'), 'error');
        // Display zero values on error
        updateCacheMetrics({ success: false });
        updateCostSavings({ success: false });
    }
}

// Update cache metrics display
function updateCacheMetrics(metrics) {
    if (!metrics || !metrics.success) {
        showToast('Failed to load cache metrics', 'error');
        // Set default values
        const elements = ['cache-hit-rate', 'cache-hits', 'cache-misses', 'cache-size'];
        elements.forEach(id => {
            const el = document.getElementById(id);
            if (el) el.textContent = '0';
        });
        return;
    }

    // Update cache hit rate
    const hitRate = ((metrics.hit_rate || 0) * 100).toFixed(1);
    const hitRateEl = document.getElementById('cache-hit-rate');
    if (hitRateEl) hitRateEl.textContent = hitRate + '%';
    
    // Update cache hits
    const hitsEl = document.getElementById('cache-hits');
    if (hitsEl) hitsEl.textContent = formatNumber(metrics.total_hits || 0);
    
    // Update cache misses
    const missesEl = document.getElementById('cache-misses');
    if (missesEl) missesEl.textContent = formatNumber(metrics.total_misses || 0);
    
    // Update cache size
    const sizeEl = document.getElementById('cache-size');
    if (sizeEl) sizeEl.textContent = formatNumber(metrics.cache_size || 0);

    // Update cache hit rate chart
    updateCacheHitRateChart(metrics);
}

// Update cache hit rate chart
function updateCacheHitRateChart(metrics) {
    const ctx = document.getElementById('cache-hit-rate-chart');
    if (!ctx) return;

    if (cacheHitRateChart) {
        cacheHitRateChart.destroy();
    }

    const total = (metrics.total_hits || 0) + (metrics.total_misses || 0);
    const hitRate = metrics.hit_rate || 0;

    cacheHitRateChart = new Chart(ctx, {
        type: 'doughnut',
        data: {
            labels: ['Cache Hits', 'Cache Misses'],
            datasets: [{
                data: [
                    metrics.total_hits || 0,
                    metrics.total_misses || 0
                ],
                backgroundColor: [
                    '#2ecc71', // Green for hits
                    '#e74c3c'  // Red for misses
                ]
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    position: 'bottom'
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            const label = context.label || '';
                            const value = context.parsed || 0;
                            const percentage = total > 0 ? ((value / total) * 100).toFixed(1) : 0;
                            return label + ': ' + formatNumber(value) + ' (' + percentage + '%)';
                        }
                    }
                },
                title: {
                    display: true,
                    text: 'Cache Hit Rate: ' + (hitRate * 100).toFixed(1) + '%'
                }
            }
        }
    });
}

// Update cost savings display
function updateCostSavings(metrics) {
    if (!metrics || !metrics.success) {
        showToast('Failed to load cost metrics', 'error');
        // Set default values
        const elements = ['opt-total-cost', 'opt-cost-savings', 'opt-savings-percentage', 
                         'opt-cache-hit-savings', 'opt-model-selection-savings', 'opt-total-requests'];
        elements.forEach(id => {
            const el = document.getElementById(id);
            if (el) {
                if (id === 'opt-savings-percentage') {
                    el.textContent = '0.0%';
                } else if (id.includes('cost') || id.includes('savings')) {
                    el.textContent = formatCurrency(0);
                } else {
                    el.textContent = '0';
                }
            }
        });
        return;
    }

    // Update total cost
    const totalCostEl = document.getElementById('opt-total-cost');
    if (totalCostEl) totalCostEl.textContent = formatCurrency(metrics.total_cost || 0);
    
    // Update cost savings
    const costSavingsEl = document.getElementById('opt-cost-savings');
    if (costSavingsEl) costSavingsEl.textContent = formatCurrency(metrics.cost_savings || 0);
    
    // Update savings percentage
    const savingsPct = (metrics.savings_percentage || 0).toFixed(1);
    const savingsPctEl = document.getElementById('opt-savings-percentage');
    if (savingsPctEl) savingsPctEl.textContent = savingsPct + '%';
    
    // Update cache hit savings
    const cacheSavingsEl = document.getElementById('opt-cache-hit-savings');
    if (cacheSavingsEl) cacheSavingsEl.textContent = formatCurrency(metrics.cache_hit_savings || 0);
    
    // Update model selection savings
    const modelSavingsEl = document.getElementById('opt-model-selection-savings');
    if (modelSavingsEl) modelSavingsEl.textContent = formatCurrency(metrics.model_selection_savings || 0);
    
    // Update total requests
    const totalRequestsEl = document.getElementById('opt-total-requests');
    if (totalRequestsEl) totalRequestsEl.textContent = formatNumber(metrics.total_requests || 0);

    // Update cost savings chart
    updateCostSavingsChart(metrics);
}

// Update cost savings chart
function updateCostSavingsChart(metrics) {
    const ctx = document.getElementById('cost-savings-chart');
    if (!ctx) return;

    if (costSavingsChart) {
        costSavingsChart.destroy();
    }

    const totalCost = metrics.total_cost || 0;
    const cacheSavings = metrics.cache_hit_savings || 0;
    const modelSavings = metrics.model_selection_savings || 0;
    const otherCost = totalCost - cacheSavings - modelSavings;

    costSavingsChart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels: ['Cost Breakdown'],
            datasets: [
                {
                    label: 'Actual Cost',
                    data: [otherCost],
                    backgroundColor: '#e74c3c'
                },
                {
                    label: 'Cache Hit Savings',
                    data: [cacheSavings],
                    backgroundColor: '#2ecc71'
                },
                {
                    label: 'Model Selection Savings',
                    data: [modelSavings],
                    backgroundColor: '#3498db'
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            scales: {
                x: {
                    stacked: true
                },
                y: {
                    stacked: true,
                    beginAtZero: true,
                    ticks: {
                        callback: function(value) {
                            return '$' + value.toFixed(2);
                        }
                    }
                }
            },
            plugins: {
                legend: {
                    position: 'bottom'
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            return context.dataset.label + ': ' + formatCurrency(context.parsed.y);
                        }
                    }
                },
                title: {
                    display: true,
                    text: 'Cost Breakdown and Savings'
                }
            }
        }
    });
}

