<?php
// VULNERABLE FILE - Contains SQL injection vulnerabilities for testing
// Sentinel should detect all issues in this file

class UserRepository {
    private $db;
    
    public function __construct($database) {
        $this->db = $database;
    }
    
    // Issue 1: Direct SQL injection vulnerability
    public function getUserById($id) {
        $query = "SELECT * FROM users WHERE id = " . $id;
        return $this->db->query($query);
    }
    
    // Issue 2: SQL injection in string concatenation
    public function getUserByEmail($email) {
        $query = "SELECT * FROM users WHERE email = '" . $email . "'";
        return $this->db->query($query);
    }
    
    // Issue 3: SQL injection with sprintf (still vulnerable)
    public function getUserByUsername($username) {
        $query = sprintf("SELECT * FROM users WHERE username = '%s'", $username);
        return $this->db->query($query);
    }
    
    // Issue 4: Multiple injections in one query
    public function searchUsers($name, $email, $role) {
        $query = "SELECT * FROM users WHERE name LIKE '%" . $name . "%' 
                  AND email = '" . $email . "' 
                  AND role = '" . $role . "'";
        return $this->db->query($query);
    }
    
    // Issue 5: Injection in ORDER BY clause
    public function getAllUsersSorted($sortColumn) {
        $query = "SELECT * FROM users ORDER BY " . $sortColumn;
        return $this->db->query($query);
    }
    
    // Issue 6: simplexml_load_string (XXE vulnerability)
    public function parseUserXml($xmlData) {
        $xml = simplexml_load_string($xmlData);
        return $xml;
    }
    
    // Issue 7: eval usage (code injection)
    public function processFormula($formula) {
        return eval("return " . $formula . ";");
    }
}

// Issue 8: Using $_GET directly in query
$userId = $_GET['id'];
$query = "SELECT * FROM users WHERE id = $userId";

// Issue 9: Using $_POST directly in query
$username = $_POST['username'];
$password = $_POST['password'];
$loginQuery = "SELECT * FROM users WHERE username = '$username' AND password = '$password'";

?>












