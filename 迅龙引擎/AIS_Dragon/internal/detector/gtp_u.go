package detector

import (
	"sync"
	"time"
)

// GTPv1 message types
type GTPv1MessageType uint8

// GTPv1 message type objects
type GTPv1Message struct {
	Type GTPv1MessageType
	Name string
}

const (
	// Echo messages
	GTPv1_ECHO_REQUEST          GTPv1MessageType = 1
	GTPv1_ECHO_RESPONSE         GTPv1MessageType = 2
	GTPv1_VERSION_NOT_SUPPORTED GTPv1MessageType = 3

	// Node alive messages
	GTPv1_NODE_ALIVE_REQUEST  GTPv1MessageType = 4
	GTPv1_NODE_ALIVE_RESPONSE GTPv1MessageType = 5

	// Redirection messages
	GTPv1_REDIRECTION_REQUEST  GTPv1MessageType = 6
	GTPv1_REDIRECTION_RESPONSE GTPv1MessageType = 7

	// PDP context messages
	GTPv1_CREATE_PDP_CONTEXT_REQUEST               GTPv1MessageType = 16
	GTPv1_CREATE_PDP_CONTEXT_RESPONSE              GTPv1MessageType = 17
	GTPv1_UPDATE_PDP_CONTEXT_REQUEST               GTPv1MessageType = 18
	GTPv1_UPDATE_PDP_CONTEXT_RESPONSE              GTPv1MessageType = 19
	GTPv1_DELETE_PDP_CONTEXT_REQUEST               GTPv1MessageType = 20
	GTPv1_DELETE_PDP_CONTEXT_RESPONSE              GTPv1MessageType = 21
	GTPv1_INITIATE_PDP_CONTEXT_ACTIVATION_REQUEST  GTPv1MessageType = 22
	GTPv1_INITIATE_PDP_CONTEXT_ACTIVATION_RESPONSE GTPv1MessageType = 23

	// Error indication
	GTPv1_ERROR_INDICATION GTPv1MessageType = 26

	// PDU notification messages
	GTPv1_PDU_NOTIFICATION_REQUEST         GTPv1MessageType = 27
	GTPv1_PDU_NOTIFICATION_RESPONSE        GTPv1MessageType = 28
	GTPv1_PDU_NOTIFICATION_REJECT_REQUEST  GTPv1MessageType = 29
	GTPv1_PDU_NOTIFICATION_REJECT_RESPONSE GTPv1MessageType = 30

	// Supported extensions header notification
	GTPv1_SUPPORTED_EXTENSIONS_HEADER_NOTIFICATION GTPv1MessageType = 31

	// Send routing for GPRS messages
	GTPv1_SEND_ROUTING_FOR_GPRS_REQUEST  GTPv1MessageType = 32
	GTPv1_SEND_ROUTING_FOR_GPRS_RESPONSE GTPv1MessageType = 33

	// Failure report messages
	GTPv1_FAILURE_REPORT_REQUEST  GTPv1MessageType = 34
	GTPv1_FAILURE_REPORT_RESPONSE GTPv1MessageType = 35

	// Note MS present messages
	GTPv1_NOTE_MS_PRESENT_REQUEST  GTPv1MessageType = 36
	GTPv1_NOTE_MS_PRESENT_RESPONSE GTPv1MessageType = 37

	// Identification messages
	GTPv1_IDENTIFICATION_REQUEST  GTPv1MessageType = 38
	GTPv1_IDENTIFICATION_RESPONSE GTPv1MessageType = 39

	// SGSN context messages
	GTPv1_SGSN_CONTEXT_REQUEST     GTPv1MessageType = 50
	GTPv1_SGSN_CONTEXT_RESPONSE    GTPv1MessageType = 51
	GTPv1_SGSN_CONTEXT_ACKNOWLEDGE GTPv1MessageType = 52

	// Forward relocation messages
	GTPv1_FORWARD_RELOCATION_REQUEST              GTPv1MessageType = 53
	GTPv1_FORWARD_RELOCATION_RESPONSE             GTPv1MessageType = 54
	GTPv1_FORWARD_RELOCATION_COMPLETE             GTPv1MessageType = 55
	GTPv1_RELOCATION_CANCEL_REQUEST               GTPv1MessageType = 56
	GTPv1_RELOCATION_CANCEL_RESPONSE              GTPv1MessageType = 57
	GTPv1_FORWARD_SRNS_CONTEXT                    GTPv1MessageType = 58
	GTPv1_FORWARD_RELOCATION_COMPLETE_ACKNOWLEDGE GTPv1MessageType = 59
	GTPv1_FORWARD_SRNS_CONTEXT_ACKNOWLEDGE        GTPv1MessageType = 60

	// UE registration messages
	GTPv1_UE_REGISTRATION_REQUEST  GTPv1MessageType = 61
	GTPv1_UE_REGISTRATION_RESPONSE GTPv1MessageType = 62

	// RAN information relay
	GTPv1_RAN_INFORMATION_RELAY GTPv1MessageType = 70

	// MBMS notification messages
	GTPv1_MBMS_NOTIFICATION_REQUEST         GTPv1MessageType = 96
	GTPv1_MBMS_NOTIFICATION_RESPONSE        GTPv1MessageType = 97
	GTPv1_MBMS_NOTIFICATION_REJECT_REQUEST  GTPv1MessageType = 98
	GTPv1_MBMS_NOTIFICATION_REJECT_RESPONSE GTPv1MessageType = 99

	// Create MBMS notification messages
	GTPv1_CREATE_MBMS_NOTIFICATION_REQUEST  GTPv1MessageType = 100
	GTPv1_CREATE_MBMS_NOTIFICATION_RESPONSE GTPv1MessageType = 101
	GTPv1_UPDATE_MBMS_NOTIFICATION_REQUEST  GTPv1MessageType = 102
	GTPv1_UPDATE_MBMS_NOTIFICATION_RESPONSE GTPv1MessageType = 103
	GTPv1_DELETE_MBMS_NOTIFICATION_REQUEST  GTPv1MessageType = 104
	GTPv1_DELETE_MBMS_NOTIFICATION_RESPONSE GTPv1MessageType = 105

	// MBMS registration messages
	GTPv1_MBMS_REGISTRATION_REQUEST     GTPv1MessageType = 112
	GTPv1_MBMS_REGISTRATION_RESPONSE    GTPv1MessageType = 113
	GTPv1_MBMS_DE_REGISTRATION_REQUEST  GTPv1MessageType = 114
	GTPv1_MBMS_DE_REGISTRATION_RESPONSE GTPv1MessageType = 115

	// MBMS session messages
	GTPv1_MBMS_SESSION_START_REQUEST   GTPv1MessageType = 116
	GTPv1_MBMS_SESSION_START_RESPONSE  GTPv1MessageType = 117
	GTPv1_MBMS_SESSION_STOP_REQUEST    GTPv1MessageType = 118
	GTPv1_MBMS_SESSION_STOP_RESPONSE   GTPv1MessageType = 119
	GTPv1_MBMS_SESSION_UPDATE_REQUEST  GTPv1MessageType = 120
	GTPv1_MBMS_SESSION_UPDATE_RESPONSE GTPv1MessageType = 121

	// MS info change messages
	GTPv1_MS_INFO_CHANGE_REQUEST  GTPv1MessageType = 128
	GTPv1_MS_INFO_CHANGE_RESPONSE GTPv1MessageType = 129

	// Data record transfer messages
	GTPv1_DATA_RECORD_TRANSFER_REQUEST  GTPv1MessageType = 240
	GTPv1_DATA_RECORD_TRANSFER_RESPONSE GTPv1MessageType = 241

	// End marker
	GTPv1_END_MARKER GTPv1MessageType = 254

	// G-PDU
	GTPv1_G_PDU GTPv1MessageType = 255
)

// 通过消息类型数字获取消息对象
func GetGTPv1Message(messageType uint8) *GTPv1Message {
	switch GTPv1MessageType(messageType) {
	case GTPv1_ECHO_REQUEST:
		return &GTPv1Message{Type: GTPv1_ECHO_REQUEST, Name: "Echo Request"}
	case GTPv1_ECHO_RESPONSE:
		return &GTPv1Message{Type: GTPv1_ECHO_RESPONSE, Name: "Echo Response"}
	case GTPv1_VERSION_NOT_SUPPORTED:
		return &GTPv1Message{Type: GTPv1_VERSION_NOT_SUPPORTED, Name: "Version Not Supported"}
	case GTPv1_NODE_ALIVE_REQUEST:
		return &GTPv1Message{Type: GTPv1_NODE_ALIVE_REQUEST, Name: "Node Alive Request"}
	case GTPv1_NODE_ALIVE_RESPONSE:
		return &GTPv1Message{Type: GTPv1_NODE_ALIVE_RESPONSE, Name: "Node Alive Response"}
	case GTPv1_REDIRECTION_REQUEST:
		return &GTPv1Message{Type: GTPv1_REDIRECTION_REQUEST, Name: "Redirection Request"}
	case GTPv1_REDIRECTION_RESPONSE:
		return &GTPv1Message{Type: GTPv1_REDIRECTION_RESPONSE, Name: "Redirection Response"}
	case GTPv1_CREATE_PDP_CONTEXT_REQUEST:
		return &GTPv1Message{Type: GTPv1_CREATE_PDP_CONTEXT_REQUEST, Name: "Create PDP Context Request"}
	case GTPv1_CREATE_PDP_CONTEXT_RESPONSE:
		return &GTPv1Message{Type: GTPv1_CREATE_PDP_CONTEXT_RESPONSE, Name: "Create PDP Context Response"}
	case GTPv1_UPDATE_PDP_CONTEXT_REQUEST:
		return &GTPv1Message{Type: GTPv1_UPDATE_PDP_CONTEXT_REQUEST, Name: "Update PDP Context Request"}
	case GTPv1_UPDATE_PDP_CONTEXT_RESPONSE:
		return &GTPv1Message{Type: GTPv1_UPDATE_PDP_CONTEXT_RESPONSE, Name: "Update PDP Context Response"}
	case GTPv1_DELETE_PDP_CONTEXT_REQUEST:
		return &GTPv1Message{Type: GTPv1_DELETE_PDP_CONTEXT_REQUEST, Name: "Delete PDP Context Request"}
	case GTPv1_DELETE_PDP_CONTEXT_RESPONSE:
		return &GTPv1Message{Type: GTPv1_DELETE_PDP_CONTEXT_RESPONSE, Name: "Delete PDP Context Response"}
	case GTPv1_INITIATE_PDP_CONTEXT_ACTIVATION_REQUEST:
		return &GTPv1Message{Type: GTPv1_INITIATE_PDP_CONTEXT_ACTIVATION_REQUEST, Name: "Initiate PDP Context Activation Request"}
	case GTPv1_INITIATE_PDP_CONTEXT_ACTIVATION_RESPONSE:
		return &GTPv1Message{Type: GTPv1_INITIATE_PDP_CONTEXT_ACTIVATION_RESPONSE, Name: "Initiate PDP Context Activation Response"}
	case GTPv1_ERROR_INDICATION:
		return &GTPv1Message{Type: GTPv1_ERROR_INDICATION, Name: "Error Indication"}
	case GTPv1_PDU_NOTIFICATION_REQUEST:
		return &GTPv1Message{Type: GTPv1_PDU_NOTIFICATION_REQUEST, Name: "PDU Notification Request"}
	case GTPv1_PDU_NOTIFICATION_RESPONSE:
		return &GTPv1Message{Type: GTPv1_PDU_NOTIFICATION_RESPONSE, Name: "PDU Notification Response"}
	case GTPv1_PDU_NOTIFICATION_REJECT_REQUEST:
		return &GTPv1Message{Type: GTPv1_PDU_NOTIFICATION_REJECT_REQUEST, Name: "PDU Notification Reject Request"}
	case GTPv1_PDU_NOTIFICATION_REJECT_RESPONSE:
		return &GTPv1Message{Type: GTPv1_PDU_NOTIFICATION_REJECT_RESPONSE, Name: "PDU Notification Reject Response"}
	case GTPv1_SUPPORTED_EXTENSIONS_HEADER_NOTIFICATION:
		return &GTPv1Message{Type: GTPv1_SUPPORTED_EXTENSIONS_HEADER_NOTIFICATION, Name: "Supported Extensions Header Notification"}
	case GTPv1_SEND_ROUTING_FOR_GPRS_REQUEST:
		return &GTPv1Message{Type: GTPv1_SEND_ROUTING_FOR_GPRS_REQUEST, Name: "Send Routing for GPRS Request"}
	case GTPv1_SEND_ROUTING_FOR_GPRS_RESPONSE:
		return &GTPv1Message{Type: GTPv1_SEND_ROUTING_FOR_GPRS_RESPONSE, Name: "Send Routing for GPRS Response"}
	case GTPv1_FAILURE_REPORT_REQUEST:
		return &GTPv1Message{Type: GTPv1_FAILURE_REPORT_REQUEST, Name: "Failure Report Request"}
	case GTPv1_FAILURE_REPORT_RESPONSE:
		return &GTPv1Message{Type: GTPv1_FAILURE_REPORT_RESPONSE, Name: "Failure Report Response"}
	case GTPv1_NOTE_MS_PRESENT_REQUEST:
		return &GTPv1Message{Type: GTPv1_NOTE_MS_PRESENT_REQUEST, Name: "Note MS Present Request"}
	case GTPv1_NOTE_MS_PRESENT_RESPONSE:
		return &GTPv1Message{Type: GTPv1_NOTE_MS_PRESENT_RESPONSE, Name: "Note MS Present Response"}
	case GTPv1_IDENTIFICATION_REQUEST:
		return &GTPv1Message{Type: GTPv1_IDENTIFICATION_REQUEST, Name: "Identification Request"}
	case GTPv1_IDENTIFICATION_RESPONSE:
		return &GTPv1Message{Type: GTPv1_IDENTIFICATION_RESPONSE, Name: "Identification Response"}
	case GTPv1_SGSN_CONTEXT_REQUEST:
		return &GTPv1Message{Type: GTPv1_SGSN_CONTEXT_REQUEST, Name: "SGSN Context Request"}
	case GTPv1_SGSN_CONTEXT_RESPONSE:
		return &GTPv1Message{Type: GTPv1_SGSN_CONTEXT_RESPONSE, Name: "SGSN Context Response"}
	case GTPv1_SGSN_CONTEXT_ACKNOWLEDGE:
		return &GTPv1Message{Type: GTPv1_SGSN_CONTEXT_ACKNOWLEDGE, Name: "SGSN Context Acknowledge"}
	case GTPv1_FORWARD_RELOCATION_REQUEST:
		return &GTPv1Message{Type: GTPv1_FORWARD_RELOCATION_REQUEST, Name: "Forward Relocation Request"}
	case GTPv1_FORWARD_RELOCATION_RESPONSE:
		return &GTPv1Message{Type: GTPv1_FORWARD_RELOCATION_RESPONSE, Name: "Forward Relocation Response"}
	case GTPv1_FORWARD_RELOCATION_COMPLETE:
		return &GTPv1Message{Type: GTPv1_FORWARD_RELOCATION_COMPLETE, Name: "Forward Relocation Complete"}
	case GTPv1_RELOCATION_CANCEL_REQUEST:
		return &GTPv1Message{Type: GTPv1_RELOCATION_CANCEL_REQUEST, Name: "Relocation Cancel Request"}
	case GTPv1_RELOCATION_CANCEL_RESPONSE:
		return &GTPv1Message{Type: GTPv1_RELOCATION_CANCEL_RESPONSE, Name: "Relocation Cancel Response"}
	case GTPv1_FORWARD_SRNS_CONTEXT:
		return &GTPv1Message{Type: GTPv1_FORWARD_SRNS_CONTEXT, Name: "Forward SRNS Context"}
	case GTPv1_FORWARD_RELOCATION_COMPLETE_ACKNOWLEDGE:
		return &GTPv1Message{Type: GTPv1_FORWARD_RELOCATION_COMPLETE_ACKNOWLEDGE, Name: "Forward Relocation Complete Acknowledge"}
	case GTPv1_FORWARD_SRNS_CONTEXT_ACKNOWLEDGE:
		return &GTPv1Message{Type: GTPv1_FORWARD_SRNS_CONTEXT_ACKNOWLEDGE, Name: "Forward SRNS Context Acknowledge"}
	case GTPv1_UE_REGISTRATION_REQUEST:
		return &GTPv1Message{Type: GTPv1_UE_REGISTRATION_REQUEST, Name: "UE Registration Request"}
	case GTPv1_UE_REGISTRATION_RESPONSE:
		return &GTPv1Message{Type: GTPv1_UE_REGISTRATION_RESPONSE, Name: "UE Registration Response"}
	case GTPv1_RAN_INFORMATION_RELAY:
		return &GTPv1Message{Type: GTPv1_RAN_INFORMATION_RELAY, Name: "RAN Information Relay"}
	case GTPv1_MBMS_NOTIFICATION_REQUEST:
		return &GTPv1Message{Type: GTPv1_MBMS_NOTIFICATION_REQUEST, Name: "MBMS Notification Request"}
	case GTPv1_MBMS_NOTIFICATION_RESPONSE:
		return &GTPv1Message{Type: GTPv1_MBMS_NOTIFICATION_RESPONSE, Name: "MBMS Notification Response"}
	case GTPv1_MBMS_NOTIFICATION_REJECT_REQUEST:
		return &GTPv1Message{Type: GTPv1_MBMS_NOTIFICATION_REJECT_REQUEST, Name: "MBMS Notification Reject Request"}
	case GTPv1_MBMS_NOTIFICATION_REJECT_RESPONSE:
		return &GTPv1Message{Type: GTPv1_MBMS_NOTIFICATION_REJECT_RESPONSE, Name: "MBMS Notification Reject Response"}
	case GTPv1_CREATE_MBMS_NOTIFICATION_REQUEST:
		return &GTPv1Message{Type: GTPv1_CREATE_MBMS_NOTIFICATION_REQUEST, Name: "Create MBMS Notification Request"}
	case GTPv1_CREATE_MBMS_NOTIFICATION_RESPONSE:
		return &GTPv1Message{Type: GTPv1_CREATE_MBMS_NOTIFICATION_RESPONSE, Name: "Create MBMS Notification Response"}
	case GTPv1_UPDATE_MBMS_NOTIFICATION_REQUEST:
		return &GTPv1Message{Type: GTPv1_UPDATE_MBMS_NOTIFICATION_REQUEST, Name: "Update MBMS Notification Request"}
	case GTPv1_UPDATE_MBMS_NOTIFICATION_RESPONSE:
		return &GTPv1Message{Type: GTPv1_UPDATE_MBMS_NOTIFICATION_RESPONSE, Name: "Update MBMS Notification Response"}
	case GTPv1_DELETE_MBMS_NOTIFICATION_REQUEST:
		return &GTPv1Message{Type: GTPv1_DELETE_MBMS_NOTIFICATION_REQUEST, Name: "Delete MBMS Notification Request"}
	case GTPv1_DELETE_MBMS_NOTIFICATION_RESPONSE:
		return &GTPv1Message{Type: GTPv1_DELETE_MBMS_NOTIFICATION_RESPONSE, Name: "Delete MBMS Notification Response"}
	case GTPv1_MBMS_REGISTRATION_REQUEST:
		return &GTPv1Message{Type: GTPv1_MBMS_REGISTRATION_REQUEST, Name: "MBMS Registration Request"}
	case GTPv1_MBMS_REGISTRATION_RESPONSE:
		return &GTPv1Message{Type: GTPv1_MBMS_REGISTRATION_RESPONSE, Name: "MBMS Registration Response"}
	case GTPv1_MBMS_DE_REGISTRATION_REQUEST:
		return &GTPv1Message{Type: GTPv1_MBMS_DE_REGISTRATION_REQUEST, Name: "MBMS De-Registration Request"}
	case GTPv1_MBMS_DE_REGISTRATION_RESPONSE:
		return &GTPv1Message{Type: GTPv1_MBMS_DE_REGISTRATION_RESPONSE, Name: "MBMS De-Registration Response"}
	case GTPv1_MBMS_SESSION_START_REQUEST:
		return &GTPv1Message{Type: GTPv1_MBMS_SESSION_START_REQUEST, Name: "MBMS Session Start Request"}
	case GTPv1_MBMS_SESSION_START_RESPONSE:
		return &GTPv1Message{Type: GTPv1_MBMS_SESSION_START_RESPONSE, Name: "MBMS Session Start Response"}
	case GTPv1_MBMS_SESSION_STOP_REQUEST:
		return &GTPv1Message{Type: GTPv1_MBMS_SESSION_STOP_REQUEST, Name: "MBMS Session Stop Request"}
	case GTPv1_MBMS_SESSION_STOP_RESPONSE:
		return &GTPv1Message{Type: GTPv1_MBMS_SESSION_STOP_RESPONSE, Name: "MBMS Session Stop Response"}
	case GTPv1_MBMS_SESSION_UPDATE_REQUEST:
		return &GTPv1Message{Type: GTPv1_MBMS_SESSION_UPDATE_REQUEST, Name: "MBMS Session Update Request"}
	case GTPv1_MBMS_SESSION_UPDATE_RESPONSE:
		return &GTPv1Message{Type: GTPv1_MBMS_SESSION_UPDATE_RESPONSE, Name: "MBMS Session Update Response"}
	case GTPv1_MS_INFO_CHANGE_REQUEST:
		return &GTPv1Message{Type: GTPv1_MS_INFO_CHANGE_REQUEST, Name: "MS Info Change Request"}
	case GTPv1_MS_INFO_CHANGE_RESPONSE:
		return &GTPv1Message{Type: GTPv1_MS_INFO_CHANGE_RESPONSE, Name: "MS Info Change Response"}
	case GTPv1_DATA_RECORD_TRANSFER_REQUEST:
		return &GTPv1Message{Type: GTPv1_DATA_RECORD_TRANSFER_REQUEST, Name: "Data Record Transfer Request"}
	case GTPv1_DATA_RECORD_TRANSFER_RESPONSE:
		return &GTPv1Message{Type: GTPv1_DATA_RECORD_TRANSFER_RESPONSE, Name: "Data Record Transfer Response"}
	case GTPv1_END_MARKER:
		return &GTPv1Message{Type: GTPv1_END_MARKER, Name: "End Marker"}
	case GTPv1_G_PDU:
		return &GTPv1Message{Type: GTPv1_G_PDU, Name: "G-PDU"}
	default:
		return nil
	}
}

//TEID爆破

// TEIDBurstDetector 用于检测TEID爆破行为
type TEIDBurstDetector struct {
	Window      time.Duration     // 滑动窗口大小
	Threshold   int               // 阈值，超过此值则视为爆破
	Events      map[uint32]*Event // 记录每个TEID的事件
	CleanupChan chan bool         // 用于触发清理操作的通道
	StopChan    chan bool         // 用于停止检测器的通道
}

// Event 表示一个TEID事件
type Event struct {
	TEID        uint32
	IP          string
	Port        uint16
	MessageType uint8
	Timestamp   time.Time
	Count       int
}

// NewTEIDBurstDetector 创建一个新的TEIDBurstDetector实例
func NewTEIDBurstDetector(window time.Duration, threshold int) *TEIDBurstDetector {
	return &TEIDBurstDetector{
		Window:      window,
		Threshold:   threshold,
		Events:      make(map[uint32]*Event),
		CleanupChan: make(chan bool),
		StopChan:    make(chan bool),
	}
}

// Start 启动检测器
func (d *TEIDBurstDetector) Start() {
	// 启动一个后台协程来执行清理操作
	go func() {
		for {
			select {
			case <-d.CleanupChan:
				d.cleanup()
			case <-d.StopChan:
				return
			}
		}
	}()
}

// Stop 停止检测器
func (d *TEIDBurstDetector) Stop() {
	d.StopChan <- true
}

// AddEvent 将一个新事件添加到检测器中
func (d *TEIDBurstDetector) AddEvent(event Event) {
	// 检查是否超过阈值
	if existingEvent, ok := d.Events[event.TEID]; ok && existingEvent.IP == event.IP {
		// 如果是同一IP，增加计数
		existingEvent.Timestamp = event.Timestamp
	} else {
		// 如果是新的IP，重置计数
		d.Events[event.TEID] = &event
	}

	// 触发清理操作
	d.CleanupChan <- true
}

// CheckBurst checks if a TEID burst occurred
func (d *TEIDBurstDetector) CheckBurst() (bool, uint32) {
	var expiredTEIDs []uint32
	now := time.Now()

	for teid, event := range d.Events {
		if now.Sub(event.Timestamp) > d.Window {
			// If the event is expired, add it to the list of expired TEIDs
			expiredTEIDs = append(expiredTEIDs, teid)
		} else {
			// If the TEID usage frequency exceeds the threshold, it is considered a burst
			if d.Events[teid] != nil && d.Events[teid].IP == event.IP {
				return true, teid
			}
		}
	}

	// Remove expired events outside the loop
	for _, teid := range expiredTEIDs {
		delete(d.Events, teid)
	}

	return false, 0
}

// cleanup 清理过期的事件
func (d *TEIDBurstDetector) cleanup() {
	now := time.Now()
	for teid, event := range d.Events {
		if now.Sub(event.Timestamp) > d.Window {
			delete(d.Events, teid)
		}
	}
}

//ddos攻击检测

// MessageKey uniquely identifies a GTP-U message event
type MessageKey struct {
	MessageType uint8
	DestIP      string
	DestPort    uint16
}

// GTPUDDoSdetector detects GTP-U DDoS attacks
type GTPUDDoSdetector struct {
	Window      time.Duration
	Threshold   int
	Events      map[MessageKey]*Event
	CleanupChan chan bool
	StopChan    chan bool
	sync.Mutex
}

// NewGTPUDDoSdetector creates a new GTPUDDoSdetector instance
func NewGTPUDDoSdetector(window time.Duration, threshold int) *GTPUDDoSdetector {
	return &GTPUDDoSdetector{
		Window:      window,
		Threshold:   threshold,
		Events:      make(map[MessageKey]*Event),
		CleanupChan: make(chan bool),
		StopChan:    make(chan bool),
	}
}

// Start begins the detector's operation
func (d *GTPUDDoSdetector) Start() {
	go func() {
		for {
			select {
			case <-d.CleanupChan:
				d.cleanup()
			case <-d.StopChan:
				return
			}
		}
	}()
}

// Stop halts the detector's operation
func (d *GTPUDDoSdetector) Stop() {
	close(d.StopChan)
}

// AddEvent adds a new event to the detector
func (d *GTPUDDoSdetector) AddEvent(messageType uint8, destIP string) {
	key := MessageKey{MessageType: messageType, DestIP: destIP}
	now := time.Now()

	d.Lock()
	defer d.Unlock()

	if event, exists := d.Events[key]; exists {
		// If the event exists, update the count and timestamp
		event.Count++
		event.Timestamp = now
	} else {
		// If the event does not exist, create a new one
		d.Events[key] = &Event{
			MessageType: messageType,
			IP:          destIP,
			Count:       1,
			Timestamp:   now,
		}
	}

	// Trigger cleanup operation
	d.CleanupChan <- true
}

// CheckDDoS checks for a GTP-U DDoS attack
func (d *GTPUDDoSdetector) CheckDDoS() (bool, MessageKey) {
	d.Lock()
	defer d.Unlock()

	for key, event := range d.Events {
		if event.Count > d.Threshold {
			return true, key
		}
	}
	return false, MessageKey{}
}

// cleanup removes expired events from the map
func (d *GTPUDDoSdetector) cleanup() {
	d.Lock()
	defer d.Unlock()

	now := time.Now()
	for key, event := range d.Events {
		if now.Sub(event.Timestamp) > d.Window {
			// If the event is expired, remove it from the map
			delete(d.Events, key)
		}
	}
}
