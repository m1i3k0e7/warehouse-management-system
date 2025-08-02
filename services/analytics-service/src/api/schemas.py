from pydantic import BaseModel
from typing import List

class MaterialMovement(BaseModel):
    material_id: str
    move_count: int

class ShelfUtilization(BaseModel):
    shelf_id: str
    utilization_rate: float

class DashboardMetrics(BaseModel):
    most_moved_materials: List[MaterialMovement]
    shelf_utilization: List[ShelfUtilization]

class Config:
    orm_mode = True
