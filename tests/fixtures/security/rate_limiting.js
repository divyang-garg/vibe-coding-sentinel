// Test fixture: Missing rate limiting middleware
// Expected: SEC-004 finding

const express = require('express');
const app = express();

// VULNERABLE: No rate limiting on API endpoints
app.get('/api/users', (req, res) => {
    res.json({ users: [] });
});

app.post('/api/login', (req, res) => {
    // No rate limiting - vulnerable to brute force
    res.json({ token: 'abc123' });
});












