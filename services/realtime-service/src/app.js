const express = require('express');
const http = require('http');
const socketIo = require('socket.io');
const cors = require('cors');

const config = require('./config');
const logger = require('./utils/logger');
const RealtimeService = require('./services/realtimeService');
const SocketController = require('./controllers/socketController');
const KafkaController = require('./controllers/kafkaController');

class App {
  constructor() {
    this.app = express();
    this.server = http.createServer(this.app);
    this.io = socketIo(this.server, {
      cors: {
        origin: config.cors.origin,
        methods: ["GET", "POST"]
      },
      transports: ['websocket', 'polling']
    });
    
    this.realtimeService = new RealtimeService(this.io);
    this.socketController = new SocketController(this.io, this.realtimeService);
    this.kafkaController = new KafkaController(this.realtimeService);
  }

  setupMiddleware() {
    this.app.use(cors(config.cors));
    this.app.use(express.json());
    
    // 健康檢查
    this.app.get('/health', (req, res) => {
      res.json({ status: 'ok', timestamp: new Date().toISOString() });
    });
  }

  setupSocketHandlers() {
    this.io.on('connection', (socket) => {
      logger.info(`Client connected: ${socket.id}`);
      this.socketController.handleConnection(socket);
    });
  }

  async start() {
    try {
      this.setupMiddleware();
      this.setupSocketHandlers();
      
      // 啟動 Kafka 消費者
      await this.kafkaController.start();
      
      this.server.listen(config.port, () => {
        logger.info(`Realtime service running on port ${config.port}`);
      });
    } catch (error) {
      logger.error('Failed to start realtime service:', error);
      process.exit(1);
    }
  }

  async stop() {
    logger.info('Shutting down realtime service...');
    
    await this.kafkaController.stop();
    this.server.close();
    
    logger.info('Realtime service stopped');
  }
}

// 優雅關閉
process.on('SIGINT', async () => {
  await app.stop();
  process.exit(0);
});

process.on('SIGTERM', async () => {
  await app.stop();
  process.exit(0);
});

const app = new App();
app.start();

module.exports = App;