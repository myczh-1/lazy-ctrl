package repository

import (
	"gorm.io/gorm"

	"github.com/myczh-1/lazy-ctrl-cloud/internal/model"
)

// DeviceRepository interface defines device data access methods
type DeviceRepository interface {
	Create(device *model.Device) error
	GetByID(deviceID string) (*model.Device, error)
	GetAll() ([]*model.Device, error)
	GetOnlineDevices() ([]*model.Device, error)
	Update(device *model.Device) error
	Delete(deviceID string) error

	// User-Device relationship methods
	CreateUserDevice(userDevice *model.UserDevice) error
	GetUserDevice(userID, deviceID string) (*model.UserDevice, error)
	GetUserDevices(userID string, onlineOnly bool) ([]*model.Device, error)
	DeleteUserDevice(userID, deviceID string) error
	DeleteAllUserDevices(deviceID string) error

	// Device Command methods
	CreateDeviceCommand(command *model.DeviceCommand) error
	GetDeviceCommand(deviceID, commandID string) (*model.DeviceCommand, error)
	GetDeviceCommands(deviceID string) ([]*model.DeviceCommand, error)
	UpdateDeviceCommand(command *model.DeviceCommand) error
	DeleteDeviceCommand(deviceID, commandID string) error
	DeleteAllDeviceCommands(deviceID string) error

	// Execution Log methods
	CreateExecutionLog(log *model.ExecutionLog) error
	GetExecutionLogs(deviceID string, limit int) ([]*model.ExecutionLog, error)
	GetUserExecutionLogs(userID string, limit int) ([]*model.ExecutionLog, error)
}

// deviceRepository implements the DeviceRepository interface
type deviceRepository struct {
	db *gorm.DB
}

// NewDeviceRepository creates a new device repository
func NewDeviceRepository(db *gorm.DB) DeviceRepository {
	return &deviceRepository{db: db}
}

// Create creates a new device
func (r *deviceRepository) Create(device *model.Device) error {
	return r.db.Create(device).Error
}

// GetByID retrieves a device by its ID
func (r *deviceRepository) GetByID(deviceID string) (*model.Device, error) {
	var device model.Device
	err := r.db.Where("id = ?", deviceID).First(&device).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

// GetAll retrieves all devices
func (r *deviceRepository) GetAll() ([]*model.Device, error) {
	var devices []*model.Device
	err := r.db.Find(&devices).Error
	return devices, err
}

// GetOnlineDevices retrieves all online devices
func (r *deviceRepository) GetOnlineDevices() ([]*model.Device, error) {
	var devices []*model.Device
	err := r.db.Where("online = ?", true).Find(&devices).Error
	return devices, err
}

// Update updates a device
func (r *deviceRepository) Update(device *model.Device) error {
	return r.db.Save(device).Error
}

// Delete deletes a device
func (r *deviceRepository) Delete(deviceID string) error {
	return r.db.Where("id = ?", deviceID).Delete(&model.Device{}).Error
}

// CreateUserDevice creates a user-device relationship
func (r *deviceRepository) CreateUserDevice(userDevice *model.UserDevice) error {
	return r.db.Create(userDevice).Error
}

// GetUserDevice retrieves a user-device relationship
func (r *deviceRepository) GetUserDevice(userID, deviceID string) (*model.UserDevice, error) {
	var userDevice model.UserDevice
	err := r.db.Where("user_id = ? AND device_id = ?", userID, deviceID).First(&userDevice).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &userDevice, nil
}

// GetUserDevices retrieves all devices for a user
func (r *deviceRepository) GetUserDevices(userID string, onlineOnly bool) ([]*model.Device, error) {
	var devices []*model.Device

	query := r.db.Table("devices").
		Joins("JOIN user_devices ON devices.id = user_devices.device_id").
		Where("user_devices.user_id = ? AND user_devices.status = ?", userID, "active")

	if onlineOnly {
		query = query.Where("devices.online = ?", true)
	}

	err := query.Find(&devices).Error
	return devices, err
}

// DeleteUserDevice deletes a user-device relationship
func (r *deviceRepository) DeleteUserDevice(userID, deviceID string) error {
	return r.db.Where("user_id = ? AND device_id = ?", userID, deviceID).Delete(&model.UserDevice{}).Error
}

// DeleteAllUserDevices deletes all user-device relationships for a device
func (r *deviceRepository) DeleteAllUserDevices(deviceID string) error {
	return r.db.Where("device_id = ?", deviceID).Delete(&model.UserDevice{}).Error
}

// CreateDeviceCommand creates a device command
func (r *deviceRepository) CreateDeviceCommand(command *model.DeviceCommand) error {
	return r.db.Create(command).Error
}

// GetDeviceCommand retrieves a device command
func (r *deviceRepository) GetDeviceCommand(deviceID, commandID string) (*model.DeviceCommand, error) {
	var command model.DeviceCommand
	err := r.db.Where("device_id = ? AND command_id = ?", deviceID, commandID).First(&command).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &command, nil
}

// GetDeviceCommands retrieves all commands for a device
func (r *deviceRepository) GetDeviceCommands(deviceID string) ([]*model.DeviceCommand, error) {
	var commands []*model.DeviceCommand
	err := r.db.Where("device_id = ?", deviceID).Find(&commands).Error
	return commands, err
}

// UpdateDeviceCommand updates a device command
func (r *deviceRepository) UpdateDeviceCommand(command *model.DeviceCommand) error {
	return r.db.Save(command).Error
}

// DeleteDeviceCommand deletes a device command
func (r *deviceRepository) DeleteDeviceCommand(deviceID, commandID string) error {
	return r.db.Where("device_id = ? AND command_id = ?", deviceID, commandID).Delete(&model.DeviceCommand{}).Error
}

// DeleteAllDeviceCommands deletes all commands for a device
func (r *deviceRepository) DeleteAllDeviceCommands(deviceID string) error {
	return r.db.Where("device_id = ?", deviceID).Delete(&model.DeviceCommand{}).Error
}

// CreateExecutionLog creates an execution log entry
func (r *deviceRepository) CreateExecutionLog(log *model.ExecutionLog) error {
	return r.db.Create(log).Error
}

// GetExecutionLogs retrieves execution logs for a device
func (r *deviceRepository) GetExecutionLogs(deviceID string, limit int) ([]*model.ExecutionLog, error) {
	var logs []*model.ExecutionLog
	query := r.db.Where("device_id = ?", deviceID).Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&logs).Error
	return logs, err
}

// GetUserExecutionLogs retrieves execution logs for a user
func (r *deviceRepository) GetUserExecutionLogs(userID string, limit int) ([]*model.ExecutionLog, error) {
	var logs []*model.ExecutionLog
	query := r.db.Where("user_id = ?", userID).Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&logs).Error
	return logs, err
}