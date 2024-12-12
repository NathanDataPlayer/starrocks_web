package main

import (
    "database/sql"
    "net/http"
    "strings"
    "fmt"
    "html/template"
    "github.com/gin-gonic/gin"
    _ "github.com/go-sql-driver/mysql"
)

type RoutineLoad struct {
    Id                   string  `json:"id"`
    Name                 string  `json:"name"`
    PauseTime            *string `json:"pause_time"`
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
    ReasonOfStateChanged string  `json:"reason_of_state_changed"`
    ErrorLogUrls         string  `json:"error_log_urls"`
    OtherMsg             string  `json:"other_msg"`
}

type MaterialView struct {
    Id             string  `json:"id"`
    Name           string  `json:"name"`
    DatabaseName   string  `json:"database_name"`
    Text           string  `json:"text"`
    Rows           int     `json:"rows"`
}

// 省略函数去重，增加相应的操作函数
var db *sql.DB

func toLower(s string) string {
    return strings.ToLower(s)
}

func pauseRoutineLoad(taskName string) error {
    query := "PAUSE ROUTINE LOAD FOR " + taskName
    _, err := db.Exec(query)
    return err
}

func resumeRoutineLoad(taskName string) error {
    query := "RESUME ROUTINE LOAD FOR " + taskName
    _, err := db.Exec(query)
    return err
}

func getMaterialViews() ([]MaterialView, error) {
    rows, err := db.Query("SHOW MATERIALIZED VIEW")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var materialViews []MaterialView
    for rows.Next() {
        var mv MaterialView
        err := rows.Scan(&mv.Id, &mv.Name, &mv.DatabaseName, &mv.Text,&mv.Rows)
        if err != nil {
            return nil, err
        }
        materialViews = append(materialViews, mv)
    }
    return materialViews, nil
}

func activateMaterialView(viewName string) error {
    query := "ALTER MATERIALIZED VIEW " + viewName + " ACTIVE"
    _, err := db.Exec(query)
    return err
}

func deactivateMaterialView(viewName string) error {
    query := "ALTER MATERIALIZED VIEW " + viewName + " INACTIVE"
    _, err := db.Exec(query)
    return err
}

func main() {
    var err error
    db, err = sql.Open("mysql", "root:root@tcp(localhost:9030)/dw")
    if err != nil {
        fmt.Println("error:", err)
        return
    }
    defer db.Close()

    r := gin.Default()
    r.SetFuncMap(template.FuncMap{
        "toLower": toLower,
    })

    r.LoadHTMLGlob("templates/*")

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

    r.GET("/", func(c *gin.Context) {
        rows, err := db.Query("SHOW ROUTINE LOAD")
        if err != nil {
            c.HTML(http.StatusInternalServerError, "index.html", gin.H{"error": err.Error()})
            return
        }
        defer rows.Close()

        var routineLoads []RoutineLoad
        for rows.Next() {
            var rl RoutineLoad
            err := rows.Scan(
                &rl.Id,
                &rl.Name,
                new(interface{}),
                &rl.PauseTime,
                new(interface{}),
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
                &rl.ReasonOfStateChanged,
                &rl.ErrorLogUrls,
                &rl.OtherMsg,
            )
            if err != nil {
                c.HTML(http.StatusInternalServerError, "index.html", gin.H{"error": err.Error()})
                return
            }
            routineLoads = append(routineLoads, rl)
        }

        c.HTML(http.StatusOK, "index.html", gin.H{
            "routineLoads": routineLoads,
        })
    })

    // 处理 GET 请求以获取物化视图
    r.GET("/material-view", func(c *gin.Context) {
        materialViews, err := getMaterialViews()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"materialViews": materialViews})
    })

    // 启动物化视图
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

    // 停止物化视图
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

    r.Run("0.0.0.0:8080")
}
