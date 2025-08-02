from sqlalchemy.orm import Session
from sqlalchemy import func

from ..models import operation_analytics as models

def calculate_shelf_utilization(db: Session):
    """Calculates the utilization rate for each shelf."""
    # Placeholder logic, as in the previous implementation.
    # A real implementation would require more data context.
    return [
        {"shelf_id": "SH-01", "utilization_rate": 0.8},
        {"shelf_id": "SH-02", "utilization_rate": 0.5},
    ]

def calculate_most_moved_materials(db: Session, limit: int = 10):
    """Calculates which materials have been moved the most."""
    return (
        db.query(
            models.MaterialEvent.material_id,
            func.count(models.MaterialEvent.id).label("move_count"),
        )
        .group_by(models.MaterialEvent.material_id)
        .order_by(func.count(models.MaterialEvent.id).desc())
        .limit(limit)
        .all()
    )
