#include <pfring.h>
#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>
#include <signal.h>
#include "nsengine.h"
#include <hiredis/hiredis.h>

#define NS_RET_ERROR  -1
#define NS_RET_SUCCESS  0

#define THREAD_CNT 64

#include <curl/curl.h>
#include <arpa/inet.h>


volatile int running = 1;


#define HOSTNAME "127.0.0.1"
#define POST 6379


#include <AIS_clickhouse.h>
void* AIS_Clickhouse;
struct Network_ NetworkFive;

redisContext *redis_conn;



static char g_ns_conf_path[512] = ".//ns_engine/conf/netstack.yaml";
static char g_ns_rule_path[512] = "./nsc$1000.309";
static char g_ns_custom_rule_path[512] = "./custom_rule";

uint64_t rx_packet_total[128] = {0};

void signal_handler(int signum) {
    running = 0;
}

redisContext* connect_redis() {

	redis_conn = redisConnect(HOSTNAME, POST);
	if (redis_conn == NULL) {
			fprintf(stderr, "Redis connection failed");
			exit(1);
   	}
   	if (redis_conn->err) {
       fprintf(stderr, "Redis connection error: %s\n", redis_conn->errstr);
       exit(1);
   	}

		return redis_conn;

}


// 发送 Redis 请求的 API 函数
int sendRedisRequest(const char *packet, size_t packetLen) {
    CURL *curl;
    CURLcode res;
    struct curl_slist *headers = NULL;

	const char *url = "http://127.0.0.1:8081/receive";
    // 初始化 CURL 句柄
    curl = curl_easy_init();
    if (curl == NULL) {
        fprintf(stderr, "Failed to initialize curl for Redis request\n");
        return 1;
    }


    // 设置 CURL 选项
    curl_easy_setopt(curl, CURLOPT_PROXY, "");
    curl_easy_setopt(curl, CURLOPT_URL, url);
    curl_easy_setopt(curl, CURLOPT_POST, 1L);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, packet);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDSIZE, packetLen);
    headers = curl_slist_append(headers, "Content-Type: application/octet-stream");
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);


    // 执行请求
    res = curl_easy_perform(curl);


    // 检查请求是否成功
    if (res!= CURLE_OK) {
        fprintf(stderr, "redis \t curl_easy_perform() failed: %s\n", curl_easy_strerror(res));
    } else {
        long http_code = 0;
        curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &http_code);
        if (http_code >= 200 && http_code < 300) {
            printf("Data sent to Redis successfully.\n");
        } else {
            fprintf(stderr, "HTTP request to Redis failed with status code %ld\n", http_code);
        }
    }


    // 清理资源
    curl_slist_free_all(headers);
    curl_easy_cleanup(curl);

}



// 发送 ClickHouse 请求的 API 函数
int sendClickHouseRequest(const char *json_str) {
    CURL *curl;
    CURLcode res;
    struct curl_slist *headers = NULL;
	
	const char *url = "http://127.0.0.1:8080/insert";

    // 初始化 CURL 句柄
    curl = curl_easy_init();
    if (curl == NULL) {
        fprintf(stderr, "Failed to initialize curl for ClickHouse request\n");
        return 1;
    }


    // 设置 CURL 选项
    curl_easy_setopt(curl, CURLOPT_PROXY, "");
    curl_easy_setopt(curl, CURLOPT_URL, url);
    curl_easy_setopt(curl, CURLOPT_POST, 1L);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json_str);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDSIZE, strlen(json_str));
    headers = curl_slist_append(headers, "Content-Type: application/json");
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
    curl_easy_setopt(curl, CURLOPT_VERBOSE, 1L);


    // 执行请求
    res = curl_easy_perform(curl);


    // 检查请求是否成功
    if (res!= CURLE_OK) {
        fprintf(stderr, "ClickHouse \t curl_easy_perform() failed: %s\n", curl_easy_strerror(res));
    } else {
        long http_code = 0;
        curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &http_code);
        if (http_code >= 200 && http_code < 300) {
            printf("Data sent to ClickHouse successfully.\n");
        } else {
            fprintf(stderr, "HTTP request to ClickHouse failed with status code %ld\n", http_code);
        }
    }


    // 清理资源
    curl_slist_free_all(headers);
    curl_easy_cleanup(curl);


}




static int32_t loadNetStackRules(NSCTX netstack)
{
    int32_t ret = 0;
    if (strlen(g_ns_rule_path) > 0) {
        ret = NSLoadPattern(netstack, g_ns_rule_path);
        if(NS_SUCCESS != ret) {
            printf("NS_LOAD_RULES failed");
            return ret;
        }
        char version[20] ={0};
        NSGetPatternVersionByFile(g_ns_rule_path, version, 20);
        printf("Netstack pattern version:%s \n", version);
    }
    if (strlen(g_ns_custom_rule_path) > 0) {
        ret = NSLoadUserPattern(netstack, g_ns_custom_rule_path);
        if(NS_SUCCESS != ret) {
            printf("NS_LOAD CUSTOM RULES failed\n");
            return ret;
        }
    }
    ret = NSBuildPattern(netstack);
    if(NS_SUCCESS != ret) {
        printf("NSBuildPattern failed");
        return ret;
    }
    return ret;
}

static void netstack_scan_callback(NSAlertInfo *data, void* user_handle)
{
/*********************************Clickhouse*******************************************/
	int calc = 1;
	if (data->rule.sid > 0){
		printf("\n---start insert clickhouse\n");

		int Src_ip = data->src.address_un_data8[3]<<24|data->src.address_un_data8[2]<<16|data->src.address_un_data8[1]<<8|data->src.address_un_data8[0];
		int Dst_ip = data->dst.address_un_data8[3]<<24|data->dst.address_un_data8[2]<<16|data->dst.address_un_data8[1]<<8|data->dst.address_un_data8[0];
			
		NetworkFive.id = calc;
		NetworkFive.protocol = data->ip_proto;

		NetworkFive.src_port = data->sp;
		NetworkFive.dst_port = data->dp;

		NetworkFive.src_ip = Src_ip;
		NetworkFive.dst_ip = Dst_ip;

		NetworkFive.rule_sid = data->rule.sid;
		NetworkFive.byte = data->buflen;

		
        //nsertClickhouse(AIS_Clickhouse, &NetworkFive, 1);
		
		char json_str[512];
        snprintf(json_str, sizeof(json_str),
                 "{\"id\":%llu,\"src_ip\":\"%s\",\"dst_ip\":\"%s\",\"src_port\":%hu,\"dst_port\":%hu,\"protocol\":%hhu,\"rule_sid\":%llu,\"byte\":%llu}",
                 NetworkFive.id,
                 inet_ntoa(*(struct in_addr *)&NetworkFive.src_ip),
                 inet_ntoa(*(struct in_addr *)&NetworkFive.dst_ip),
                 NetworkFive.src_port,
                 NetworkFive.dst_port,
                 NetworkFive.protocol,
                 NetworkFive.rule_sid,
                 NetworkFive.byte);

		
		// 调用 API 函数发送 ClickHouse 请求

		sendClickHouseRequest(json_str);
			

/*********************************Clickhouse*******************************************/


	   printf("%u.%u.%u.%u -> %u.%u.%u.%u -> alert sid: %u\n",
			  data->src.address_un_data8[0], data->src.address_un_data8[1],
			  data->src.address_un_data8[2], data->src.address_un_data8[3],
			  data->dst.address_un_data8[0], data->dst.address_un_data8[1],
			  data->dst.address_un_data8[2], data->dst.address_un_data8[3],
			  data->rule.sid);
		if (1 == data->ip_proto) //ICMP协议
		{
			printf("ICMP Type:%u Code:%u ", data->icmp_s.type, data->icmp_s.code);
			printf("%u.%u.%u.%u->", data->src.address_un_data8[0], data->src.address_un_data8[1], data->src.address_un_data8[2], data->src.address_un_data8[3]);
			printf("%u.%u.%u.%u->", data->dst.address_un_data8[0], data->dst.address_un_data8[1], data->dst.address_un_data8[2], data->dst.address_un_data8[3]);
		}
	  else
		{
				
			printf("%s ", data->ip_proto == 6 ? "TCP" : (data->ip_proto == 17 ? "UDP" : "-"));
			printf("%u.%u.%u.%u:%u->", data->src.address_un_data8[0], data->src.address_un_data8[1], data->src.address_un_data8[2], data->src.address_un_data8[3], data->sp);
			printf("%u.%u.%u.%u:%u->", data->dst.address_un_data8[0], data->dst.address_un_data8[1], data->dst.address_un_data8[2], data->dst.address_un_data8[3], data->dp);
		}
	}
}

static int32_t initNS(void **ns_dev, int thread_type)
{
    int32_t ret = NSInit(ns_dev);

    if(NS_RET_SUCCESS != ret){
        printf("NSInit error: %d\n", NSStrError(ret));
        ns_dev = NULL;
        return NS_RET_ERROR;
    }

    ret = loadNetStackRules(*ns_dev);
    if(NS_RET_SUCCESS != ret){
        printf("NSLoadPattern error: %d\n", NSStrError(ret));
        return NS_RET_ERROR;
    }

    ret = NSSetAlertCallBackFunc(*ns_dev, netstack_scan_callback);
    if(NS_RET_SUCCESS != ret){
        printf("NSSetCallback error: %d\n", NSStrError(ret));
        return NS_RET_ERROR;
    }

    ret = NSSetConfig(*ns_dev, NS_CONFIG_LINK_TYPE, LINKTYPE_ETHERNET);
    if(ret != NS_SUCCESS){
        printf("NSSetConfig error: %d line:%d\n", NSStrError(ret), __LINE__);
        return NS_RET_ERROR;
    }

    ret = NSSetConfig(*ns_dev, NS_CONFIG_FAST_MODE, 1);
    if(ret != NS_SUCCESS){
        printf("NSSetConfig error: %d line:%d\n", NSStrError(ret), __LINE__);
        return NS_RET_ERROR;
    }

    /* Thread type ---> ct.ct_type
     * offline mode: 0
     * pcapfile mode: 1
     */
    ret = NSSetConfig(*ns_dev, NS_CONFIG_THREAD_TYPE, thread_type);
    if(ret != NS_SUCCESS){
        printf("NSSetConfig error: %d line:%d\n", NSStrError(ret), __LINE__);
        return NS_RET_ERROR;
    }

    ret = NSSetConfig(*ns_dev, NS_CONFIG_TCP_ONEWAY_BYTE, 10000);
    if(ret != NS_SUCCESS){
        printf("NSSetConfig error: %d line:%d\n", NSStrError(ret), __LINE__);
        return NS_RET_ERROR;
    }
    printf("SET TCP_ONEWAY_SCAN_SIZE:%d\n", 10000);

    return NS_RET_SUCCESS;
}

void* packet_capture_thread(void* arg) {
    int thread_id = *((int*)arg);
    void* ns_dev;
    char *device = "em3"; // 你想要捕获数据包的网络接口
    u_int32_t flags = PF_RING_PROMISC;

    // 打开网络接口
    pfring *ring = pfring_open(device, 1500 /* snaplen */, flags);
    if (ring == NULL) {
        fprintf(stderr, "Error opening device %s\n", device);
        return NULL;
    }
    // 设置信号处理器以便在按下 Ctrl+C 时停止捕获
    pfring_set_direction(ring, (packet_direction)rx_and_tx_direction);
    pfring_set_cluster(ring, 99, (cluster_type)cluster_per_inner_flow_2_tuple);
    pfring_set_poll_watermark(ring, 1);

    // 开始捕获数据包
    if (pfring_enable_ring(ring) != 0) {
        fprintf(stderr, "Error enabling ring\n");
        pfring_close(ring);
        return NULL;
    }

    printf("Capturing packets on %s\n", device);


    if (NS_RET_SUCCESS != initNS(&ns_dev, 1)){
	    printf("init ns error.\n");
        return NULL;
    }
    
    struct pfring_pkthdr hdr;
    uint8_t *packet = NULL;
    struct timespec sleep_ts;
    sleep_ts.tv_sec = 0;
    sleep_ts.tv_nsec = 200;
    int ret;
    uint64_t no_pkt_count = 0;

    while (running) {
        // 捕获数据包
        ret = pfring_recv(ring, &packet, 0, &hdr, 1);
        if (ret == 1) {
            struct timeval tv = {0,0};
            gettimeofday(&tv, NULL);
            NSPacket ns_packet;
            ns_packet.len = hdr.caplen;
            ns_packet.ts.tv_sec = tv.tv_sec;
            ns_packet.ts.tv_usec = tv.tv_usec;
            ns_packet.data = packet;
            ns_packet.user_handle =NULL;


            NSScanPacket(ns_dev, &ns_packet);
            rx_packet_total[thread_id]++;
            //printf("thread:%d, packet:%lu\n", thread_id, rx_packet_total[thread_id]);
			/*no_pkt_count++;
			printf("packet:%lu\n",no_pkt_count);


			*/
	
			

			// 调用 API 函数发送 Redis 请求
			sendRedisRequest(packet, hdr.caplen);

        } 
    }

    // 关闭网络接口
    pfring_close(ring);
    NSQuit(ns_dev);
    return NULL;
}

static int32_t loadGlobalNetStackRules()
{
    int32_t ret = 0;
    if (strlen(g_ns_rule_path) > 0) {
        ret = NSGlobalLoadPattern(g_ns_rule_path);
        if(NS_SUCCESS != ret) {
            printf("NS_LOAD_RULES failed\n");
            return ret;
        }
        char version[20] ={0};
        NSGetPatternVersionByFile(g_ns_rule_path, version, 20);
        printf("Netstack pattern version:%s\n", version);
    }
    if (strlen(g_ns_custom_rule_path) > 0) {
        ret = NSGlobalLoadUserPattern(g_ns_custom_rule_path);
        if(NS_SUCCESS != ret) {
            printf("NS_LOAD CUSTOM RULES failed\n");
            return ret;
        }
    }
    ret = NSGlobalBuildPattern();
    if(NS_SUCCESS != ret) {
        printf("NSBuildPattern failed\n");
        return ret;
    }
    return ret;
}

void init_ns_global()
{
    int i;

    if (NSLoadYamlConfig(g_ns_conf_path) != NS_SUCCESS)
    {
        printf("Load yaml error");
    }
    NSSetEngineAlertMode(ENGINE_MODE_ALERT_FLOW);
    if (NS_SUCCESS != NSFeedbackSwapDirPath("/opt", 0))
    {
        printf("Init sfb dir path error.");
    }
    NSSetGlobalConfig(NS_CONFIG_SU_DIR, "/opt", 256);
    NSGlobalInit();

	NSSetRuleIp(HOME_NET, "any");

    if (NS_SUCCESS != NSSetRuleIp(HOME_NET, "any")) {
        printf("IDS init any addr fail.");
    }


    loadGlobalNetStackRules();

    //NSSETRuleIp must set before NSGlobalLoadPattern and after NSGlobalInit
    //ruleip any did not work at this place
    //NSSetRuleIp(HOME_NET, (char*)"any");
    //NSSetRuleIp(EXTERNAL_NET, (char*)"any");
    NSSetThreadNumber(THREAD_CNT);
}

int main() {

    init_ns_global();

		redisContext* redis_conn = connect_redis();

		redisEnableKeepAlive(redis_conn);
  	
  	/*****************************clickhouse***********************************/


    // 创建多个线程来捕获数据包
    int num_threads = THREAD_CNT; // 你可以根据需要调整线程数量
    pthread_t threads[num_threads];
    int thread_ids[num_threads];
    for (int i = 0; i < num_threads; ++i) {
        thread_ids[i] = i;
        pthread_create(&threads[i], NULL, packet_capture_thread, &thread_ids[i]);
    }

    // 等待所有线程完成
    for (int i = 0; i < num_threads; ++i) {
        pthread_join(threads[i], NULL);
    }

		redisFree(redis_conn);


    
    return 0;
}
