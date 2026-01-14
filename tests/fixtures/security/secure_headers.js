// Test fixture: Missing secure headers middleware
// Expected: SEC-007 finding

const express = require('express');
const app = express();

// VULNERABLE: No secure headers middleware (helmet)
app.get('/api/data', (req, res) => {
    res.json({ data: 'sensitive' });
});












