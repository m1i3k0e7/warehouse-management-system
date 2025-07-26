package mqtt

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    mqtt "github.com/eclipse/paho.mqtt.golang"
    "warehouse/internal/domain/services"
    "warehouse/pkg/logger"
)

// MQTT 消息結構定義
type ShelfEvent struct {
    ShelfID     string    `json:"shelf_id"`
    SlotID      string    `json:"slot_id"`
    EventType   string    `json:"event_type"` // "material_detected", "material_removed", "slot_error"
    MaterialBarcode string `json:"material_barcode,omitempty"`
    Timestamp   int64     `json:"timestamp"`
    SensorData  SensorData `json:"sensor_data,omitempty"`
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
    client           mqtt.Client
    inventoryService *services.InventoryService
    topicPrefix      string
    retryService     *services.RetryService
}

func NewMQTTHandler(brokerURL string, inventoryService *services.InventoryService, retryService *services.RetryService) *MQTTHandler {
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
        client:           client,
        inventoryService: inventoryService,
        topicPrefix:      "warehouse/shelf",
        retryService:     retryService,
    }
}

func (h *MQTTHandler) Connect() error {
    if token := h.client.Connect(); token.Wait() && token.Error() != nil {
        return token.Error()
    }
    
    // 訂閱料架事件
    eventTopic := fmt.Sprintf("%s/+/events", h.topicPrefix)
    if token := h.client.Subscribe(eventTopic, 1, h.handleShelfEvent); token.Wait() && token.Error() != nil {
        return token.Error()
    }
    
    // 訂閱料架狀態
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
    
    // 使用重試機制處理事件
    h.retryService.ExecuteWithRetry(func() error {
        return h.processShelfEvent(&event)
    }, 3, time.Second*2)
}

func (h *MQTTHandler) processShelfEvent(event *ShelfEvent) error {
    ctx := context.WithTimeout(context.Background(), 30*time.Second)
    defer ctx.Done()
    
    switch event.EventType {
    case "material_detected":
        cmd := services.PlaceMaterialCommand{
            MaterialBarcode: event.MaterialBarcode,
            SlotID:         event.SlotID,
            OperatorID:     "SHELF_SYSTEM",
            SensorData: &services.SensorData{
                Weight:      event.SensorData.Weight,
                Temperature: event.SensorData.Temperature,
                Humidity:    event.SensorData.Humidity,
            },
        }
        return h.inventoryService.PlaceMaterial(ctx, cmd)
        
    case "material_removed":
        cmd := services.RemoveMaterialCommand{
            SlotID:     event.SlotID,
            OperatorID: "SHELF_SYSTEM",
        }
        return h.inventoryService.RemoveMaterial(ctx, cmd)
        
    case "slot_error":
        return h.inventoryService.HandleSlotError(ctx, event.SlotID, "sensor_error")
        
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
    h.inventoryService.UpdateShelfStatus(ctx, status.ShelfID, status.Status)
}