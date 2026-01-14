// VULNERABLE FILE - Contains security issues for testing
// Sentinel should detect all issues in this file

// Issue 1: Hardcoded API key
const API_KEY = "sk-1234567890abcdef1234567890abcdef";

// Issue 2: Hardcoded AWS credentials
const AWS_ACCESS_KEY = "AKIAIOSFODNN7EXAMPLE";
const AWS_SECRET_KEY = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY";

// Issue 3: Hardcoded password
const DB_PASSWORD = "super_secret_password_123";

// Issue 4: JWT secret in code
const JWT_SECRET = "my-256-bit-secret-key-for-jwt-signing";

// Issue 5: Private key in code
const PRIVATE_KEY = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF8PbnGy...
-----END RSA PRIVATE KEY-----`;

// Issue 6: Console.log statements (warning level)
console.log("Debug: User data:", { password: DB_PASSWORD });
console.log("API Key:", API_KEY);

// Issue 7: Alert statement
alert("This should not be in production!");

// Issue 8: Debugger statement
debugger;

// Function that uses the secrets
function makeApiCall(endpoint) {
  console.log("Making API call to:", endpoint);
  
  return fetch(endpoint, {
    headers: {
      'Authorization': `Bearer ${API_KEY}`,
      'X-AWS-Key': AWS_ACCESS_KEY
    }
  });
}

// Export for testing
module.exports = {
  API_KEY,
  AWS_ACCESS_KEY,
  makeApiCall
};












