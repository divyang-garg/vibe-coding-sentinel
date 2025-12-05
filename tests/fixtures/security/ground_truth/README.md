# Ground Truth Test Suite for Security Detection Rate Validation

This directory contains labeled test cases for validating security rule detection rates.

## Structure

Each test file should be labeled with:
- **Rule ID**: Which security rule it tests (SEC-001 to SEC-008)
- **Expected Finding**: Whether a finding should be detected (true/false)
- **Severity**: The severity of the vulnerability (critical/high/medium/low)
- **Type**: The type of vulnerability (vulnerable/safe/edge_case)

## Label Format

Files should be named: `{rule_id}_{type}_{description}.{ext}`

Examples:
- `SEC-001_vulnerable_missing_ownership.js` - Should detect SEC-001
- `SEC-002_safe_parameterized_query.js` - Should NOT detect SEC-002
- `SEC-005_vulnerable_md5_hash.js` - Should detect SEC-005

## Detection Rate Calculation

Detection rate = (True Positives + True Negatives) / Total Test Cases

Target: 85%+ detection rate

