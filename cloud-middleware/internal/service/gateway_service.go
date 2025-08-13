package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	controllerPb "github.com/myczh-1/lazy-ctrl-agent/proto"
)

// DeviceConnection represents a gRPC connection to a specific device
type DeviceConnection struct {
	DeviceID     string
	Address      string
	Connection   *grpc.ClientConn
	Client       controllerPb.ControllerServiceClient
	LastPing     time.Time
	IsHealthy    bool
	ConnectedAt  time.Time
	mutex        sync.RWMutex
}

// GatewayService manages gRPC connections to multiple devices
type GatewayService struct {
	connections map[string]*DeviceConnection
	mutex       sync.RWMutex
	
	// Connection pool settings
	maxConnections int
	connectTimeout time.Duration
	pingInterval   time.Duration
	
	// Health check settings
	healthCheckInterval time.Duration
	maxRetries          int
}

// NewGatewayService creates a new gateway service instance
func NewGatewayService() *GatewayService {
	return &GatewayService{
		connections:         make(map[string]*DeviceConnection),
		maxConnections:      100,
		connectTimeout:      10 * time.Second,
		pingInterval:        30 * time.Second,
		healthCheckInterval: 60 * time.Second,
		maxRetries:          3,
	}
}

// AddDevice adds a new device connection
func (gs *GatewayService) AddDevice(deviceID, address string) error {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	if len(gs.connections) >= gs.maxConnections {
		return fmt.Errorf("maximum number of connections reached")
	}

	// Check if device already exists
	if _, exists := gs.connections[deviceID]; exists {
		return fmt.Errorf("device %s already connected", deviceID)
	}

	// Create new connection
	ctx, cancel := context.WithTimeout(context.Background(), gs.connectTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to device %s at %s: %w", deviceID, address, err)
	}

	client := controllerPb.NewControllerServiceClient(conn)

	deviceConn := &DeviceConnection{
		DeviceID:    deviceID,
		Address:     address,
		Connection:  conn,
		Client:      client,
		LastPing:    time.Now(),
		IsHealthy:   true,
		ConnectedAt: time.Now(),
	}

	gs.connections[deviceID] = deviceConn

	// Start health checking for this device
	go gs.healthCheckWorker(deviceID)

	log.Printf("Device %s connected at %s", deviceID, address)
	return nil
}

// RemoveDevice removes a device connection
func (gs *GatewayService) RemoveDevice(deviceID string) error {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	conn, exists := gs.connections[deviceID]
	if !exists {
		return fmt.Errorf("device %s not found", deviceID)
	}

	// Close the connection
	if err := conn.Connection.Close(); err != nil {
		log.Printf("Error closing connection for device %s: %v", deviceID, err)
	}

	delete(gs.connections, deviceID)
	log.Printf("Device %s disconnected", deviceID)
	return nil
}

// GetDeviceClient returns the gRPC client for a device
func (gs *GatewayService) GetDeviceClient(deviceID string) (controllerPb.ControllerServiceClient, error) {
	gs.mutex.RLock()
	defer gs.mutex.RUnlock()

	conn, exists := gs.connections[deviceID]
	if !exists {
		return nil, fmt.Errorf("device %s not connected", deviceID)
	}

	if !conn.IsHealthy {
		return nil, fmt.Errorf("device %s is not healthy", deviceID)
	}

	return conn.Client, nil
}

// ListConnectedDevices returns a list of connected device IDs
func (gs *GatewayService) ListConnectedDevices() []string {
	gs.mutex.RLock()
	defer gs.mutex.RUnlock()

	devices := make([]string, 0, len(gs.connections))
	for deviceID, conn := range gs.connections {
		if conn.IsHealthy {
			devices = append(devices, deviceID)
		}
	}

	return devices
}

// GetDeviceStatus returns the status of a specific device
func (gs *GatewayService) GetDeviceStatus(deviceID string) (*DeviceConnection, error) {
	gs.mutex.RLock()
	defer gs.mutex.RUnlock()

	conn, exists := gs.connections[deviceID]
	if !exists {
		return nil, fmt.Errorf("device %s not found", deviceID)
	}

	return conn, nil
}

// healthCheckWorker performs periodic health checks for a device
func (gs *GatewayService) healthCheckWorker(deviceID string) {
	ticker := time.NewTicker(gs.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			gs.performHealthCheck(deviceID)
		}

		// Check if device still exists
		gs.mutex.RLock()
		_, exists := gs.connections[deviceID]
		gs.mutex.RUnlock()

		if !exists {
			return // Device was removed, stop health checking
		}
	}
}

// performHealthCheck performs a health check on a specific device
func (gs *GatewayService) performHealthCheck(deviceID string) {
	gs.mutex.RLock()
	conn, exists := gs.connections[deviceID]
	gs.mutex.RUnlock()

	if !exists {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection state
	state := conn.Connection.GetState()
	if state == connectivity.TransientFailure || state == connectivity.Shutdown {
		conn.mutex.Lock()
		conn.IsHealthy = false
		conn.mutex.Unlock()
		log.Printf("Device %s health check failed: connection state %v", deviceID, state)
		return
	}

	// Try to ping the device
	_, err := conn.Client.HealthCheck(ctx, &controllerPb.HealthCheckRequest{})
	
	conn.mutex.Lock()
	if err != nil {
		conn.IsHealthy = false
		log.Printf("Device %s health check failed: %v", deviceID, err)
	} else {
		conn.IsHealthy = true
		conn.LastPing = time.Now()
	}
	conn.mutex.Unlock()
}

// ExecuteCommand executes a command on a specific device
func (gs *GatewayService) ExecuteCommand(deviceID, commandID string, timeout int32) (*controllerPb.ExecuteCommandResponse, error) {
	client, err := gs.GetDeviceClient(deviceID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &controllerPb.ExecuteCommandRequest{
		CommandId:      commandID,
		TimeoutSeconds: timeout,
	}

	return client.ExecuteCommand(ctx, req)
}

// ListCommands retrieves all commands from a device
func (gs *GatewayService) ListCommands(deviceID string) (*controllerPb.ListCommandsResponse, error) {
	client, err := gs.GetDeviceClient(deviceID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &controllerPb.ListCommandsRequest{}
	return client.ListCommands(ctx, req)
}

// ReloadConfig reloads configuration on a device
func (gs *GatewayService) ReloadConfig(deviceID string) (*controllerPb.ReloadConfigResponse, error) {
	client, err := gs.GetDeviceClient(deviceID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &controllerPb.ReloadConfigRequest{}
	return client.ReloadConfig(ctx, req)
}

// Stop gracefully stops the gateway service
func (gs *GatewayService) Stop() error {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	var lastErr error
	for deviceID, conn := range gs.connections {
		if err := conn.Connection.Close(); err != nil {
			log.Printf("Error closing connection for device %s: %v", deviceID, err)
			lastErr = err
		}
	}

	// Clear all connections
	gs.connections = make(map[string]*DeviceConnection)
	return lastErr
}