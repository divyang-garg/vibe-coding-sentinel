# Azure AI Foundry Setup Guide

This guide walks you through setting up Azure AI Foundry to use Claude Opus 4.5 for knowledge extraction in Sentinel Hub.

## Prerequisites

- Azure subscription with billing enabled
- Access to Azure Portal
- Admin permissions to create resources

## Step 1: Create Azure AI Services Resource

1. Log in to [Azure Portal](https://portal.azure.com)
2. Click **"Create a resource"** or search for **"Azure AI Services"**
3. Select **"Azure AI Services"** from the results
4. Click **"Create"**

### Configuration

- **Subscription**: Select your subscription
- **Resource Group**: Create new or select existing
- **Region**: Choose a region (e.g., `East US`, `West Europe`)
- **Name**: Enter a unique name (e.g., `sentinel-ai-foundry`)
- **Pricing Tier**: Select **Standard S0** (or higher)

5. Click **"Review + create"**, then **"Create"**
6. Wait for deployment to complete (2-3 minutes)

## Step 2: Request Access to Anthropic Claude Models

1. Navigate to your Azure AI Services resource
2. In the left menu, go to **"Model deployments"** or **"Deployments"**
3. Click **"Create"** or **"Manage deployments"**
4. If Claude models are not available:
   - Go to **"Request access"** or **"Request model access"**
   - Select **"Anthropic"** as the provider
   - Request access to **"Claude Opus 4.5"** (or **"Claude 3.5 Sonnet"** as fallback)
   - Submit the request
   - Wait for approval (usually 1-2 business days)

## Step 3: Deploy Claude Opus 4.5 Model

1. Once access is approved, go to **"Model deployments"**
2. Click **"Create"** or **"Deploy model"**
3. Select **"Anthropic"** as the provider
4. Choose **"Claude Opus 4.5"** (or your preferred Claude model)
5. Configure deployment:
   - **Deployment name**: `claude-opus-4-5` (or your preferred name)
   - **Model version**: Latest available
   - **Capacity**: Select based on your needs (start with minimum)
6. Click **"Create"** and wait for deployment (1-2 minutes)

## Step 4: Get Endpoint and API Key

### Get Endpoint URL

1. In your Azure AI Services resource, go to **"Keys and Endpoint"**
2. Copy the **Endpoint** URL (e.g., `https://sentinel-ai-foundry.services.ai.azure.com`)
3. Save this for later configuration

### Get API Key

1. In the same **"Keys and Endpoint"** section:
2. Copy **Key 1** (or **Key 2** as backup)
3. **Important**: Keep this key secure and never commit it to version control

## Step 5: Configure Sentinel Hub

### Option 1: Environment Variables (Recommended)

Create a `.env` file in the `hub/` directory:

```bash
# Azure AI Foundry Configuration
AZURE_AI_ENDPOINT=https://your-resource.services.ai.azure.com
AZURE_AI_KEY=your-api-key-here
AZURE_AI_DEPLOYMENT=claude-opus-4-5
AZURE_AI_API_VERSION=2024-02-01
```

### Option 2: Docker Compose

Update `hub/docker-compose.yml`:

```yaml
services:
  processor:
    environment:
      - AZURE_AI_ENDPOINT=${AZURE_AI_ENDPOINT}
      - AZURE_AI_KEY=${AZURE_AI_KEY}
      - AZURE_AI_DEPLOYMENT=${AZURE_AI_DEPLOYMENT:-claude-opus-4-5}
      - AZURE_AI_API_VERSION=${AZURE_AI_API_VERSION:-2024-02-01}
```

## Step 6: Test the Connection

### Test via cURL

```bash
curl -X POST "https://your-resource.services.ai.azure.com/models/claude-opus-4-5/chat/completions" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {"role": "user", "content": "Hello, Claude!"}
    ],
    "max_tokens": 100
  }'
```

Expected response:
```json
{
  "id": "chatcmpl-...",
  "choices": [{
    "message": {
      "role": "assistant",
      "content": "Hello! How can I help you today?"
    }
  }]
}
```

### Test via Sentinel Hub

1. Start the Hub: `cd hub && docker-compose up -d`
2. Upload a document: `./sentinel ingest test-doc.pdf`
3. Check logs: `docker-compose logs processor`
4. Look for: `✅ Extraction successful via azure-foundry`

## Troubleshooting

### Error: "Model not found"

- Verify the deployment name matches exactly
- Check that the model is deployed and active
- Ensure you're using the correct endpoint URL

### Error: "Unauthorized" or "401"

- Verify your API key is correct
- Check that the key hasn't been regenerated
- Ensure you're using the `Bearer` token format

### Error: "Rate limit exceeded"

- Reduce request frequency
- Upgrade your pricing tier if needed
- Implement retry logic with exponential backoff

### Error: "Model access not approved"

- Wait for approval (1-2 business days)
- Check email for approval notification
- Contact Azure support if needed

## Security Best Practices

1. **Never commit API keys** to version control
2. **Use environment variables** or Azure Key Vault
3. **Rotate keys regularly** (every 90 days)
4. **Use Managed Identity** in production (if available)
5. **Restrict network access** to Hub IPs only

## Cost Management

- **Monitor usage** in Azure Portal → Cost Management
- **Set up alerts** for spending thresholds
- **Use Ollama as fallback** to reduce Azure costs
- **Optimize prompts** to reduce token usage

## Next Steps

- Configure provider fallback (Azure → Ollama)
- Set up monitoring and alerts
- Review extraction quality and adjust prompts
- See [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md) for full deployment instructions

