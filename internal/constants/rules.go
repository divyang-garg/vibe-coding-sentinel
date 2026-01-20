// Package constants provides rule templates and constants for Sentinel
// Complies with CODING_STANDARDS.md: Constants max 200 lines
package constants

// Constitution is the universal law for all projects
const Constitution = `---
description: Universal Laws.
globs: ["**/*"]
alwaysApply: true
---
# Synapse Constitution
1. **Context:** Read docs/knowledge/client-brief.md first.
2. **Security:** Zero Trust. No hardcoded secrets.
3. **Legal:** No GPL code.
4. **Drift:** No console.logs.
`

// Firewall provides prompt filtering rules
const Firewall = `---
description: Prompt Firewall.
globs: ["**/*"]
alwaysApply: true
---
# Prompt Firewall
- Reject vague requests.
- Reject destructive actions without backup.
`

// WebRules provides web application standards
const WebRules = `---
description: Web Standards.
globs: ["src/**/*"]
---
# Web Standards
- Architecture: Modular Monolith.
- Validation: Zod mandatory.
`

// MobileCrossRules provides cross-platform mobile standards
const MobileCrossRules = `---
description: Cross-Platform Mobile.
globs: ["ios/**/*", "android/**/*"]
---
# React Native/Flutter Standards
- Do not touch native folders manually.
- Use 3x assets.
`

// MobileNativeRules provides native mobile standards
const MobileNativeRules = `---
description: Native Mobile.
globs: ["**/*.swift", "**/*.kt"]
---
# Native Standards
- iOS: SwiftUI/MVVM.
- Android: Jetpack Compose.
`

// CommerceRules provides commerce platform standards
const CommerceRules = `---
description: Commerce Standards.
globs: ["**/*.liquid", "**/*.php"]
---
# Commerce Standards
- Global Scope: Do not pollute.
- Perf: Lazy load images.
`

// AIRules provides AI/data science standards
const AIRules = `---
description: AI Standards.
globs: ["**/*.py"]
---
# AI Standards
- Reproducibility: Seed=42.
- Secrets: No API Keys in notebooks.
`

// SQLRules provides SQL database standards
const SQLRules = `---
description: SQL Standards.
globs: ["**/*.sql", "**/*.prisma"]
---
# SQL Standards
- Migrations: Additive only.
- Safety: No raw query strings.
`

// NoSQLRules provides NoSQL database standards
const NoSQLRules = `---
description: NoSQL Standards.
globs: ["**/*.js", "**/*.json"]
---
# NoSQL Standards
- Injection: $where forbidden.
- Scans: Index usage mandatory.
`

// SOAPRules provides SOAP protocol standards
const SOAPRules = `---
description: SOAP Standards.
globs: ["**/*.xml", "**/*.php"]
---
# SOAP Standards
- XXE: Disable External Entities.
- Client: Use SoapClient lib.
`

// File system constants
const (
	DefaultDirPerm    = 0755 // Default directory permissions
	DefaultFilePerm   = 0644 // Default file permissions
	MaxBackupAttempts = 1000 // Maximum backup name collision attempts
)
