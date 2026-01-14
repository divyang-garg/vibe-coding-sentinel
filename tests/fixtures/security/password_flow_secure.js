// Test fixture: Password variable flows to secure hashing (bcrypt)
// Expected: No SEC-005 finding

const bcrypt = require('bcrypt');

function createUser(req, res) {
    const password = req.body.password; // User input
    
    // SECURE: Using bcrypt for password hashing
    const hashedPassword = bcrypt.hashSync(password, 10);
    
    // Save to database
    db.users.insert({ password: hashedPassword });
}












