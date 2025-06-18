const express = require('express');
const Device = require('../model/Device');
const { v4: uuidv4 } = require('uuid');

const router = express.Router();

router.get('/', async (req, res) => {
  try {
    const devices = await Device.find({ userId: req.user.userId });
    res.json(devices);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

router.post('/', async (req, res) => {
  try {
    const { name, description, type } = req.body;
    
    const device = new Device({
      id: uuidv4(),
      name,
      description,
      type,
      userId: req.user.userId,
      status: 'offline'
    });
    
    await device.save();
    res.status(201).json(device);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

router.get('/:deviceId', async (req, res) => {
  try {
    const device = await Device.findOne({
      id: req.params.deviceId,
      userId: req.user.userId
    });
    
    if (!device) {
      return res.status(404).json({ error: 'Device not found' });
    }
    
    res.json(device);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

router.put('/:deviceId', async (req, res) => {
  try {
    const { name, description } = req.body;
    
    const device = await Device.findOneAndUpdate(
      {
        id: req.params.deviceId,
        userId: req.user.userId
      },
      { name, description },
      { new: true }
    );
    
    if (!device) {
      return res.status(404).json({ error: 'Device not found' });
    }
    
    res.json(device);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

router.delete('/:deviceId', async (req, res) => {
  try {
    const device = await Device.findOneAndDelete({
      id: req.params.deviceId,
      userId: req.user.userId
    });
    
    if (!device) {
      return res.status(404).json({ error: 'Device not found' });
    }
    
    res.json({ message: 'Device deleted successfully' });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

router.post('/:deviceId/commands', async (req, res) => {
  try {
    const { commandId, args } = req.body;
    
    const device = await Device.findOne({
      id: req.params.deviceId,
      userId: req.user.userId
    });
    
    if (!device) {
      return res.status(404).json({ error: 'Device not found' });
    }
    
    if (device.status !== 'online') {
      return res.status(400).json({ error: 'Device is offline' });
    }
    
    // This will be handled by the relay service through WebSocket
    res.json({ 
      message: 'Command sent',
      commandId,
      deviceId: device.id
    });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

module.exports = router;