---
description: Project Business Knowledge (Auto-Generated)
globs: ["**/*"]
alwaysApply: true
---

# Business Knowledge

This file contains extracted and approved business knowledge.
Last updated: 2025-12-04 22:24

## Business Rules

### Order Cancellation Policy

Orders can only be cancelled within 24 hours of creation and before the shipping status is set. Premium users have an extended cancellation window of 48 hours.

### Refund Processing Time

Refunds are processed within 5-7 business days after the cancellation is approved.

## Domain Entities

### User

Represents a registered customer with the following attributes: id (UUID), email (string, unique), name (string), role (enum: admin, user, premium), created_at (datetime).

### Order

Represents a customer order with attributes: id (UUID), user_id (UUID, foreign key), status (enum: pending, processing, shipped, delivered, cancelled), total (decimal), created_at (datetime).

## Glossary

| Term | Definition |
|------|------------|
| **Premium User** | A user with an active premium subscription that provides additional benefits such as extended cancellation windows and priority support. |
| **SKU** | Stock Keeping Unit - A unique identifier for each distinct product and service that can be purchased. |

