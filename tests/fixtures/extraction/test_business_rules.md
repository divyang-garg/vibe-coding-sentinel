# Business Requirements Document

## Order Processing Rules

The system must validate all orders before processing.
Orders must be processed within 24 hours of submission.
Users must not cancel orders after shipment has been initiated.

## Authentication Rules

The user must authenticate before accessing protected resources.
Sessions shall expire after 30 minutes of inactivity.
Administrators must use two-factor authentication for sensitive operations.

## Payment Processing

All payments must be encrypted during transmission.
The system must validate payment amounts before processing.
Refunds must be processed within 5 business days.

## Data Protection

Personal data must not be stored longer than 7 years.
The system must encrypt sensitive data at rest.
Users must consent before data collection.
