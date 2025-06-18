const jwt = require('jsonwebtoken');
const Device = require('../model/Device');

const JWT_SECRET = process.env.JWT_SECRET || 'your-secret-key';

class RelayService {
  constructor(io, logger) {
    this.io = io;
    this.logger = logger;
    this.connectedDevices = new Map(); // deviceId -> socket
    this.connectedClients = new Map(); // userId -> socket
  }

  async authenticateSocket(socket, token) {
    try {
      const decoded = jwt.verify(token, JWT_SECRET);
      socket.userId = decoded.userId;
      socket.authenticated = true;
      
      this.logger.info(`Socket authenticated for user: ${decoded.userId}`);
      socket.emit('authenticated', { success: true });
    } catch (error) {
      this.logger.error('Socket authentication failed:', error);
      socket.emit('authenticated', { success: false, error: 'Invalid token' });
    }
  }

  async registerDevice(socket, data) {
    if (!socket.authenticated) {
      socket.emit('error', { message: 'Not authenticated' });
      return;
    }

    try {
      const { deviceId, deviceInfo } = data;
      
      // Update device status in database
      await Device.findOneAndUpdate(
        { id: deviceId, userId: socket.userId },
        { 
          status: 'online',
          lastSeen: new Date(),
          ...deviceInfo
        }
      );

      // Store device connection
      this.connectedDevices.set(deviceId, socket);
      socket.deviceId = deviceId;

      this.logger.info(`Device registered: ${deviceId}`);
      socket.emit('device-registered', { success: true, deviceId });
    } catch (error) {
      this.logger.error('Device registration failed:', error);
      socket.emit('device-registered', { success: false, error: error.message });
    }
  }

  async relayCommand(socket, data) {
    if (!socket.authenticated) {
      socket.emit('error', { message: 'Not authenticated' });
      return;
    }

    try {
      const { deviceId, commandId, args, requestId } = data;
      
      // Check if device is connected
      const deviceSocket = this.connectedDevices.get(deviceId);
      if (!deviceSocket) {
        socket.emit('command-error', { 
          requestId,
          error: 'Device not connected' 
        });
        return;
      }

      // Verify device ownership
      const device = await Device.findOne({
        id: deviceId,
        userId: socket.userId
      });

      if (!device) {
        socket.emit('command-error', { 
          requestId,
          error: 'Device not found or access denied' 
        });
        return;
      }

      // Forward command to device
      deviceSocket.emit('execute-command', {
        commandId,
        args,
        requestId,
        fromUser: socket.userId
      });

      this.logger.info(`Command relayed to device ${deviceId}: ${commandId}`);
    } catch (error) {
      this.logger.error('Command relay failed:', error);
      socket.emit('command-error', { 
        requestId: data.requestId,
        error: error.message 
      });
    }
  }

  async relayResult(socket, data) {
    if (!socket.authenticated || !socket.deviceId) {
      socket.emit('error', { message: 'Not authenticated or not a device' });
      return;
    }

    try {
      const { result, requestId, fromUser } = data;
      
      // Find user socket to send result back
      const userSocket = Array.from(this.connectedClients.values())
        .find(s => s.userId === fromUser);

      if (userSocket) {
        userSocket.emit('command-result', {
          requestId,
          deviceId: socket.deviceId,
          result
        });
      }

      this.logger.info(`Result relayed for request ${requestId}`);
    } catch (error) {
      this.logger.error('Result relay failed:', error);
    }
  }

  async handleDisconnect(socket) {
    // Remove from connected devices if it's a device
    if (socket.deviceId) {
      this.connectedDevices.delete(socket.deviceId);
      
      // Update device status in database
      try {
        await Device.findOneAndUpdate(
          { id: socket.deviceId },
          { 
            status: 'offline',
            lastSeen: new Date()
          }
        );
        this.logger.info(`Device disconnected: ${socket.deviceId}`);
      } catch (error) {
        this.logger.error('Error updating device status:', error);
      }
    }

    // Remove from connected clients
    if (socket.userId) {
      this.connectedClients.delete(socket.userId);
      this.logger.info(`User disconnected: ${socket.userId}`);
    }
  }
}

module.exports = RelayService;