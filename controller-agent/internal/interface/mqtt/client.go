package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"

	"github.com/myczh-1/lazy-ctrl-agent/internal/config"
	"github.com/myczh-1/lazy-ctrl-agent/internal/command/service"
	"github.com/myczh-1/lazy-ctrl-agent/internal/infrastructure/executor"
	"github.com/myczh-1/lazy-ctrl-agent/internal/infrastructure/security"
)

// Client represents the MQTT client
type Client struct {
	config          *config.Config
	logger          *logrus.Logger
	commandService  *service.CommandService
	executorService *executor.Service
	securityService *security.Service
	client          mqtt.Client
}

// ExecuteRequest represents MQTT execute request
type ExecuteRequest struct {
	CommandID string `json:"commandId"`
	Pin       string `json:"pin,omitempty"`
}

// ExecuteResponse represents MQTT execute response
type ExecuteResponse struct {
	Success  bool   `json:"success"`
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
	ExitCode int    `json:"exitCode"`
}

// NewClient creates a new MQTT client instance
func NewClient(
	cfg *config.Config,
	logger *logrus.Logger,
	commandService *service.CommandService,
	executorService *executor.Service,
	securityService *security.Service,
) *Client {
	return &Client{
		config:          cfg,
		logger:          logger,
		commandService:  commandService,
		executorService: executorService,
		securityService: securityService,
	}
}

// Start starts the MQTT client
func (c *Client) Start() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", c.config.MQTT.Broker, c.config.MQTT.Port))
	opts.SetClientID(c.config.MQTT.ClientID)
	
	if c.config.MQTT.Username != "" {
		opts.SetUsername(c.config.MQTT.Username)
	}
	if c.config.MQTT.Password != "" {
		opts.SetPassword(c.config.MQTT.Password)
	}

	opts.SetDefaultPublishHandler(c.messageHandler)
	opts.OnConnect = c.onConnect
	opts.OnConnectionLost = c.onConnectionLost

	c.client = mqtt.NewClient(opts)

	c.logger.Info("Connecting to MQTT broker")
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	return nil
}

// Stop stops the MQTT client
func (c *Client) Stop() {
	c.logger.Info("Disconnecting from MQTT broker")
	if c.client != nil && c.client.IsConnected() {
		c.client.Disconnect(250)
	}
}

// onConnect handles MQTT connection event
func (c *Client) onConnect(client mqtt.Client) {
	c.logger.Info("MQTT client connected")
	
	// Subscribe to execute topic
	executeTopic := fmt.Sprintf("%s/execute", c.config.MQTT.TopicBase)
	if token := client.Subscribe(executeTopic, 1, c.executeHandler); token.Wait() && token.Error() != nil {
		c.logger.WithError(token.Error()).Error("Failed to subscribe to execute topic")
	} else {
		c.logger.WithField("topic", executeTopic).Info("Subscribed to MQTT topic")
	}
	
	// Subscribe to commands topic
	commandsTopic := fmt.Sprintf("%s/commands", c.config.MQTT.TopicBase)
	if token := client.Subscribe(commandsTopic, 1, c.commandsHandler); token.Wait() && token.Error() != nil {
		c.logger.WithError(token.Error()).Error("Failed to subscribe to commands topic")
	} else {
		c.logger.WithField("topic", commandsTopic).Info("Subscribed to MQTT topic")
	}
}

// onConnectionLost handles MQTT connection lost event
func (c *Client) onConnectionLost(client mqtt.Client, err error) {
	c.logger.WithError(err).Warn("MQTT connection lost")
}

// messageHandler handles default MQTT messages
func (c *Client) messageHandler(client mqtt.Client, msg mqtt.Message) {
	c.logger.WithFields(logrus.Fields{
		"topic":   msg.Topic(),
		"payload": string(msg.Payload()),
	}).Debug("Received MQTT message")
}

// executeHandler handles execute command requests
func (c *Client) executeHandler(client mqtt.Client, msg mqtt.Message) {
	c.logger.WithField("topic", msg.Topic()).Info("Received execute request")
	
	var req ExecuteRequest
	if err := json.Unmarshal(msg.Payload(), &req); err != nil {
		c.logger.WithError(err).Error("Failed to parse execute request")
		c.publishError(client, "invalid_request", "Failed to parse request", msg.Topic())
		return
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Execute the command
	response := c.executeCommand(ctx, req)
	
	// Publish response
	responseTopic := fmt.Sprintf("%s/response", c.config.MQTT.TopicBase)
	responseData, _ := json.Marshal(response)
	
	if token := client.Publish(responseTopic, 1, false, responseData); token.Wait() && token.Error() != nil {
		c.logger.WithError(token.Error()).Error("Failed to publish execute response")
	}
}

// commandsHandler handles get commands requests
func (c *Client) commandsHandler(client mqtt.Client, msg mqtt.Message) {
	c.logger.WithField("topic", msg.Topic()).Info("Received commands request")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	commands, err := c.commandService.GetAllCommands(ctx)
	if err != nil {
		c.publishError(client, "internal_error", "Failed to get commands", msg.Topic())
		return
	}
	
	// Convert to simple format
	simpleCommands := make([]map[string]interface{}, len(commands))
	for i, cmd := range commands {
		simpleCommands[i] = map[string]interface{}{
			"id":          cmd.ID,
			"name":        cmd.Name,
			"description": cmd.Description,
			"category":    cmd.Category,
			"platform":    cmd.Platform,
			"available":   cmd.IsAvailableOnPlatform(),
			"requiresPin": cmd.RequiresPin(),
		}
	}
	
	// Publish response
	responseTopic := fmt.Sprintf("%s/response", c.config.MQTT.TopicBase)
	responseData, _ := json.Marshal(map[string]interface{}{
		"success":  true,
		"commands": simpleCommands,
	})
	
	if token := client.Publish(responseTopic, 1, false, responseData); token.Wait() && token.Error() != nil {
		c.logger.WithError(token.Error()).Error("Failed to publish commands response")
	}
}

// executeCommand executes a command and returns response
func (c *Client) executeCommand(ctx context.Context, req ExecuteRequest) ExecuteResponse {
	// Get command
	cmd, err := c.commandService.GetCommand(ctx, req.CommandID)
	if err != nil {
		return ExecuteResponse{
			Success:  false,
			Error:    fmt.Sprintf("Command not found: %s", req.CommandID),
			ExitCode: -1,
		}
	}
	
	// PIN verification if required
	if cmd.RequiresPin() {
		if !c.securityService.ValidatePin(req.Pin) {
			return ExecuteResponse{
				Success:  false,
				Error:    "Invalid or missing PIN",
				ExitCode: -1,
			}
		}
	}
	
	// Get platform command
	platformCommand, err := c.commandService.GetPlatformCommand(ctx, req.CommandID)
	if err != nil {
		return ExecuteResponse{
			Success:  false,
			Error:    fmt.Sprintf("Command not available: %s", err.Error()),
			ExitCode: -1,
		}
	}
	
	// Execute with timeout
	executeCtx, cancel := context.WithTimeout(ctx, time.Duration(cmd.GetTimeout())*time.Millisecond)
	defer cancel()
	
	result, err := c.executorService.Execute(executeCtx, platformCommand)
	if err != nil {
		return ExecuteResponse{
			Success:  false,
			Error:    fmt.Sprintf("Execution failed: %s", err.Error()),
			ExitCode: -1,
		}
	}
	
	return ExecuteResponse{
		Success:  result.Success,
		Output:   result.Output,
		Error:    result.Error,
		ExitCode: result.ExitCode,
	}
}

// publishError publishes an error response
func (c *Client) publishError(client mqtt.Client, code, message, originalTopic string) {
	errorResponse := map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	}
	
	responseData, _ := json.Marshal(errorResponse)
	responseTopic := fmt.Sprintf("%s/response", c.config.MQTT.TopicBase)
	
	if token := client.Publish(responseTopic, 1, false, responseData); token.Wait() && token.Error() != nil {
		c.logger.WithError(token.Error()).Error("Failed to publish error response")
	}
}