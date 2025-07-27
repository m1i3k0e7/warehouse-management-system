const logger = require('../utils/logger');

class SocketController {
  constructor(io, realtimeService) {
    this.io = io;
    this.realtimeService = realtimeService;
  }

  handleConnection(socket) {
    // handle new connection
    socket.on('join_shelf', (data) => {
      const { shelfId, operatorId } = data;
      this.realtimeService.joinShelfRoom(socket, shelfId, operatorId);
    });

    // handle shelf operation requests
    socket.on('operation_request', (data) => {
      this.realtimeService.handleOperationRequest(socket, data);
    });

    // handle shelf disconnection
    socket.on('disconnect', () => {
      this.realtimeService.handleDisconnect(socket);
    });

    // handle ping-pong for heartbeat
    socket.on('ping', () => {
      socket.emit('pong');
    });
  }
}

module.exports = SocketController;