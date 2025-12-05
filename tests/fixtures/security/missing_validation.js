// SEC-006: Missing Input Validation
// This file contains handlers without input validation

const express = require('express');
const router = express.Router();

// VULNERABLE: No input validation
router.post('/api/users', (req, res) => {
    const { username, email, age } = req.body;
    // Missing validation - direct use of user input
    return db.query("INSERT INTO users (username, email, age) VALUES (?, ?, ?)", 
        [username, email, age])
        .then(() => res.json({ success: true }));
});

// VULNERABLE: No validation on update
router.put('/api/users/:id', (req, res) => {
    // Missing validation library usage
    return db.query("UPDATE users SET ? WHERE id = ?", [req.body, req.params.id])
        .then(() => res.json({ success: true }));
});

// SAFE: Input validation with joi (should not be flagged)
const Joi = require('joi');
const userSchema = Joi.object({
    username: Joi.string().alphanum().min(3).max(30).required(),
    email: Joi.string().email().required(),
    age: Joi.number().integer().min(0).max(150)
});

router.post('/api/users-safe', (req, res) => {
    const { error, value } = userSchema.validate(req.body);
    if (error) {
        return res.status(400).json({ error: error.details[0].message });
    }
    return db.query("INSERT INTO users (username, email, age) VALUES (?, ?, ?)", 
        [value.username, value.email, value.age])
        .then(() => res.json({ success: true }));
});

