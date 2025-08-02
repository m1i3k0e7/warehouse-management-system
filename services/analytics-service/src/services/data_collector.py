from sqlalchemy.orm import Session

from ..config.database import SessionLocal
from ..models.operation_analytics import MaterialEvent

def save_event_data(event_data: dict):
    """Saves a single event to the database."""
    db: Session = SessionLocal()
    try:
        # TODO: Add validation here using Pydantic schemas
        db_event = MaterialEvent(
            material_id=event_data.get("material_id"),
            event_type=event_data.get("event_type"),
            timestamp=event_data.get("timestamp"),
            shelf_id=event_data.get("shelf_id"),
            slot_id=event_data.get("slot_id"),
            worker_id=event_data.get("worker_id")
        )
        db.add(db_event)
        db.commit()
    finally:
        db.close()
