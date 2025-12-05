"""
Sample Python utility module with snake_case naming.
Used for pattern detection tests.
"""
from datetime import datetime
from typing import Any, Dict, List, Optional
import json


def format_date(date: datetime) -> str:
    """
    Format a datetime object to ISO string.
    
    Args:
        date: The datetime to format
        
    Returns:
        Formatted date string (YYYY-MM-DD)
    """
    return date.strftime('%Y-%m-%d')


def calculate_sum(numbers: List[float]) -> float:
    """
    Calculate the sum of a list of numbers.
    
    Args:
        numbers: List of numbers to sum
        
    Returns:
        Sum of all numbers
    """
    return sum(numbers)


def deep_clone(obj: Dict[str, Any]) -> Dict[str, Any]:
    """
    Deep clone a dictionary.
    
    Args:
        obj: Dictionary to clone
        
    Returns:
        Deep cloned dictionary
    """
    return json.loads(json.dumps(obj))


def validate_email(email: str) -> bool:
    """
    Validate an email address format.
    
    Args:
        email: Email address to validate
        
    Returns:
        True if valid, False otherwise
    """
    import re
    pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
    return bool(re.match(pattern, email))


def sanitize_input(text: str) -> str:
    """
    Sanitize user input by removing dangerous characters.
    
    Args:
        text: Input text to sanitize
        
    Returns:
        Sanitized text
    """
    dangerous_chars = ['<', '>', '&', '"', "'"]
    result = text
    for char in dangerous_chars:
        result = result.replace(char, '')
    return result


# Constants
MAX_ITEMS = 100
DEFAULT_TIMEOUT = 5000
API_VERSION = "v1"

