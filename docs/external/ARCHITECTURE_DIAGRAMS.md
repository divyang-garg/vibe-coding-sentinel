# Sentinel Hub Architecture Diagrams

## System Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Sentinel Hub API                               │
│                    Production-Ready Code Analysis Platform              │
└─────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           API Gateway Layer                             │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    HTTP Router & Middleware                     │    │
│  │  • Chi Router with RESTful endpoints                          │    │
│  │  • Authentication & Authorization middleware                   │    │
│  │  • Rate limiting & CORS support                               │    │
│  │  • Request logging & error handling                           │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Handler Layer                                  │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                 HTTP Request Handlers                           │    │
│  │  • TaskHandler - CRUD operations                              │    │
│  │  • DocumentHandler - File processing                          │    │
│  │  • WorkflowHandler - Orchestration                            │    │
│  │  • OrganizationHandler - Multi-tenancy                        │    │
│  │  • CodeAnalysisHandler - Static analysis                      │    │
│  │  • MonitoringHandler - Observability                          │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Service Layer                                  │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                 Business Logic Services                         │    │
│  │  • TaskService - Task lifecycle management                     │    │
│  │  • DocumentService - Content processing                        │    │
│  │  • WorkflowService - Process orchestration                     │    │
│  │  • OrganizationService - Tenant management                     │    │
│  │  • CodeAnalysisService - Code quality analysis                 │    │
│  │  • MonitoringService - System observability                    │    │
│  │  • APIVersionService - Version compatibility                   │    │
│  │  • RepositoryService - Repository analysis                     │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         Repository Layer                                │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                 Data Access Layer                               │    │
│  │  • TaskRepository - PostgreSQL task operations                │    │
│  │  • DocumentRepository - File system operations                │    │
│  │  • WorkflowRepository - Workflow persistence                  │    │
│  │  • OrganizationRepository - Tenant data access                │    │
│  │  • DependencyAnalyzer - Relationship analysis                 │    │
│  │  • ImpactAnalyzer - Change impact assessment                  │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Data Layer                                     │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                   PostgreSQL Database                            │    │
│  │  • ACID transactions                                           │    │
│  │  • Connection pooling                                          │    │
│  │  • Migration support                                           │    │
│  │  • JSON document storage                                       │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                   File System Storage                           │    │
│  │  • Document storage                                            │    │
│  │  • Temporary file handling                                     │    │
│  │  • Secure file operations                                      │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
```

## Component Architecture

### Clean Architecture Implementation

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Entities & Value Objects                      │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  • Task, Document, Workflow, Organization                     │    │
│  │  • TaskStatus, DocumentStatus, WorkflowStatus                 │    │
│  │  • Validation rules and business constraints                 │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           Use Cases / Services                          │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  • Application-specific business rules                        │    │
│  │  • Workflow orchestration                                     │    │
│  │  • Code analysis algorithms                                   │    │
│  │  • Cross-cutting concerns (logging, caching)                 │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                       Interface Adapters                               │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  • HTTP Controllers (Handlers)                                │    │
│  │  • Repository Implementations                                │    │
│  │  • External API Clients                                      │    │
│  │  • Message Queue Producers/Consumers                         │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           Frameworks & Drivers                          │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  • Database drivers (PostgreSQL)                              │    │
│  │  • Web frameworks (Chi router)                               │    │
│  │  • File system operations                                    │    │
│  │  • External service clients                                  │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
```

## Data Flow Architecture

### Request Processing Flow

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Client    │───▶│  Router    │───▶│  Handler   │───▶│  Service   │
│             │    │             │    │             │    │             │
│ HTTP Request│    │ • Auth      │    │ • Validate  │    │ • Business │
│             │    │ • Rate Lim  │    │ • Parse     │    │ • Logic     │
└─────────────┘    │ • CORS      │    │ • Route     │    │ • Rules     │
                   └─────────────┘    └─────────────┘    └─────────────┘
                                                                 │
                                                                 ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ Repository  │───▶│  Database  │    │   Cache     │    │  External  │
│             │    │             │    │             │    │   APIs     │
│ • Query     │    │ • PostgreSQL│    │ • Redis     │    │ • LLM      │
│ • Persist   │    │ • ACID      │    │ • In-memory │    │ • GitHub   │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

### Error Handling Flow

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Service    │───▶│ Repository  │───▶│  Database  │    │   Client    │
│  Error      │    │  Error      │    │  Error      │    │   Response  │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
       ▲                   ▲                   ▲                   │
       │                   │                   │                   ▼
       └───────────────────┴───────────────────┴─────────────▶ HTTP Error
                   Error Wrapping with Context
```

## Security Architecture

### Authentication & Authorization

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Security Layers                                  │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                Network Security Layer                           │    │
│  │  • TLS 1.3 encryption                                         │    │
│  │  • Certificate validation                                     │    │
│  │  • IP allowlisting (optional)                                 │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                        │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                Application Security Layer                       │    │
│  │  • API Key Authentication                                      │    │
│  │  • Rate Limiting (100 req/10sec)                              │    │
│  │  • CORS Policy Enforcement                                     │    │
│  │  • Input Validation & Sanitization                             │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                        │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                Data Security Layer                              │    │
│  │  • SQL Injection Prevention                                    │    │
│  │  • XSS Protection                                             │    │
│  │  • CSRF Protection                                            │    │
│  │  • Secure File Upload Validation                              │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
```

### Security Middleware Chain

```
HTTP Request ──▶ [TLS Termination] ──▶ [Rate Limiting] ──▶ [CORS]
                   │                        │                      │
                   ▼                        ▼                      ▼
              [Certificate]           [Token Bucket]          [Origin Check]
                   │                        │                      │
                   └─────────┬──────────────┴──────────────┬───────┘
                             ▼                             ▼
                       [API Key Auth]               [Security Headers]
                             │                             │
                             └─────────────┬───────────────┘
                                           ▼
                                    [Request Processing]
```

## Deployment Architecture

### Production Deployment

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Load Balancer                                  │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    NGINX / HAProxy                              │    │
│  │  • SSL/TLS termination                                        │    │
│  │  • Load balancing                                              │    │
│  │  • Health checks                                               │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
                                     │
                    ┌────────────────┼────────────────┐
                    │                │                │
                    ▼                ▼                ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Application Servers                            │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    Sentinel Hub API                              │    │
│  │  • Go 1.21+ runtime                                           │    │
│  │  • 2-4 CPU cores                                              │    │
│  │  • 4-8 GB RAM                                                 │    │
│  │  • Health check endpoints                                      │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Data Layer                                     │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    PostgreSQL Cluster                            │    │
│  │  • Primary + Read replicas                                     │    │
│  │  • Connection pooling                                          │    │
│  │  • Automated backups                                           │    │
│  │  • Monitoring & alerting                                       │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                        │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    Redis Cache (Optional)                        │    │
│  │  • Session storage                                             │    │
│  │  • Response caching                                            │    │
│  │  • Pub/Sub for async operations                                │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                        │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    File Storage                                  │    │
│  │  • Local file system                                           │    │
│  │  • S3-compatible storage                                       │    │
│  │  • CDN integration                                             │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
```

### Development Deployment

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Development Environment                          │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    Docker Compose                                │    │
│  │  • API Server (Go)                                             │    │
│  │  • PostgreSQL database                                         │    │
│  │  • Redis cache (optional)                                       │    │
│  │  • Local file storage                                          │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
```

## Monitoring & Observability

### Metrics Collection

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Application Metrics                              │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                Business Metrics                                 │    │
│  │  • API request count by endpoint                               │    │
│  │  • Response time percentiles                                   │    │
│  │  • Error rate by category                                      │    │
│  │  • Task completion rate                                        │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                        │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                System Metrics                                   │    │
│  │  • CPU usage                                                   │    │
│  │  • Memory usage                                                │    │
│  │  • Disk I/O                                                    │    │
│  │  • Network I/O                                                 │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                        │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                Database Metrics                                 │    │
│  │  • Connection pool usage                                       │    │
│  │  • Query execution time                                        │    │
│  │  • Transaction count                                           │    │
│  │  • Lock contention                                             │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Metrics Pipeline                                 │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                Collection & Transport                           │    │
│  │  • Prometheus metrics                                          │    │
│  │  • Structured logging                                          │    │
│  │  • Health check endpoints                                       │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Monitoring Stack                                 │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                Visualization & Alerting                          │    │
│  │  • Grafana dashboards                                          │    │
│  │  • Prometheus alerting                                          │    │
│  │  • ELK stack for logs                                          │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
```

## Component Interaction Diagram

### Service Dependencies

```
TaskService
├── TaskRepository (PostgreSQL)
├── DependencyAnalyzer
└── ImpactAnalyzer

DocumentService
├── DocumentRepository
└── SearchEngine (Elasticsearch)

WorkflowService
├── WorkflowRepository
└── StepExecutor

CodeAnalysisService
├── AST Parser
├── Pattern Matcher
└── Quality Scorer

MonitoringService
├── ErrorRepository
├── MetricsCollector
└── AlertManager

APIVersionService
└── VersionRepository

RepositoryService
├── GitClient
├── ImpactCalculator
└── NetworkAnalyzer
```

### Data Flow Between Components

```
HTTP Request
    ↓
Router (Chi)
    ↓
Middleware Chain:
  - Authentication
  - Rate Limiting
  - CORS
  - Logging
    ↓
Handler Layer
    ↓
Validation & Parsing
    ↓
Service Layer
    ↓
Business Logic
    ↓
Repository Layer
    ↓
Database/File System
    ↓
Response Generation
    ↓
Middleware (Response)
    ↓
HTTP Response
```

---

## Implementation Notes

### Design Principles Followed

1. **Single Responsibility**: Each component has one clear purpose
2. **Dependency Injection**: Clean separation of concerns
3. **Interface Segregation**: Focused, minimal interfaces
4. **Error Handling**: Consistent error wrapping and propagation
5. **Testing**: Comprehensive unit and integration test coverage

### Performance Considerations

- Connection pooling for database operations
- In-memory caching for frequently accessed data
- Asynchronous processing for long-running tasks
- Rate limiting to prevent resource exhaustion
- Efficient data structures and algorithms

### Scalability Features

- Horizontal scaling through stateless design
- Database read replicas for query offloading
- CDN integration for static asset delivery
- Message queues for async processing
- Microservice-ready architecture

---

*Architecture diagrams last updated: January 14, 2026*