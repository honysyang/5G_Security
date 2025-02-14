#include <pfring.h>
#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>
#include <signal.h>
#include "nsengine.h"
#include <hiredis/hiredis.h>

#define NS_RET_ERROR  -1
#define NS_RET_SUCCESS  0

#define THREAD_CNT 8

volatile int running = 1;


#define HOSTNAME "10.1.200.14"
#define POST 6379


#include <AIS_clickhouse.h>
void* AIS_Clickhouse;
struct Network_ NetworkFive;

redisContext *redis_conn;

static int batch_count = 0;
static NetworkFive batch[100]


static char g_ns_conf_path[512] = ".//ns_engine/conf/netstack.yaml";
static char g_ns_rule_path[512] = "./nsc$1000.309";
static char g_ns_custom_rule_path[512] = "./custom_rule";

uint64_t rx_packet_total[128] = {0};



// 定义连接池结构
typedef struct {
    redisContext **connections;
    int count;
    int current_index;
    pthread_mutex_t mutex;
} RedisConnectionPool;

// 连接池实例
RedisConnectionPool* redis_pool;



void signal_handler(int signum) {
    running = 0;
}

// 创建 Redis 连接池
RedisConnectionPool* create_redis_connection_pool(const char* hostname, int port, int pool_size) {
    RedisConnectionPool* pool = (RedisConnectionPool*)malloc(sizeof(RedisConnectionPool));
    pool->count = pool_size;
    pool->current_index = 0;
    pool->connections = (redisContext**)malloc(sizeof(redisContext*) * pool_size);
    pthread_mutex_init(&pool->mutex, NULL);
    for (int i = 0; i < pool_size; i++) {
        redisContext *c = redisConnect(hostname, port);
        if (c == NULL || c->err) {
            if (c) {
                fprintf(stderr, "Redis connection failed: %s\n", c->errstr);
                redisFree(c);
                free(pool);
                return NULL;
            }
        }
        pool->connections[i] = c;
    }
    return pool;
}


// 从连接池获取连接
redisContext* get_redis_connection(RedisConnectionPool* pool) {
    pthread_mutex_lock(&pool->mutex);
    redisContext* conn = pool->connections[pool->current_index];
    pool->current_index = (pool->current_index + 1) % pool->count;
    pthread_mutex_unlock(&pool->mutex);
    return conn;
}


// 释放 Redis 连接池
void destroy_redis_connection_pool(RedisConnectionPool* pool) {
    for (int i = 0; i < pool->count; i++) {
        redisFree(pool->connections[i]);
    }
    free(pool->connections);
    pthread_mutex_destroy(&pool->mutex);
    free(pool);
}

// 连接 Redis 并添加到连接池
redisContext* connect_redis(RedisConnectionPool* pool) {
    return get_redis_connection(pool);
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
	pthread_mutex_lock(&mux);
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

		batch[batch_count++] = NetworkFive;
		if (batch_count > 0){	
			InsertClickhouse(AIS_Clickhouse, (char *)&NetworkFive, 1);
		}
	}
	pthread_mutex_unlock(&mux);

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
    char *device = "ens33"; // 你想要捕获数据包的网络接口
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
	
	redisContext* redis_conn = connect_redis(redis_pool);

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
            printf("thread:%d, packet:%lu\n", thread_id, rx_packet_total[thread_id]);

					
			// 使用 Redis 管道
            redisReply *reply;
            redisAppendCommand(redis_conn, "PUBLISH packet_channel %b", ns_packet.data, ns_packet.len);
            if (redisGetReply(redis_conn, (void**)&reply) == REDIS_OK) {
                if (reply) {
                    freeReplyObject(reply);
                }
            }
					
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

	// 创建 Redis 连接池
    redis_pool = create_redis_connection_pool(HOSTNAME, POST, 10);
  	
  	/*****************************clickhouse***********************************/

		AIS_Clickhouse = DatabaseInit();
    if(AIS_Clickhouse != NULL){
		  printf("databaseInit success! \n");
	  }
	  DatabaseCreate(AIS_Clickhouse);

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

	destroy_redis_connection_pool(redis_pool);

    
    return 0;
}
