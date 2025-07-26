const logger = require('../utils/logger');

class SocketController {
  constructor(io, realtimeService) {
    this.io = io;
    this.realtimeService = realtimeService;
  }

  handleConnection(socket) {
    // 處理加入料架房間
    socket.on('join_shelf', (data) => {
      const { shelfId, operatorId } = data;
      this.realtimeService.joinShelfRoom(socket, shelfId, operatorId);
    });

    // 處理操作請求
    socket.on('operation_request', (data) => {
      this.realtimeService.handleOperationRequest(socket, data);
    });

    // 處理斷開連接
    socket.on('disconnect', () => {
      this.realtimeService.handleDisconnect(socket);
    });

    // 處理心跳
    socket.on('ping', () => {
      socket.emit('pong');
    });
  }
}

module.exports = SocketController;