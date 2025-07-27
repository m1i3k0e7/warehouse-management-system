const express = require('express');
const http = require('http');
const socketIo = require('socket.io');
const cors = require('cors');

const { createAdapter } = require('@socket.io/redis-adapter');
const { createClient } = require('redis');

const config = require('./config');
const logger = require('./utils/logger');

const SocketController = require('./controllers/socketController');
const KafkaController = require('./controllers/kafkaController');

const authMiddleware = require('./middleware/auth');
const rateLimitMiddleware = require('./middleware/rateLimit');

const NotificationService = require('./services/notificationService');
const RoomService = require('./services/roomService');
const RealtimeService = require('./services/realtimeService');

class App {
  constructor() {
    this.app = express();
    this.server = http.createServer(this.app);
    this.io = new socketIo.Server(this.server, {
      cors: {
        origin: config.cors.origin,
        methods: ["GET", "POST"]
      },
      transports: ['websocket', 'polling']
    });

    // Apply middleware
    this.io.use(authMiddleware);
    this.io.use(rateLimitMiddleware);

    const pubClient = createClient({ url: config.redis.url });
    const subClient = pubClient.duplicate();

    Promise.all([pubClient.connect(), subClient.connect()]).then(() => {
      this.io.adapter(createAdapter(pubClient, subClient));
    });

    this.notificationService = new NotificationService();
    this.roomService = new RoomService(this.io);
    this.realtimeService = new RealtimeService(this.io, this.notificationService, this.roomService);
    this.socketController = new SocketController(this.io, this.realtimeService);
    this.kafkaController = new KafkaController(this.realtimeService);
  }

  setupMiddleware() {
    this.app.use(cors(config.cors));
    this.app.use(express.json());
    
    // health check endpoint
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
      
      // start the realtime service
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

// handle shutdown signals
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