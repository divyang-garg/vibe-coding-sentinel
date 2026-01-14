// SEC-002: SQL Injection Vulnerability
// This file contains vulnerable SQL query patterns

// VULNERABLE: String concatenation in SQL
function getUserById(userId) {
    const query = "SELECT * FROM users WHERE id = " + userId;
    return db.query(query);
}

// VULNERABLE: Template literal with user input
function searchUsers(searchTerm) {
    return db.query(`SELECT * FROM users WHERE name LIKE '%${searchTerm}%'`);
}

// VULNERABLE: SQL Server dynamic SQL
function executeDynamicSQL(sql) {
    return db.query("EXEC(@sql)", { sql: sql });
}

// SAFE: Parameterized query (should not be flagged)
function getUserByIdSafe(userId) {
    return db.query("SELECT * FROM users WHERE id = $1", [userId]);
}

// SAFE: Named parameters
function searchUsersSafe(searchTerm) {
    return db.query("SELECT * FROM users WHERE name LIKE :search", { search: `%${searchTerm}%` });
}












