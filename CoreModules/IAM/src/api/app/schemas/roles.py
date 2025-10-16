from typing import Optional, List, Dict, Any
from pydantic import BaseModel, Field, validator
from uuid import UUID
from datetime import datetime

class RoleBase(BaseModel):
    code: str
    name: str
    description: Optional[str] = None
    
class RoleCreate(RoleBase):
    is_system_role: bool = False
    parent_role_id: Optional[UUID] = None
    metadata: Optional[Dict[str, Any]] = None
    permissions: Optional[List[UUID]] = None

class RoleUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    is_system_role: Optional[bool] = None
    parent_role_id: Optional[UUID] = None
    is_active: Optional[bool] = None
    metadata: Optional[Dict[str, Any]] = None

class RoleSummary(BaseModel):
    id: UUID
    code: str
    name: str
    is_system_role: bool
    created_at: datetime
    
    class Config:
        orm_mode = True

class Role(RoleSummary):
    organization_id: UUID
    description: Optional[str] = None
    parent_role_id: Optional[UUID] = None
    is_active: bool
    updated_at: Optional[datetime] = None
    metadata: Optional[Dict[str, Any]] = None
    permissions: Optional[List[Dict[str, Any]]] = None
    
    class Config:
        orm_mode = True

class RolePermissionAssignment(BaseModel):
    permission_id: UUID

class RoleHierarchyNode(BaseModel):
    id: UUID
    code: str
    name: str
    description: Optional[str] = None
    is_system_role: bool
    children: List['RoleHierarchyNode'] = []

RoleHierarchyNode.update_forward_refs()

class RoleWithHierarchy(Role):
    children: List[RoleHierarchyNode] = []
    parent: Optional[RoleSummary] = None
    
    class Config:
        orm_mode = True
