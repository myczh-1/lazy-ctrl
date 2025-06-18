const express = require('express');
const router = express.Router();

// Get relay status
router.get('/status', (req, res) => {
  res.json({ status: 'active' });
});

// Execute command through relay
router.post('/execute', (req, res) => {
  const { command, targetDeviceId } = req.body;
  
  if (!command || !targetDeviceId) {
    return res.status(400).json({ error: 'Command and targetDeviceId are required' });
  }
  
  // Command will be handled by the relay service via WebSocket
  res.json({ 
    success: true, 
    message: 'Command queued for execution',
    command,
    targetDeviceId
  });
});

module.exports = router;