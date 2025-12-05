# Azure Blob Storage Migration Guide

## LLM Instructions

**IMPORTANT FOR AI ASSISTANTS**: The code structures and snippets in this document are for reference only. They represent the intended architecture and approach, but the actual implementation must be verified against the current codebase during vibe coding. Always:

1. Read the actual current code files before making changes
2. Verify function signatures, imports, and package structures match
3. Test each change incrementally
4. Ensure backward compatibility with existing local filesystem mode
5. Handle errors gracefully and provide fallback mechanisms

**Do not blindly copy-paste code from this document.** Use it as a guide and adapt to the actual codebase structure.

---

## Overview

This document describes the migration from local filesystem storage to Azure Blob Storage for document storage in the Sentinel Hub. This migration enables serverless deployment on Azure App Services where local filesystem storage is not available.

## Current Architecture (Local Filesystem)

### File Storage Location

- **Path Structure**: `/data/documents/{project.ID}/{docID}/{filename}`
- **Storage Type**: Docker volume (`document_storage:/data/documents`)
- **Shared Between**: API server and Processor worker
- **Database Storage**: Full filesystem path stored in `documents.file_path` column

### Current Flow

```
1. API receives upload
   → Saves to: /data/documents/{project}/{doc-id}/file.pdf
   → Stores path in DB: /data/documents/{project}/{doc-id}/file.pdf

2. Processor queries DB
   → Gets file_path from documents table
   → Reads file: os.ReadFile(file_path)
   → Parses and extracts knowledge
```

### Limitations

- ❌ Not compatible with Azure App Services (no shared filesystem)
- ❌ Files are ephemeral in serverless environments
- ❌ Cannot scale across multiple instances
- ❌ No built-in redundancy or backup

---

## Target Architecture (Azure Blob Storage)

### File Storage Location

- **Path Structure**: `https://{account}.blob.core.windows.net/{container}/{project.ID}/{docID}/{filename}`
- **Storage Type**: Azure Blob Storage (Standard LRS)
- **Access**: Via Azure SDK from both API and Processor
- **Database Storage**: Blob URL stored in `documents.file_path` column

### Target Flow

```
1. API receives upload
   → Uploads to Azure Blob Storage
   → Gets blob URL: https://{account}.blob.core.windows.net/{container}/{project}/{doc-id}/file.pdf
   → Stores URL in DB: https://{account}.blob.core.windows.net/{container}/{project}/{doc-id}/file.pdf

2. Processor queries DB
   → Gets blob URL from documents table
   → Downloads blob to temp file (for parsing tools that need filesystem)
   → Parses and extracts knowledge
   → Cleans up temp file
```

### Benefits

- ✅ Compatible with Azure App Services
- ✅ Persistent storage across deployments
- ✅ Scales automatically
- ✅ Built-in redundancy and backup options
- ✅ Can enable CDN for public access (if needed)

---

## Implementation Plan

### Phase 1: Add Azure Blob Storage SDK

**Files to Modify:**
- `hub/api/go.mod`
- `hub/processor/go.mod`

**Changes:**
Add Azure Blob Storage SDK dependencies to both modules.

**Reference Implementation:**
```go
// hub/api/go.mod
require (
    github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.4.0
    github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.2.1
    // ... existing dependencies
)

// hub/processor/go.mod
require (
    github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.4.0
    github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.2.1
    // ... existing dependencies
)
```

**Verification Steps:**
1. Run `go mod tidy` in both directories
2. Verify no dependency conflicts
3. Check that imports resolve correctly

---

### Phase 2: Update Configuration

**Files to Modify:**
- `hub/api/main.go` (Config struct and loadConfig function)
- `hub/processor/main.go` (Config struct and loadConfig function)
- `hub/docker-compose.yml` (environment variables)

**Changes:**
1. Add blob storage configuration fields to Config structs
2. Add environment variable loading for blob storage credentials
3. Add feature flag `USE_BLOB_STORAGE` to enable/disable blob storage
4. Maintain backward compatibility with local filesystem

**Reference Implementation:**

```go
// hub/api/main.go - Config struct
type Config struct {
    // ... existing fields ...
    // Azure Blob Storage
    StorageAccountName string
    StorageAccountKey  string
    StorageContainer   string
    UseBlobStorage     bool  // Feature flag
}

// hub/api/main.go - loadConfig()
func loadConfig() *Config {
    useBlob := getEnv("USE_BLOB_STORAGE", "false") == "true"
    
    return &Config{
        // ... existing config ...
        StorageAccountName: getEnv("STORAGE_ACCOUNT_NAME", ""),
        StorageAccountKey:  getEnv("STORAGE_ACCOUNT_KEY", ""),
        StorageContainer:   getEnv("STORAGE_CONTAINER", "documents"),
        UseBlobStorage:     useBlob,
    }
}
```

**Verification Steps:**
1. Verify config loads correctly with `USE_BLOB_STORAGE=false` (default)
2. Verify config loads correctly with blob storage enabled
3. Test that missing credentials are handled gracefully

---

### Phase 3: Implement Blob Storage Client (API)

**Files to Modify:**
- `hub/api/main.go` (add blob storage functions)

**Changes:**
1. Add blob storage client initialization function
2. Add blob upload function
3. Add container creation/verification
4. Handle errors gracefully

**Reference Implementation:**

```go
// hub/api/main.go - Add after imports
import (
    // ... existing imports ...
    "context"
    "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

// Global blob container client
var blobContainerClient *azblob.ContainerClient

// Initialize blob storage
func initBlobStorage(config *Config) error {
    if !config.UseBlobStorage {
        return nil // Use local filesystem
    }

    if config.StorageAccountName == "" || config.StorageAccountKey == "" {
        return fmt.Errorf("blob storage enabled but credentials missing")
    }

    credential, err := azblob.NewSharedKeyCredential(
        config.StorageAccountName, 
        config.StorageAccountKey,
    )
    if err != nil {
        return fmt.Errorf("failed to create blob credential: %w", err)
    }

    serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net", 
        config.StorageAccountName)
    serviceClient, err := azblob.NewServiceClientWithSharedKey(
        serviceURL, credential, nil,
    )
    if err != nil {
        return fmt.Errorf("failed to create blob service client: %w", err)
    }

    blobContainerClient = serviceClient.NewContainerClient(config.StorageContainer)

    // Ensure container exists
    ctx := context.Background()
    _, err = blobContainerClient.Create(ctx, nil)
    if err != nil {
        // Container might already exist
        var storageErr *azblob.StorageError
        if err, ok := err.(*azblob.StorageError); ok && 
           err.ErrorCode == azblob.StorageErrorCodeContainerAlreadyExists {
            log.Println("Blob container already exists")
        } else {
            return fmt.Errorf("failed to create blob container: %w", err)
        }
    }

    log.Println("✅ Azure Blob Storage initialized")
    return nil
}

// Upload file to blob storage
func uploadToBlob(ctx context.Context, projectID, docID, filename string, 
    file io.Reader, size int64) (string, error) {
    blobPath := fmt.Sprintf("%s/%s/%s", projectID, docID, filename)
    blobClient := blobContainerClient.NewBlockBlobClient(blobPath)

    _, err := blobClient.UploadStream(ctx, file, azblob.UploadStreamOptions{})
    if err != nil {
        return "", fmt.Errorf("failed to upload to blob: %w", err)
    }

    // Return blob URL
    blobURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s",
        blobContainerClient.AccountName(), 
        blobContainerClient.Name(), 
        blobPath)
    
    return blobURL, nil
}
```

**Verification Steps:**
1. Test blob storage initialization with valid credentials
2. Test blob storage initialization with invalid credentials (should error gracefully)
3. Test container creation (should handle existing container)
4. Test file upload to blob storage

---

### Phase 4: Update API Upload Handler

**Files to Modify:**
- `hub/api/main.go` (uploadDocumentHandler function)

**Changes:**
1. Check `config.UseBlobStorage` flag
2. If enabled: Upload to blob storage, store blob URL in DB
3. If disabled: Use existing local filesystem logic
4. Maintain same database schema (file_path column stores either path or URL)

**Reference Implementation:**

```go
// hub/api/main.go - uploadDocumentHandler
func uploadDocumentHandler(config *Config) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        project := r.Context().Value("project").(*Project)

        // Parse multipart form (max 100MB)
        if err := r.ParseMultipartForm(100 << 20); err != nil {
            http.Error(w, "File too large", http.StatusBadRequest)
            return
        }

        file, header, err := r.FormFile("file")
        if err != nil {
            http.Error(w, "No file provided", http.StatusBadRequest)
            return
        }
        defer file.Close()

        docID := uuid.New().String()
        var filePath string
        var fileSize int64 = header.Size

        if config.UseBlobStorage {
            // Upload to Azure Blob Storage
            ctx := r.Context()
            blobURL, err := uploadToBlob(ctx, project.ID, docID, 
                header.Filename, file, header.Size)
            if err != nil {
                http.Error(w, "Storage error: "+err.Error(), 
                    http.StatusInternalServerError)
                return
            }
            filePath = blobURL
            log.Printf("Uploaded to blob: %s", blobURL)
        } else {
            // Local filesystem (original behavior)
            storageDir := filepath.Join(config.DocumentStorage, 
                project.ID, docID)
            if err := os.MkdirAll(storageDir, 0755); err != nil {
                http.Error(w, "Storage error", http.StatusInternalServerError)
                return
            }

            filePath = filepath.Join(storageDir, header.Filename)
            dst, err := os.Create(filePath)
            if err != nil {
                http.Error(w, "Storage error", http.StatusInternalServerError)
                return
            }
            defer dst.Close()

            if _, err := io.Copy(dst, file); err != nil {
                http.Error(w, "Storage error", http.StatusInternalServerError)
                return
            }
        }

        // ... rest of handler (MIME type detection, DB insert) ...
        // Store filePath (either local path or blob URL) in database
    }
}
```

**Verification Steps:**
1. Test upload with `USE_BLOB_STORAGE=false` (should work as before)
2. Test upload with `USE_BLOB_STORAGE=true` (should upload to blob)
3. Verify blob URL is stored correctly in database
4. Verify file is accessible via blob URL

---

### Phase 5: Implement Blob Storage Reader (Processor)

**Files to Modify:**
- `hub/processor/main.go` (add blob storage download functions)

**Changes:**
1. Add blob storage client initialization (similar to API)
2. Add blob download function
3. Add function to download blob to temp file (for parsing tools)
4. Handle cleanup of temp files

**Reference Implementation:**

```go
// hub/processor/main.go - Add after imports
import (
    // ... existing imports ...
    "context"
    "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

// Global processor blob container client
var processorBlobContainerClient *azblob.ContainerClient

// Initialize blob storage for processor
func initProcessorBlobStorage(config *Config) error {
    if !config.UseBlobStorage {
        return nil
    }

    if config.StorageAccountName == "" || config.StorageAccountKey == "" {
        return fmt.Errorf("blob storage enabled but credentials missing")
    }

    credential, err := azblob.NewSharedKeyCredential(
        config.StorageAccountName, 
        config.StorageAccountKey,
    )
    if err != nil {
        return fmt.Errorf("failed to create blob credential: %w", err)
    }

    serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net", 
        config.StorageAccountName)
    serviceClient, err := azblob.NewServiceClientWithSharedKey(
        serviceURL, credential, nil,
    )
    if err != nil {
        return fmt.Errorf("failed to create blob service client: %w", err)
    }

    processorBlobContainerClient = serviceClient.NewContainerClient(
        config.StorageContainer)
    log.Println("✅ Processor: Azure Blob Storage initialized")
    return nil
}

// Download blob to temp file (for parsing tools that need filesystem)
func downloadBlobToTempFile(ctx context.Context, blobURL string) (string, error) {
    // Parse blob URL: https://{account}.blob.core.windows.net/{container}/{path}
    parts := strings.SplitN(blobURL, "/", 5)
    if len(parts) < 5 {
        return "", fmt.Errorf("invalid blob URL format: %s", blobURL)
    }
    
    containerName := parts[3]
    blobPath := parts[4]
    
    // Verify container matches
    if containerName != processorBlobContainerClient.Name() {
        return "", fmt.Errorf("container mismatch: expected %s, got %s",
            processorBlobContainerClient.Name(), containerName)
    }
    
    blobClient := processorBlobContainerClient.NewBlockBlobClient(blobPath)
    
    downloadResponse, err := blobClient.DownloadStream(ctx, nil)
    if err != nil {
        return "", fmt.Errorf("failed to download from blob: %w", err)
    }
    defer downloadResponse.Body.Close()
    
    // Create temp file
    tmpFile, err := os.CreateTemp("", "sentinel-doc-*")
    if err != nil {
        return "", fmt.Errorf("failed to create temp file: %w", err)
    }
    
    // Copy blob content to temp file
    _, err = io.Copy(tmpFile, downloadResponse.Body)
    if err != nil {
        tmpFile.Close()
        os.Remove(tmpFile.Name())
        return "", fmt.Errorf("failed to copy blob to temp file: %w", err)
    }
    
    tmpFile.Close()
    return tmpFile.Name(), nil
}
```

**Verification Steps:**
1. Test blob download with valid URL
2. Test blob download with invalid URL (should error gracefully)
3. Test temp file creation and cleanup
4. Verify temp files are deleted after processing

---

### Phase 6: Update Processor Document Parsing

**Files to Modify:**
- `hub/processor/main.go` (parseDocument function and worker function)

**Changes:**
1. Detect if `file_path` is a blob URL (starts with `https://`)
2. If blob URL: Download to temp file, parse, then cleanup
3. If local path: Use existing filesystem logic
4. Ensure temp files are always cleaned up (use defer)

**Reference Implementation:**

```go
// hub/processor/main.go - parseDocument function
func parseDocument(doc *Document, config *Config) (string, error) {
    var filePath string
    var isTempFile bool
    
    // Check if file_path is a blob URL
    if strings.HasPrefix(doc.FilePath, "https://") && 
       strings.Contains(doc.FilePath, ".blob.core.windows.net") {
        // Download from blob to temp file
        ctx := context.Background()
        tempPath, err := downloadBlobToTempFile(ctx, doc.FilePath)
        if err != nil {
            return "", fmt.Errorf("failed to download from blob: %w", err)
        }
        filePath = tempPath
        isTempFile = true
        defer func() {
            if isTempFile {
                os.Remove(tempPath) // Clean up temp file
            }
        }()
    } else {
        // Local filesystem path
        filePath = doc.FilePath
        isTempFile = false
    }
    
    ext := strings.ToLower(filepath.Ext(filePath))

    var text string
    var err error
    
    switch ext {
    case ".txt", ".md", ".markdown":
        text, err = parseTextFile(filePath)
    case ".pdf":
        text, err = parsePDF(filePath)
    case ".docx":
        text, err = parseDOCX(filePath)
    case ".xlsx":
        text, err = parseXLSX(filePath)
    case ".eml":
        text, err = parseEmail(filePath)
    case ".png", ".jpg", ".jpeg":
        text, err = parseImage(filePath)
    default:
        return "", fmt.Errorf("unsupported file type: %s", ext)
    }
    
    return text, err
}
```

**Verification Steps:**
1. Test parsing with local filesystem path (should work as before)
2. Test parsing with blob URL (should download and parse)
3. Verify temp files are cleaned up after parsing
4. Test error handling when blob download fails

---

### Phase 7: Update Main Functions

**Files to Modify:**
- `hub/api/main.go` (main function)
- `hub/processor/main.go` (main function)

**Changes:**
1. Initialize blob storage if enabled
2. Fall back to local filesystem if blob storage disabled
3. Handle initialization errors gracefully

**Reference Implementation:**

```go
// hub/api/main.go - main function
func main() {
    config := loadConfig()

    // Initialize database
    log.Println("Connecting to database...")
    if err := initDB(config.DatabaseURL); err != nil {
        log.Fatalf("Database connection failed: %v", err)
    }
    defer db.Close()

    // Initialize blob storage if enabled
    if config.UseBlobStorage {
        log.Println("Initializing Azure Blob Storage...")
        if err := initBlobStorage(config); err != nil {
            log.Fatalf("Blob storage initialization failed: %v", err)
        }
    } else {
        // Create local storage directory
        if err := os.MkdirAll(config.DocumentStorage, 0755); err != nil {
            log.Fatalf("Failed to create storage directory: %v", err)
        }
        log.Println("Using local filesystem storage")
    }

    // Run migrations
    log.Println("Running migrations...")
    if err := runMigrations(); err != nil {
        log.Fatalf("Migrations failed: %v", err)
    }

    // ... rest of main function ...
}

// hub/processor/main.go - main function
func main() {
    config := loadConfig()

    log.Println("Connecting to database...")
    if err := initDB(config.DatabaseURL); err != nil {
        log.Fatalf("Database connection failed: %v", err)
    }
    defer db.Close()

    // Initialize blob storage if enabled
    if config.UseBlobStorage {
        log.Println("Initializing Azure Blob Storage...")
        if err := initProcessorBlobStorage(config); err != nil {
            log.Fatalf("Blob storage initialization failed: %v", err)
        }
    }

    log.Printf("🔧 Sentinel Document Processor starting with %d workers", 
        config.WorkerCount)

    // Start workers
    for i := 0; i < config.WorkerCount; i++ {
        go worker(i, config)
    }

    // Keep main goroutine alive
    select {}
}
```

**Verification Steps:**
1. Test startup with `USE_BLOB_STORAGE=false` (should use local filesystem)
2. Test startup with `USE_BLOB_STORAGE=true` and valid credentials
3. Test startup with `USE_BLOB_STORAGE=true` and invalid credentials (should fail gracefully)

---

### Phase 8: Update Docker Compose

**Files to Modify:**
- `hub/docker-compose.yml`

**Changes:**
Add environment variables for blob storage configuration to both `api` and `processor` services.

**Reference Implementation:**

```yaml
services:
  api:
    environment:
      # ... existing environment variables ...
      - USE_BLOB_STORAGE=${USE_BLOB_STORAGE:-false}
      - STORAGE_ACCOUNT_NAME=${STORAGE_ACCOUNT_NAME:-}
      - STORAGE_ACCOUNT_KEY=${STORAGE_ACCOUNT_KEY:-}
      - STORAGE_CONTAINER=${STORAGE_CONTAINER:-documents}

  processor:
    environment:
      # ... existing environment variables ...
      - USE_BLOB_STORAGE=${USE_BLOB_STORAGE:-false}
      - STORAGE_ACCOUNT_NAME=${STORAGE_ACCOUNT_NAME:-}
      - STORAGE_ACCOUNT_KEY=${STORAGE_ACCOUNT_KEY:-}
      - STORAGE_CONTAINER=${STORAGE_CONTAINER:-documents}
```

**Verification Steps:**
1. Test docker-compose with default values (should use local filesystem)
2. Test docker-compose with blob storage enabled
3. Verify environment variables are passed correctly to containers

---

## Testing Strategy

### Unit Tests

1. **Blob Storage Client Tests**
   - Test credential creation
   - Test container creation
   - Test file upload
   - Test file download
   - Test error handling

2. **Configuration Tests**
   - Test config loading with blob storage enabled
   - Test config loading with blob storage disabled
   - Test missing credentials handling

3. **Integration Tests**
   - Test full upload → process → knowledge extraction flow with blob storage
   - Test backward compatibility with local filesystem

### Manual Testing Checklist

- [ ] Upload document with `USE_BLOB_STORAGE=false` (local filesystem)
- [ ] Upload document with `USE_BLOB_STORAGE=true` (blob storage)
- [ ] Verify document is accessible via blob URL
- [ ] Verify processor can download and parse blob-stored documents
- [ ] Verify knowledge extraction works with blob-stored documents
- [ ] Verify temp files are cleaned up after processing
- [ ] Test error handling (invalid credentials, network failures)
- [ ] Test with existing documents in database (backward compatibility)

---

## Migration Path

### Option 1: Big Bang Migration

1. Deploy new code with blob storage enabled
2. Migrate existing documents to blob storage (one-time script)
3. Update database `file_path` values from local paths to blob URLs

### Option 2: Gradual Migration

1. Deploy new code with `USE_BLOB_STORAGE=false` (backward compatible)
2. Run migration script to upload existing documents to blob storage
3. Update database `file_path` values
4. Enable `USE_BLOB_STORAGE=true` for new uploads
5. Eventually remove local filesystem support

### Migration Script (Reference)

```go
// migrate_to_blob.go - One-time migration script
// This script should be run once to migrate existing documents

package main

import (
    // ... imports ...
)

func main() {
    // Connect to database
    // Connect to blob storage
    // Query all documents with local file_path
    // For each document:
    //   1. Read file from local filesystem
    //   2. Upload to blob storage
    //   3. Update file_path in database to blob URL
    //   4. Optionally delete local file
}
```

---

## Environment Variables Reference

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `USE_BLOB_STORAGE` | Enable blob storage | `false` | No |
| `STORAGE_ACCOUNT_NAME` | Azure storage account name | - | Yes (if blob enabled) |
| `STORAGE_ACCOUNT_KEY` | Azure storage account key | - | Yes (if blob enabled) |
| `STORAGE_CONTAINER` | Blob container name | `documents` | No |
| `DOCUMENT_STORAGE` | Local filesystem path (fallback) | `/data/documents` | No |

---

## Troubleshooting

### Issue: Blob storage initialization fails

**Symptoms**: Error on startup: "blob storage initialization failed"

**Solutions**:
1. Verify `STORAGE_ACCOUNT_NAME` and `STORAGE_ACCOUNT_KEY` are set correctly
2. Check Azure storage account exists and is accessible
3. Verify container name is correct
4. Check network connectivity to Azure

### Issue: Upload fails with "Storage error"

**Symptoms**: Document upload returns 500 error

**Solutions**:
1. Check blob storage credentials are valid
2. Verify container exists or can be created
3. Check file size limits (blob storage supports large files)
4. Review application logs for detailed error messages

### Issue: Processor can't download blob

**Symptoms**: Document processing fails with "failed to download from blob"

**Solutions**:
1. Verify blob URL format is correct
2. Check blob storage credentials in processor
3. Verify blob exists at the URL
4. Check network connectivity from processor to Azure

### Issue: Temp files not cleaned up

**Symptoms**: Disk space filling up on processor

**Solutions**:
1. Verify `defer os.Remove()` is called in parseDocument
2. Check error paths also clean up temp files
3. Consider using `os.TempDir()` with automatic cleanup

---

## Security Considerations

1. **Credentials Management**
   - Never commit storage account keys to version control
   - Use Azure Key Vault for production
   - Use managed identities when possible (instead of keys)

2. **Access Control**
   - Use private blob containers (not public)
   - Implement SAS tokens for time-limited access if needed
   - Use Azure RBAC for fine-grained access control

3. **Data Encryption**
   - Enable encryption at rest (default in Azure)
   - Use HTTPS for all blob operations (default in SDK)

---

## Performance Considerations

1. **Temp File Management**
   - Temp files are created for each document processing
   - Ensure sufficient disk space on processor
   - Consider streaming for large files (future enhancement)

2. **Blob Storage Performance**
   - Use appropriate storage tier (Hot/Cool/Archive)
   - Consider CDN for frequently accessed documents
   - Monitor blob storage metrics for optimization

3. **Concurrent Processing**
   - Multiple workers can process documents simultaneously
   - Each worker downloads its own temp file
   - Ensure sufficient network bandwidth

---

## Future Enhancements

1. **Direct Streaming**
   - Stream blob content directly to parsing tools (avoid temp files)
   - Requires parser tools to support streaming

2. **Blob Storage Events**
   - Use Azure Event Grid to trigger processing
   - Eliminate polling from database

3. **Multi-Region Support**
   - Replicate documents across regions
   - Process from nearest region

4. **Lifecycle Management**
   - Auto-archive old documents
   - Delete processed temp files automatically

---

## Related Documentation

- [Azure Deployment Guide](./AZURE_DEPLOYMENT_GUIDE.md) - Complete Azure App Services deployment
- [Architecture Document Processing](./ARCHITECTURE_DOCUMENT_PROCESSING.md) - Document processing architecture
- [Deployment Guide](./DEPLOYMENT_GUIDE.md) - General deployment strategies

---

**Last Updated**: 2024-12-XX  
**Status**: Reference Implementation - Verify against actual codebase

