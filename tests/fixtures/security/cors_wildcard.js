// SEC-008: CORS Wildcard Configuration
// This file contains insecure CORS configurations

const express = require('express');
const app = express();

// VULNERABLE: CORS with wildcard origin
const cors = require('cors');
app.use(cors({
    origin: '*'  // SEC-008: Wildcard origin allows any domain
}));

// VULNERABLE: Manual CORS headers with wildcard
app.use((req, res, next) => {
    res.header('Access-Control-Allow-Origin', '*');
    res.header('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE');
    next();
});

// SAFE: CORS with specific origins (should not be flagged)
app.use(cors({
    origin: ['https://example.com', 'https://app.example.com'],
    credentials: true
}));

// SAFE: Environment-based CORS
app.use(cors({
    origin: process.env.ALLOWED_ORIGINS?.split(',') || [],
    credentials: true
}));












