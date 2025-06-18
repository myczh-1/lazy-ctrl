const express = require('express');
const http = require('http');
const socketIo = require('socket.io');
const cors = require('cors');
const helmet = require('helmet');
const rateLimit = require('express-rate-limit');
const mongoose = require('mongoose');
const winston = require('winston');

const authController = require('./controller/auth');
const deviceController = require('./controller/device');
const relayController = require('./controller/relay');
const authMiddleware = require('./middleware/auth');

const app = express();
const server = http.createServer(app);
const io = socketIo(server, {
  cors: {
    origin: "*",
    methods: ["GET", "POST"]
  }
});

const logger = winston.createLogger({
  level: 'info',
  format: winston.format.json(),
  transports: [
    new winston.transports.File({ filename: 'logs/error.log', level: 'error' }),
    new winston.transports.File({ filename: 'logs/combined.log' }),
    new winston.transports.Console({
      format: winston.format.simple()
    })
  ]
});

const limiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100 // limit each IP to 100 requests per windowMs
});

app.use(helmet());
app.use(cors());
app.use(limiter);
app.use(express.json());

mongoose.connect(process.env.MONGODB_URI || 'mongodb://localhost:27017/lazy-ctrl', {
  useNewUrlParser: true,
  useUnifiedTopology: true
});

app.use('/api/auth', authController);
app.use('/api/devices', authMiddleware, deviceController);
app.use('/api/relay', authMiddleware, relayController);

const RelayService = require('./service/relay');
const relayService = new RelayService(io, logger);

io.on('connection', (socket) => {
  logger.info('Client connected:', socket.id);
  
  socket.on('authenticate', (token) => {
    relayService.authenticateSocket(socket, token);
  });
  
  socket.on('register-device', (data) => {
    relayService.registerDevice(socket, data);
  });
  
  socket.on('execute-command', (data) => {
    relayService.relayCommand(socket, data);
  });
  
  socket.on('command-result', (data) => {
    relayService.relayResult(socket, data);
  });
  
  socket.on('disconnect', () => {
    logger.info('Client disconnected:', socket.id);
    relayService.handleDisconnect(socket);
  });
});

const PORT = process.env.PORT || 3000;
server.listen(PORT, () => {
  logger.info(`Cloud service running on port ${PORT}`);
});