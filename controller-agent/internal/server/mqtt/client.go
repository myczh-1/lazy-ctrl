package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/myczh-1/lazy-ctrl-agent/internal/config"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/command"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/executor"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/security"
	"github.com/sirupsen/logrus"
)

type Client struct {
	config          *config.Config
	logger          *logrus.Logger
	commandService  *command.Service
	executorService *executor.Service
	securityService *security.Service
	mqttClient      mqtt.Client
	ctx             context.Context
	cancel          context.CancelFunc
}

type CommandMessage struct {
	CommandID string `json:"command_id"`
	Timeout   int    `json:"timeout,omitempty"`
	Pin       string `json:"pin,omitempty"`
}

type ResponseMessage struct {
	Success       bool   `json:"success"`
	Output        string `json:"output,omitempty"`
	Error         string `json:"error,omitempty"`
	ExitCode      int    `json:"exit_code"`
	ExecutionTime int64  `json:"execution_time_ms"`
	Timestamp     int64  `json:"timestamp"`
}

func NewClient(
	config *config.Config,
	logger *logrus.Logger,
	commandService *command.Service,
	executorService *executor.Service,
	securityService *security.Service,
) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Client{
		config:          config,
		logger:          logger,
		commandService:  commandService,
		executorService: executorService,
		securityService: securityService,
		ctx:             ctx,
		cancel:          cancel,
	}
}

func (c *Client) Start() error {
	if !c.config.MQTT.Enabled {
		c.logger.Info("MQTT client disabled")
		return nil
	}

	opts := mqtt.NewClientOptions()
	broker := fmt.Sprintf("tcp://%s:%d", c.config.MQTT.Broker, c.config.MQTT.Port)
	opts.AddBroker(broker)
	opts.SetClientID(c.config.MQTT.ClientID)
	
	if c.config.MQTT.Username != "" {
		opts.SetUsername(c.config.MQTT.Username)
		opts.SetPassword(c.config.MQTT.Password)
	}

	opts.SetDefaultPublishHandler(c.defaultMessageHandler)
	opts.OnConnect = c.onConnect
	opts.OnConnectionLost = c.onConnectionLost
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)

	c.mqttClient = mqtt.NewClient(opts)
	
	c.logger.WithField("broker", broker).Info("Connecting to MQTT broker")
	
	if token := c.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	return nil
}

func (c *Client) onConnect(client mqtt.Client) {
	c.logger.Info("Connected to MQTT broker")
	
	// 订阅命令主题
	commandTopic := fmt.Sprintf("%s/command", c.config.MQTT.TopicBase)
	if token := client.Subscribe(commandTopic, 1, c.handleCommandMessage); token.Wait() && token.Error() != nil {
		c.logger.WithError(token.Error()).Error("Failed to subscribe to command topic")
	} else {
		c.logger.WithField("topic", commandTopic).Info("Subscribed to command topic")
	}
}

func (c *Client) onConnectionLost(client mqtt.Client, err error) {
	c.logger.WithError(err).Warn("MQTT connection lost")
}

func (c *Client) defaultMessageHandler(client mqtt.Client, msg mqtt.Message) {
	c.logger.WithFields(logrus.Fields{
		"topic":   msg.Topic(),
		"payload": string(msg.Payload()),
	}).Debug("Received MQTT message")
}

func (c *Client) handleCommandMessage(client mqtt.Client, msg mqtt.Message) {
	c.logger.WithFields(logrus.Fields{
		"topic":   msg.Topic(),
		"payload": string(msg.Payload()),
	}).Info("Received command message")

	var cmdMsg CommandMessage
	if err := json.Unmarshal(msg.Payload(), &cmdMsg); err != nil {
		c.logger.WithError(err).Error("Failed to parse command message")
		c.publishResponse(msg.Topic(), &ResponseMessage{
			Success:   false,
			Error:     "Invalid JSON format",
			ExitCode:  1,
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// 安全检查
	clientID := c.securityService.GetClientID("mqtt", "mqtt")
	if err := c.securityService.CheckRateLimit(clientID); err != nil {
		c.publishResponse(msg.Topic(), &ResponseMessage{
			Success:   false,
			Error:     err.Error(),
			ExitCode:  1,
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// PIN验证
	if c.config.Security.PinRequired {
		if !c.securityService.ValidatePin(cmdMsg.Pin) {
			c.publishResponse(msg.Topic(), &ResponseMessage{
				Success:   false,
				Error:     "Invalid or missing PIN",
				ExitCode:  1,
				Timestamp: time.Now().Unix(),
			})
			return
		}
	}

	// 验证命令访问权限
	if err := c.securityService.ValidateCommandAccess(cmdMsg.CommandID); err != nil {
		c.publishResponse(msg.Topic(), &ResponseMessage{
			Success:   false,
			Error:     err.Error(),
			ExitCode:  1,
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// 获取并执行命令
	c.executeCommand(msg.Topic(), &cmdMsg)
}

func (c *Client) executeCommand(responseTopic string, cmdMsg *CommandMessage) {
	startTime := time.Now()

	// 获取命令
	cmd, ok := c.commandService.GetCommand(cmdMsg.CommandID)
	if !ok {
		c.publishResponse(responseTopic, &ResponseMessage{
			Success:       false,
			Error:         fmt.Sprintf("Command not found: %s", cmdMsg.CommandID),
			ExitCode:      1,
			ExecutionTime: time.Since(startTime).Milliseconds(),
			Timestamp:     time.Now().Unix(),
		})
		return
	}

	// 获取平台特定命令
	platformCmd, ok := c.commandService.GetPlatformCommand(cmd)
	if !ok {
		c.publishResponse(responseTopic, &ResponseMessage{
			Success:       false,
			Error:         "Command not supported on this platform",
			ExitCode:      1,
			ExecutionTime: time.Since(startTime).Milliseconds(),
			Timestamp:     time.Now().Unix(),
		})
		return
	}

	// 执行命令
	var result *executor.ExecutionResult
	var err error

	if cmdMsg.Timeout > 0 {
		result, err = c.executorService.ExecuteWithTimeout(platformCmd, time.Duration(cmdMsg.Timeout)*time.Second)
	} else {
		result, err = c.executorService.ExecuteWithTimeout(platformCmd, 30*time.Second)
	}

	response := &ResponseMessage{
		Timestamp: time.Now().Unix(),
	}

	if err != nil {
		response.Success = false
		response.Error = err.Error()
		response.ExitCode = 1
		response.ExecutionTime = time.Since(startTime).Milliseconds()
	} else {
		response.Success = result.Success
		response.Output = result.Output
		response.Error = result.Error
		response.ExitCode = result.ExitCode
		response.ExecutionTime = result.ExecutionTime.Milliseconds()
	}

	c.publishResponse(responseTopic, response)
}

func (c *Client) publishResponse(requestTopic string, response *ResponseMessage) {
	// 构造响应主题
	responseTopic := fmt.Sprintf("%s/response", c.config.MQTT.TopicBase)
	
	responseData, err := json.Marshal(response)
	if err != nil {
		c.logger.WithError(err).Error("Failed to marshal response")
		return
	}

	if token := c.mqttClient.Publish(responseTopic, 1, false, responseData); token.Wait() && token.Error() != nil {
		c.logger.WithError(token.Error()).Error("Failed to publish response")
	} else {
		c.logger.WithField("topic", responseTopic).Debug("Published response")
	}
}

func (c *Client) Stop() {
	c.logger.Info("Stopping MQTT client")
	c.cancel()
	
	if c.mqttClient != nil && c.mqttClient.IsConnected() {
		c.mqttClient.Disconnect(250)
	}
}