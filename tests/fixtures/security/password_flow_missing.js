// Test fixture: Password variable from user input but no hashing
// Expected: SEC-005 finding

function createUser(req, res) {
    const password = req.body.password; // User input
    
    // VULNERABLE: Password stored without hashing
    db.users.insert({ password: password });
}

