// Test fixture: Password variable flows to insecure hashing (MD5)
// Expected: SEC-005 finding

function createUser(req, res) {
    const password = req.body.password; // User input
    
    // VULNERABLE: Using MD5 for password hashing
    const hashedPassword = crypto.createHash('md5').update(password).digest('hex');
    
    // Save to database
    db.users.insert({ password: hashedPassword });
}












