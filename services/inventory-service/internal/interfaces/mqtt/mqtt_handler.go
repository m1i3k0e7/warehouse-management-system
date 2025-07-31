package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"WMS/services/inventory-service/internal/application/commands"
	"WMS/services/inventory-service/internal/domain/services"
	"WMS/services/inventory-service/pkg/utils/logger"
)

// MQTT message structure for shelf events and status updates
type ShelfEvent struct {
	ShelfID         string     `json:"shelf_id"`
	SlotID          string     `json:"slot_id"`
	EventType       string     `json:"event_type"` // "material_detected", "material_removed", "slot_error"
	MaterialBarcode string     `json:"material_barcode,omitempty"`
	Timestamp       int64      `json:"timestamp"`
	SensorData      *SensorData `json:"sensor_data,omitempty"`
}

type SensorData struct {
	Weight      float64 `json:"weight,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	Humidity    float64 `json:"humidity,omitempty"`
	LightLevel  int     `json:"light_level,omitempty"`
}

type ShelfStatus struct {
	ShelfID   string `json:"shelf_id"`
	Status    string `json:"status"` // "online", "offline", "maintenance"
	Timestamp int64  `json:"timestamp"`
}

type MQTTHandler struct {
	client                   mqtt.Client
	placeMaterialHandler     *commands.PlaceMaterialCommandHandler
	removeMaterialHandler    *commands.RemoveMaterialCommandHandler
	handleSlotErrorHandler   *commands.HandleSlotErrorCommandHandler
	updateShelfStatusHandler *commands.UpdateShelfStatusCommandHandler
	inventoryService         *services.InventoryService // New dependency
	topicPrefix              string
	retryService             *services.RetryService
}

func NewMQTTHandler(
	brokerURL string,
	placeMaterialHandler *commands.PlaceMaterialCommandHandler,
	removeMaterialHandler *commands.RemoveMaterialCommandHandler,
	handleSlotErrorHandler *commands.HandleSlotErrorCommandHandler,
	updateShelfStatusHandler *commands.UpdateShelfStatusCommandHandler,
	inventoryService *services.InventoryService, // New parameter
	retryService *services.RetryService,
) *MQTTHandler {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID("inventory-service")
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		logger.Error("MQTT connection lost", err)
	})

	opts.SetReconnectingHandler(func(client mqtt.Client, options *mqtt.ClientOptions) {
		logger.Info("MQTT reconnecting...")
	})

	client := mqtt.NewClient(opts)

	return &MQTTHandler{
		client:                   client,
		placeMaterialHandler:     placeMaterialHandler,
		removeMaterialHandler:    removeMaterialHandler,
		handleSlotErrorHandler:   handleSlotErrorHandler,
		updateShelfStatusHandler: updateShelfStatusHandler,
		inventoryService:         inventoryService, // Initialize new dependency
		topicPrefix:              "WMS/services/inventory-service/shelf",
		retryService:             retryService,
	}
}

func (h *MQTTHandler) Connect() error {
	if token := h.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	// subscribe to shelf events and status updates
	eventTopic := fmt.Sprintf("%s/+/events", h.topicPrefix)
	if token := h.client.Subscribe(eventTopic, 1, h.handleShelfEvent); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	// subscribe to shelf status updates
	statusTopic := fmt.Sprintf("%s/+/status", h.topicPrefix)
	if token := h.client.Subscribe(statusTopic, 1, h.handleShelfStatus); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	logger.Info("MQTT Handler connected and subscribed")
	return nil
}

func (h *MQTTHandler) handleShelfEvent(client mqtt.Client, msg mqtt.Message) {
	var event ShelfEvent
	if err := json.Unmarshal(msg.Payload(), &event); err != nil {
		logger.Error("Failed to unmarshal shelf event", err)
		return
	}

	// handle the event with retry logic
	h.retryService.Execute(func() error {
		return h.processShelfEvent(&event)
	})
}

func (h *MQTTHandler) processShelfEvent(event *ShelfEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch event.EventType {
		case services.EventTypeMaterialDetected:
			return h.inventoryService.HandleMaterialDetectedEvent(ctx, event.SlotID, event.MaterialBarcode)

		case services.EventTypeMaterialRemoved:
			return h.inventoryService.HandleMaterialRemovedEvent(ctx, event.SlotID, event.MaterialBarcode)

		// case services.EventTypeSlotError:
		// 	cmd := commands.HandleSlotErrorCommand{
		// 		SlotID:    event.SlotID,
		// 		ErrorType: string(entities.AlertTypeSlotError),
		// 	}
		// 	return h.handleSlotErrorHandler.Handle(ctx, cmd)

		default:
			return fmt.Errorf("unknown event type: %s", event.EventType)
	}
}

func (h *MQTTHandler) handleShelfStatus(client mqtt.Client, msg mqtt.Message) {
	var status ShelfStatus
	if err := json.Unmarshal(msg.Payload(), &status); err != nil {
		logger.Error("Failed to unmarshal shelf status", err)
		return
	}

	ctx := context.Background()
	cmd := commands.UpdateShelfStatusCommand{
		ShelfID: status.ShelfID,
		Status:  status.Status,
	}
	if err := h.updateShelfStatusHandler.Handle(ctx, cmd); err != nil {
		logger.Error("Failed to handle shelf status update", err)
	}
}
