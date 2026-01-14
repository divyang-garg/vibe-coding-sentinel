// Usage Dashboard Management
let providerChart = null;
let modelChart = null;
let trendsChart = null;

// Initialize usage dashboard
async function initUsageDashboard() {
    setupUsageEventHandlers();
    await loadUsageData();
}

// Setup event handlers for usage dashboard
function setupUsageEventHandlers() {
    document.getElementById('period-select').addEventListener('change', loadUsageData);
    document.getElementById('refresh-usage').addEventListener('click', loadUsageData);
    document.getElementById('export-csv').addEventListener('click', exportToCSV);
    document.getElementById('export-json').addEventListener('click', exportToJSON);
}

// Load usage data
async function loadUsageData() {
    const projectId = getProjectId();
    if (!projectId) {
        showToast('Project ID not found', 'error');
        return;
    }

    const period = document.getElementById('period-select').value;
    
    try {
        // Load stats
        const statsResponse = await getUsageStats(projectId, period);
        updateUsageOverview(statsResponse);

        // Load cost breakdown
        const breakdownResponse = await getCostBreakdown(projectId, period);
        updateCostCharts(breakdownResponse);

        // Load trends
        const trendsResponse = await getUsageTrends(projectId, period, 'day');
        updateTrendsChart(trendsResponse);

        // Load usage report for table
        const endDate = new Date().toISOString().split('T')[0];
        let startDate;
        switch (period) {
            case 'daily':
                startDate = new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString().split('T')[0];
                break;
            case 'weekly':
                startDate = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
                break;
            case 'monthly':
                startDate = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
                break;
            case 'yearly':
                startDate = new Date(Date.now() - 365 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
                break;
            default:
                startDate = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
        }

        const reportResponse = await getUsageReport(projectId, startDate, endDate);
        updateUsageTable(reportResponse);
    } catch (error) {
        showToast('Failed to load usage data: ' + error.message, 'error');
    }
}

// Update usage overview cards
function updateUsageOverview(stats) {
    document.getElementById('total-tokens').textContent = formatNumber(stats.total_tokens || 0);
    document.getElementById('total-cost').textContent = formatCurrency(stats.total_cost || 0);
    document.getElementById('average-cost').textContent = formatCurrency(stats.average_cost || 0);
    document.getElementById('total-requests').textContent = formatNumber(stats.total_requests || 0);
}

// Update cost charts
function updateCostCharts(breakdown) {
    // Provider pie chart
    const providerCtx = document.getElementById('provider-chart').getContext('2d');
    if (providerChart) {
        providerChart.destroy();
    }

    const providerLabels = Object.keys(breakdown.by_provider || {});
    const providerData = Object.values(breakdown.by_provider || {});

    providerChart = new Chart(providerCtx, {
        type: 'pie',
        data: {
            labels: providerLabels,
            datasets: [{
                data: providerData,
                backgroundColor: [
                    '#3498db',
                    '#2ecc71',
                    '#e74c3c',
                    '#f39c12',
                    '#9b59b6',
                ]
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    position: 'bottom',
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            const label = context.label || '';
                            const value = context.parsed || 0;
                            const percentage = breakdown.provider_percentages?.[label] || 0;
                            return `${label}: ${formatCurrency(value)} (${percentage.toFixed(1)}%)`;
                        }
                    }
                }
            }
        }
    });

    // Model bar chart
    const modelCtx = document.getElementById('model-chart').getContext('2d');
    if (modelChart) {
        modelChart.destroy();
    }

    const modelLabels = Object.keys(breakdown.by_model || {}).slice(0, 10); // Top 10
    const modelData = modelLabels.map(label => breakdown.by_model[label] || 0);

    modelChart = new Chart(modelCtx, {
        type: 'bar',
        data: {
            labels: modelLabels,
            datasets: [{
                label: 'Cost',
                data: modelData,
                backgroundColor: '#3498db',
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: false,
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            return formatCurrency(context.parsed.y);
                        }
                    }
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        callback: function(value) {
                            return formatCurrency(value);
                        }
                    }
                }
            }
        }
    });
}

// Update trends chart
function updateTrendsChart(trendsResponse) {
    const trendsCtx = document.getElementById('trends-chart').getContext('2d');
    if (trendsChart) {
        trendsChart.destroy();
    }

    const trends = trendsResponse.trends || [];
    const labels = trends.map(t => t.label);
    const costs = trends.map(t => t.cost);

    trendsChart = new Chart(trendsCtx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'Daily Cost',
                data: costs,
                borderColor: '#3498db',
                backgroundColor: 'rgba(52, 152, 219, 0.1)',
                tension: 0.4,
                fill: true,
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: false,
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            return formatCurrency(context.parsed.y);
                        }
                    }
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        callback: function(value) {
                            return formatCurrency(value);
                        }
                    }
                }
            }
        }
    });
}

// Update usage table
function updateUsageTable(report) {
    const tbody = document.querySelector('#usage-table tbody');
    tbody.innerHTML = '';

    if (!report.daily_usage || report.daily_usage.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" class="loading">No usage data available</td></tr>';
        return;
    }

    // Flatten daily usage with provider/model details (simplified - in real implementation, you'd need to query detailed data)
    report.daily_usage.forEach(daily => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${daily.date}</td>
            <td>${Object.keys(report.usage_by_provider || {}).join(', ') || 'N/A'}</td>
            <td>${Object.keys(report.usage_by_model || {}).join(', ') || 'N/A'}</td>
            <td>${formatNumber(daily.tokens)}</td>
            <td>${formatCurrency(daily.cost)}</td>
        `;
        tbody.appendChild(row);
    });

    // Add sorting functionality
    setupTableSorting();
}

// Setup table sorting
function setupTableSorting() {
    const headers = document.querySelectorAll('#usage-table th');
    headers.forEach((header, index) => {
        header.addEventListener('click', () => {
            sortTable(index);
        });
    });
}

// Sort table
function sortTable(columnIndex) {
    const table = document.getElementById('usage-table');
    const tbody = table.querySelector('tbody');
    const rows = Array.from(tbody.querySelectorAll('tr'));
    
    const isNumeric = columnIndex === 3 || columnIndex === 4; // Tokens or Cost columns
    
    rows.sort((a, b) => {
        const aText = a.cells[columnIndex].textContent.trim();
        const bText = b.cells[columnIndex].textContent.trim();
        
        if (isNumeric) {
            const aNum = parseFloat(aText.replace(/[^0-9.-]/g, '')) || 0;
            const bNum = parseFloat(bText.replace(/[^0-9.-]/g, '')) || 0;
            return bNum - aNum; // Descending by default
        } else {
            return aText.localeCompare(bText);
        }
    });
    
    rows.forEach(row => tbody.appendChild(row));
}

// Export to CSV
function exportToCSV() {
    const projectId = getProjectId();
    if (!projectId) {
        showToast('Project ID not found', 'error');
        return;
    }

    const period = document.getElementById('period-select').value;
    const endDate = new Date().toISOString().split('T')[0];
    let startDate;
    switch (period) {
        case 'daily':
            startDate = new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString().split('T')[0];
            break;
        case 'weekly':
            startDate = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
            break;
        case 'monthly':
            startDate = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
            break;
        case 'yearly':
            startDate = new Date(Date.now() - 365 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
            break;
        default:
            startDate = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    }

    getUsageReport(projectId, startDate, endDate)
        .then(report => {
            let csv = 'Date,Provider,Model,Tokens,Cost\n';
            
            if (report.daily_usage) {
                report.daily_usage.forEach(daily => {
                    csv += `${daily.date},${Object.keys(report.usage_by_provider || {}).join(';')},${Object.keys(report.usage_by_model || {}).join(';')},${daily.tokens},${daily.cost}\n`;
                });
            }
            
            const blob = new Blob([csv], { type: 'text/csv' });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `usage-report-${startDate}-to-${endDate}.csv`;
            a.click();
            window.URL.revokeObjectURL(url);
            
            showToast('CSV exported successfully', 'success');
        })
        .catch(error => {
            showToast('Failed to export CSV: ' + error.message, 'error');
        });
}

// Export to JSON
function exportToJSON() {
    const projectId = getProjectId();
    if (!projectId) {
        showToast('Project ID not found', 'error');
        return;
    }

    const period = document.getElementById('period-select').value;
    const endDate = new Date().toISOString().split('T')[0];
    let startDate;
    switch (period) {
        case 'daily':
            startDate = new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString().split('T')[0];
            break;
        case 'weekly':
            startDate = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
            break;
        case 'monthly':
            startDate = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
            break;
        case 'yearly':
            startDate = new Date(Date.now() - 365 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
            break;
        default:
            startDate = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    }

    getUsageReport(projectId, startDate, endDate)
        .then(report => {
            const json = JSON.stringify(report, null, 2);
            const blob = new Blob([json], { type: 'application/json' });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `usage-report-${startDate}-to-${endDate}.json`;
            a.click();
            window.URL.revokeObjectURL(url);
            
            showToast('JSON exported successfully', 'success');
        })
        .catch(error => {
            showToast('Failed to export JSON: ' + error.message, 'error');
        });
}









