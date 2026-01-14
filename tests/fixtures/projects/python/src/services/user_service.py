"""
Sample Python service module with snake_case naming.
Tests: naming conventions, import patterns, class structure.
"""
from datetime import datetime
from typing import Dict, List, Optional, Any
from dataclasses import dataclass

from src.utils.helpers import format_date, deep_clone, validate_email


@dataclass
class User:
    """User data class."""
    id: str
    email: str
    name: str
    created_at: datetime
    updated_at: Optional[datetime] = None


class UserService:
    """
    User Service - handles user operations.
    
    Attributes:
        http_client: HTTP client for API calls
        cache: In-memory cache for users
    """
    
    API_BASE_URL = '/api/v1'
    
    def __init__(self, http_client: Any) -> None:
        """
        Initialize the user service.
        
        Args:
            http_client: HTTP client instance
        """
        self.http_client = http_client
        self._cache: Dict[str, User] = {}
    
    async def get_user_by_id(self, user_id: str) -> Optional[User]:
        """
        Get a user by their ID.
        
        Args:
            user_id: The user's ID
            
        Returns:
            User object if found, None otherwise
        """
        if user_id in self._cache:
            return self._cache[user_id]
        
        response = await self.http_client.get(
            f"{self.API_BASE_URL}/users/{user_id}"
        )
        
        if response.status_code == 200:
            user = User(**response.json())
            self._cache[user_id] = user
            return user
        
        return None
    
    async def create_user(self, user_data: Dict[str, Any]) -> User:
        """
        Create a new user.
        
        Args:
            user_data: User data dictionary
            
        Returns:
            Created User object
            
        Raises:
            ValueError: If email is invalid
        """
        if not validate_email(user_data.get('email', '')):
            raise ValueError('Invalid email address')
        
        payload = deep_clone(user_data)
        payload['created_at'] = format_date(datetime.now())
        
        response = await self.http_client.post(
            f"{self.API_BASE_URL}/users",
            json=payload
        )
        
        return User(**response.json())
    
    async def update_user(
        self, 
        user_id: str, 
        updates: Dict[str, Any]
    ) -> Optional[User]:
        """
        Update a user's information.
        
        Args:
            user_id: The user's ID
            updates: Dictionary of updates
            
        Returns:
            Updated User object
        """
        response = await self.http_client.patch(
            f"{self.API_BASE_URL}/users/{user_id}",
            json=updates
        )
        
        # Invalidate cache
        self._cache.pop(user_id, None)
        
        if response.status_code == 200:
            return User(**response.json())
        
        return None
    
    async def delete_user(self, user_id: str) -> bool:
        """
        Delete a user.
        
        Args:
            user_id: The user's ID
            
        Returns:
            True if deleted successfully
        """
        response = await self.http_client.delete(
            f"{self.API_BASE_URL}/users/{user_id}"
        )
        
        self._cache.pop(user_id, None)
        
        return response.status_code == 204
    
    def clear_cache(self) -> None:
        """Clear the user cache."""
        self._cache.clear()












