package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQL数据库连接
var mysqlDB *sql.DB
var mysqlOnce sync.Once

// GetMySQLDB returns a singleton instance of the MySQL database connection
func GetMySQLDB() (*sql.DB, error) {
	mysqlOnce.Do(func() {
		var err error
		mysqlDB, err = connectToMySQL()
		if err != nil {
			fmt.Println("Error connecting to MySQL:", err)
			return
		}
		// Configure the connection pool
		mysqlDB.SetMaxOpenConns(50)                 // Set the maximum number of open connections to the database
		mysqlDB.SetMaxIdleConns(25)                 // Set the maximum number of connections in the idle connection pool
		mysqlDB.SetConnMaxLifetime(5 * time.Minute) // Set the maximum amount of time a connection may be reused
	})
	if mysqlDB == nil {
		log.Printf("Failed to get MySQL connection")
		return nil, errors.New("failed to connect to MySQL")
	}
	return mysqlDB, nil
}

// connectToMySQL tries to connect to MySQL with retries
func connectToMySQL() (*sql.DB, error) {
	attempts := 5
	delay := 2 * time.Second
	for i := 0; i < attempts; i++ {
		db, err := sql.Open("mysql", "yx:Gsycl3541@tcp(10.1.200.13:3306)/AIS")
		if err != nil {
			fmt.Printf("Attempt %d/%d failed to connect to MySQL: %v\n", i+1, attempts, err)
			time.Sleep(delay)
			log.Printf("Attempt %d/%d failed to connect to MySQL: %v\n", i+1, attempts, err)
			continue
		}
		return db, nil
	}
	return nil, fmt.Errorf("all attempts to connect to MySQL failed")
}

// PFCDetectionResult represents the data structure for the detection result
type PFCDetectionResult struct {
	GlobalID                string
	StampTime               time.Time
	SrcIP                   string
	DstIP                   string
	SrcPort                 int
	DstPort                 int
	IMSI                    string
	SubscriberIdentity      string
	EquipmentIdentity       string
	Protocol                string
	Length                  int
	PFCPVersion             int
	SEID                    string
	MessageType             string
	RuleID                  string
	PFCPPDRResult           int
	PFCPFDRResult           int
	PFCPDeleteSessionResult int
	PFCPN4SMFResult         int
	PFCPFormResult          int
	DisposeStatus           string
	AddressNetwork          string
	AISScene                string
	LocationNE              string
	AlterStage              string
	DangerLevel             string
	PACPInfo                string
}

// InsertPFCDetectionResult inserts a new detection result into the database
func InsertPFCDetectionResult(db *sql.DB, result PFCDetectionResult) error {
	stmt, err := db.Prepare(`
		INSERT INTO AIS_PFCP_DETECTION_RESULT (
			globel_id,
			stamptime,
			src_ip,
			dst_ip,
			src_port,
			dst_port,
			imsi,
			subscriber_identity,
			equipment_identity,
			protocol,
			length,
			pfcp_version,
			seid,
			message_type,
			rule_id,
			pfcp_pdr_result,
			pfcp_fdr_result,
			pfcp_delete_session_result,
			pfcp_n4_smf_result,
			pfcp_form_result,
			dispose_status,
			address_network,
			ais_scene,
			location_ne,
			alter_stage,
			danger_level,
			pacp_info
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Printf("Failed to prepare insert statement: %v", err)
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// 使用预处理语句来避免SQL注入和提高性能
	_, err = stmt.Exec(
		result.GlobalID,
		result.StampTime,
		result.SrcIP,
		result.DstIP,
		result.SrcPort,
		result.DstPort,
		result.IMSI,
		result.SubscriberIdentity,
		result.EquipmentIdentity,
		result.Protocol,
		result.Length,
		result.PFCPVersion,
		result.SEID,
		result.MessageType,
		result.RuleID,
		result.PFCPPDRResult,
		result.PFCPFDRResult,
		result.PFCPDeleteSessionResult,
		result.PFCPN4SMFResult,
		result.PFCPFormResult,
		result.DisposeStatus,
		result.AddressNetwork,
		result.AISScene,
		result.LocationNE,
		result.AlterStage,
		result.DangerLevel,
		result.PACPInfo,
	)
	if err != nil {
		log.Printf("Failed to execute insert statement: %v", err)
		return fmt.Errorf("failed to insert detection result: %w", err)
	}

	return nil
}

// GTPUDetectionResult represents the data structure for the GTP-U detection result
type GTPUDetectionResult struct {
	GlobalID           string
	StampTime          time.Time
	SrcIP              string
	DstIP              string
	SrcPort            int
	DstPort            int
	IMSI               string
	SubscriberIdentity string
	EquipmentIdentity  string
	Protocol           string
	Length             int
	GTPVersion         int
	TEID               string
	MessageType        string
	GTPTEIDResult      int
	GTPPayloadResult   int
	GTPGTPinGTPResult  int
	GTPFormResult      int
	RuleID             string
	DisposeStatus      string
	AddressNetwork     string
	AISScene           string
	LocationNE         string
	AlterStage         string
	DangerLevel        string
	PACPInfo           string
}

// InsertGTPUDetectionResult inserts a new detection result into the database
func InsertGTPUDetectionResult(db *sql.DB, result GTPUDetectionResult) error {
	stmt, err := db.Prepare(`
		INSERT INTO AIS_GTP_U_DETECTION_RESULT (
			globel_id,
			stamptime,
			src_ip,
			dst_ip,
			src_port,
			dst_port,
			imsi,
			subscriber_identity,
			equipment_identity,
			protocol,
			length,
			gtp_version,
			teid,
			message_type,
			gtp_teid_result,
			gtp_payload_result,
			gtp_gtp_in_gtp_result,
			gtp_form_result,
			rule_id,
			dispose_status,
			address_network,
			ais_scene,
			location_ne,
			alter_stage,
			danger_level,
			pacp_info
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Printf("Failed to prepare insert statement: %v", err)
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// 使用预处理语句来避免SQL注入和提高性能
	_, err = stmt.Exec(
		result.GlobalID,
		result.StampTime,
		result.SrcIP,
		result.DstIP,
		result.SrcPort,
		result.DstPort,
		result.IMSI,
		result.SubscriberIdentity,
		result.EquipmentIdentity,
		result.Protocol,
		result.Length,
		result.GTPVersion,
		result.TEID,
		result.MessageType,
		result.GTPTEIDResult,
		result.GTPPayloadResult,
		result.GTPGTPinGTPResult,
		result.GTPFormResult,
		result.RuleID,
		result.DisposeStatus,
		result.AddressNetwork,
		result.AISScene,
		result.LocationNE,
		result.AlterStage,
		result.DangerLevel,
		result.PACPInfo,
	)
	if err != nil {
		log.Printf("Failed to execute insert statement: %v", err)
		return fmt.Errorf("failed to insert detection result: %w", err)
	}

	return nil
}

type GTPCDetectionResult struct {
	GlobalID           string
	StampTime          time.Time
	SrcIP              string
	DstIP              string
	SrcPort            int
	DstPort            int
	IMSI               string
	SubscriberIdentity string
	EquipmentIdentity  string
	Protocol           string
	Length             int
	GTPVersion         int
	TEID               string
	MessageType        string
	GTPFormResult      int
	RuleID             string
	DisposeStatus      string
	AddressNetwork     string
	AISScene           string
	LocationNE         string
	AlterStage         string
	DangerLevel        string
	PACPInfo           string
}

// InsertGTPCDetectionResult inserts a new detection result into the database
func InsertGTPCDetectionResult(db *sql.DB, result GTPCDetectionResult) error {
	stmt, err := db.Prepare(`
		INSERT INTO AIS_GTP_C_DETECTION_RESULT (
			globel_id,
			stamptime,
			src_ip,
			dst_ip,
			src_port,
			dst_port,
			imsi,
			subscriber_identity,
			equipment_identity,
			protocol,
			length,
			gtp_version,
			teid,
			message_type,
			gtp_form_result,
			rule_id,
			dispose_status,
			address_network,
			ais_scene,
			location_ne,
			alter_stage,
			danger_level,
			pacp_info
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Printf("Failed to prepare insert statement: %v", err)
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// 使用预处理语句来避免SQL注入和提高性能
	_, err = stmt.Exec(
		result.GlobalID,
		result.StampTime,
		result.SrcIP,
		result.DstIP,
		result.SrcPort,
		result.DstPort,
		result.IMSI,
		result.SubscriberIdentity,
		result.EquipmentIdentity,
		result.Protocol,
		result.Length,
		result.GTPVersion,
		result.TEID,
		result.MessageType,
		result.GTPFormResult,
		result.RuleID,
		result.DisposeStatus,
		result.AddressNetwork,
		result.AISScene,
		result.LocationNE,
		result.AlterStage,
		result.DangerLevel,
		result.PACPInfo,
	)
	if err != nil {
		log.Printf("Failed to execute insert statement: %v", err)
		return fmt.Errorf("failed to insert detection result: %w", err)
	}

	return nil
}

type SCTPDetectionResult struct {
	GlobalID                     string
	StampTime                    time.Time
	SrcIP                        string
	DstIP                        string
	SrcPort                      int
	DstPort                      int
	IMSI                         string
	SubscriberIdentity           string
	EquipmentIdentity            string
	Protocol                     string
	Length                       int
	ChunksType                   string
	Sctp_verificationTag         string
	SCTPFourHandshakesDDOSResult int
	SCTPSuperResult              int
	SCTPMultichunkResult         int
	SCTPInitFloodResult          int
	RuleID                       string
	DisposeStatus                string
	AddressNetwork               string
	AISScene                     string
	LocationNE                   string
	AlterStage                   string
	DangerLevel                  string
	PACPInfo                     string
}

// InsertSCTPDetectionResult inserts a new detection result into the database
func InsertSCTPDetectionResult(db *sql.DB, result SCTPDetectionResult) error {
	stmt, err := db.Prepare(`
		INSERT INTO AIS_SCTP_DETECTION_RESULT (
			globel_id,
			stamptime,
			src_ip,
			dst_ip,
			src_port,
			dst_port,
			imsi,
			subscriber_identity,
			equipment_identity,
			protocol,
			length,
			chunks_type,
			sctp_verificationTag,
			sctp_four_handshakes_ddos_result,
			sctp_super_result,
			sctp_multichunk_result,
			sctp_init_flood_result,
			rule_id,
			dispose_status,
			address_network,
			ais_scene,
			location_ne,
			alter_stage,
			danger_level,
			pacp_info
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Printf("Failed to prepare insert statement: %v", err)
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// 使用预处理语句来避免SQL注入和提高性能
	_, err = stmt.Exec(
		result.GlobalID,
		result.StampTime,
		result.SrcIP,
		result.DstIP,
		result.SrcPort,
		result.DstPort,
		result.IMSI,
		result.SubscriberIdentity,
		result.EquipmentIdentity,
		result.Protocol,
		result.Length,
		result.ChunksType,
		result.Sctp_verificationTag,
		result.SCTPFourHandshakesDDOSResult,
		result.SCTPSuperResult,
		result.SCTPMultichunkResult,
		result.SCTPInitFloodResult,
		result.RuleID,
		result.DisposeStatus,
		result.AddressNetwork,
		result.AISScene,
		result.LocationNE,
		result.AlterStage,
		result.DangerLevel,
		result.PACPInfo,
	)
	if err != nil {
		log.Printf("Failed to execute insert statement: %v", err)
		return fmt.Errorf("failed to insert detection result: %w", err)
	}

	return nil
}

type NGAPDetectionResult struct {
	GlobalID                 string
	StampTime                time.Time
	SrcIP                    string
	DstIP                    string
	SrcPort                  int
	DstPort                  int
	IMSI                     string
	SubscriberIdentity       string
	EquipmentIdentity        string
	Protocol                 string
	Length                   int
	MessageType              string
	ProcedureCode            string
	Criticlity               string
	NGAPReleaseRequestResult int
	Ngap_from_result         int
	RuleID                   string
	DisposeStatus            string
	AddressNetwork           string
	AISScene                 string
	LocationNE               string
	AlterStage               string
	DangerLevel              string
	PACPInfo                 string
}

// InsertNGAPDetectionResult inserts a new detection result into the database
func InsertNGAPDetectionResult(db *sql.DB, result NGAPDetectionResult) error {

	stmt, err := db.Prepare(`
		INSERT INTO AIS_NGAP_DETECTION_RESULT (
			globel_id,
			stamptime,
			src_ip,
			dst_ip,
			src_port,
			dst_port,
			imsi,
			subscriber_identity,
			equipment_identity,
			protocol,
			length,
			message_type,
			procedure_code,
			criticlity,
			ngap_releaserequest_result,
			ngap_from_result,
			rule_id,
			dispose_status,
			address_network,
			ais_scene,
			location_ne,
			alter_stage,
			danger_level,
			pacp_info
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,?, ?)
	`)
	if err != nil {
		log.Printf("Failed to prepare insert statement: %v", err)
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// 使用预处理语句来避免SQL注入和提高性能
	_, err = stmt.Exec(
		result.GlobalID,
		result.StampTime,
		result.SrcIP,
		result.DstIP,
		result.SrcPort,
		result.DstPort,
		result.IMSI,
		result.SubscriberIdentity,
		result.EquipmentIdentity,
		result.Protocol,
		result.Length,
		result.MessageType,
		result.ProcedureCode,
		result.Criticlity,
		result.NGAPReleaseRequestResult,
		result.Ngap_from_result,
		result.RuleID,
		result.DisposeStatus,
		result.AddressNetwork,
		result.AISScene,
		result.LocationNE,
		result.AlterStage,
		result.DangerLevel,
		result.PACPInfo,
	)
	if err != nil {
		log.Printf("Failed to execute insert statement: %v", err)
		return fmt.Errorf("failed to insert detection result: %w", err)
	}

	return nil
}

type SignalStormDetectionResult struct {
	GlobalID                                string
	StampTime                               time.Time
	SrcIP                                   string
	DstIP                                   string
	SrcPort                                 int
	DstPort                                 int
	IMSI                                    string
	SubscriberIdentity                      string
	EquipmentIdentity                       string
	Protocol                                string
	Length                                  int
	TEID                                    string
	SEID                                    string
	Byte                                    int
	Threshold                               int
	SignalStormMulUELoginResult             int
	SignalStormMulAccessResult              int
	SignalStormNFFaultyResult               int
	SignalStormGTPUSynDDOSResult            int
	SignalStormPFCPN4Result                 int
	SignalStormSCTPInitFloodResult          int
	SignalStormSCTPFourHandshakesDDOSResult int
	RuleID                                  string
	DisposeStatus                           string
	AddressNetwork                          string
	AISScene                                string
	LocationNE                              string
	AlterStage                              string
	DangerLevel                             string
	PACPInfo                                string
}

// InsertSignalStormDetectionResult inserts a new detection result into the database
func InsertSignalStormDetectionResult(db *sql.DB, result SignalStormDetectionResult) error {

	stmt, err := db.Prepare(`
		INSERT INTO AIS_SIGNAL_STORM_DETECTION_RESULT (
			globel_id,
			stamptime,
			src_ip,
			dst_ip,
			src_port,
			dst_port,
			imsi,
			subscriber_identity,
			equipment_identity,
			protocol,
			length,
			teid,
			seid,
			byte,
			threshold,
			signal_storm_mul_ue_login_result,
			signal_storm_mul_access_result,
			signal_storm_nf_faulty_result,
			signal_storm_gtp_u_syn_ddos_result,
			signal_storm_pfcp_n4_result,
			signal_storm_sctp_init_flood_result,
			signal_storm_sctp_four_handshakes_ddos_result,
			rule_id,
			dispose_status,
			address_network,
			ais_scene,
			location_ne,
			alter_stage,
			danger_level,
			pacp_info
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,?,?)
	`)
	if err != nil {
		log.Printf("Failed to prepare insert statement: %v", err)
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// 使用预处理语句来避免SQL注入和提高性能
	_, err = stmt.Exec(
		result.GlobalID,
		result.StampTime,
		result.SrcIP,
		result.DstIP,
		result.SrcPort,
		result.DstPort,
		result.IMSI,
		result.SubscriberIdentity,
		result.EquipmentIdentity,
		result.Protocol,
		result.Length,
		result.TEID,
		result.SEID,
		result.Byte,
		result.Threshold,
		result.SignalStormMulUELoginResult,
		result.SignalStormMulAccessResult,
		result.SignalStormNFFaultyResult,
		result.SignalStormGTPUSynDDOSResult,
		result.SignalStormPFCPN4Result,
		result.SignalStormSCTPInitFloodResult,
		result.SignalStormSCTPFourHandshakesDDOSResult,
		result.RuleID,
		result.DisposeStatus,
		result.AddressNetwork,
		result.AISScene,
		result.LocationNE,
		result.AlterStage,
		result.DangerLevel,
		result.PACPInfo,
	)
	if err != nil {
		log.Printf("Failed to execute insert statement: %v", err)
		return fmt.Errorf("failed to insert detection result: %w", err)
	}

	return nil
}

type NFInformation struct {
	NfName              string
	NfType              string
	NfIP                string
	NfMac               string
	NfPorts             string
	NfProtocols         string
	NfConfigurePolicies string
}

func GetNFInformationByIP(db *sql.DB, ip string) (NFInformation, error) {
	var nf NFInformation

	query := "SELECT nf_name, nf_type, nf_ip, nf_mac, nf_ports, nf_protocols, nf_configure_policies FROM nf_Information WHERE nf_ip = ?"
	err := db.QueryRow(query, ip).Scan(
		&nf.NfName,
		&nf.NfType,
		&nf.NfIP,
		&nf.NfMac,
		&nf.NfPorts,
		&nf.NfProtocols,
		&nf.NfConfigurePolicies,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No record found for IP %s", ip)
			return nf, fmt.Errorf("no record found for IP %s", ip)
		}
		return nf, fmt.Errorf("error querying database: %w", err)
	}

	return nf, nil
}

type TerminalInformation struct {
	Imsi   string
	Msisdn string
	Imei   string
	Guti   string
}

func GetTerminalInformationByIP(db *sql.DB, ip string) (TerminalInformation, error) {
	var terminalInfo TerminalInformation

	query := "SELECT imsi, msisdn, imei, guti FROM 5G_Terminal_Information WHERE ip_address = ?"
	err := db.QueryRow(query, ip).Scan(
		&terminalInfo.Imsi,
		&terminalInfo.Msisdn,
		&terminalInfo.Imei,
		&terminalInfo.Guti,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No record found for IP %s", ip)
			return terminalInfo, fmt.Errorf("no record found for IP %s", ip)
		}
		return terminalInfo, fmt.Errorf("error querying database: %w", err)
	}

	return terminalInfo, nil
}
