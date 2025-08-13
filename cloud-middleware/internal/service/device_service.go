package service

import (
	"fmt"
	"time"

	"github.com/myczh-1/lazy-ctrl-cloud/internal/model"
	"github.com/myczh-1/lazy-ctrl-cloud/internal/repository"
)

// DeviceService handles device-related business logic
type DeviceService struct {
	deviceRepo repository.DeviceRepository
}

// NewDeviceService creates a new device service
func NewDeviceService(deviceRepo repository.DeviceRepository) *DeviceService {
	return &DeviceService{
		deviceRepo: deviceRepo,
	}
}

// RegisterDevice registers a new device
func (ds *DeviceService) RegisterDevice(userID, deviceID, deviceName, deviceType, platform, agentVersion string, systemInfo map[string]interface{}) (*model.Device, error) {
	// Check if device already exists
	existingDevice, err := ds.deviceRepo.GetByID(deviceID)
	if err == nil && existingDevice != nil {
		return nil, fmt.Errorf("device with ID %s already exists", deviceID)
	}

	// Create new device
	device := &model.Device{
		ID:           deviceID,
		DeviceName:   deviceName,
		DeviceType:   deviceType,
		Platform:     platform,
		AgentVersion: agentVersion,
		Online:       true,
		LastSeen:     time.Now(),
		SystemInfo:   systemInfo,
		Settings: &model.DeviceSettings{
			AutoStart:               false,
			AllowRemoteShutdown:     false,
			RequirePinForExecution:  true,
			SecurityLevel:           "medium",
		},
	}

	// Save device
	if err := ds.deviceRepo.Create(device); err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	// Bind device to user
	userDevice := &model.UserDevice{
		UserID:   userID,
		DeviceID: deviceID,
		Role:     "owner",
		Status:   "active",
	}

	if err := ds.deviceRepo.CreateUserDevice(userDevice); err != nil {
		return nil, fmt.Errorf("failed to bind device to user: %w", err)
	}

	return device, nil
}

// BindDeviceToUser binds an existing device to a user
func (ds *DeviceService) BindDeviceToUser(userID, deviceID, role string) error {
	// Check if device exists
	device, err := ds.deviceRepo.GetByID(deviceID)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}
	if device == nil {
		return fmt.Errorf("device %s does not exist", deviceID)
	}

	// Check if user is already bound to this device
	userDevice, err := ds.deviceRepo.GetUserDevice(userID, deviceID)
	if err == nil && userDevice != nil {
		return fmt.Errorf("user %s is already bound to device %s", userID, deviceID)
	}

	// Create user-device relationship
	userDevice = &model.UserDevice{
		UserID:   userID,
		DeviceID: deviceID,
		Role:     role,
		Status:   "active",
	}

	return ds.deviceRepo.CreateUserDevice(userDevice)
}

// UnbindDeviceFromUser removes the binding between a user and a device
func (ds *DeviceService) UnbindDeviceFromUser(userID, deviceID string) error {
	return ds.deviceRepo.DeleteUserDevice(userID, deviceID)
}

// GetUserDevices returns all devices associated with a user
func (ds *DeviceService) GetUserDevices(userID string, onlineOnly bool) ([]*model.Device, error) {
	return ds.deviceRepo.GetUserDevices(userID, onlineOnly)
}

// GetDeviceByID returns a device by its ID
func (ds *DeviceService) GetDeviceByID(deviceID string) (*model.Device, error) {
	return ds.deviceRepo.GetByID(deviceID)
}

// UpdateDeviceStatus updates the online status and last seen time
func (ds *DeviceService) UpdateDeviceStatus(deviceID string, online bool) error {
	device, err := ds.deviceRepo.GetByID(deviceID)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}

	device.Online = online
	device.LastSeen = time.Now()

	return ds.deviceRepo.Update(device)
}

// UpdateDeviceInfo updates device information
func (ds *DeviceService) UpdateDeviceInfo(deviceID, deviceName string, settings *model.DeviceSettings) (*model.Device, error) {
	device, err := ds.deviceRepo.GetByID(deviceID)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	if deviceName != "" {
		device.DeviceName = deviceName
	}

	if settings != nil {
		device.Settings = settings
	}

	if err := ds.deviceRepo.Update(device); err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	return device, nil
}

// UpdateSystemInfo updates device system information
func (ds *DeviceService) UpdateSystemInfo(deviceID string, systemInfo map[string]interface{}) error {
	device, err := ds.deviceRepo.GetByID(deviceID)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}

	device.SystemInfo = systemInfo
	device.LastSeen = time.Now()

	return ds.deviceRepo.Update(device)
}

// DeleteDevice removes a device and all its associations
func (ds *DeviceService) DeleteDevice(deviceID string) error {
	// Delete all user-device relationships first
	if err := ds.deviceRepo.DeleteAllUserDevices(deviceID); err != nil {
		return fmt.Errorf("failed to delete user-device relationships: %w", err)
	}

	// Delete all device commands
	if err := ds.deviceRepo.DeleteAllDeviceCommands(deviceID); err != nil {
		return fmt.Errorf("failed to delete device commands: %w", err)
	}

	// Delete the device
	return ds.deviceRepo.Delete(deviceID)
}

// CheckUserDevicePermission checks if a user has permission to access a device
func (ds *DeviceService) CheckUserDevicePermission(userID, deviceID, requiredRole string) (bool, error) {
	userDevice, err := ds.deviceRepo.GetUserDevice(userID, deviceID)
	if err != nil {
		return false, err
	}

	if userDevice == nil {
		return false, nil
	}

	if userDevice.Status != "active" {
		return false, nil
	}

	// Role hierarchy: owner > admin > user > viewer
	roleHierarchy := map[string]int{
		"viewer": 1,
		"user":   2,
		"admin":  3,
		"owner":  4,
	}

	userRoleLevel, exists := roleHierarchy[userDevice.Role]
	if !exists {
		return false, nil
	}

	requiredRoleLevel, exists := roleHierarchy[requiredRole]
	if !exists {
		return false, nil
	}

	return userRoleLevel >= requiredRoleLevel, nil
}

// GetAllDevices returns all devices (admin function)
func (ds *DeviceService) GetAllDevices() ([]*model.Device, error) {
	return ds.deviceRepo.GetAll()
}

// GetOnlineDevices returns all online devices
func (ds *DeviceService) GetOnlineDevices() ([]*model.Device, error) {
	return ds.deviceRepo.GetOnlineDevices()
}

// UpdateDeviceLastSeen updates the last seen timestamp
func (ds *DeviceService) UpdateDeviceLastSeen(deviceID string) error {
	device, err := ds.deviceRepo.GetByID(deviceID)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}

	device.LastSeen = time.Now()
	return ds.deviceRepo.Update(device)
}