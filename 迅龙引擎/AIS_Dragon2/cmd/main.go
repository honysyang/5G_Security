package main

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"fmt"
	"reflect"

	"github.com/IBM/sarama"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/honysyang/kafka/internal/db"
	"github.com/honysyang/kafka/internal/detector"

	"github.com/free5gc/pfcp"

	"github.com/free5gc/tlv"

	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"regexp"
	"strconv"

	"github.com/google/uuid"

	"github.com/free5gc/aper"
	"github.com/free5gc/ngap/ngapType"

	"github.com/go-redis/redis/v8"
)

//定义用来传参存储的参数

// 网络五元组
var src_ip string
var dst_ip string
var src_port uint64
var dst_port uint64
var src_mac string
var dst_mac string

// 网元信息
var src_nf string
var dst_nf string

// 终端信息
var user_imsi string
var user_subscriber_identity string
var user_ip string
var user_equipment_identity string

// 切片信息
var network_slice_st string
var network_slice_ssd string

// 协议信息
var protocol_5g = "NO"
var ip_layer = "IPv4"
var protocol_transport_layer string
var protocol_application_layer string
var length uint64

// sctp协议信息
var sctp_chunks_flags string
var sctp_verificationTag uint64
var sctp_chunks_type string
var sctp_ie_length uint64

// pfcp协议信息
var pfcp_message_type string
var pfcp_ie_length uint64
var pfcp_header_flag_apare string
var pfcp_version int
var seid_value string

// gtpv1U协议信息
var gtp_u_message_type string
var gtp_u_ie_length uint64
var gtp_u_header_flag_apare string
var gtp_u_reserved string
var GTPU_version int
var teid_value string

// gtpv2C协议信息
var gtp_c_message_type string
var gtp_c_ie_length uint64
var gtp_c_header_flag_apare string
var gtp_c_ie_reserved string

// ngap协议信息
var ngap_message_type string
var ngap_criticlity string
var NgapprocedureCode uint64

// nas协议信息
var nas_message_type string
var nas_extended_protocol_discriminator string
var nas_security_header_type string

// http2协议信息
var http2Method string
var http2SbiApi string
var http2Scheme string

// 其他信息
var ruleId = "0"

// action的值，只能是1 或0,   0表示pass，1表示alert, 2表示block
var action = "0"

// 告警等级：0表示正常，1表示低危， 2表示中危，3表示高危
var danger_level = "0"

// 处置状态： 0表示不需处理，1表示有告警未处理，2表示有告警已处理
var dispose_status = "0"

var pacp_info string

// 网络位置：   Web网络，5G接入网，5G核心网
var address_network = "Web网络"

var ais_scene string

// 5G网元位置：WEB、UE~gNB、UE~AMF、gNB~AMF、gNB~UPF、UPF~SMF、UPF~DN、AMF~SMF、AMF~AUSF、AMF~NRF、Known
var location_ne = "WEB"

var alter_stage = ""

var pfcpPDRResult int
var pfcpFDRResult int
var pfcpN4SMFResult int
var pfcpFormResult int
var pfcfDeleteSessionResult int

var gtpuTEIDResult int
var gtpuPayloadResult int
var gtpuGTPinGTPResult int
var gtpuFormResult int

var sctpFourHandshakesDDOSResult int
var sctpSuperResult int
var sctpMultichunkResult int
var sctpInitFloodResult int

var ngapReleaseRequestResult int
var ngap_from_result int

var signalStormMulUELoginResult int
var signalStormMulAccessResult int
var signalStormNFFaultyResult int
var signalStormGTPUSynDDOSResult int
var signalStormPFCPN4Result int
var signalStormSCTPInitFloodResult int
var signalStormSCTPFourHandshakesDDOSResult int

var signalStorm uint

var isNgap bool

var db_clickhouse *sql.DB
var db_mysql *sql.DB

var err error

// 定义一些规则ID
const (
	RuleIdTEIDBruteForce     = "10001" //TEID的爆破
	RuleIdContentTooLong     = "10002"
	RuleIdContentMalicious   = "10003"
	RuleIdGTPINGTP           = "10004"
	RuleIdGTPv1UWithoutUDP   = "10005"
	RuleIdGTPv1UMESSAGENIL   = "10006"
	RuleIdGTPv1UMESSAGLENGTH = "10007"
	RuleIdGTPv1UMESSAGTYPE   = "10008"
	RuleIdGTPv1UBurst        = "10009" //GTP-U的DDOS

	PFCPsessionEstablishmentRequestPDR    = "20001"
	PFCPsessionEstablishmentRequestFAR    = "20002"
	PFCPsessionModificationRequestPDR     = "20003"
	PFCPsessionModificationRequestPDR1    = "20004"
	PFCPsessionModificationRequestPDR2    = "20005"
	PFCPsessionModificationRequestFAR     = "20006"
	PFCPsessionModificationRequestFAR1    = "20007"
	PFCPsessionModificationRequestFAR2    = "20008"
	PFCPsessionDeletionRequest1           = "20009"
	PFCPsessionDeletionRequest2           = "20010"
	PFCPassocUpdateRequest                = "20011"
	PFCPassocReleaseRequest               = "20012"
	PFCPassocReleaseResponse              = "20013"
	PFCPsessionSetDeletionRequest         = "20014"
	PFCPsessionSetDeletionResponse        = "20015"
	PFCPsessionModificationRequest        = "20016"
	PFCPsessionModificationResponse       = "20017"
	RuleIdPFCPVersionNotSupportedResponse = "20018"
	RuleIdPFCPInvalidLength               = "20019"
	PFCPinvaliadmessagetype               = "20020"

	RuleIdPFCPReleaseRequest              = "20021"
	RuleIdPFCPDeletionRequest             = "20022"
	RuleIdPFCPSessionEstablishmentRequest = "20023"

	RuleIdSCTPOverTCPorUDP          = "30001"
	RuleIdSCTPShutdownCompleteFlood = "30002" //ddos
	RuleIdSCTPInitFlood             = "30003"
	RuleIdSCTPDataFlood             = "30004"
	RuleIdSCTPInitackFlood          = "30005"
	RuleIdSCTPSackFlood             = "30006"
	RuleIdSCTPHeartbeatFlood        = "30007"
	RuleIdSCTPHeartbeatAckFlood     = "30008"
	RuleIdSCTPShutdownFlood         = "30009"
	RuleIdSCTPErrorFlood            = "30010"
	RuleIdSCTPShutdownAckFlood      = "30011"
	RuleIdSCTPCookieEchoFlood       = "30012"
	RuleIdSCTPCookieAckFlood        = "30013"

	RuleIdUEContextReleaseRequest = "40001"
	RuleIdNGAPUnknown             = "40002"
)

// 设置信号风暴结果
func setSignalStormResults(ruleId string) {
	// 初始化所有结果为 0

	pfcpPDRResult = 0
	pfcpFDRResult = 0
	pfcpN4SMFResult = 0
	pfcpFormResult = 0
	pfcfDeleteSessionResult = 0

	gtpuTEIDResult = 0
	gtpuPayloadResult = 0
	gtpuGTPinGTPResult = 0
	gtpuFormResult = 0

	sctpFourHandshakesDDOSResult = 0
	sctpSuperResult = 0
	sctpMultichunkResult = 0
	sctpInitFloodResult = 0

	ngapReleaseRequestResult = 0
	ngap_from_result = 0

	signalStormMulUELoginResult = 0
	signalStormMulAccessResult = 0
	signalStormNFFaultyResult = 0
	signalStormGTPUSynDDOSResult = 0
	signalStormPFCPN4Result = 0
	signalStormSCTPInitFloodResult = 0
	signalStormSCTPFourHandshakesDDOSResult = 0
	alter_stage = ""
	danger_level = "0"
	action = "0"

	switch ruleId {
	case RuleIdTEIDBruteForce:
		gtpuTEIDResult = 1
		signalStormGTPUSynDDOSResult = 1
		alter_stage = "1"
		danger_level = "3"
		action = "1"
	case RuleIdGTPv1UBurst:
		signalStormGTPUSynDDOSResult = 1
		alter_stage = "0"
		danger_level = "1"
		action = "1"
	case RuleIdContentTooLong:
		gtpuFormResult = 1
		alter_stage = "0"
		danger_level = "1"
		action = "1"
	case RuleIdContentMalicious:
		gtpuPayloadResult = 1
		alter_stage = "2"
		danger_level = "3"
		action = "1"
	case RuleIdGTPINGTP:
		gtpuGTPinGTPResult = 1
		alter_stage = "2"
		danger_level = "3"
		action = "1"
	case RuleIdGTPv1UWithoutUDP:
		gtpuFormResult = 1
		alter_stage = "0"
		danger_level = "2"
		action = "1"
	case RuleIdGTPv1UMESSAGENIL:
		gtpuFormResult = 1
		alter_stage = "0"
		danger_level = "1"
		action = "1"
	case RuleIdGTPv1UMESSAGLENGTH:
		gtpuFormResult = 1
		alter_stage = "0"
		danger_level = "1"
		action = "1"
	case RuleIdGTPv1UMESSAGTYPE:
		gtpuFormResult = 1
		alter_stage = "0"
		danger_level = "1"
		action = "1"
	case PFCPsessionEstablishmentRequestPDR:
		pfcpPDRResult = 1
		alter_stage = "3"
		danger_level = "2"
		action = "1"
	case PFCPsessionEstablishmentRequestFAR:
		pfcpFDRResult = 1
		alter_stage = "3"
		danger_level = "2"
		action = "1"
	case PFCPsessionModificationRequestPDR:
		pfcpPDRResult = 1
		alter_stage = "3"
		danger_level = "1"
		action = "1"
	case PFCPsessionModificationRequestPDR1:
		pfcpPDRResult = 1
		alter_stage = "3"
		danger_level = "1"
		action = "1"
	case PFCPsessionModificationRequestPDR2:
		pfcpPDRResult = 1
		alter_stage = "3"
		danger_level = "1"
		action = "1"
	case PFCPsessionModificationRequestFAR:
		pfcpFDRResult = 1
		alter_stage = "3"
		danger_level = "1"
		action = "1"
	case PFCPsessionModificationRequestFAR1:
		pfcpFDRResult = 1
		alter_stage = "3"
		danger_level = "1"
		action = "1"
	case PFCPsessionModificationRequestFAR2:
		pfcpFDRResult = 1
		alter_stage = "3"
		danger_level = "1"
		action = "1"
	case PFCPsessionDeletionRequest1:
		pfcfDeleteSessionResult = 1
		alter_stage = "3"
		danger_level = "3"
		action = "1"
	case PFCPsessionDeletionRequest2:
		pfcfDeleteSessionResult = 1
		alter_stage = "3"
		danger_level = "3"
		action = "1"
	case PFCPassocUpdateRequest:
		pfcpN4SMFResult = 1
		alter_stage = "3"
		danger_level = "1"
		action = "1"
	case PFCPassocReleaseRequest:
		pfcpN4SMFResult = 1
		alter_stage = "3"
		danger_level = "1"
		action = "1"
	case PFCPassocReleaseResponse:
		pfcpN4SMFResult = 1
		alter_stage = "3"
		danger_level = "3"
		action = "1"
	case PFCPsessionSetDeletionRequest:
		pfcfDeleteSessionResult = 1
		alter_stage = "3"
		danger_level = "3"
		action = "1"
	case PFCPsessionSetDeletionResponse:
		pfcfDeleteSessionResult = 1
		alter_stage = "3"
		danger_level = "3"
		action = "1"
	case PFCPsessionModificationRequest:
		pfcpN4SMFResult = 1
		alter_stage = "3"
		danger_level = "1"
		action = "1"
	case PFCPsessionModificationResponse:
		pfcpN4SMFResult = 1
		alter_stage = "3"
		danger_level = "1"
		action = "1"
	case RuleIdPFCPVersionNotSupportedResponse:
		pfcpFormResult = 1
		alter_stage = "3"
		danger_level = "1"
		action = "1"
	case RuleIdPFCPInvalidLength:
		pfcpFormResult = 1
		alter_stage = "3"
		danger_level = "2"
		action = "1"
	case PFCPinvaliadmessagetype:
		pfcpFormResult = 1
		alter_stage = "3"
		danger_level = "1"
		action = "1"
	case RuleIdPFCPReleaseRequest:
		pfcpFormResult = 1
		alter_stage = "3"
		danger_level = "2"
		action = "1"
	case RuleIdPFCPDeletionRequest:
		pfcfDeleteSessionResult = 1
		alter_stage = "3"
		danger_level = "2"
		action = "1"
	case RuleIdPFCPSessionEstablishmentRequest:
		pfcpN4SMFResult = 1
		alter_stage = "3"
		danger_level = "1"
		action = "1"
	case RuleIdSCTPOverTCPorUDP:
		sctpSuperResult = 1
		alter_stage = "2"
		danger_level = "2"
		action = "1"
	case RuleIdSCTPShutdownCompleteFlood:
		sctpInitFloodResult = 1
		alter_stage = "2"
		danger_level = "3"
		signalStormSCTPInitFloodResult = 1
		action = "1"
	case RuleIdSCTPInitFlood:
		sctpInitFloodResult = 1
		signalStormSCTPInitFloodResult = 1
		alter_stage = "2"
		danger_level = "3"
		action = "1"
	case RuleIdSCTPDataFlood:
		signalStormSCTPInitFloodResult = 1
		alter_stage = "2"
		danger_level = "3"
		action = "1"
	case RuleIdSCTPInitackFlood:
		signalStormSCTPInitFloodResult = 1
		alter_stage = "2"
		danger_level = "3"
		action = "1"
	case RuleIdSCTPSackFlood:
		signalStormSCTPInitFloodResult = 1
		alter_stage = "2"
		danger_level = "3"
		action = "1"
	case RuleIdSCTPHeartbeatFlood:
		signalStormSCTPInitFloodResult = 1
		alter_stage = "2"
		danger_level = "3"
		action = "1"
	case RuleIdSCTPHeartbeatAckFlood:
		signalStormSCTPInitFloodResult = 1
		alter_stage = "2"
		danger_level = "3"
		action = "1"
	case RuleIdSCTPShutdownFlood:
		signalStormSCTPInitFloodResult = 1
		alter_stage = "2"
		danger_level = "3"
		action = "1"
	case RuleIdSCTPErrorFlood:
		signalStormSCTPInitFloodResult = 1
		alter_stage = "2"
		danger_level = "3"
		action = "1"
	case RuleIdSCTPShutdownAckFlood:
		signalStormSCTPInitFloodResult = 1
		alter_stage = "2"
		danger_level = "3"
		action = "1"
	case RuleIdSCTPCookieEchoFlood:
		signalStormSCTPInitFloodResult = 1
		alter_stage = "2"
		danger_level = "3"
		action = "1"
	case RuleIdSCTPCookieAckFlood:
		signalStormSCTPInitFloodResult = 1
		alter_stage = "2"
		danger_level = "3"
		action = "1"
	case RuleIdUEContextReleaseRequest:
		ngapReleaseRequestResult = 1
		alter_stage = "2"
		danger_level = "3"
		action = "1"
	case RuleIdNGAPUnknown:
		alter_stage = "2"
		danger_level = "1"
		ngap_from_result = 1
		action = "1"
	default:
		log.Printf("unknown rule id: %s", ruleId)
	}
}

// 定义恶意字符的正则表达式
var maliciousPattern = regexp.MustCompile(`[\x00-\x1F\x7F-\x9F]|['"();--]|<[^>]*>`)

// 创建一个检测器实例，窗口大小为10秒，阈值为100     //TEID检测器
var detector_teid = detector.NewTEIDBurstDetector(10*time.Second, 10)

// 创建一个检测器实例，窗口大小为5秒，阈值为100     //DDOS检测器
var detector_gtpddos = detector.NewGTPUDDoSdetector(5*time.Second, 10)

// 创建一个检测器实例，窗口大小为5秒，阈值为1000    //sctp init flood检测器
var detector_sctpinit = detector.NewSCTPINITFloodDetector(5*time.Second, 10)

func main() {

	logFile, err := os.OpenFile("../log/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	defer logFile.Close()

	// 设置log输出到文件
	log.SetOutput(logFile)

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "10.1.200.14:6379", // Redis 服务器地址和端口
		Password: "",                 // Redis 访问密码，如果没有可以为空字符串
		DB:       0,                  // 使用的 Redis 数据库编号，默认为 0
	})

	// 使用 Ping() 方法测试是否成功连接到 Redis 服务器
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Failed to connect to Redis:", err)
		return
	}
	fmt.Println("Connected to Redis:", pong)

	// 订阅频道
	pubsub := rdb.Subscribe(ctx, "packet_channel")
	defer pubsub.Close()

	// 启动TEID检测器
	detector_teid.Start()
	// 启动gtp ddos检测器
	detector_gtpddos.Start()

	//启动SCTP init检测器
	detector_sctpinit.Start()

	// 获取clickhouse数据库连接,单例
	db_clickhouse, err = db.GetDB()

	if err != nil {
		// 处理错误
		log.Printf("Error getting database connection: %v", err)
	}

	defer db_clickhouse.Close()

	//获取mysql数据库连接,单例
	db_mysql, err = db.GetMySQLDB()

	if err != nil {
		log.Printf("Error getting database connection: %v", err)
		return
	}

	defer db_mysql.Close()

	// 处理消息
	// 等待消息
	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Fatalf("Failed to receive message: %v", err)
		}
		fmt.Printf("Received message from channel '%s': %02x\n", msg.Channel, msg.Payload)

		timestamp := msg.Timestamp
		// 将时间戳格式化为你所期望的格式
		formattedTime := timestamp.Format(time.RFC3339)
		//fmt.Printf("Message at Offset %d, Key = %s, Value (Hex) = %02x, Received message at time: %s\n ", msg.Offset, string(msg.Key), msg.Payload, formattedTime)

		// 将Value中的每个字节以16进制形式打印，以便更好地观察TCP报文的二进制内容
		// for i := 0; i < len(msg.Payload); i++ {
		// 	fmt.Printf("%02x ", msg.Payload[i])
		// }
		// fmt.Println() // 换行

		// 解析数据包
		parseNetworkPacket(msg.Payload)
		fmt.Println("Network packet parsed")

		pacp_info = ""

		pacp_info = fmt.Sprintf("%02x", msg.Payload)

		length = uint64(len(msg.Payload))

		globel_id, err := uuid.NewUUID()

		srcip := src_ip // 您想要查询的 IP 地址
		nfInfo, err := db.GetNFInformationByIP(db_mysql, srcip)
		if err != nil {
			fmt.Printf("Error getting NF information: %v\n", err) //此时可以判断为异常，说明nf无法识别，可以用来关联分析
			src_nf = "Unknown"

		} else {
			src_nf = nfInfo.NfType
		}

		dstip := dst_ip // 您想要查询的 IP 地址
		nfInfo, err = db.GetNFInformationByIP(db_mysql, dstip)
		if err != nil {
			fmt.Printf("Error getting NF information: %v\n", err) //此时可以判断为异常，说明nf无法识别，可以用来关联分析
			dst_nf = "Unknown"

		} else {
			dst_nf = nfInfo.NfType
		}

		if protocol_5g != "" {
			user_ip = "192.168.1.1"
			address_network = "5G网络"
		}

		if protocol_5g == "GTP-U" {
			location_ne = "N3"
		} else if protocol_5g == "PFCP" {
			location_ne = "N4"
		} else if protocol_5g == "SCTP" {
			location_ne = "N1、N2"
		} else if protocol_5g == "NGAP" {
			location_ne = "N1、N2"
		} else if protocol_5g == "GTP-C" {
			location_ne = "N3"
		} else if protocol_5g == "NAS" {
			location_ne = "N1"
		}

		userip := user_ip

		terminalInfo, err := db.GetTerminalInformationByIP(db_mysql, userip)
		if err != nil {
			user_ip = ""
			user_imsi = ""
			user_subscriber_identity = ""
			user_equipment_identity = ""
		} else {
			user_ip = userip
			user_imsi = terminalInfo.Imsi
			user_subscriber_identity = terminalInfo.Guti
			user_equipment_identity = terminalInfo.Imei
		}

		setSignalStormResults(ruleId)

		// 将 uint64 转换为 string
		sctpVerificationTagStr := strconv.FormatUint(sctp_verificationTag, 10) // 10 表示十进制格式

		// 将 uint64 转换为 string
		NgapprocedureCodeStr := strconv.FormatUint(NgapprocedureCode, 10) // 10 表示十进制格式

		// 日志数据
		logEntry := db.LogDetectionResultModel{
			GlobelID:                         globel_id.String(),
			SrcIP:                            src_ip,
			DstIP:                            dst_ip,
			SrcMac:                           src_mac,
			DstMac:                           dst_mac,
			SrcPort:                          src_port,
			DstPort:                          dst_port,
			SrcNf:                            src_nf,
			DstNf:                            dst_nf,
			UserIP:                           user_ip,
			UserImsI:                         user_imsi,
			UserSubscriberIdentity:           user_subscriber_identity,
			UserEquipmentIdentity:            user_equipment_identity,
			NetworkSliceSsd:                  network_slice_ssd,
			NetworkSliceSt:                   network_slice_st,
			IPLayer:                          ip_layer,
			ProtocolApplicationLayer:         protocol_application_layer,
			ProtocolTransportLayer:           protocol_transport_layer,
			Protocol5g:                       protocol_5g,
			Length:                           length,
			SctpChunksFlags:                  sctp_chunks_flags,
			SctpVerificationTag:              sctp_verificationTag,
			SctpChunksType:                   sctp_chunks_type,
			SctpIeLength:                     sctp_ie_length,
			PfcpMessageType:                  pfcp_message_type,
			PfcpIeLength:                     pfcp_ie_length,
			PfcpHeaderFlagApare:              pfcp_header_flag_apare,
			GtpUMessageType:                  gtp_u_message_type,
			GtpUIeLength:                     gtp_u_ie_length,
			GtpUHeaderFlagApare:              gtp_u_header_flag_apare,
			GtpUIeReserved:                   gtp_u_reserved,
			GtpCMessageType:                  gtp_c_message_type,
			GtpCIeLength:                     gtp_c_ie_length,
			GtpCHeaderFlagApare:              gtp_c_header_flag_apare,
			GtpCIeReserved:                   gtp_c_ie_reserved,
			NgapProcedureCode:                NgapprocedureCode,
			NgapCriticlity:                   ngap_criticlity,
			NgapMessageType:                  ngap_message_type,
			NasMessageType:                   nas_message_type,
			NasExtendedProtocolDiscriminator: nas_extended_protocol_discriminator,
			NasSecurityHeaderType:            nas_security_header_type,
			Http2Method:                      http2Method,
			Http2SbiApi:                      http2SbiApi,
			Http2Scheme:                      http2Scheme,
			Action:                           action,
			RuleId:                           ruleId,
			AddressNetwork:                   address_network,
			AisScene:                         ais_scene,
			LocationNe:                       location_ne,
			Alter_stage:                      alter_stage,
			DisposeStatus:                    dispose_status,
			DangerLevel:                      danger_level,
			PacpInfo:                         pacp_info,
		}

		startTime := time.Now() // 记录事务开始时间
		// 开始事务
		tx, err := db_clickhouse.Begin()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Transaction begun")

		// 存储日志数据
		err = db.StoreLog(tx, logEntry)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}
		fmt.Println("Log data stored")

		// 提交事务
		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}

		endTime := time.Now() // 记录事务提交时间
		fmt.Printf("Transaction committed at %v\n", endTime)

		// 输出插入数据库所花费的时间
		duration := endTime.Sub(startTime)
		fmt.Printf("Time taken to insert data: %v\n", duration)

		fmt.Println("Transaction committed")

		//下面根据协议进行存储
		// 准备要插入的数据
		result_PFCP := db.PFCDetectionResult{
			GlobalID:                globel_id.String(),
			StampTime:               time.Now().UTC(),
			SrcIP:                   src_ip,
			DstIP:                   dst_ip,
			SrcPort:                 int(src_port),
			DstPort:                 int(dst_port),
			IMSI:                    user_imsi,
			SubscriberIdentity:      user_subscriber_identity,
			EquipmentIdentity:       user_equipment_identity,
			Protocol:                protocol_5g,
			Length:                  int(length),
			PFCPVersion:             pfcp_version,
			SEID:                    seid_value,
			MessageType:             pfcp_message_type,
			RuleID:                  ruleId,
			PFCPPDRResult:           pfcpPDRResult,
			PFCPFDRResult:           pfcpFDRResult,
			PFCPDeleteSessionResult: pfcfDeleteSessionResult,
			PFCPN4SMFResult:         pfcpN4SMFResult,
			PFCPFormResult:          pfcpFormResult,
			DisposeStatus:           dispose_status,
			AddressNetwork:          address_network,
			AISScene:                ais_scene,
			LocationNE:              location_ne,
			AlterStage:              alter_stage,
			DangerLevel:             danger_level,
			PACPInfo:                pacp_info,
		}

		result_GTPU := db.GTPUDetectionResult{
			GlobalID:           globel_id.String(),
			StampTime:          time.Now().UTC(),
			SrcIP:              src_ip,
			DstIP:              dst_ip,
			SrcPort:            int(src_port),
			DstPort:            int(dst_port),
			IMSI:               user_imsi,
			SubscriberIdentity: user_subscriber_identity,
			EquipmentIdentity:  user_equipment_identity,
			Protocol:           protocol_5g,
			Length:             int(length),
			GTPVersion:         GTPU_version,
			TEID:               teid_value,
			MessageType:        gtp_u_message_type,
			RuleID:             ruleId,
			GTPTEIDResult:      gtpuTEIDResult,
			GTPPayloadResult:   gtpuPayloadResult,
			GTPGTPinGTPResult:  gtpuGTPinGTPResult,
			GTPFormResult:      gtpuFormResult,
			DisposeStatus:      dispose_status,
			AddressNetwork:     address_network,
			AISScene:           ais_scene,
			LocationNE:         location_ne,
			AlterStage:         alter_stage,
			DangerLevel:        danger_level,
			PACPInfo:           pacp_info,
		}

		result_GTPC := db.GTPCDetectionResult{
			GlobalID:           globel_id.String(),
			StampTime:          time.Now().UTC(),
			SrcIP:              src_ip,
			DstIP:              dst_ip,
			SrcPort:            int(src_port),
			DstPort:            int(dst_port),
			IMSI:               user_imsi,
			SubscriberIdentity: user_subscriber_identity,
			EquipmentIdentity:  user_equipment_identity,
			Protocol:           protocol_5g,
			Length:             int(length),
			GTPVersion:         GTPU_version,
			TEID:               teid_value,
			MessageType:        gtp_c_message_type,
			RuleID:             ruleId,
			GTPFormResult:      0,
			DisposeStatus:      dispose_status,
			AddressNetwork:     address_network,
			AISScene:           ais_scene,
			LocationNE:         location_ne,
			AlterStage:         alter_stage,
			DangerLevel:        danger_level,
			PACPInfo:           pacp_info,
		}

		result_SCTP := db.SCTPDetectionResult{
			GlobalID:                     globel_id.String(),
			StampTime:                    time.Now().UTC(),
			SrcIP:                        src_ip,
			DstIP:                        dst_ip,
			SrcPort:                      int(src_port),
			DstPort:                      int(dst_port),
			IMSI:                         user_imsi,
			SubscriberIdentity:           user_subscriber_identity,
			EquipmentIdentity:            user_equipment_identity,
			Protocol:                     protocol_5g,
			Length:                       int(length),
			ChunksType:                   sctp_chunks_type,
			Sctp_verificationTag:         sctpVerificationTagStr,
			RuleID:                       ruleId,
			SCTPFourHandshakesDDOSResult: sctpFourHandshakesDDOSResult,
			SCTPSuperResult:              sctpSuperResult,
			SCTPMultichunkResult:         sctpMultichunkResult,
			SCTPInitFloodResult:          sctpInitFloodResult,
			DisposeStatus:                dispose_status,
			AddressNetwork:               address_network,
			AISScene:                     ais_scene,
			LocationNE:                   location_ne,
			AlterStage:                   alter_stage,
			DangerLevel:                  danger_level,
			PACPInfo:                     pacp_info,
		}

		result_NGAP := db.NGAPDetectionResult{
			GlobalID:                 globel_id.String(),
			StampTime:                time.Now().UTC(),
			SrcIP:                    src_ip,
			DstIP:                    dst_ip,
			SrcPort:                  int(src_port),
			DstPort:                  int(dst_port),
			IMSI:                     user_imsi,
			SubscriberIdentity:       user_subscriber_identity,
			EquipmentIdentity:        user_equipment_identity,
			Protocol:                 protocol_5g,
			Length:                   int(length),
			MessageType:              ngap_message_type,
			ProcedureCode:            NgapprocedureCodeStr,
			Criticlity:               ngap_criticlity,
			RuleID:                   ruleId,
			NGAPReleaseRequestResult: ngapReleaseRequestResult,
			Ngap_from_result:         ngap_from_result,
			DisposeStatus:            dispose_status,
			AddressNetwork:           address_network,
			AISScene:                 ais_scene,
			LocationNE:               location_ne,
			AlterStage:               alter_stage,
			DangerLevel:              danger_level,
			PACPInfo:                 pacp_info,
		}

		result_Signal := db.SignalStormDetectionResult{
			GlobalID:                                globel_id.String(),
			StampTime:                               time.Now().UTC(),
			SrcIP:                                   src_ip,
			DstIP:                                   dst_ip,
			SrcPort:                                 int(src_port),
			DstPort:                                 int(dst_port),
			IMSI:                                    user_imsi,
			SubscriberIdentity:                      user_subscriber_identity,
			EquipmentIdentity:                       user_equipment_identity,
			Protocol:                                protocol_5g,
			Length:                                  int(length),
			TEID:                                    teid_value,
			SEID:                                    seid_value,
			Byte:                                    int(length),
			Threshold:                               10,
			SignalStormMulUELoginResult:             signalStormMulUELoginResult,
			SignalStormMulAccessResult:              signalStormMulAccessResult,
			SignalStormNFFaultyResult:               signalStormNFFaultyResult,
			SignalStormGTPUSynDDOSResult:            signalStormGTPUSynDDOSResult,
			SignalStormPFCPN4Result:                 signalStormPFCPN4Result,
			SignalStormSCTPInitFloodResult:          signalStormSCTPInitFloodResult,
			SignalStormSCTPFourHandshakesDDOSResult: signalStormSCTPFourHandshakesDDOSResult,
			RuleID:                                  ruleId,
			DisposeStatus:                           dispose_status,
			AddressNetwork:                          address_network,
			AISScene:                                ais_scene,
			LocationNE:                              location_ne,
			AlterStage:                              alter_stage,
			DangerLevel:                             danger_level,
			PACPInfo:                                pacp_info,
		}

		if protocol_5g == "PFCP" {
			// 插入数据到数据库
			if err := db.InsertPFCDetectionResult(db_mysql, result_PFCP); err != nil {
				log.Printf("Error inserting data PFCP : %v\n", err)
				return
			}
		}

		if protocol_5g == "GTP-U" {
			// 插入数据到数据库
			if err := db.InsertGTPUDetectionResult(db_mysql, result_GTPU); err != nil {
				log.Printf("Error inserting data GTP-U: %v\n", err)
				return
			}
		}

		if protocol_5g == "SCTP" {
			// 插入数据到数据库
			if err := db.InsertSCTPDetectionResult(db_mysql, result_SCTP); err != nil {
				log.Printf("Error inserting data SCTP: %v\n", err)
				return
			}
		}

		if protocol_5g == "NGAP" {
			// 插入数据到数据库
			if err := db.InsertNGAPDetectionResult(db_mysql, result_NGAP); err != nil {
				log.Printf("Error inserting data NGAP: %v\n", err)
				return
			}
		}

		if protocol_5g == "GTP-C" {
			// 插入数据到数据库
			if err := db.InsertGTPCDetectionResult(db_mysql, result_GTPC); err != nil {
				log.Printf("Error inserting data GTP-C: %v\n", err)
				return
			}
		}

		if signalStormMulUELoginResult == 1 || signalStormMulAccessResult == 1 || signalStormNFFaultyResult == 1 || signalStormGTPUSynDDOSResult == 1 || signalStormPFCPN4Result == 1 || signalStormSCTPInitFloodResult == 1 || signalStormSCTPFourHandshakesDDOSResult == 1 {
			// 插入数据到数据库
			if err := db.InsertSignalStormDetectionResult(db_mysql, result_Signal); err != nil {
				log.Printf("Error inserting data signal: %v\n", err)
				return
			}
		}

		fmt.Printf("Log stored successfully!")

	}
	// 停止TEID检测器
	detector_teid.Stop()
	// 停止DDOS GTP检测器
	detector_gtpddos.Stop()

	detector_sctpinit.Stop()

}

// 解析网络数据包
func parseNetworkPacket(packetData []byte) {

	//定义用来传参存储的参数

	// 网络五元组
	src_ip = ""
	dst_ip = ""
	src_port = uint64(0)
	dst_port = uint64(0)
	src_mac = ""
	dst_mac = ""

	// 切片信息
	network_slice_st = "0x01"
	network_slice_ssd = ""

	//0x01: eMBB（增强型移动宽带）
	//0x02: mMTC（大规模机器类型通信）
	//0x03: URLLC（超可靠低延迟通信）

	protocol_5g = ""

	// sctp协议信息
	sctp_chunks_flags = ""
	sctp_verificationTag = uint64(0)
	sctp_chunks_type = ""
	sctp_ie_length = uint64(0)

	// pfcp协议信息
	pfcp_message_type = ""
	pfcp_ie_length = uint64(0)
	pfcp_header_flag_apare = ""

	// gtpv1U协议信息
	gtp_u_message_type = ""
	gtp_u_ie_length = uint64(0)
	gtp_u_header_flag_apare = ""
	gtp_u_reserved = ""

	// gtpv2C协议信息
	gtp_c_message_type = ""
	gtp_c_ie_length = uint64(0)
	gtp_c_header_flag_apare = ""
	gtp_c_ie_reserved = ""

	// ngap协议信息
	ngap_message_type = ""
	ngap_criticlity = ""
	NgapprocedureCode = uint64(0)

	// nas协议信息
	nas_message_type = ""
	nas_extended_protocol_discriminator = ""
	nas_security_header_type = ""

	// http2协议信息
	http2Method = ""
	http2SbiApi = ""
	http2Scheme = ""

	// 其他信息
	ruleId = "0"

	// action的值，只能是1 或0,   0表示pass，1表示alert, 2表示block
	action = "0"

	// 告警等级：0表示正常，1表示低危， 2表示中危，3表示高危
	danger_level = "0"

	// 处置状态： 0表示不需处理，1表示有告警未处理，2表示有告警已处理
	dispose_status = "0"

	// 网络位置：   Web网络，5G接入网，5G核心网
	address_network = "Web网络"

	ais_scene = "eMBB"

	// 5G网元位置：WEB、UE~gNB、UE~AMF、gNB~AMF、gNB~UPF、UPF~SMF、UPF~DN、AMF~SMF、AMF~AUSF、AMF~NRF、Known
	location_ne = "WEB"
	// 使用gopacket解析数据包
	parsedPacket := gopacket.NewPacket(packetData, layers.LayerTypeEthernet, gopacket.DecodeOptions{})

	// 打印解析结果
	PrintPacketInfo(parsedPacket)
}

// 打印出当前检测出的所有层
func printPacket_layers(packet gopacket.Packet) {
	// 检测 GTP-IN-GTP
	var gtpCount = 0
	fmt.Println("All packet layers:")
	for i, layer := range packet.Layers() {
		fmt.Printf("  层类型: %s\n", layer.LayerType())
		fmt.Printf("  层数据: %02x\n", layer.LayerContents())
		if layer.LayerType() == layers.LayerTypeGTPv1U {
			gtpCount++
			// 检查 GTPv1-U 的上一层是否为 UDP
			if i > 0 {
				previousLayer := packet.Layers()[i-1]
				if previousLayer.LayerType() != layers.LayerTypeUDP {
					// 如果不是 UDP，则设置 ruleId
					ruleId = RuleIdGTPv1UWithoutUDP
					fmt.Println("GTPv1-U layer is not preceded by UDP, possible security threat.")
				}
			}
		}

		// 检查SCTP上层是否是TCP或UDP
		if layer.LayerType() == layers.LayerTypeSCTP {
			if i > 0 {
				previousLayer := packet.Layers()[i-1]
				if previousLayer.LayerType() == layers.LayerTypeTCP || previousLayer.LayerType() == layers.LayerTypeUDP {
					ruleId = RuleIdSCTPOverTCPorUDP
					fmt.Println("SCTP layer is detected over TCP or UDP, which is not expected and may indicate a security threat.")
				}
			}
		}
	}

	// 如果 GTP 层数量大于 1，则可能存在 GTP-in-GTP 攻击
	if gtpCount > 1 {
		ruleId = RuleIdGTPINGTP
		fmt.Println("Multiple GTP layers detected, possible GTP-in-GTP attack.")
	}
}

// 解析pfcp协议
func ParsePFCP(payload []byte) {
	//解析pfcp协议
	// 创建一个缓冲区读取器
	reader := bytes.NewReader(payload)

	// 定义一个缓冲区来存储从 reader 中读取的数据
	var buf bytes.Buffer

	// 将 reader 中的所有数据读取到缓冲区 buf 中
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return
	}

	// 现在 buf.Bytes() 提供了原始字节切片
	data := buf.Bytes()

	// 定义PFCP头部变量
	var header pfcp.Header

	// 解析PFCP包头,,,,,解析pfcp协议
	if err := header.UnmarshalBinary(data); err != nil {
		fmt.Println("Failed to unmarshal PFCP header", err)
		return
	}

	// 验证PFCP版本
	if header.Version != pfcp.PfcpVersion {
		fmt.Println("Invalid PFCP version:", header.Version)
		return
	}

	protocol_5g = "PFCP"

	pfcp_version = int(header.Version)
	seid_value = fmt.Sprintf("SEID:%d", header.SEID)

	// 打印或处理PFCP头部信息
	fmt.Printf("PFCP Version: %d\n", header.Version)
	fmt.Printf("Message Priority: %d\n", header.MessagePriority)
	fmt.Printf("Message Type: %d\n", header.MessageType)
	fmt.Printf("Message Length: %d\n", header.MessageLength)
	fmt.Printf("Sequence Number: %d\n", header.SequenceNumber)
	fmt.Printf("SEID: %d\n", header.SEID)

	//打印pfcp协议的ie部分
	//通过pfcp.MessageType来判断是哪个类型的pfcp消息，然后解析对应的ie部分

	// 从缓冲区读取器中跳过头部长度
	_, _ = reader.Seek(int64(header.Len()), 0)

	pfcp_header_flag_apare = fmt.Sprintf("%04x", header.SequenceNumber)
	pfcp_ie_length = uint64(header.MessageLength) - uint64(header.Len())

	if header.Len() != 16 || header.Len() != 12 {
		fmt.Println("Invalid PFCP message length")
		ruleId = RuleIdPFCPInvalidLength
	}

	if header.MessageType == pfcp.PFCP_VERSION_NOT_SUPPORTED_RESPONSE {
		fmt.Println("PFCP_VERSION_NOT_SUPPORTED_RESPONSE")
		ruleId = RuleIdPFCPVersionNotSupportedResponse
	}
	// 根据PFCP消息类型解析IE部分
	switch header.MessageType {
	case pfcp.PFCP_HEARTBEAT_REQUEST:
		// 解析HeartbeatRequest消息的IE
		pfcp_message_type = "PFCP_HEARTBEAT_REQUEST"

		var heartbeatRequest pfcp.HeartbeatRequest
		if err := tlv.Unmarshal(data, &heartbeatRequest); err != nil {
			log.Println("Failed to unmarshal HeartbeatRequest IE:", err)
			return
		}
		// 打印或处理HeartbeatRequest中的IE
		fmt.Println("HeartbeatRequest:", heartbeatRequest)

	case pfcp.PFCP_HEARTBEAT_RESPONSE:
		// 解析HeartbeatResponse消息的IE

		pfcp_message_type = "PFCP_HEARTBEAT_RESPONSE"
		var heartbeatResponse pfcp.HeartbeatResponse
		if err := tlv.Unmarshal(data, &heartbeatResponse); err != nil {
			log.Println("Failed to unmarshal HeartbeatResponse IE:", err)
			return
		}
		// 打印或处理HeartbeatResponse中的IE
		fmt.Println("HeartbeatResponse:", heartbeatResponse)

	case pfcp.PFCP_ASSOCIATION_SETUP_REQUEST:
		// 解析AssociationSetupRequest消息的IE
		pfcp_message_type = "PFCP_ASSOCIATION_SETUP_REQUEST"
		var assocSetupRequest pfcp.PFCPAssociationSetupRequest
		if err := tlv.Unmarshal(data, &assocSetupRequest); err != nil {
			log.Println("Failed to unmarshal AssociationSetupRequest IE:", err)
			return
		}
		// 打印或处理AssociationSetupRequest中的IE
		fmt.Println("AssociationSetupRequest:", assocSetupRequest)

	case pfcp.PFCP_ASSOCIATION_SETUP_RESPONSE:
		// 解析AssociationSetupResponse消息的IE
		pfcp_message_type = "PFCP_ASSOCIATION_SETUP_RESPONSE"

		var assocSetupResponse pfcp.PFCPAssociationSetupResponse
		if err := tlv.Unmarshal(data, &assocSetupResponse); err != nil {
			log.Println("Failed to unmarshal AssociationSetupResponse IE:", err)
			return
		}
		//打印或处理AssociationSetupResponse中的IE
		fmt.Println("AssociationSetupResponse:", assocSetupResponse)

	case pfcp.PFCP_ASSOCIATION_UPDATE_REQUEST:
		// 解析AssociationUpdateRequest消息的IE
		pfcp_message_type = "PFCP_ASSOCIATION_UPDATE_REQUEST"

		var assocUpdateRequest pfcp.PFCPAssociationUpdateRequest
		if err := tlv.Unmarshal(data, &assocUpdateRequest); err != nil {
			log.Println("Failed to unmarshal AssociationUpdateRequest IE:", err)
			return
		}
		// 打印或处理AssociationUpdateRequest中的IE
		fmt.Println("AssociationUpdateRequest:", assocUpdateRequest)
		ruleId = PFCPassocUpdateRequest

	case pfcp.PFCP_ASSOCIATION_UPDATE_RESPONSE:
		// 解析AssociationUpdateResponse消息的IE
		pfcp_message_type = "PFCP_ASSOCIATION_UPDATE_RESPONSE"

		var assocUpdateResponse pfcp.PFCPAssociationUpdateResponse
		if err := tlv.Unmarshal(data, &assocUpdateResponse); err != nil {
			log.Println("Failed to unmarshal AssociationUpdateResponse IE:", err)
			return
		}
		// 打印或处理AssociationUpdateResponse中的IE
		fmt.Println("AssociationUpdateResponse:", assocUpdateResponse)
		ruleId = PFCPassocUpdateRequest

	case pfcp.PFCP_ASSOCIATION_RELEASE_REQUEST:
		// 解析AssociationReleaseRequest消息的IE
		pfcp_message_type = "PFCP_ASSOCIATION_RELEASE_REQUEST"

		var assocReleaseRequest pfcp.PFCPAssociationReleaseRequest
		if err := tlv.Unmarshal(data, &assocReleaseRequest); err != nil {
			log.Println("Failed to unmarshal AssociationReleaseRequest IE:", err)
			return
		}

		//检测PFCP的释放请求
		detector_sctpinit.AddEvent(uint8(pfcp.PFCP_ASSOCIATION_RELEASE_REQUEST), dst_ip, uint16(dst_port))

		if attackDetected, key := detector_sctpinit.CheckINITFlood(); attackDetected {
			fmt.Printf("SCTP INIT Flood attack detected to IP: %s on port: %d\n", key.DestIP, key.DestPort)
			ruleId = RuleIdPFCPReleaseRequest
		}

		// 打印或处理AssociationReleaseRequest中的IE
		fmt.Println("AssociationReleaseRequest:", assocReleaseRequest)
		ruleId = PFCPassocReleaseRequest

	case pfcp.PFCP_ASSOCIATION_RELEASE_RESPONSE:
		// 解析AssociationReleaseResponse消息的IE
		pfcp_message_type = "PFCP_ASSOCIATION_RELEASE_RESPONSE"

		var assocReleaseResponse pfcp.PFCPAssociationReleaseResponse
		if err := tlv.Unmarshal(data, &assocReleaseResponse); err != nil {
			log.Println("Failed to unmarshal AssociationReleaseResponse IE:", err)
			return
		}
		// 打印或处理AssociationReleaseResponse中的IE
		fmt.Println("AssociationReleaseResponse:", assocReleaseResponse)
		ruleId = PFCPassocReleaseResponse

	case pfcp.PFCP_NODE_REPORT_REQUEST:
		// 解析NodeReportRequest消息的IE
		pfcp_message_type = "PFCP_NODE_REPORT_REQUEST"
		var nodeReportRequest pfcp.PFCPNodeReportRequest
		if err := tlv.Unmarshal(data, &nodeReportRequest); err != nil {
			log.Println("Failed to unmarshal NodeReportRequest IE:", err)
			return
		}
		// 打印或处理NodeReportRequest中的IE
		fmt.Println("NodeReportRequest:", nodeReportRequest)

	case pfcp.PFCP_NODE_REPORT_RESPONSE:
		// 解析NodeReportResponse消息的IE
		pfcp_message_type = "PFCP_NODE_REPORT_RESPONSE"
		var nodeReportResponse pfcp.PFCPNodeReportResponse
		if err := tlv.Unmarshal(data, &nodeReportResponse); err != nil {
			log.Println("Failed to unmarshal NodeReportResponse IE:", err)
			return
		}
		// 打印或处理NodeReportResponse中的IE
		fmt.Println("NodeReportResponse:", nodeReportResponse)

	case pfcp.PFCP_SESSION_SET_DELETION_REQUEST:
		// 解析SessionSetDeletionRequest消息的IE
		pfcp_message_type = "PFCP_SESSION_SET_DELETION_REQUEST"
		var sessionSetDeletionRequest pfcp.PFCPSessionSetDeletionRequest
		if err := tlv.Unmarshal(data, &sessionSetDeletionRequest); err != nil {
			log.Println("Failed to unmarshal SessionSetDeletionRequest IE:", err)
			return
		}
		//检测PFCP的释放请求
		detector_sctpinit.AddEvent(uint8(pfcp.PFCP_SESSION_SET_DELETION_REQUEST), dst_ip, uint16(dst_port))

		if attackDetected, key := detector_sctpinit.CheckINITFlood(); attackDetected {
			fmt.Printf("SCTP INIT Flood attack detected to IP: %s on port: %d\n", key.DestIP, key.DestPort)
			ruleId = RuleIdPFCPDeletionRequest
		}
		// 打印或处理SessionSetDeletionRequest中的IE
		fmt.Println("SessionSetDeletionRequest:", sessionSetDeletionRequest)
		ruleId = PFCPsessionSetDeletionRequest

	case pfcp.PFCP_SESSION_SET_DELETION_RESPONSE:
		// 解析SessionSetDeletionResponse消息的IE
		pfcp_message_type = "PFCP_SESSION_SET_DELETION_RESPONSE"
		var sessionSetDeletionResponse pfcp.PFCPSessionSetDeletionResponse
		if err := tlv.Unmarshal(data, &sessionSetDeletionResponse); err != nil {
			log.Println("Failed to unmarshal SessionSetDeletionResponse IE:", err)
			return
		}
		//打印或处理SessionSetDeletionResponse中的IE
		fmt.Println("SessionSetDeletionResponse:", sessionSetDeletionResponse)
		ruleId = PFCPsessionSetDeletionResponse

	case pfcp.PFCP_SESSION_ESTABLISHMENT_REQUEST:
		// 解析SessionEstablishmentRequest消息的IE
		pfcp_message_type = "PFCP_SESSION_ESTABLISHMENT_REQUEST"
		var sessionEstablishmentRequest pfcp.PFCPSessionEstablishmentRequest
		if err := tlv.Unmarshal(data, &sessionEstablishmentRequest); err != nil {
			log.Println("Failed to unmarshal SessionEstablishmentRequest IE:", err)
			return
		}
		//检测PFCP的释放请求
		detector_sctpinit.AddEvent(uint8(pfcp.PFCP_SESSION_ESTABLISHMENT_REQUEST), dst_ip, uint16(dst_port))

		if attackDetected, key := detector_sctpinit.CheckINITFlood(); attackDetected {
			fmt.Printf("SCTP INIT Flood attack detected to IP: %s on port: %d\n", key.DestIP, key.DestPort)
			ruleId = RuleIdPFCPSessionEstablishmentRequest
		}
		// 打印或处理SessionEstablishmentRequest中的IE
		fmt.Println("SessionEstablishmentRequest:", sessionEstablishmentRequest)

		if sessionEstablishmentRequest.CreatePDR == nil {
			ruleId = PFCPsessionEstablishmentRequestPDR
			break
		}

		if sessionEstablishmentRequest.CreateFAR == nil {
			ruleId = PFCPsessionEstablishmentRequestFAR
			break
		}

	case pfcp.PFCP_SESSION_ESTABLISHMENT_RESPONSE:
		// 解析SessionEstablishmentResponse消息的IE
		pfcp_message_type = "PFCP_SESSION_ESTABLISHMENT_RESPONSE"
		var sessionEstablishmentResponse pfcp.PFCPSessionEstablishmentResponse
		if err := tlv.Unmarshal(data, &sessionEstablishmentResponse); err != nil {
			log.Println("Failed to unmarshal SessionEstablishmentResponse IE:", err)
			return
		}
		// 打印或处理SessionEstablishmentResponse中的IE
		fmt.Println("SessionEstablishmentResponse:", sessionEstablishmentResponse)

	case pfcp.PFCP_SESSION_MODIFICATION_REQUEST:
		// 解析SessionModificationRequest消息的IE
		pfcp_message_type = "PFCP_SESSION_MODIFICATION_REQUEST"
		var sessionModificationRequest pfcp.PFCPSessionModificationRequest
		if err := tlv.Unmarshal(data, &sessionModificationRequest); err != nil {
			log.Println("Failed to unmarshal SessionModificationRequest IE:", err)
			return
		}
		// 打印或处理SessionModificationRequest中的IE
		fmt.Println("SessionModificationRequest:", sessionModificationRequest)

		if sessionModificationRequest.CreatePDR == nil {
			ruleId = PFCPsessionModificationRequestPDR
			break
		}

		if sessionModificationRequest.UpdatePDR == nil {
			ruleId = PFCPsessionModificationRequestPDR1
			break
		}

		if sessionModificationRequest.RemovePDR == nil {
			ruleId = PFCPsessionModificationRequestPDR2
			break
		}

		if sessionModificationRequest.CreateFAR == nil {
			ruleId = PFCPsessionModificationRequestFAR
			break
		}

		if sessionModificationRequest.RemoveFAR == nil {
			ruleId = PFCPsessionModificationRequestFAR1
			break
		}

		if sessionModificationRequest.UpdateFAR == nil {
			ruleId = PFCPsessionModificationRequestFAR2
			break
		}
		ruleId = PFCPsessionModificationRequest

	case pfcp.PFCP_SESSION_MODIFICATION_RESPONSE:
		// 解析SessionModificationResponse消息的IE
		pfcp_message_type = "PFCP_SESSION_MODIFICATION_RESPONSE"
		var sessionModificationResponse pfcp.PFCPSessionModificationResponse
		if err := tlv.Unmarshal(data, &sessionModificationResponse); err != nil {
			log.Println("Failed to unmarshal SessionModificationResponse IE:", err)
			return
		}
		// 打印或处理SessionModificationResponse中的IE
		fmt.Println("SessionModificationResponse:", sessionModificationResponse)
		ruleId = PFCPsessionModificationResponse

	case pfcp.PFCP_SESSION_DELETION_REQUEST:
		// 解析SessionDeletionRequest消息的IE
		pfcp_message_type = "PFCP_SESSION_DELETION_REQUEST"
		var sessionDeletionRequest pfcp.PFCPSessionDeletionRequest
		if err := tlv.Unmarshal(data, &sessionDeletionRequest); err != nil {
			log.Println("Failed to unmarshal SessionDeletionRequest IE:", err)
			return
		}
		// 打印或处理SessionDeletionRequest中的IE
		fmt.Println("SessionDeletionRequest:", sessionDeletionRequest)
		ruleId = PFCPsessionDeletionRequest1

	case pfcp.PFCP_SESSION_DELETION_RESPONSE:
		// 解析SessionDeletionResponse消息的IE
		pfcp_message_type = "PFCP_SESSION_DELETION_RESPONSE"
		var sessionDeletionResponse pfcp.PFCPSessionDeletionResponse
		if err := tlv.Unmarshal(data, &sessionDeletionResponse); err != nil {
			log.Println("Failed to unmarshal SessionDeletionResponse IE:", err)
			return
		}
		// 打印或处理SessionDeletionResponse中的IE
		fmt.Println("SessionDeletionResponse:", sessionDeletionResponse)
		ruleId = PFCPsessionDeletionRequest2

	case pfcp.PFCP_SESSION_REPORT_REQUEST:
		// 解析SessionReportRequest消息的IE
		pfcp_message_type = "PFCP_SESSION_REPORT_REQUEST"
		var sessionReportRequest pfcp.PFCPSessionReportRequest
		if err := tlv.Unmarshal(data, &sessionReportRequest); err != nil {
			log.Println("Failed to unmarshal SessionReportRequest IE:", err)
			return
		}
		// 打印或处理SessionReportRequest中的IE
		fmt.Println("SessionReportRequest:", sessionReportRequest)

	case pfcp.PFCP_SESSION_REPORT_RESPONSE:
		// 解析SessionReportResponse消息的IE
		pfcp_message_type = "PFCP_SESSION_REPORT_RESPONSE"
		var sessionReportResponse pfcp.PFCPSessionReportResponse
		if err := tlv.Unmarshal(data, &sessionReportResponse); err != nil {
			log.Println("Failed to unmarshal SessionReportResponse IE:", err)
			return
		}
		// 打印或处理SessionReportResponse中的IE
		fmt.Println("SessionReportResponse:", sessionReportResponse)

	default:
		log.Println("Unsupported PFCP message type:", header.MessageType)
		ruleId = PFCPinvaliadmessagetype
		return
	}

}

// parseNGAPMessage 解析NGAP协议消息
func parseNGAPMessage(data []byte) (*ngapType.NGAPPDU, error) {
	pdu := &ngapType.NGAPPDU{}
	err := aper.UnmarshalWithParams(data, pdu, "valueExt,valueLB:0,valueUB:2")
	if err != nil {
		return nil, fmt.Errorf("failed to parse NGAP message: %w", err)
	}
	return pdu, nil
}

// printNGAPFieldValues 以树状结构打印NGAP结构体字段值
func printNGAPFieldValues(v interface{}, indent string) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	printStructValues(val, indent)
}

// printStructValues 打印结构体字段值
func printStructValues(val reflect.Value, indent string) {
	typ := val.Type()
	if typ.Kind() != reflect.Struct {
		fmt.Println("provided value is not a struct")
		return
	}

	for i := 0; i < typ.NumField(); i++ {
		field := val.Field(i)
		fieldType := field.Type()
		fieldName := typ.Field(i).Name

		fmt.Printf("%sField: %s\n", indent, fieldName)
		if field.CanInterface() {
			printValue(field, indent+"  ")
		} else {
			fmt.Printf("%s  (inaccessible field of type %s)\n", indent, fieldType)
		}
	}
}

// printValue 打印字段值
func printValue(field reflect.Value, indent string) {
	switch field.Kind() {
	case reflect.Ptr:
		if field.IsNil() {
			fmt.Printf("%sNil pointer\n", indent)
		} else {
			printNGAPFieldValues(field.Interface(), indent)
		}
	case reflect.Struct:
		printStructValues(field, indent)
	case reflect.Slice:
		for i := 0; i < field.Len(); i++ {
			fmt.Printf("%sElement %d:\n", indent, i)
			printValue(field.Index(i), indent+"  ")
		}
	case reflect.String:
		fmt.Printf("%s%s\n", indent, field.String())
	default:
		if field.CanInterface() {
			fmt.Printf("%s%v\n", indent, field.Interface())
		} else {
			fmt.Printf("%s(inaccessible value)\n", indent)
		}
	}
}

// getValueFromStruct 通过反射从结构体中获取特定字段的值
func getValueFromStruct(v interface{}, fieldName string) interface{} {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		field := val.FieldByName(fieldName)
		if field.IsValid() && field.CanInterface() {
			return field.Interface()
		}
	}
	return nil
}

func PrintPacketInfo(packet gopacket.Packet) {
	// Iterate over all layers, printing out each layer type
	//打印出当前所有检测出的层
	printPacket_layers(packet)
	///.......................................................

	// 判断数据包是否为以太网数据包，可解析出源mac地址、目的mac地址、以太网类型（如ip类型）等
	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethernetLayer != nil {
		log.Println("Ethernet layer detected.")
		ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)

		src_mac = ethernetPacket.SrcMAC.String()
		dst_mac = ethernetPacket.DstMAC.String()
	}
	// Let's see if the packet is IP (even though the ether type told us)

	// 判断数据包是否为IPv4数据包，可解析出源ip、目的ip、协议号等
	ip4Layer := packet.Layer(layers.LayerTypeIPv4)
	if ip4Layer != nil {
		log.Println("IPv4 layer detected.")

		ip4Packet, _ := ip4Layer.(*layers.IPv4)
		fmt.Printf("From %s to %s\n", ip4Packet.SrcIP, ip4Packet.DstIP)
		fmt.Printf("IP procotl: %d\n", ip4Packet.Protocol)

		src_ip = ip4Packet.SrcIP.String()
		dst_ip = ip4Packet.DstIP.String()
		protocol_transport_layer = ip4Packet.Protocol.String()
		ip_layer = "IPv4"

	}

	// 判断数据包是否为IPv6数据包，可解析出源ip、目的ip、协议号等
	ip6Layer := packet.Layer(layers.LayerTypeIPv6)
	if ip6Layer != nil {
		log.Println("IPv6 layer detected.")
		ip6Packet, _ := ip6Layer.(*layers.IPv6)

		src_ip = ip6Packet.SrcIP.String()
		dst_ip = ip6Packet.DstIP.String()
		protocol_transport_layer = ip6Packet.NextHeader.String()
		ip_layer = "IPv6"
	}

	//判断是否存在TCP协议
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		log.Println("TCP layer detected.")
		tcp, ok := tcpLayer.(*layers.TCP)
		if !ok {
			fmt.Println("Error converting to TCP layer.")
			return
		}
		src_port = uint64(tcp.SrcPort)
		dst_port = uint64(tcp.DstPort)
		protocol_transport_layer = "TCP"
	}

	//判断数据包的传输层
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if ethernetLayer != nil && ip4Layer != nil && udpLayer != nil {
		// 解析 UDP 数据包
		log.Println("UDP layer detected.")

		payload := udpLayer.(*layers.UDP).Payload

		src_port = uint64(udpLayer.(*layers.UDP).SrcPort)
		dst_port = uint64(udpLayer.(*layers.UDP).DstPort)
		protocol_transport_layer = "UDP"

		ParsePFCP(payload)

		if pfcp_message_type != "" {
			protocol_5g = "PFCP"
		}

	}

	//判断是否存在GTPv1U协议
	gtpv1uLayer := packet.Layer(layers.LayerTypeGTPv1U)
	if gtpv1uLayer != nil {
		gtpv1u, ok := gtpv1uLayer.(*layers.GTPv1U)
		if !ok {
			log.Println("Error converting to GTPv1U layer.")
			return
		}

		message := detector.GetGTPv1Message(gtpv1u.MessageType)
		// 检测GTPv1U的Message Type不符合3GPP的规范：获取message-type，然后其是否不存在，或者是否为保留值
		if message == nil || message.Name == "Reserved" {
			ruleId = RuleIdGTPv1UMESSAGENIL
		}
		if gtpv1u.MessageType == 3 {
			ruleId = RuleIdGTPv1UMESSAGTYPE
		}

		gtp_u_message_type = message.Name
		protocol_5g = "GTP-U"
		gtp_u_ie_length = uint64(gtpv1u.MessageLength)
		gtp_u_header_flag_apare = fmt.Sprintf("%t", gtpv1u.ExtensionHeaderFlag)
		gtp_u_reserved = fmt.Sprintf("%d", gtpv1u.Reserved)

		GTPU_version = int(gtpv1u.Version)
		teid_value = fmt.Sprintf("%d", gtpv1u.TEID)

		// 打印GTPv1U层的所有字段
		fmt.Printf("GTPv1U Layer Information:\n")
		fmt.Printf("Version: %d\n", gtpv1u.Version)
		fmt.Printf("Protocol Type: %d\n", gtpv1u.ProtocolType)
		fmt.Printf("Reserved: %d\n", gtpv1u.Reserved)
		fmt.Printf("Extension Header Flag: %t\n", gtpv1u.ExtensionHeaderFlag)
		fmt.Printf("Sequence Number Flag: %t\n", gtpv1u.SequenceNumberFlag)
		fmt.Printf("NPDU Flag: %t\n", gtpv1u.NPDUFlag)
		fmt.Printf("Message Type: %d\n", gtpv1u.MessageType)
		fmt.Printf("Message Length: %d\n", gtpv1u.MessageLength)
		fmt.Printf("TEID: %d\n", gtpv1u.TEID)
		fmt.Printf("Sequence Number: %d\n", gtpv1u.SequenceNumber)
		fmt.Printf("NPDU: %d\n", gtpv1u.NPDU)

		// 检测GTPv1U的Message Length
		if gtpv1u.MessageLength < 8 || gtpv1u.MessageLength > 300 {
			ruleId = RuleIdGTPv1UMESSAGLENGTH
		}

		// 打印GTP Extension Headers信息
		if len(gtpv1u.GTPExtensionHeaders) > 0 {
			fmt.Println("Extension Headers:")
			for i, eh := range gtpv1u.GTPExtensionHeaders {
				fmt.Printf("  Extension Header %d:\n", i+1)
				fmt.Printf("    Type: %d\n", eh.Type)
				fmt.Printf("    Content: %02x\n", eh.Content)
				fmt.Printf("    Content Length: %d bytes\n", len(eh.Content))

				// 检测GTPv1U的Content
				// 1. 长度
				if len(eh.Content) > 100 {
					fmt.Println("Content is too long to display.")
					ruleId = RuleIdContentTooLong
					break
				}
				// 2. 内容，如果Content存在恶意字符，则标记
				if maliciousPattern.Match(eh.Content) {
					ruleId = RuleIdContentMalicious
					break
				}

			}
		} else {
			fmt.Println("No Extension Headers found.")
		}

		//检测GTPv1U的TEID爆破
		events := []detector.Event{
			{TEID: gtpv1u.TEID, IP: dst_ip, MessageType: gtpv1u.MessageType},
		}

		// 添加事件并检查爆破
		for _, event := range events {
			detector_teid.AddEvent(event)
			// 添加事件到 DDoS 检测器
			detector_gtpddos.AddEvent(gtpv1u.MessageType, dst_ip)
			// 检测GTPv1U的DDOS爆破
			gtp_u_ddos, key := detector_gtpddos.CheckDDoS()
			// 检测GTPv1U的TEID爆破
			burst, teid := detector_teid.CheckBurst()
			if burst {
				fmt.Printf("Detected TEID burst for TEID %d\n", teid)
				ruleId = RuleIdTEIDBruteForce
				break
			}
			if gtp_u_ddos {
				log.Printf("Detected GTPv1U ddos for key %s\n", key.DestIP)
				ruleId = RuleIdGTPv1UBurst
				break
			}
		}
	}

	//判断是否存在DNS协议
	dnsLayer := packet.Layer(layers.LayerTypeDNS)
	if dnsLayer != nil {
		dns, ok := dnsLayer.(*layers.DNS)
		if !ok {
			fmt.Println("Error converting to DNS layer.")
			return
		}

		protocol_application_layer = "DNS"

		// 打印DNS层的所有字段
		fmt.Printf("DNS Layer Information:\n")
		fmt.Printf("ID: %d\n", dns.ID)
		fmt.Printf("QR: %t\n", dns.QR)
		fmt.Printf("Opcode: %d\n", dns.OpCode)
		fmt.Printf("QDCount: %d\n", dns.QDCount)
	}

	//解析ICMP层
	icmpLayer := packet.Layer(layers.LayerTypeICMPv4)
	if icmpLayer != nil {
		icmp, ok := icmpLayer.(*layers.ICMPv4)
		if !ok {
			fmt.Println("Error converting to ICMPv4 layer.")
			return
		}
		// 打印ICMP层的所有字段
		fmt.Printf("ICMP Layer Information:\n")
		fmt.Printf("ICMP Layer Contents (Hex): %02x\n", icmpLayer.LayerContents())
		fmt.Printf("Type: %d\n", icmp.TypeCode.Type())
		fmt.Printf("Code: %d\n", icmp.TypeCode.Code())
		fmt.Printf("Checksum: %d\n", icmp.Checksum)
		switch icmp.TypeCode.Type() {
		case layers.ICMPv4TypeEchoReply:
			fmt.Printf("Identifier: %d\n", icmp.Id)
			fmt.Printf("Sequence: %d\n", icmp.Seq)
		case layers.ICMPv4TypeEchoRequest:
			fmt.Printf("Identifier: %d\n", icmp.Id)
			fmt.Printf("Sequence: %d\n", icmp.Seq)
			fmt.Printf("Data: %02x\n", icmp.Payload)
		}
	}

	//定义一个字段用来标记ngap
	var isNgap bool = false
	//接收ngap层的数据
	var ngapPayload []byte

	// 获取 SCTP 层
	sctpLayer := packet.Layer(layers.LayerTypeSCTP)
	if sctpLayer != nil {
		sctp, ok := sctpLayer.(*layers.SCTP)
		if !ok {
			fmt.Println("Failed to cast SCTP layer.")
			return
		}

		log.Println("SCTP Layer:")
		fmt.Printf("SCTP Layer Contents (Hex): %02x\n", sctpLayer.LayerContents())
		fmt.Printf("  SrcPort: %d\n", sctp.SrcPort)
		fmt.Printf("  DstPort: %d\n", sctp.DstPort)
		fmt.Printf("  VerificationTag: %d\n", sctp.VerificationTag)
		fmt.Printf("  Checksum: %d\n", sctp.Checksum)

		protocol_5g = "SCTP"
		src_port = uint64(sctp.SrcPort)
		dst_port = uint64(sctp.DstPort)

		sctp_verificationTag = uint64(sctp.VerificationTag)
		// 解析 SCTP 块
		for _, layer := range packet.Layers() {

			switch layer.LayerType() {
			case layers.LayerTypeSCTPData:
				sctpData, ok := layer.(*layers.SCTPData)
				if !ok {
					fmt.Println("Failed to cast SCTP Data chunk.")
					continue
				}
				sctp_chunks_type = "SCTP Data Chunk"
				sctp_chunks_flags = fmt.Sprintf("%02x", sctpData.Flags)
				sctp_ie_length = uint64(sctpData.ActualLength)

				fmt.Println("SCTP Data Chunk:")
				fmt.Printf("  TSN: %d\n", sctpData.TSN)
				fmt.Printf("  StreamId: %d\n", sctpData.StreamId)
				fmt.Printf("  StreamSequence: %d\n", sctpData.StreamSequence)
				fmt.Printf("  PayloadProtocol: %s\n", sctpData.PayloadProtocol)
				// 打印数据块的详细信息
				fmt.Printf("  Payload: %v\n", sctpData.Payload)

				// 打印数据块的详细信息
				payloadHex := hex.EncodeToString(sctpData.Payload)
				fmt.Printf("  Payload (hex): %s\n", payloadHex)
				ngapPayload = sctpData.Payload
				isNgap = true // 标记为ngap

				detector_sctpinit.AddEvent(uint8(layers.LayerTypeSCTPData), dst_ip, uint16(dst_port))

				if attackDetected, key := detector_sctpinit.CheckINITFlood(); attackDetected {
					fmt.Printf("SCTP INIT Flood attack detected to IP: %s on port: %d\n", key.DestIP, key.DestPort)
					ruleId = RuleIdSCTPDataFlood
				}

			case layers.LayerTypeSCTPInit:
				sctpInit, ok := layer.(*layers.SCTPInit)
				if !ok {
					fmt.Println("Failed to cast SCTP Init chunk.")
					continue
				}

				detector_sctpinit.AddEvent(uint8(layers.LayerTypeSCTPInit), dst_ip, uint16(dst_port))

				if attackDetected, key := detector_sctpinit.CheckINITFlood(); attackDetected {
					fmt.Printf("SCTP INIT Flood attack detected to IP: %s on port: %d\n", key.DestIP, key.DestPort)
					ruleId = RuleIdSCTPInitFlood
				}

				sctp_chunks_type = "SCTP Init Chunk"
				sctp_chunks_flags = fmt.Sprintf("%02x", sctpInit.Flags)
				sctp_ie_length = uint64(sctpInit.ActualLength)

				fmt.Println("SCTP Init Chunk:")
				fmt.Printf("  Initiate Tag: %d\n", sctpInit.InitiateTag)
				fmt.Printf("  Advertised Receiver Window Credit (a_rwnd): %d\n", sctpInit.AdvertisedReceiverWindowCredit)
				fmt.Printf("  Number of Outbound Streams (OS): %d\n", sctpInit.OutboundStreams)
				fmt.Printf("  Number of Inbound Streams (MIS): %d\n", sctpInit.InboundStreams)
				fmt.Printf("  Initial TSN: %d\n", sctpInit.InitialTSN)

			case layers.LayerTypeSCTPInitAck:
				sctpInitack, ok := layer.(*layers.SCTPInit)
				if !ok {
					fmt.Println("Failed to cast SCTP Init chunk.")
					continue
				}

				detector_sctpinit.AddEvent(uint8(layers.LayerTypeSCTPInitAck), dst_ip, uint16(dst_port))

				if attackDetected, key := detector_sctpinit.CheckINITFlood(); attackDetected {
					fmt.Printf("SCTP INIT Flood attack detected to IP: %s on port: %d\n", key.DestIP, key.DestPort)
					ruleId = RuleIdSCTPInitackFlood
				}

				sctp_chunks_type = "SCTP Initack Chunk"
				sctp_chunks_flags = fmt.Sprintf("%02x", sctpInitack.Flags)
				sctp_ie_length = uint64(sctpInitack.ActualLength)

				fmt.Println("SCTP Initack Chunk:")
				fmt.Printf("  Initiate Tag: %d\n", sctpInitack.InitiateTag)
				fmt.Printf("  Advertised Receiver Window Credit (a_rwnd): %d\n", sctpInitack.AdvertisedReceiverWindowCredit)
				fmt.Printf("  Number of Outbound Streams (OS): %d\n", sctpInitack.OutboundStreams)
				fmt.Printf("  Number of Inbound Streams (MIS): %d\n", sctpInitack.InboundStreams)
				fmt.Printf("  Initial TSN: %d\n", sctpInitack.InitialTSN)

			case layers.LayerTypeSCTPSack:
				sctpSack, ok := layer.(*layers.SCTPSack)
				if !ok {
					fmt.Println("Failed to cast SCTP Sack chunk.")
					continue
				}

				detector_sctpinit.AddEvent(uint8(layers.LayerTypeSCTPSack), dst_ip, uint16(dst_port))

				if attackDetected, key := detector_sctpinit.CheckINITFlood(); attackDetected {
					fmt.Printf("SCTP INIT Flood attack detected to IP: %s on port: %d\n", key.DestIP, key.DestPort)
					ruleId = RuleIdSCTPSackFlood
				}

				sctp_chunks_type = "SCTP Sack Chunk"
				sctp_chunks_flags = fmt.Sprintf("%02x", sctpSack.Flags)
				sctp_ie_length = uint64(sctpSack.ActualLength)

				fmt.Println("SCTP Sack Chunk:")
				fmt.Printf("  Cumulative TSN Ack: %d\n", sctpSack.CumulativeTSNAck)
				fmt.Printf("  Advertised Receiver Window Credit (a_rwnd): %d\n", sctpSack.AdvertisedReceiverWindowCredit)
				fmt.Printf("  Number of Gap ACKs: %d\n", sctpSack.NumGapACKs)
				fmt.Printf("  Number of Duplicate TSNs: %d\n", sctpSack.NumDuplicateTSNs)
				// 打印已接收的数据块的信息
				for _, ack := range sctpSack.GapACKs {
					fmt.Printf("  Gap ACK: TSN=%d\n", ack)
				}
				// 打印间隙的信息
				for _, tsn := range sctpSack.DuplicateTSNs {
					fmt.Printf("  Duplicate TSN: %d\n", tsn)
				}

			case layers.LayerTypeSCTPHeartbeat:
				sctpHeartbeat, ok := layer.(*layers.SCTPHeartbeat)
				if !ok {
					fmt.Println("Failed to cast SCTP Heartbeat chunk.")
					continue
				}

				detector_sctpinit.AddEvent(uint8(layers.LayerTypeSCTPHeartbeat), dst_ip, uint16(dst_port))

				if attackDetected, key := detector_sctpinit.CheckINITFlood(); attackDetected {
					fmt.Printf("SCTP INIT Flood attack detected to IP: %s on port: %d\n", key.DestIP, key.DestPort)
					ruleId = RuleIdSCTPHeartbeatFlood
				}

				sctp_chunks_type = "SCTP Heartbeat Chunk"
				sctp_chunks_flags = fmt.Sprintf("%02x", sctpHeartbeat.Flags)
				sctp_ie_length = uint64(sctpHeartbeat.ActualLength)

				fmt.Println("SCTP Heartbeat Chunk:")
				// 打印 SCTP 心跳块的信息
				for _, param := range sctpHeartbeat.Parameters {
					fmt.Printf("  Parameter Type: %d\n", param.Type)
					fmt.Printf("  Parameter Length: %d\n", param.Length)
					fmt.Printf("  Parameter Value: %v\n", param.Value)
				}

			case layers.LayerTypeSCTPHeartbeatAck:
				sctpHeartbeatAck, ok := layer.(*layers.SCTPHeartbeat)
				if !ok {
					fmt.Println("Failed to cast SCTP Heartbeat Ack chunk.")
					continue
				}

				detector_sctpinit.AddEvent(uint8(layers.LayerTypeSCTPHeartbeatAck), dst_ip, uint16(dst_port))

				if attackDetected, key := detector_sctpinit.CheckINITFlood(); attackDetected {
					fmt.Printf("SCTP INIT Flood attack detected to IP: %s on port: %d\n", key.DestIP, key.DestPort)
					ruleId = RuleIdSCTPHeartbeatAckFlood
				}

				sctp_chunks_type = "SCTP Heartbeat Ack Chunk"
				sctp_chunks_flags = fmt.Sprintf("%02x", sctpHeartbeatAck.Flags)
				sctp_ie_length = uint64(sctpHeartbeatAck.ActualLength)

				fmt.Println("SCTP Heartbeat Ack Chunk:")
				// 打印 SCTP 心跳确认块的信息
				for _, param := range sctpHeartbeatAck.Parameters {
					fmt.Printf("  Parameter Type: %d\n", param.Type)
					fmt.Printf("  Parameter Length: %d\n", param.Length)
					fmt.Printf("  Parameter Value: %v\n", param.Value)
				}

			case layers.LayerTypeSCTPError:
				sctpError, ok := layer.(*layers.SCTPError)
				if !ok {
					log.Println("Failed to cast SCTP Error chunk.")
					continue
				}

				detector_sctpinit.AddEvent(uint8(layers.LayerTypeSCTPError), dst_ip, uint16(dst_port))

				if attackDetected, key := detector_sctpinit.CheckINITFlood(); attackDetected {
					fmt.Printf("SCTP INIT Flood attack detected to IP: %s on port: %d\n", key.DestIP, key.DestPort)
					ruleId = RuleIdSCTPErrorFlood
				}

				sctp_chunks_type = "SCTP Error Chunk"
				sctp_chunks_flags = fmt.Sprintf("%02x", sctpError.Flags)
				sctp_ie_length = uint64(sctpError.ActualLength)

				fmt.Println("SCTP Error Chunk:")
				// 打印 SCTP 错误块的信息
				for _, param := range sctpError.Parameters {
					fmt.Printf("  Parameter Type: %d\n", param.Type)
					fmt.Printf("  Parameter Length: %d\n", param.Length)
					fmt.Printf("  Parameter Value: %v\n", param.Value)
				}

			case layers.LayerTypeSCTPAbort:
				sctpAbort, ok := layer.(*layers.SCTPError)
				if !ok {
					fmt.Println("Failed to cast SCTP Abort chunk.")
					continue
				}

				sctp_chunks_type = "SCTP Abort Chunk"
				sctp_chunks_flags = fmt.Sprintf("%02x", sctpAbort.Flags)
				sctp_ie_length = uint64(sctpAbort.ActualLength)

				fmt.Println("SCTP Abort Chunk:")
				// 打印 SCTP 终止块的信息
				for _, param := range sctpAbort.Parameters {
					fmt.Printf("  Parameter Type: %d\n", param.Type)
					fmt.Printf("  Parameter Length: %d\n", param.Length)
					fmt.Printf("  Parameter Value: %v\n", param.Value)
				}

			case layers.LayerTypeSCTPShutdown:
				sctpShutdown, ok := layer.(*layers.SCTPShutdown)
				if !ok {
					fmt.Println("Failed to cast SCTP Shutdown chunk.")
					continue
				}

				detector_sctpinit.AddEvent(uint8(layers.LayerTypeSCTPShutdown), dst_ip, uint16(dst_port))

				if attackDetected, key := detector_sctpinit.CheckINITFlood(); attackDetected {
					fmt.Printf("SCTP INIT Flood attack detected to IP: %s on port: %d\n", key.DestIP, key.DestPort)
					ruleId = RuleIdSCTPShutdownFlood
				}

				sctp_chunks_type = "SCTP Shutdown Chunk"
				sctp_chunks_flags = fmt.Sprintf("%02x", sctpShutdown.Flags)
				sctp_ie_length = uint64(sctpShutdown.ActualLength)
				fmt.Println("SCTP Shutdown Chunk:")
				fmt.Printf("  Cumulative TSN Ack: %d\n", sctpShutdown.CumulativeTSNAck)

			case layers.LayerTypeSCTPShutdownAck:
				sctpShutdownAck, ok := layer.(*layers.SCTPShutdownAck)
				if !ok {
					fmt.Println("Failed to cast SCTP Shutdown Ack chunk.")
					continue
				}

				detector_sctpinit.AddEvent(uint8(layers.LayerTypeSCTPShutdownAck), dst_ip, uint16(dst_port))

				if attackDetected, key := detector_sctpinit.CheckINITFlood(); attackDetected {
					fmt.Printf("SCTP INIT Flood attack detected to IP: %s on port: %d\n", key.DestIP, key.DestPort)
					ruleId = RuleIdSCTPShutdownAckFlood
				}

				sctp_chunks_type = "SCTP Shutdown Ack Chunk"
				sctp_chunks_flags = fmt.Sprintf("%02x", sctpShutdownAck.Flags)
				sctp_ie_length = uint64(sctpShutdownAck.ActualLength)

			case layers.LayerTypeSCTPCookieEcho:
				sctpCookieEcho, ok := layer.(*layers.SCTPCookieEcho)
				if !ok {
					fmt.Println("Failed to cast SCTP Cookie Echo chunk.")
					continue
				}

				detector_sctpinit.AddEvent(uint8(layers.LayerTypeSCTPCookieEcho), dst_ip, uint16(dst_port))

				if attackDetected, key := detector_sctpinit.CheckINITFlood(); attackDetected {
					fmt.Printf("SCTP INIT Flood attack detected to IP: %s on port: %d\n", key.DestIP, key.DestPort)
					ruleId = RuleIdSCTPCookieEchoFlood
				}

				sctp_chunks_type = "SCTP Cookie Echo Chunk"
				sctp_chunks_flags = fmt.Sprintf("%02x", sctpCookieEcho.Flags)
				sctp_ie_length = uint64(sctpCookieEcho.ActualLength)

				fmt.Println("SCTP Cookie Echo Chunk:")
				// 打印 SCTP Cookie Echo 块的信息
				fmt.Printf("  Cookie: %v\n", sctpCookieEcho.Cookie)

			case layers.LayerTypeSCTPCookieAck:
				sctpCookieAck, ok := layer.(*layers.SCTPEmptyLayer)
				if !ok {
					fmt.Println("Failed to cast SCTP Cookie Ack chunk.")
					continue
				}

				detector_sctpinit.AddEvent(uint8(layers.LayerTypeSCTPCookieAck), dst_ip, uint16(dst_port))

				if attackDetected, key := detector_sctpinit.CheckINITFlood(); attackDetected {
					fmt.Printf("SCTP INIT Flood attack detected to IP: %s on port: %d\n", key.DestIP, key.DestPort)
					ruleId = RuleIdSCTPCookieAckFlood
				}

				sctp_chunks_type = "SCTP Cookie Ack Chunk"
				sctp_chunks_flags = fmt.Sprintf("%02x", sctpCookieAck.Flags)
				sctp_ie_length = uint64(sctpCookieAck.ActualLength)

			case layers.LayerTypeSCTPShutdownComplete:
				sctpShutdownComplete, ok := layer.(*layers.SCTPEmptyLayer)
				if !ok {
					fmt.Println("Failed to cast SCTP Shutdown Complete chunk.")
					continue
				}

				detector_sctpinit.AddEvent(uint8(layers.LayerTypeSCTPShutdownComplete), dst_ip, uint16(dst_port))

				if attackDetected, key := detector_sctpinit.CheckINITFlood(); attackDetected {
					fmt.Printf("SCTP INIT Flood attack detected to IP: %s on port: %d\n", key.DestIP, key.DestPort)
					ruleId = RuleIdSCTPShutdownCompleteFlood
				}

				sctp_chunks_type = "SCTP Shutdown Complete Chunk"
				sctp_chunks_flags = fmt.Sprintf("%02x", sctpShutdownComplete.Flags)
				sctp_ie_length = uint64(sctpShutdownComplete.ActualLength)

			default:
				fmt.Printf("Unknown SCTP chunk type: %s\n", layer.LayerType())
			}
		}
	} else {
		fmt.Println("No SCTP layer found.")
	}

	//ngap协议解析
	var transportPayload []byte
	if udpLayer != nil {
		udp, _ := udpLayer.(*layers.UDP)
		transportPayload = udp.Payload
	} else if isNgap == true {
		transportPayload = ngapPayload //sctp data payload
	}

	//如果保护恶意字符，如%*//<>等

	if len(transportPayload) > 0 {
		fmt.Println("NGAP Message:")
		// 解析NGAP消息
		ngapPdu, err := parseNGAPMessage(transportPayload)

		if err != nil {
			log.Printf("Failed to parse NGAP message: %v", err)
			return
		}
		if protocol_5g == "" {
			protocol_5g = "NGAP"
		}
		if protocol_5g == "SCTP" {
			protocol_5g = "NGAP"
		}

		if len(transportPayload) > 300 {
			ruleId = RuleIdNGAPUnknown
		}

		// 输出字段的值
		//printNGAPFieldValues(ngapPdu, "")
		//如何ngapPdu的date存在恶意字符，则告警

		// 输出NGAP字段的值
		switch ngapPdu.Present {
		case ngapType.NGAPPDUPresentInitiatingMessage:
			// 获取ngap message type
			ngap_message_type = "InitiatingMessage"

			procedureCode := ngapPdu.InitiatingMessage.ProcedureCode.Value

			NgapprocedureCode, ok := detector.ProcedureCodeToStringMap[int(procedureCode)]
			if ok {
				fmt.Println("ProcedureCode:", NgapprocedureCode)
			} else {
				fmt.Println("Unknown ProcedureCode:", NgapprocedureCode)
			}

			// 获取Criticality
			criticality := ngapPdu.InitiatingMessage.Criticality.Value
			ngap_criticlity, ok := detector.CriticalityToStringMap[int(criticality)]
			if ok {
				fmt.Println("Criticality:", ngap_criticlity)
			} else {
				fmt.Println("Unknown Criticality:", ngap_criticlity)
			}

			if int(ngapType.ProcedureCodeUEContextReleaseRequest) != 0 {
				// 进一步检查是否为UEContextReleaseReqUEst消息
				// 这里需要根据你的协议定义来检查具体的消息类型
				// 例如，可能需要检查InitiatingMessage的另一个字段，如MessageType
				ruleId = RuleIdUEContextReleaseRequest
			}

		case ngapType.NGAPPDUPresentSuccessfulOutcome:
			// 获取ngap message type
			ngap_message_type = "SuccessfulOutcome"

			// 获取ProcedureCode
			procedureCode := ngapPdu.SuccessfulOutcome.ProcedureCode.Value

			NgapprocedureCode, ok := detector.ProcedureCodeToStringMap[int(procedureCode)]
			if ok {
				fmt.Println("ProcedureCode:", NgapprocedureCode)
			} else {
				fmt.Println("Unknown ProcedureCode:", NgapprocedureCode)
			}

			// 获取Criticality
			criticality := ngapPdu.SuccessfulOutcome.Criticality.Value
			ngap_criticlity, ok := detector.CriticalityToStringMap[int(criticality)]
			if ok {
				fmt.Println("Criticality:", ngap_criticlity)
			} else {
				fmt.Println("Unknown Criticality:", ngap_criticlity)
			}

		case ngapType.NGAPPDUPresentUnsuccessfulOutcome:
			// 获取ngap message type
			ngap_message_type = "UnsuccessfulOutcome"

			// 获取ProcedureCode
			procedureCode := ngapPdu.UnsuccessfulOutcome.ProcedureCode.Value

			NgapprocedureCode, ok := detector.ProcedureCodeToStringMap[int(procedureCode)]
			if ok {
				fmt.Println("ProcedureCode:", NgapprocedureCode)
			} else {
				fmt.Println("Unknown ProcedureCode:", NgapprocedureCode)
			}

			// 获取Criticality
			criticality := ngapPdu.UnsuccessfulOutcome.Criticality.Value
			ngap_criticlity, ok := detector.CriticalityToStringMap[int(criticality)]
			if ok {
				fmt.Println("Criticality:", ngap_criticlity)
			} else {
				fmt.Println("Unknown Criticality:", ngap_criticlity)
			}

		default:
			fmt.Println("Unknown NGAP PDU type")
			ruleId = RuleIdNGAPUnknown
		}
	}

	// Check for errors
	if err := packet.ErrorLayer(); err != nil {
		fmt.Println("Error decoding some part of the packet:", err)
	}
}
