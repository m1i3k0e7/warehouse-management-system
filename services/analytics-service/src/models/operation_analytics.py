from sqlalchemy import Column, Integer, String, DateTime
import datetime

from ..config.database import Base

class MaterialEvent(Base):
    __tablename__ = "material_events"

    id = Column(Integer, primary_key=True, index=True)
    material_id = Column(String, index=True)
    event_type = Column(String) # e.g., "PLACED", "PICKED", "MOVED"
    timestamp = Column(DateTime, default=datetime.datetime.utcnow)
    shelf_id = Column(String)
    slot_id = Column(String)
    worker_id = Column(String, nullable=True)
