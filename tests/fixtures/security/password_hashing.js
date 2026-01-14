// SEC-005: Password Hashing Vulnerability
// This file contains insecure password hashing patterns

const crypto = require('crypto');

// VULNERABLE: MD5 hashing
function hashPasswordMD5(password) {
    return crypto.createHash('md5').update(password).digest('hex');
}

// VULNERABLE: SHA1 hashing
function hashPasswordSHA1(password) {
    return crypto.createHash('sha1').update(password).digest('hex');
}

// VULNERABLE: Plain text password storage
function createUser(username, password) {
    return db.query("INSERT INTO users (username, password) VALUES (?, ?)", [username, password]);
}

// SAFE: bcrypt hashing (should not be flagged)
const bcrypt = require('bcrypt');
function hashPasswordSafe(password) {
    return bcrypt.hashSync(password, 10);
}

// SAFE: argon2 hashing
const argon2 = require('argon2');
async function hashPasswordArgon2(password) {
    return await argon2.hash(password);
}












