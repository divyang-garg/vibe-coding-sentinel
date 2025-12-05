// VULNERABLE FILE - Contains NoSQL injection vulnerabilities for testing
// Sentinel should detect all issues in this file

const mongoose = require('mongoose');

class UserController {
  // Issue 1: NoSQL injection with $where
  async findUsersWithWhere(condition) {
    return await User.find({
      $where: condition  // DANGEROUS: $where with user input
    });
  }
  
  // Issue 2: Another $where injection
  async searchUsers(searchQuery) {
    return await User.find({
      $where: `this.name.includes('${searchQuery}')`
    });
  }
  
  // Issue 3: NoSQL injection via object injection
  async login(username, password) {
    // If password is { $gt: "" }, this bypasses authentication
    return await User.findOne({
      username: username,
      password: password  // Should use bcrypt.compare instead
    });
  }
  
  // Issue 4: Aggregation injection
  async aggregateUsers(pipeline) {
    // User-controlled pipeline is dangerous
    return await User.aggregate(pipeline);
  }
  
  // Issue 5: MapReduce injection (deprecated but still detected)
  async mapReduceUsers(mapFunction, reduceFunction) {
    return await User.mapReduce(
      mapFunction,   // User-controlled
      reduceFunction // User-controlled
    );
  }
  
  // Issue 6: $regex injection
  async searchByName(namePattern) {
    return await User.find({
      name: { $regex: namePattern }  // Can cause ReDoS
    });
  }
  
  // Issue 7: $expr injection
  async complexQuery(expression) {
    return await User.find({
      $expr: expression  // User-controlled expression
    });
  }
}

// Issue 8: Console.log with sensitive data
console.log("User password:", user.password);

// Issue 9: Debugger statement
debugger;

// Issue 10: NOLOCK usage (SQL but detected in JS too)
const query = "SELECT * FROM users WITH (NOLOCK)";

module.exports = UserController;

