import logging
from apscheduler.schedulers.background import BackgroundScheduler

from ..services import report_generator

scheduler = BackgroundScheduler()

def start_scheduler():
    """Starts the scheduler and adds jobs."""
    # Example: Generate a daily summary report every day at midnight
    scheduler.add_job(
        report_generator.generate_daily_summary,
        'cron',
        hour=0,
        minute=0,
        id="daily_summary_report",
        replace_existing=True
    )
    scheduler.start()
    logging.info("Scheduler started with daily jobs.")
