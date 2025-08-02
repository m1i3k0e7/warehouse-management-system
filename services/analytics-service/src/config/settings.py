from pydantic import BaseSettings

class Settings(BaseSettings):
    APP_NAME: str = "Analytics Service"
    DATABASE_URL: str = "postgresql://user:password@localhost:5433/analytics_db"
    KAFKA_BOOTSTRAP_SERVERS: str = "localhost:9092"
    KAFKA_TOPIC: str = "wms_events"

    class Config:
        env_file = ".env"

settings = Settings()
