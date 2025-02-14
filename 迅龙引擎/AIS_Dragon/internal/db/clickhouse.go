package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/ClickHouse/clickhouse-go"
)

// 连接clickhosue数据库
var db *sql.DB
var once sync.Once

// GetDB returns a singleton instance of the database connection
// GetDB returns a singleton instance of the database connection
func GetDB() (*sql.DB, error) {
	once.Do(func() {
		var err error
		db, err = connectToClickHouse()
		if err != nil {
			fmt.Println("Error connecting to ClickHouse:", err)
			return
		}
		// Configure the connection pool
		db.SetMaxOpenConns(50)                 // Set the maximum number of open connections to the database
		db.SetMaxIdleConns(25)                 // Set the maximum number of connections in the idle connection pool
		db.SetConnMaxLifetime(5 * time.Minute) // Set the maximum amount of time a connection may be reused
	})
	if db == nil {
		log.Printf("db is nil")
		return nil, errors.New("failed to connect to ClickHouse")
	}
	return db, nil
}

// connectToClickHouse tries to connect to ClickHouse with retries
func connectToClickHouse() (*sql.DB, error) {
	attempts := 5
	delay := 2 * time.Second
	for i := 0; i < attempts; i++ {
		db, err := sql.Open("clickhouse", "tcp://10.1.200.14:9000?debug=false")
		if err != nil {
			fmt.Printf("Attempt %d/%d failed to connect to ClickHouse: %v\n", i+1, attempts, err)
			time.Sleep(delay)
			log.Printf("Attempt %d/%d failed to connect to ClickHouse: %v\n", i+1, attempts, err)
			continue
		}
		return db, nil
	}
	return nil, fmt.Errorf("all attempts to connect to ClickHouse failed")
}

// LogDetectionResultModel 代表日志数据模型
type LogDetectionResultModel struct {
	GlobelID                         string `json:"globel_id"`
	SrcIP                            string `json:"src_ip"`
	DstIP                            string `json:"dst_ip"`
	SrcMac                           string `json:"src_mac"`
	DstMac                           string `json:"dst_mac"`
	SrcPort                          uint64 `json:"src_port"`
	DstPort                          uint64 `json:"dst_port"`
	SrcNf                            string `json:"src_nf"`
	DstNf                            string `json:"dst_nf"`
	UserIP                           string `json:"user_ip"`
	UserImsI                         string `json:"user_imsi"`
	UserSubscriberIdentity           string `json:"user_subscriber_identity"`
	UserEquipmentIdentity            string `json:"user_equipment_identity"`
	NetworkSliceSsd                  string `json:"network_slice_ssd"`
	NetworkSliceSt                   string `json:"network_slice_st"`
	IPLayer                          string `json:"ip_layer"`
	ProtocolApplicationLayer         string `json:"protocol_application_layer"`
	ProtocolTransportLayer           string `json:"protocol_transport_layer"`
	Protocol5g                       string `json:"protocol_5g"`
	Length                           uint64 `json:"length"`
	SctpChunksFlags                  string `json:"sctp_chunks_flags"`
	SctpVerificationTag              uint64 `json:"sctp_verificationTag"`
	SctpChunksType                   string `json:"sctp_chunks_type"`
	SctpIeLength                     uint64 `json:"sctp_ie_length"`
	PfcpMessageType                  string `json:"pfcp_message_type"`
	PfcpIeLength                     uint64 `json:"pfcp_ie_length"`
	PfcpHeaderFlagApare              string `json:"pfcp_header_flag_apare"`
	GtpUMessageType                  string `json:"gtp_u_message_type"`
	GtpUIeLength                     uint64 `json:"gtp_u_ie_length"`
	GtpUHeaderFlagApare              string `json:"gtp_u_header_flag_apare"`
	GtpUIeReserved                   string `json:"gtp_u_ie_reserved"`
	GtpCMessageType                  string `json:"gtp_c_message_type"`
	GtpCIeLength                     uint64 `json:"gtp_c_ie_length"`
	GtpCHeaderFlagApare              string `json:"gtp_c_header_flag_apare"`
	GtpCIeReserved                   string `json:"gtp_c_ie_reserved"`
	NgapProcedureCode                uint64 `json:"ngap_procedure_code"`
	NgapCriticlity                   string `json:"ngap_criticlity"`
	NgapMessageType                  string `json:"ngap_message_type"`
	NasMessageType                   string `json:"nas_message_type"`
	NasExtendedProtocolDiscriminator string `json:"nas_extended_protocol_discriminator"`
	NasSecurityHeaderType            string `json:"nas_security_header_type"`
	Http2Method                      string `json:"http2_method"`
	Http2SbiApi                      string `json:"http2_sbi_api"`
	Http2Scheme                      string `json:"http2_scheme"`
	Action                           string `json:"action"`
	RuleId                           string `json:"rule_id"`
	AddressNetwork                   string `json:"address_network"`
	AisScene                         string `json:"ais_scene"`
	LocationNe                       string `json:"location_ne"`
	Alter_stage                      string `json:"alter_stage"`
	DisposeStatus                    string `json:"dispose_status"`
	DangerLevel                      string `json:"danger_level"`
	PacpInfo                         string `json:"pcap_info"`
}

// StoreLog 快速存储日志接口

var err error

func StoreLog(tx *sql.Tx, logEntry LogDetectionResultModel) error {
	// Construct the SQL INSERT statement, including the PfcpInfo field
	insertStatement := `
		INSERT INTO AIS.AIS_LOG_DETECTION_RESULT (
			id,globel_id, stamptime, src_ip, dst_ip, src_mac, dst_mac, src_port, dst_port, src_nf, dst_nf,
			user_ip, user_imsi, user_subscriber_identity, user_equipment_identity, Network_slice_ssd, Network_slice_st,
			ip_layer, protocol_Application_layer, protocol_Transport_layer, protocol_5g, length, sctp_chunks_flags, sctp_verificationTag,
			sctp_chunks_type, sctp_ie_length, pfcp_message_type, pfcp_ie_length, pfcp_header_flag_apare, gtp_u_message_type,
			gtp_u_ie_length, gtp_u_header_flag_apare, gtp_u_ie_reserved, gtp_c_message_type, gtp_c_ie_length,
			gtp_c_header_flag_apare, gtp_c_ie_reserved, ngap_procedure_code, ngap_criticlity, ngap_message_type,
			nas_message_type, nas_extended_protocol_discriminator, nas_security_header_type, http2_method, http2_sbi_api,
			http2_scheme, action, rule_id, address_network, ais_scene, location_ne,alter_stage, dispose_status, danger_level,
			pcap_info
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?,?
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?,?,?
		)
	`
	currentTime := time.Now()
	currentTime1 := time.Now().UTC()
	// globelID, err := uuid.NewUUID()
	// if err != nil {
	// 	return fmt.Errorf("error generating UUID: %w", err)
	// }

	id := currentTime.UnixNano() / int64(time.Millisecond)

	// Execute the SQL statement within the transaction, including the PfcpInfo value
	_, err = tx.Exec(
		insertStatement,
		id, logEntry.GlobelID, currentTime1, logEntry.SrcIP, logEntry.DstIP, logEntry.SrcMac, logEntry.DstMac, logEntry.SrcPort, logEntry.DstPort, logEntry.SrcNf, logEntry.DstNf,
		logEntry.UserIP, logEntry.UserImsI, logEntry.UserSubscriberIdentity, logEntry.UserEquipmentIdentity, logEntry.NetworkSliceSsd, logEntry.NetworkSliceSt,
		logEntry.IPLayer, logEntry.ProtocolApplicationLayer, logEntry.ProtocolTransportLayer, logEntry.Protocol5g, logEntry.Length, logEntry.SctpChunksFlags, logEntry.SctpVerificationTag,
		logEntry.SctpChunksType, logEntry.SctpIeLength, logEntry.PfcpMessageType, logEntry.PfcpIeLength, logEntry.PfcpHeaderFlagApare, logEntry.GtpUMessageType,
		logEntry.GtpUIeLength, logEntry.GtpUHeaderFlagApare, logEntry.GtpUIeReserved, logEntry.GtpCMessageType, logEntry.GtpCIeLength,
		logEntry.GtpCHeaderFlagApare, logEntry.GtpCIeReserved, logEntry.NgapProcedureCode, logEntry.NgapCriticlity, logEntry.NgapMessageType,
		logEntry.NasMessageType, logEntry.NasExtendedProtocolDiscriminator, logEntry.NasSecurityHeaderType, logEntry.Http2Method, logEntry.Http2SbiApi,
		logEntry.Http2Scheme, logEntry.Action, logEntry.RuleId, logEntry.AddressNetwork, logEntry.AisScene, logEntry.LocationNe, logEntry.Alter_stage, logEntry.DisposeStatus, logEntry.DangerLevel,
		logEntry.PacpInfo, // Include PfcpInfo in the Exec call
	)
	if err != nil {
		log.Println("Error executing insert statement:", err)
		return fmt.Errorf("error executing insert statement: %w", err)
	}

	return nil
}
