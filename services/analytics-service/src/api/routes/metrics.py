from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from ...config.database import get_db
from .. import schemas
from ...analytics import utilization_analyzer

router = APIRouter()

@router.get("/dashboard", response_model=schemas.DashboardMetrics, tags=["Metrics"])
def get_dashboard_metrics(db: Session = Depends(get_db)):
    """Retrieve key metrics for the main dashboard."""
    moved_materials_data = utilization_analyzer.calculate_most_moved_materials(db)
    shelf_utilization_data = utilization_analyzer.calculate_shelf_utilization(db)

    # Convert data to schema models
    moved_materials = [
        schemas.MaterialMovement(material_id=row[0], move_count=row[1]) 
        for row in moved_materials_data
    ]

    return schemas.DashboardMetrics(
        most_moved_materials=moved_materials,
        shelf_utilization=shelf_utilization_data,
    )
