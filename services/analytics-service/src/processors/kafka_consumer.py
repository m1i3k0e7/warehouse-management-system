import json
import threading
import logging
from kafka import KafkaConsumer

from ..config.settings import settings
from ..services.data_collector import save_event_data

logging.basicConfig(level=logging.INFO)

def consume_events():
    """Consumes events from the Kafka topic and processes them."""
    consumer = KafkaConsumer(
        settings.KAFKA_TOPIC,
        bootstrap_servers=settings.KAFKA_BOOTSTRAP_SERVERS,
        auto_offset_reset='earliest',
        group_id='analytics-service-group',
        value_deserializer=lambda m: json.loads(m.decode('utf-8'))
    )
    
    logging.info(f"Subscribed to Kafka topic: {settings.KAFKA_TOPIC}")

    for message in consumer:
        try:
            event_data = message.value
            logging.info(f"Received event: {event_data}")
            # This function will handle saving the data to the DB
            save_event_data(event_data)
        except Exception as e:
            logging.error(f"Error processing message: {e}")

def start_consumer_thread():
    """Starts the Kafka consumer in a separate thread."""
    consumer_thread = threading.Thread(target=consume_events, daemon=True)
    consumer_thread.start()
    logging.info("Kafka consumer thread started.")
