// SEC-003: Missing Authentication Middleware
// This file contains routes without authentication

const express = require('express');
const router = express.Router();

// VULNERABLE: Protected route without auth middleware
router.get('/api/users/:id', (req, res) => {
    const userId = req.params.id;
    // Missing authentication check
    return db.query("SELECT * FROM users WHERE id = ?", [userId])
        .then(user => res.json(user));
});

// VULNERABLE: Admin route without auth
router.delete('/api/users/:id', (req, res) => {
    // Missing authentication and authorization
    return db.query("DELETE FROM users WHERE id = ?", [req.params.id])
        .then(() => res.json({ success: true }));
});

// SAFE: Route with authentication middleware (should not be flagged)
const auth = require('./middleware/auth');
router.get('/api/profile', auth, (req, res) => {
    return res.json(req.user);
});

// SAFE: Route with JWT middleware
const jwt = require('jsonwebtoken');
router.post('/api/posts', authenticateToken, (req, res) => {
    // Protected route
    return res.json({ success: true });
});

function authenticateToken(req, res, next) {
    const token = req.headers['authorization'];
    if (!token) return res.sendStatus(401);
    jwt.verify(token, process.env.JWT_SECRET, (err, user) => {
        if (err) return res.sendStatus(403);
        req.user = user;
        next();
    });
}












