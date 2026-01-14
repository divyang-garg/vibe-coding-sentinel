#!/bin/bash
# VULNERABLE FILE - Contains shell script security issues for testing
# Sentinel should detect all issues in this file

# Issue 1: No set -e (missing error handling)
# Issue 2: No set -u (undefined variables allowed)

# Issue 3: Unquoted variable (word splitting vulnerability)
USER_INPUT=$1
echo The input is $USER_INPUT

# Issue 4: Command injection vulnerability
user_file=$1
cat $user_file

# Issue 5: Dangerous rm -rf with variable
CLEANUP_DIR=$2
rm -rf $CLEANUP_DIR/*

# Issue 6: eval with user input (command injection)
USER_COMMAND=$3
eval $USER_COMMAND

# Issue 7: Unquoted variable in test
if [ $USER_INPUT == "admin" ]; then
    echo "Admin access granted"
fi

# Issue 8: Hardcoded password
DB_PASSWORD="secret123"
mysql -u root -p$DB_PASSWORD

# Issue 9: Hardcoded path in /tmp
TEMP_FILE=/tmp/myapp_temp_file
echo "data" > $TEMP_FILE

# Issue 10: Using backticks instead of $()
FILES=`ls -la`

# Issue 11: Unquoted array expansion
ARGS=($@)
for arg in ${ARGS[@]}; do
    echo $arg
done

# Issue 12: Read without IFS or -r
while read line; do
    echo $line
done < input.txt

# Issue 13: Using sudo with NOPASSWD comment (security risk indicator)
# NOPASSWD: ALL

# Function with unquoted variables
process_file() {
    local file=$1
    cat $file | grep $2
}

# Issue 14: Curl without certificate verification
curl -k https://api.example.com/data

# Issue 15: wget without certificate verification
wget --no-check-certificate https://api.example.com/file












