package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

// NetworkData 是我们期望从客户端接收到的 JSON 数据结构
type NetworkData struct {
	ID            uint64    `json:"id"`
	SrcIP         string    `json:"src_ip"` // IPv4 地址
	DstIP         string    `json:"dst_ip"` // IPv4 地址
	SrcPort       uint16    `json:"src_port"`
	DstPort       uint16    `json:"dst_port"`
	RuleSID       uint64    `json:"rule_sid"`
	Timestamp     time.Time `json:"timestamp,omitempty"`      // 可选字段，默认为当前时间
	Status        uint8     `json:"status,omitempty"`         // 流量状态,0表示allow,1表示alert,2表示block
	DisposeStatus uint8     `json:"dispose_status,omitempty"` // 可选字段，默认为 1
}

// MySQL 连接信息
var (
	mysqlDSN = "root:Gsycl3541@tcp(10.10.50.107:3306)/ai_getway?parseTime=true"
)

// db 是全局变量，用于存储 MySQL 的数据库连接
var db *sql.DB
var once sync.Once

// 定义一个通道，用于接收需要插入的数据
var dataQueue chan NetworkData
var wg sync.WaitGroup

// GetDB 返回 MySQL 数据库的单例连接
func GetDB() (*sql.DB, error) {
	var err error
	once.Do(func() {
		if mysqlDSN == "" {
			log.Fatalf("MySQL DSN is not set in environment variable MYSQL_DSN")
		}
		log.Println("Initializing MySQL database connection...")
		db, err = sql.Open("mysql", mysqlDSN)
		if err != nil {
			log.Fatalf("Failed to connect to MySQL: %v", err)
		}
		// 配置连接池
		db.SetMaxOpenConns(50)                 // 设置最大打开连接数
		db.SetMaxIdleConns(25)                 // 设置最大空闲连接数
		db.SetConnMaxLifetime(5 * time.Minute) // 设置连接的最大生命周期
		log.Println("MySQL database connection initialized successfully.")
	})
	if db == nil {
		return nil, fmt.Errorf("failed to initialize database connection")
	}
	return db, nil
}

// batchInsertData 批量插入数据到 MySQL
func batchInsertData(dataList []NetworkData) error {
	if len(dataList) == 0 {
		log.Println("No data to insert.")
		return nil
	}

	log.Printf("Preparing to insert batch of %d records...", len(dataList))

	// 构造批量插入的 SQL 语句
	query := `
        INSERT INTO flow_log (system_create_time, system_update_time, system_status, id, src_ip, src_port, dst_ip, dst_port, rule_id, time, status, dispose_status)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	// 批量插入
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		log.Printf("Failed to prepare statement: %v", err)
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	
	for _, data := range dataList {

		ruleIDStr := strconv.FormatUint(data.RuleSID, 10)   //将rule_id转换为字符串类型

		timestamp := data.Timestamp
		if timestamp.IsZero() {
			timestamp = time.Now()
		}

		status := data.Status
		if status == 0 {
			status = 1 // 默认设置为 alert
		}

		disposeStatus := data.DisposeStatus
		if disposeStatus == 0 {
			disposeStatus = 1 // 默认设置为已处置
		}

		system_update_time := time.Now()
		system_create_time := time.Now()
		system_status := 1

		_, err := stmt.Exec(
			system_update_time,
			system_create_time,
			system_status,
			data.ID,
			data.SrcIP,
			data.SrcPort,
			data.DstIP,
			data.DstPort,
			ruleIDStr,
			timestamp,
			status,
			disposeStatus,
		)
		if err != nil {
			tx.Rollback()
			log.Printf("Failed to execute statement for data ID %d: %v", data.ID, err)
			return fmt.Errorf("failed to execute statement for data ID %d: %v", data.ID, err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		log.Printf("Failed to commit transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("Successfully inserted batch of %d records.", len(dataList))
	return nil
}

// worker 是一个后台 goroutine，负责从队列中取出数据并批量插入到数据库
func worker(id int, queue <-chan NetworkData, done chan<- bool) {
	defer wg.Done()
	batchSize := 100
	flushInterval := 5 * time.Second // 定时触发批量插入的时间间隔
	var dataList []NetworkData

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case data := <-queue:
			log.Printf("Worker %d received data ID %d from queue", id, data.ID)
			dataList = append(dataList, data)
			if len(dataList) >= batchSize {
				log.Printf("Worker %d: Batch size reached, inserting %d records...", id, len(dataList))
				err := batchInsertData(dataList)
				if err != nil {
					log.Printf("Worker %d failed to insert data: %v", id, err)
				} else {
					log.Printf("Worker %d successfully inserted %d records", id, len(dataList))
				}
				dataList = nil
			}
		case <-ticker.C:
			if len(dataList) > 0 {
				log.Printf("Worker %d: Timer triggered, inserting %d records...", id, len(dataList))
				err := batchInsertData(dataList)
				if err != nil {
					log.Printf("Worker %d failed to insert timed-out data: %v", id, err)
				} else {
					log.Printf("Worker %d successfully inserted %d records", id, len(dataList))
				}
				dataList = nil
			}
		}
	}
}

// handleInsert 处理 /insert 路由的 POST 请求
func handleInsert(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Printf("Received non-POST request on /insert from %s", r.RemoteAddr)
		return
	}

	var data NetworkData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		log.Printf("Invalid JSON payload from %s: %v", r.RemoteAddr, err)
		return
	}
	defer r.Body.Close()

	log.Printf("Received request from %s: %+v", r.RemoteAddr, data)

	// 将数据放入队列
	select {
	case dataQueue <- data:
		// 数据成功放入队列
		log.Printf("Data queued for insertion: %+v", data)
	default:
		http.Error(w, "Queue is full, try again later", http.StatusServiceUnavailable)
		log.Printf("Queue is full, rejected request from %s", r.RemoteAddr)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Data received and queued for insertion.\n")

	log.Printf("Request processed in %v", time.Since(startTime))
}

// healthCheck 处理 /health 路由的 GET 请求
func healthCheck(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Printf("Received non-GET request on /health from %s", r.RemoteAddr)
		return
	}

	// 检查数据库连接是否正常
	err := db.Ping()
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		log.Printf("Health check failed: %v", err)
		return
	}

	// 返回健康状态
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Server is healthy.\n")

	log.Printf("Health check passed in %v", time.Since(startTime))
}

func main() {
	// 初始化 MySQL 数据库连接
	db, err := GetDB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	defer db.Close()

	// 初始化数据队列和工作 goroutine
	dataQueue = make(chan NetworkData, 10000) // 队列容量为 10000
	numWorkers := 4                           // 启动 4 个工作 goroutine
	done := make(chan bool, numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(id, dataQueue, done)
		}(i)
	}

	// 注册路由
	http.HandleFunc("/insert", handleInsert)
	http.HandleFunc("/health", healthCheck)

	// 启动 HTTP 服务器
	port := "0.0.0.0:8080"
	log.Printf("Starting server on port %s", port)
	server := &http.Server{
		Addr:              port,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待所有工作 goroutine 完成
	wg.Wait()
	close(done)

	log.Println("All workers have finished. Shutting down server gracefully.")

	// 关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}
