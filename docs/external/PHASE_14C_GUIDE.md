# Phase 14C: Hub Configuration Interface Guide

## Overview

Phase 14C implements a complete Hub UI for LLM provider configuration, API key management, cost optimization settings, and usage tracking. This guide explains how to use the configuration interface and usage dashboard.

## Accessing the Dashboard

The Hub Dashboard is available at:
- URL: `http://localhost:8080/dashboard/` (or your Hub URL)
- Redirect: `http://localhost:8080/` automatically redirects to `/dashboard/`

## Configuration Tab

### Setting Up LLM Provider Configuration

1. **Select Provider**: Choose from OpenAI, Anthropic, or Azure OpenAI
2. **Select Model**: Choose a model for the selected provider
3. **Enter API Key**: Provide your API key (encrypted and stored securely)
4. **Configure Cost Optimization**:
   - Enable/disable caching
   - Set cache TTL (hours)
   - Enable progressive depth analysis
   - Set maximum cost per request (optional)

5. **Test Connection**: Click "Test Connection" to verify your API key works
6. **Save Configuration**: Click "Save Configuration" to store your settings

### Managing Configurations

- **View All Configurations**: See all configured providers in the table below the form
- **Edit Configuration**: Click "Edit" to modify an existing configuration
- **Delete Configuration**: Click "Delete" to remove a configuration (with confirmation)

### API Key Security

- API keys are encrypted using AES-256-GCM before storage
- API keys are masked in the UI (only last 4 characters shown)
- All configuration changes are logged in the audit log

## Usage Dashboard Tab

### Viewing Usage Statistics

The Usage Dashboard provides comprehensive insights into LLM usage:

1. **Overview Cards**: 
   - Total Tokens Used
   - Total Cost
   - Average Cost per Request
   - Total Requests

2. **Cost Breakdown Charts**:
   - Pie chart showing cost by provider
   - Bar chart showing cost by model

3. **Cost Trends**:
   - Line chart showing cost over time
   - Filter by period (daily, weekly, monthly, yearly)

4. **Usage Table**:
   - Detailed usage data by date
   - Sortable columns
   - Export to CSV or JSON

### Filtering Usage Data

- Select period from dropdown: Daily, Weekly, Monthly, or Yearly
- Click "Refresh" to reload data
- Export data using "Export CSV" or "Export JSON" buttons

## API Endpoints

### Configuration Endpoints

- `POST /api/v1/llm/config` - Create new configuration
- `GET /api/v1/llm/config/{id}` - Get configuration by ID
- `PUT /api/v1/llm/config/{id}` - Update configuration
- `DELETE /api/v1/llm/config/{id}` - Delete configuration
- `GET /api/v1/llm/config/project/{projectId}` - List all configurations for a project

### Metadata Endpoints

- `GET /api/v1/llm/providers` - Get list of supported providers
- `GET /api/v1/llm/models/{provider}` - Get list of models for a provider

### Validation Endpoint

- `POST /api/v1/llm/config/validate` - Test API key and model connection

### Usage Reporting Endpoints

- `GET /api/v1/llm/usage/report` - Get detailed usage report
- `GET /api/v1/llm/usage/stats` - Get aggregated usage statistics
- `GET /api/v1/llm/usage/cost-breakdown` - Get cost breakdown by provider/model
- `GET /api/v1/llm/usage/trends` - Get usage trends over time

## Troubleshooting

### Common Issues

1. **"Project ID not found"**
   - Ensure `project_id` is provided in URL: `?project_id=YOUR_PROJECT_ID`
   - Or set in localStorage: `localStorage.setItem('sentinel_project_id', 'YOUR_PROJECT_ID')`

2. **"Failed to load providers"**
   - Check that the Hub API is running
   - Verify network connectivity
   - Check browser console for errors

3. **"Connection test failed"**
   - Verify API key is correct
   - Check provider/model combination is valid
   - Ensure API key has necessary permissions
   - Check network connectivity to provider API

4. **"Failed to save configuration"**
   - Verify all required fields are filled
   - Check API key format is correct for provider
   - Ensure project_id is valid

5. **Charts not displaying**
   - Check browser console for JavaScript errors
   - Verify Chart.js library is loaded
   - Ensure usage data exists for the selected period

### API Key Format Requirements

- **OpenAI**: Must start with `sk-`
- **Anthropic**: Minimum 20 characters
- **Azure**: Minimum 20 characters

## Security Considerations

- API keys are encrypted at rest using AES-256-GCM
- Encryption key should be stored securely (environment variable)
- All configuration changes are logged in audit log
- Rate limiting is applied to all endpoints
- CORS is configured for dashboard origin

## Cost Optimization Tips

1. **Enable Caching**: Reduces redundant LLM calls
2. **Use Progressive Depth**: Start with fast checks, only use LLM when needed
3. **Set Max Cost**: Prevent accidental high-cost requests
4. **Monitor Usage**: Regularly check usage dashboard to identify cost trends
5. **Choose Appropriate Models**: Use cheaper models for non-critical tasks

## Best Practices

1. **Test Before Saving**: Always test connection before saving configuration
2. **Regular Monitoring**: Check usage dashboard weekly to track costs
3. **Rotate API Keys**: Regularly rotate API keys for security
4. **Use Project-Specific Keys**: Use different API keys for different projects
5. **Review Audit Logs**: Periodically review audit logs for unauthorized changes









