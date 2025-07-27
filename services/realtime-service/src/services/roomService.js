const logger = require('../utils/logger');

class RoomService {
  constructor(io) {
    this.io = io;
  }

  /**
   * add a socket to a specified room
   * @param {object} socket - Socket.IO socket object
   * @param {string} roomName - room name to join
   */
  joinRoom(socket, roomName) {
    socket.join(roomName);
    logger.info(`Socket ${socket.id} joined room: ${roomName}`);
  }

  /**
   * remove a socket from a specified room
   * @param {object} socket - Socket.IO socket object
   * @param {string} roomName - room name to leave
   */
  leaveRoom(socket, roomName) {
    socket.leave(roomName);
    logger.info(`Socket ${socket.id} left room: ${roomName}`);
  }

  /**
   * broadcast an event to all sockets in a specified room
   * @param {string} roomName - room name to broadcast to
   * @param {string} eventName - event name to broadcast
   * @param {object} data - broadcast data
   */
  broadcastToRoom(roomName, eventName, data) {
    this.io.to(roomName).emit(eventName, data);
    logger.debug(`Broadcasted event '${eventName}' to room '${roomName}' with data:`, data);
  }

  /**
   * get all sockets in a specified room
   * @param {string} roomName - room name to get sockets from
   */
  async getRoomOccupancy(roomName) {
    const sockets = await this.io.in(roomName).allSockets();
    return sockets.size;
  }
}

module.exports = RoomService;
