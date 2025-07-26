const config = require('../config');

const LOG_LEVELS = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3,
};

const currentLogLevel = LOG_LEVELS[config.logLevel] || LOG_LEVELS.info;

const logger = {
  debug: (message, ...args) => {
    if (currentLogLevel <= LOG_LEVELS.debug) {
      console.log(`[DEBUG] ${message}`, ...args);
    }
  },
  info: (message, ...args) => {
    if (currentLogLevel <= LOG_LEVELS.info) {
      console.log(`[INFO] ${message}`, ...args);
    }
  },
  warn: (message, ...args) => {
    if (currentLogLevel <= LOG_LEVELS.warn) {
      console.warn(`[WARN] ${message}`, ...args);
    }
  },
  error: (message, ...args) => {
    if (currentLogLevel <= LOG_LEVELS.error) {
      console.error(`[ERROR] ${message}`, ...args);
    }
  },
};

module.exports = logger;
