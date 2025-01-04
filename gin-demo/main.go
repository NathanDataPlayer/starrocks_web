package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"net/http"
	"strings"
)

// 任务状态结构体
type RoutineLoad struct {
	Id                   string  `json:"id"`
	Name                 string  `json:"name"`
	CreateTime           string  `json:"create_time"`
	PauseTime            *string `json:"pause_time"`
	EndTime              *string `json:"end_time"`
	DbName               string  `json:"db_name"`
	TableName            string  `json:"table_name"`
	State                string  `json:"state"`
	DataSourceType       string  `json:"data_source_type"`
	CurrentTaskNum       int     `json:"current_task_num"`
	JobProperties        string  `json:"job_properties"`
	DataSourceProperties string  `json:"data_source_properties"`
	CustomProperties     string  `json:"custom_properties"`
	Statistic            string  `json:"statistic"`
	Progress             string  `json:"progress"`
	TimestampProgress    string  `json:"timestamp_progress"`
	ReasonOfStateChanged *string `json:"reason_of_state_changed"` // 可能为 NULL
	ErrorLogUrls         *string `json:"error_log_urls"`         // 可能为 NULL
	TrackingSQL          *string `json:"tracking_sql"`           // 可能为 NULL
	OtherMsg             string  `json:"other_msg"`
	LatestSourcePosition string  `json:"latest_source_position"` // 结构是 "{"0":"810","1":"816"}"
}

// 物化视图结构体
type MaterialView struct {
	Id                          string  `json:"id"`
	DatabaseName                string  `json:"database_name"`
	Name                        string  `json:"name"`
	RefreshType                 string  `json:"refresh_type"`
	IsActive                    bool    `json:"is_active"`
	InactiveReason              *string `json:"inactive_reason"`
	PartitionType               string  `json:"partition_type"`
	TaskId                      int     `json:"task_id"`
	TaskName                    string  `json:"task_name"`
	LastRefreshStartTime        string  `json:"last_refresh_start_time"`
	LastRefreshFinishedTime     string  `json:"last_refresh_finished_time"`
	LastRefreshDuration         float64 `json:"last_refresh_duration"`
	LastRefreshState            string  `json:"last_refresh_state"`
	LastRefreshForceRefresh     bool    `json:"last_refresh_force_refresh"`
	LastRefreshStartPartition    *string `json:"last_refresh_start_partition"`
	LastRefreshEndPartition      *string `json:"last_refresh_end_partition"`
	LastRefreshBaseRefreshPartitions string `json:"last_refresh_base_refresh_partitions"`
	LastRefreshMvRefreshPartitions string `json:"last_refresh_mv_refresh_partitions"`
	LastRefreshErrorCode        int     `json:"last_refresh_error_code"`
	LastRefreshErrorMessage     *string `json:"last_refresh_error_message"`
	Rows                        int     `json:"rows"`
	Text                        string  `json:"text"`
	ExtraMessage                string  `json:"extra_message"`
	QueryRewriteStatus          string  `json:"query_rewrite_status"`
	Creator                     string  `json:"creator"`
}

// 物化视图依赖结构体
type MaterialViewDependency struct {
	ObjectID       int    `json:"object_id"`
	ObjectName     string `json:"object_name"`
	ObjectDatabase string `json:"object_database"`
	ObjectType     string `json:"object_type"`
	RefObjectID    int    `json:"ref_object_id"`
	RefObjectName  string `json:"ref_object_name"`
	RefObjectDatabase string `json:"ref_object_database"`
	RefObjectType  string `json:"ref_object_type"`
}

// 数据库连接
var db *sql.DB

// 转换为小写
func toLower(s string) string {
	return strings.ToLower(s)
}

func getDatabases() ([]string, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, err
		}
		databases = append(databases, dbName)
	}

	return databases, nil
}

// 
func pauseRoutineLoad(taskName string) error {
	query := "PAUSE ROUTINE LOAD FOR " + taskName
	_, err := db.Exec(query)
	return err
}

// 
func resumeRoutineLoad(taskName string) error {
	query := "RESUME ROUTINE LOAD FOR " + taskName
	_, err := db.Exec(query)
	return err
}

// 
func getRoutineLoads(dbName string) ([]RoutineLoad, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := fmt.Sprintf("SHOW ROUTINE LOAD FROM %s", dbName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routineLoads []RoutineLoad
	for rows.Next() {
		var rl RoutineLoad
		var pauseTime, endTime, reasonOfStateChanged, errorLogUrls, trackingSQL sql.NullString

		err := rows.Scan(
			&rl.Id,
			&rl.Name,
			&rl.CreateTime,
			&pauseTime,
			&endTime,
			&rl.DbName,
			&rl.TableName,
			&rl.State,
			&rl.DataSourceType,
			&rl.CurrentTaskNum,
			&rl.JobProperties,
			&rl.DataSourceProperties,
			&rl.CustomProperties,
			&rl.Statistic,
			&rl.Progress,
			&rl.TimestampProgress,
			&reasonOfStateChanged,
			&errorLogUrls,
			&trackingSQL,
			&rl.OtherMsg,
			&rl.LatestSourcePosition,
		)

		if err != nil {
			return nil, err
		}

		// 处理可能为NULL的字段
		if pauseTime.Valid {
			rl.PauseTime = &pauseTime.String
		}
		if endTime.Valid {
			rl.EndTime = &endTime.String
		}
		if reasonOfStateChanged.Valid {
			rl.ReasonOfStateChanged = &reasonOfStateChanged.String
		}
		if errorLogUrls.Valid {
			rl.ErrorLogUrls = &errorLogUrls.String
		}
		if trackingSQL.Valid {
			rl.TrackingSQL = &trackingSQL.String
		}

		routineLoads = append(routineLoads, rl)
	}

	return routineLoads, nil
}

// 获取物化视图
func getMaterialViews(dbName string) ([]MaterialView, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := fmt.Sprintf("SHOW MATERIALIZED VIEWS FROM %s", dbName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var materialViews []MaterialView
	for rows.Next() {
		var mv MaterialView
		var inactiveReason, lastRefreshStartPartition, lastRefreshEndPartition, lastRefreshErrorMessage sql.NullString

		err = rows.Scan(
			&mv.Id,
			&mv.DatabaseName,
			&mv.Name,
			&mv.RefreshType,
			&mv.IsActive,
			&inactiveReason,
			&mv.PartitionType,
			&mv.TaskId,
			&mv.TaskName,
			&mv.LastRefreshStartTime,
			&mv.LastRefreshFinishedTime,
			&mv.LastRefreshDuration,
			&mv.LastRefreshState,
			&mv.LastRefreshForceRefresh,
			&lastRefreshStartPartition,
			&lastRefreshEndPartition,
			&mv.LastRefreshBaseRefreshPartitions,
			&mv.LastRefreshMvRefreshPartitions,
			&mv.LastRefreshErrorCode,
			&lastRefreshErrorMessage,
			&mv.Rows,
			&mv.Text,
			&mv.ExtraMessage,
			&mv.QueryRewriteStatus,
			&mv.Creator,
		)

		if err != nil {
			return nil, err
		}

		if inactiveReason.Valid {
			mv.InactiveReason = &inactiveReason.String
		}
		if lastRefreshStartPartition.Valid {
			mv.LastRefreshStartPartition = &lastRefreshStartPartition.String
		}
		if lastRefreshEndPartition.Valid {
			mv.LastRefreshEndPartition = &lastRefreshEndPartition.String
		}
		if lastRefreshErrorMessage.Valid {
			mv.LastRefreshErrorMessage = &lastRefreshErrorMessage.String
		}

		materialViews = append(materialViews, mv)
	}
	return materialViews, nil
}

// 获取物化视图依赖关系
func getMaterialViewDependencies(dbName string) ([]MaterialViewDependency, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := fmt.Sprintf("SELECT object_id, object_name, object_database, object_type, ref_object_id, ref_object_name, ref_object_database, ref_object_type FROM sys.object_dependencies WHERE object_database='%s'", dbName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dependencies []MaterialViewDependency
	for rows.Next() {
		var dep MaterialViewDependency
		err := rows.Scan(
			&dep.ObjectID,
			&dep.ObjectName,
			&dep.ObjectDatabase,
			&dep.ObjectType,
			&dep.RefObjectID,
			&dep.RefObjectName,
			&dep.RefObjectDatabase,
			&dep.RefObjectType,
		)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, dep)
	}
	return dependencies, nil
}

// 激活物化视图
func activateMaterialView(viewName string) error {
	if db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	query := "ALTER MATERIALIZED VIEW " + viewName + " ACTIVE"
	_, err := db.Exec(query)
	return err
}

// 停用物化视图
func deactivateMaterialView(viewName string) error {
	if db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	query := "ALTER MATERIALIZED VIEW " + viewName + " INACTIVE"
	_, err := db.Exec(query)
	return err
}

// 连接数据库
func connectDatabase(username, password, dbName string) (*sql.DB, error) {
	connStr := fmt.Sprintf("%s:%s@tcp(localhost:9030)/%s", username, password, dbName)
	tempDB, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, err
	}

	if err := tempDB.Ping(); err != nil {
		return nil, err
	}

	return tempDB, nil
}

func main() {
	username := "root"
	password := ""

	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(localhost:9030)/", username, password))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer db.Close() // 在程序结束时关闭数据库连接

	// 测试连接
	if err := db.Ping(); err != nil {
		fmt.Println("error on ping:", err)
		return
	}

	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"toLower": toLower,
	})

	r.LoadHTMLGlob("templates/*")


	r.GET("/", func(c *gin.Context) {
		databases, err := getDatabases()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
			return
		}
		c.HTML(http.StatusOK, "index.html", gin.H{
			"databases": databases,
		})
	})

	// 连接数据库的路由
	r.GET("/connect", func(c *gin.Context) {
		dbName := c.Query("db")
		if dbName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "database name is required"})
			return
		}

		// 构建连接字符串
		connStr := fmt.Sprintf("%s:%s@tcp(localhost:9030)/%s", username, password, dbName)

		// 在这里尝试连接
		tempDB, err := sql.Open("mysql", connStr)
		defer tempDB.Close() // 关闭临时数据库连接

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "failed to connect to database"})
			return
		}

		// 测试数据库连接
		if err := tempDB.Ping(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "failed to ping database"})
			return
		}

		// 如果连接成功，可以为后续操作更新全局db
		db, err = tempDB, nil
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "connected to database " + dbName})
	})

	r.GET("/databases", func(c *gin.Context) {
		databases, err := getDatabases()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"databases": databases})
	})

	r.POST("/routine-load/pause", func(c *gin.Context) {
		var requestBody struct {
			Tasks []string `json:"tasks"`
		}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "invalid request"})
			return
		}

		for _, taskName := range requestBody.Tasks {
			err := pauseRoutineLoad(taskName)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/routine-load/resume", func(c *gin.Context) {
		var requestBody struct {
			Tasks []string `json:"tasks"`
		}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "invalid request"})
			return
		}

		for _, taskName := range requestBody.Tasks {
			err := resumeRoutineLoad(taskName)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/routine-load", func(c *gin.Context) {
		dbName := c.Query("db")
		if dbName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "database name is required"})
			return
		}

		routineLoads, err := getRoutineLoads(dbName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"routineLoads": routineLoads})
	})

	r.GET("/material-view", func(c *gin.Context) {
		dbName := c.Query("db")
		if dbName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "database name is required"})
			return
		}

		materialViews, err := getMaterialViews(dbName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"materialViews": materialViews})
	})

	r.GET("/material-view-dependencies", func(c *gin.Context) {
		dbName := c.Query("db")
		if dbName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "database name is required"})
			return
		}

		dependencies, err := getMaterialViewDependencies(dbName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"dependencies": dependencies})
	})

	r.POST("/material-view/activate", func(c *gin.Context) {
		var requestBody struct {
			Views []string `json:"views"`
		}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "invalid request"})
			return
		}

		for _, viewName := range requestBody.Views {
			err := activateMaterialView(viewName)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/material-view/deactivate", func(c *gin.Context) {
		var requestBody struct {
			Views []string `json:"views"`
		}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "invalid request"})
			return
		}

		for _, viewName := range requestBody.Views {
			err := deactivateMaterialView(viewName)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.Run("0.0.0.0:7080")
}