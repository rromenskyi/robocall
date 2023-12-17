package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"bytes"
	"encoding/csv"
	"io/ioutil"
//	"strconv"
    "html/template"    
    "database/sql"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
)

//global var for allowed ips to access webß
var allowedIPs []string

func main() {

	uniqueID := generateRandomID()
	fmt.Println("Generated unique ID:", uniqueID)


	var config Config
	config, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

//set IPS
    allowedIPs = strings.Split(config.Global.AllowedIP, ",")  
//    log.Println("config allowed IPS:",config.Global.AllowedIP)


	setupDatabase(config.Database)  // передаем конфигурацию базы данных
	defer db.Close()
//process tasks table to calls_queue
    go processTasks(db)


    if config.AMI.Enabled == "1" {
    		fmt.Println("AMI Enabled")
    		go initAMI(config.AMI)
    } else {
    		fmt.Println("AMI Disabled")
    }
    if config.OAMI.Enabled == "1" {
    		fmt.Println("OAMI Enabled")
    		go initOAMI(config.OAMI)
		//go processlimits
			go processLimits(db)
		//go callqueue
		    go processCallQueueue(db, config.Global)
    } else {
    		fmt.Println("OAMI Disabled")
    }
    	r := setupRouter()
        if config.Global.HttpPort != "" {
    	   r.Run(config.Global.HttpPort)
        }
        if config.Global.HttpsPort != "" {
            err := r.RunTLS(config.Global.HttpsPort, config.Global.HttpsPubKey, config.Global.HttpsPriKey)
            if err != nil {
                log.Fatalf("Failed to run server: %v", err)
            }
        }
}

func shorten(s string, length int) string {
    if len(s) > length {
        return s[:length]
    }
    return s
}


func uploadPage(c *gin.Context) {
    ivrNames, err := getIVRNames()
    if err != nil {
        c.JSON(500, gin.H{"message": "Error fetching IVR names"})
        return
    }


    geos, err := fetchGeos()
    if err != nil {
        c.JSON(500, gin.H{"message": "Error fetching geo names"})
        return
    }

    c.HTML(http.StatusOK, "upload.html", gin.H{
        "ivrNames": ivrNames,
        "geos": geos,
     })

 }

func getTasks() ([]Task, error) {
    var tasks []Task

    rows, err := db.Query("SELECT id, phone_number, nlines, cps, ivr, retries, call_timeout, retry_time, uid, type, ready, gmt FROM tasks order by id desc")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var t Task
        if err := rows.Scan(&t.ID, &t.PhoneNumber, &t.NLines, &t.CPS, &t.IVR, &t.Retries, &t.Dial_time, &t.Retry_time, &t.UID, &t.Type, &t.Ready, &t.GMT); err != nil {
            return nil, err
        }
        tasks = append(tasks, t)
    }
    if err = rows.Err(); err != nil {
        return nil, err
    }
    return tasks, nil
}

func AuthRequired(c *gin.Context) {

//    allowedIPs := strings.Split(config.Global.AllowedIP, ",")  

    clientIP := c.ClientIP()
    

    isAllowedIP := false
    for _, allowedIP := range allowedIPs {
        if clientIP == allowedIP {
            isAllowedIP = true
            break
        }
    }

    if !isAllowedIP {
        c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Access forbidden: IP is not allowed"})
        return
    }

    session := sessions.Default(c)
    user := session.Get("user")
    if user == nil {
        c.Redirect(http.StatusSeeOther, "/login")
        c.Abort()
    } else {
        c.Next()
    }
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
	    "shorten": shorten,
	    "countForUID": countForUID,
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	r.NoRoute(func(c *gin.Context) {
	    c.Redirect(http.StatusSeeOther, "/")
	})

	r.GET("/login", func(c *gin.Context) {
	    c.HTML(http.StatusOK, "login.html", nil)
	})

	r.POST("/login", func(c *gin.Context) {
	    username := c.PostForm("username")
	    password := c.PostForm("password")

	    if isValidUser(username, password) {
	        session := sessions.Default(c)
	        session.Set("user", username)
	        session.Save()
	        c.Redirect(http.StatusSeeOther, "/")
	    } else {
	        c.HTML(http.StatusOK, "login.html", gin.H{"Error": "Неверное имя пользователя или пароль"})
	    }
	})

	r.GET("/logout", func(c *gin.Context) {
	    session := sessions.Default(c)
	    session.Delete("user")
	    session.Save()
	    c.Redirect(http.StatusSeeOther, "/login")
	})

	r.GET("/private", AuthRequired, func(c *gin.Context) {
	    c.String(http.StatusOK, "Приватная информация!")
	})

	r.GET("/", AuthRequired, func(c *gin.Context) {
	    c.HTML(http.StatusOK, "main.html", gin.H{
		    "title": "Auto-dialer",
	    })
	})

	r.GET("/ping", func(c *gin.Context) {
	    c.JSON(200, gin.H{
	        "message": "pong",
	    })
	})

	r.GET("/tasks", AuthRequired, func(c *gin.Context) {
		AuthRequired(c)
		if !c.IsAborted() {
		        tasks, err := getTasks()
					if err != nil {
					    c.JSON(500, gin.H{"message": "Error fetching tasks", "error": err.Error()})
					    return
					}
		    c.HTML(http.StatusOK, "tasks.html", gin.H{
		    "title": "Tasks",
		    "tasks": tasks,
		    })
		}
	})


	r.GET("/upload", AuthRequired, func(c *gin.Context) {
		AuthRequired(c)
		if !c.IsAborted() {
		    ivrNames, err := getIVRNames()
		    if err != nil {
		        c.JSON(500, gin.H{"message": "Error fetching IVR names"})
		        return
		    }
		    geos, err := fetchGeos()
		    if err != nil {
		        c.JSON(500, gin.H{"message": "Error fetching geo names"})
		        return
		    }

		    c.HTML(http.StatusOK, "upload.html", gin.H{
		    "title": "Upload",
	        "ivrNames": ivrNames,
	        "geos": geos,
		    })
		}
	})

r.GET("/branches", AuthRequired, func(c *gin.Context) {
    branches, err := fetchBranches()
    if err != nil {
        c.JSON(500, gin.H{
            "message": "Error fetching branches",
            "error":   err.Error(),
        })
        return
    }
    
    c.HTML(http.StatusOK, "branches.html", gin.H{
        "title":    "Branches",
        "branches": branches,
        "callEntry": callEntry,
    })
})

r.GET("/users", AuthRequired, func(c *gin.Context) {
    users, err := fetchUsers()
    if err != nil {
        c.JSON(500, gin.H{
            "message": "Error fetching branches",
            "error":   err.Error(),
        })
        return
    }
    
    c.HTML(http.StatusOK, "users.html", gin.H{
        "title":    "Users",
        "users": users,
    })
})


r.GET("/user/:id", func(c *gin.Context) {
    userId := c.Param("id")
    var user User

    err := db.QueryRow("SELECT id, username, password, accesslevel, access, comment FROM users WHERE id=?", userId).Scan(&user.ID, &user.UserName, &user.Password, &user.AccessLevel, &user.Access, &user.Comment)
    
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get user details"})
        return
    }

    // Если вы не хотите передавать пароль в ответе (что является хорошей практикой), установите его в пустую строку.
    user.Password = ""

    c.JSON(http.StatusOK, user)
})

r.POST("/user-edit/:id", AuthRequired, func(c *gin.Context) {
    taskID := c.Param("id")
    username := c.PostForm("username")
    accesslevel := c.PostForm("accesslevel")
    access := c.PostForm("access")
    comment := c.PostForm("comment")
    password := ""
   
    if c.PostForm("password") != "" {
       password = hashPassword(c.PostForm("password"))
    } else {
    }

if taskID == "0" {
//        log.Println("INSERT INTO users (username, password, accesslevel, access, comment) VALUES ('?','?','?','?','?')", username, password, accesslevel, access, comment)
        // Обновите значение в базе данных
        _, err := db.Exec("INSERT INTO users (username, password, accesslevel, access, comment) VALUES (?,?,?,?,?)", username, password, accesslevel, access, comment)
        if err != nil {
            log.Println("update users error: ", err)
            c.JSON(http.StatusInternalServerError, gin.H{"message": "Error updating user"})
            return
        } 

    } else {
//        log.Println("UPDATE users SET password = COALESCE(NULLIF('?', ''), password), accesslevel = '?', access = '?', comment='?' WHERE id = '?'", password, accesslevel, access, comment, taskID)
        // Обновите значение в базе данных
        _, err := db.Exec("UPDATE users SET password = COALESCE(NULLIF(?, ''), password), accesslevel = ?, access = ?, comment=? WHERE id = ?", password, accesslevel, access, comment, taskID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"message": "Error updating user"})
            log.Println("update users error: ", err)
            return
        } 

    }

    c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
})

r.GET("/report-provider-min", AuthRequired, func(c *gin.Context) {

    c.HTML(http.StatusOK, "report-provider-min.html", gin.H{
        "title":    "Provider-Minutes",
    })
})


r.POST("/get-report-data", AuthRequired, func(c *gin.Context) {
    var formData struct {
        StartdatetimeDate string `json:"startdatetimeDate"`
        StopdatetimeDate  string `json:"stopdatetimeDate"`
    }
    
    if err := c.BindJSON(&formData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
        return
    }

    reportData, err := GetReportData(formData.StartdatetimeDate, formData.StopdatetimeDate)
//    log.Println("debug reportdata: ", reportData)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching report data"})
        return
    }

    c.JSON(http.StatusOK, reportData)
})

r.GET("/geos", AuthRequired, func(c *gin.Context) {
    geos, err := fetchGeos()
    if err != nil {
        c.JSON(500, gin.H{
            "message": "Error fetching branches",
            "error":   err.Error(),
        })
        return
    }
    
    c.HTML(http.StatusOK, "geos.html", gin.H{
        "title":    "Geo",
        "geos": geos,
    })
})


r.GET("/geo/:id", func(c *gin.Context) {
    geoId := c.Param("id")
    var geo Geo

    err := db.QueryRow("SELECT id, geo, geo2, code, prefix, src, provider, nlines, cps, comment FROM geos WHERE id=?", geoId).Scan(&geo.ID, &geo.Geo, &geo.Geo2, &geo.Code, &geo.Prefix, &geo.Src, &geo.Provider, &geo.NLines, &geo.CPS, &geo.Comment)
    
    if err != nil {
       log.Println("get geo error: ", err)
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"message": "Geo not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get geo details"})
        return
    }

    c.JSON(http.StatusOK, geo)
})

r.POST("/geo-edit/:id", AuthRequired, func(c *gin.Context) {
    taskID := c.Param("id")
    geo := c.PostForm("geoname")
    geo2 := c.PostForm("geoname2")
    code := c.PostForm("geocode")
    prefix := c.PostForm("geoprefix")
    src := c.PostForm("geosrc")
    provider := c.PostForm("geoprovider")
    nlines := c.PostForm("geonlines")
    cps := c.PostForm("geocps")
    comment := c.PostForm("geocomment")
   

if taskID == "0" {
        // Обновите значение в базе данных
        _, err := db.Exec("INSERT INTO geos (geo, geo2, code, prefix, src, provider, nlines, cps, comment) VALUES (?,?,?,?,?,?,?,?,?)", geo, geo2, code, prefix, src, provider, nlines, cps, comment)
        if err != nil {
            log.Println("create geo error: ", err)
            c.JSON(http.StatusInternalServerError, gin.H{"message": "Error creating geo"})
            return
        } 

    } else {
        // Обновите значение в базе данных
        _, err := db.Exec("UPDATE geos SET geo=?, geo2=?, code=?, prefix=?, src=?, provider=?, nlines=?, cps=?, comment=? WHERE id = ?", geo, geo2, code, prefix, src, provider, nlines, cps, comment, taskID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"message": "Error updating geo"})
            log.Println("update geo error: ", err)
            return
        } 

    }

    c.JSON(http.StatusOK, gin.H{"message": "Geo updated successfully"})
})



r.POST("/task-edit/:id", AuthRequired, func(c *gin.Context) {
    taskID := c.Param("id")
    
    // Обновите значение в базе данных
    _, err := db.Exec("UPDATE tasks SET ready = 1 WHERE id = ?", taskID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Error updating task"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully"})
})

r.POST("/task-delete/:id", AuthRequired, func(c *gin.Context) {
    taskID := c.Param("id")
    
    // Обновите значение в базе данных
    _, err := db.Exec("DELETE FROM tasks WHERE id = ?", taskID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Error updating task"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
})

r.POST("/branches-stat/:id", AuthRequired, func(c *gin.Context) {
    branchID := c.Param("id")


    // Выполняем запрос.
    rows, err := db.Query("SELECT sip_cause,hangupcause AS cause, count(*) AS total_calls, CONCAT(ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM cdr WHERE TYPE = 'out-ivr' AND userfield in (select branch from dialer_branches where id=? ) )),'%') AS percentage  FROM cdr WHERE TYPE='out-ivr' AND userfield in (select branch from dialer_branches where id=? ) GROUP BY sip_cause,hangupcause", branchID, branchID)

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failure",
            "message": "Ошибка при выполнении SQL запроса: " + err.Error(),
        })
        return
    }
    defer rows.Close()

var sipCause, hangupCause, totalCalls, percentage string

var result []string
for rows.Next() {
    if err := rows.Scan(&sipCause, &hangupCause, &totalCalls, &percentage); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failure",
            "message": "Ошибка при сканировании строки: " + err.Error(),
        })
        return
    }
    rowStr := fmt.Sprintf("sip: %s, hangup: %s, calls: %s - %s", sipCause, hangupCause, totalCalls, percentage)
    result = append(result, rowStr)
}

    combinedResult := strings.Join(result, "\n")

    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": combinedResult,
    })
})

r.POST("/branches-edit/:id", AuthRequired, func(c *gin.Context) {
    branchID := c.Param("id")
///
bodyBytes, err := ioutil.ReadAll(c.Request.Body)
if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
        "status": "failure",
        "message": "Can't read request body",
    })
    return
}

// Печать тела запроса
fmt.Println("Request body:", string(bodyBytes))

// Восстановление тела запроса для дальнейшего использования
c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
    
    ///
    // Структура для хранения данных из JSON
    var requestData struct {
        Nlines int `json:"nlines"`
        Cps    int `json:"cps"`
    }
    
    // Пробуем получить и разобрать JSON из запроса
    if err := c.BindJSON(&requestData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "failure",
            "message": "Invalid request data",
        })
        return
    }

    // Выполняем запрос на обновление данных в базе
    _, err = db.Exec("UPDATE tasks t JOIN dialer_branches b ON t.uid=b.branch SET t.nlines = ?, t.cps = ? WHERE b.id=?", requestData.Nlines, requestData.Cps, branchID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failure",
            "message": "Database update error: " + err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Successfully updated sliders' values",
    })
})


r.GET("/branches-csv", AuthRequired, generateCSV)


//	r.GET("/upload", AuthRequired, uploadPage)
	r.POST("/upload", AuthRequired, handleUpload)


	// ... добавьте другие роуты по мере необходимости

	return r
}

func fetchBranches() ([]Branch, error) {
    // Создание слайса для хранения веток
    var branches []Branch

    // Выполнение запроса
    rows, err := db.Query("SELECT b.id, b.branch, b.start_time, b.stop_time, b.rows_processed, b.rows_total, b.rows_ok, t.nlines, t.cps, t.starttime, t.stoptime, b.parent_id, b.comment FROM dialer_branches b JOIN tasks t ON b.branch=t.uid order by id desc")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // Цикл по каждой строке результатов
    for rows.Next() {
        var b Branch
        if err := rows.Scan(&b.ID, &b.BranchName, &b.StartTime, &b.StopTime, &b.RowsProcessed, &b.RowsTotal, &b.RowsOK, &b.NLines, &b.CPS, &b.TODStart, &b.TODStop, &b.ParentID, &b.Comment); err != nil {
            return nil, err
        }
        // проверка на null
        if b.ParentID == nil {
        zero := 0
        b.ParentID = &zero
        }
        // проверка на null
        if b.RowsProcessed == nil {
        zero := 0
        b.RowsProcessed = &zero
        }
        // проверка на null
        if b.RowsTotal == nil {
        zero := 0
        b.RowsTotal = &zero
        }
        // проверка на null
        if b.RowsOK == nil {
        zero := 0
        b.RowsOK = &zero
        }
        if !b.Comment.Valid {
        b.Comment.String = "" // или любое другое значение по умолчанию
        }
        branches = append(branches, b)
    }

    // Возвращаем branches и nil в качестве ошибки
    return branches, nil
}


func fetchUsers() ([]User, error) {
    // Создание слайса для хранения веток
    var users []User

    // Выполнение запроса
    rows, err := db.Query("SELECT id, username, password, accesslevel, access, comment FROM users")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // Цикл по каждой строке результатов
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.UserName, &u.Password, &u.AccessLevel, &u.Access, &u.Comment); err != nil {
            return nil, err
        }
        if !u.Comment.Valid {
        u.Comment.String = "" // или любое другое значение по умолчанию
        }
        users = append(users, u)
    }

    // Возвращаем users и nil в качестве ошибки
    return users, nil
}

func fetchGeos() ([]Geo, error) {
    // Создание слайса для хранения веток
    var geos []Geo

    // Выполнение запроса
    rows, err := db.Query("SELECT id, geo, geo2, code, prefix, src, provider, nlines, cps, comment FROM geos order by id")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // Цикл по каждой строке результатов
    for rows.Next() {
        var g Geo
        // проверка на null
        if err := rows.Scan(&g.ID, &g.Geo, &g.Geo2, &g.Code, &g.Prefix, &g.Src, &g.Provider, &g.NLines, &g.CPS, &g.Comment); err != nil {
            return nil, err
        }
        if !g.Comment.Valid {
        g.Comment.String = "" // или любое другое значение по умолчанию
        }
        geos = append(geos, g)
    }

    // Возвращаем branches и nil в качестве ошибки
    return geos, nil
}


func generateCSV(c *gin.Context) {
    // Выполнение SQL-запроса
    rows, err := db.Query("SELECT accountcode, calldate, src, dst, billsec FROM cdr WHERE disposition = 'ANSWERED'")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failure",
            "message": "Ошибка при выполнении SQL запроса: " + err.Error(),
        })
        return
    }
    defer rows.Close()

    // Инициализация CSV writer
    buffer := &bytes.Buffer{}
    writer := csv.NewWriter(buffer)
    writer.Comma = ';' // Установка разделителя

    // Запись заголовков
    headers := []string{"accountcode", "calldate", "src", "dst", "billsec"}
    if err := writer.Write(headers); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "failure",
            "message": "Ошибка при записи в CSV: " + err.Error(),
        })
        return
    }

    // Запись данных
    for rows.Next() {
        var accountcode, calldate, src, dst, billsec string
        if err := rows.Scan(&accountcode, &calldate, &src, &dst, &billsec); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "status":  "failure",
                "message": "Ошибка при сканировании строки: " + err.Error(),
            })
            return
        }
        if err := writer.Write([]string{accountcode, calldate, src, dst, billsec}); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "status":  "failure",
                "message": "Ошибка при записи в CSV: " + err.Error(),
            })
            return
        }
    }

    // Завершение записи
    writer.Flush()

    // Отправка CSV файла пользователю
    c.Header("Content-Description", "File Transfer")
    c.Header("Content-Disposition", "attachment; filename=records.csv")
    c.Data(http.StatusOK, "text/csv", buffer.Bytes())
}
