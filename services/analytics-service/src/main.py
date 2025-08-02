from fastapi import FastAPI
from .config.settings import settings
from .config.database import Base, engine
from .api.routes import metrics
from .processors.kafka_consumer import start_consumer_thread
from .tasks.scheduled_jobs import start_scheduler

# Create database tables on startup
Base.metadata.create_all(bind=engine)

app = FastAPI(
    title=settings.APP_NAME,
    version="2.0.0",
    description="Provides analytics and insights on warehouse operations with a robust architecture."
)

@app.on_event("startup")
def startup_event():
    """Tasks to run when the application starts."""
    start_consumer_thread()
    start_scheduler()

# Include API routers
app.include_router(metrics.router, prefix="/api/v1/metrics")

@app.get("/health", tags=["Monitoring"])
def health_check():
    """Check the health of the service."""
    return {"status": "ok"}
