package main

import (
	"fmt"
	"log"
	"time"
	"strings"
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
//	"bufio"
//	"strings"
	"io"
	"sync"
	"strconv"
    "encoding/csv"
   	"math/rand"
   	 "golang.org/x/crypto/bcrypt"


//    "errors"
)

var db *sql.DB

type Call struct {
	ID          int    `db:"id"`
	Lead string `db:"lead"`
	UID  string `db:"uid"`
	src string `db:"src"`
	dst string `db:"dst"`
	IVR  string    `db:"ivr"`
	R_left int   `db:"retries_left"`
	StartTime string `db:"starttime"`
	StopTime string  `db:"stoptime"`
	Type	int `db:"type"`
	GeoID   int `db:"geo_id"`
	RetryTime int `db:"retry_time"`
	QueueName string `db:"queue_name"`
	Timeout int `db:"call_timeout"`
}

type Task struct {
	ID          int    `db:"id"`
	PhoneNumber string `db:"phone_number"`
	NLines  int    `db:"nlines"`
	CPS  int    `db:"cps"`
	IVR  string    `db:"ivr"`
	Retries int   `db:"retries"`
	Retry_time int   `db:"retries"`
	Dial_time int    `db:"call_timeout"`
	UID  sql.NullString `db:"uid"`
	Type  int    `db:"type"` //1 autoinformer, 2 autodialer
	Ready int    `db:"ready"`
	GMT  sql.NullString `db:"gmt"`
	GeoID          int    `db:"geo_id"`
	StartTime 	string `db:"start"`
	StopTime	string `db:"stop"`
	Src string
}

type Branch struct {
    ID            int       `db:"id"`
    BranchName    string    `db:"branch"`
    StartTime     string `db:"start_time"`
    StopTime      string `db:"stop_time"`
    RowsProcessed *int       `db:"rows_processed"`
    RowsTotal     *int       `db:"rows_total"`
    RowsOK        *int       `db:"rows_ok"`
    ParentID      *int       `db:"parent_id"`
//    Comment       string    `db:"comment"`
	NLines  int    `db:"nlines"`
	CPS  int       `db:"cps"`
    Comment        sql.NullString    `db:"comment"`
    TODStart string
    TODStop string
}

//type User struct {
//    ID            int       `db:"id"`
//    UserName     string    `db:"username"`
//    Password     string `db:"password"`
//    AccessLevel		int `db:"accesslevel"`
//    Access		string `db:"access"`
//    Comment        sql.NullString    `db:"comment"`
//}

type User struct {
    ID            int       `db:"id" json:"id"`
    UserName      string    `db:"username" json:"username"`
    Password      string    `db:"password" json:"password"`
    AccessLevel   int       `db:"accesslevel" json:"accesslevel"`
    Access        string    `db:"access" json:"access"`
    Comment       sql.NullString    `db:"comment" json:"comment"`
}


type Geo struct {
    ID            int       `db:"id"`
    Geo           string    `db:"geo"`
    Geo2          string `db:"geo2"`
    Code          string `db:"code"`
    Prefix        string `db:"prefix"`
    Src          string `db:"src"`
    Provider          string `db:"provider"`
	NLines  int    `db:"nlines"`
	CPS  int       `db:"cps"`
    Comment        sql.NullString    `db:"comment"`
}


type CallQueue struct {
    ID     int
    Lead   string
    Branch string
    Dst    string
    Src    string
}


type LimitsItem struct {
	NLines int
	CPS    int
}

var defaultLimits = LimitsItem{
	NLines: 300,  // Здесь укажите значения по умолчанию
	CPS:    30, // и здесь
}

type CallEntryItem struct {
	ID       int
	Lead     string
	Src      string
	Dst      string
	IVR      string
	State    int
	UID      string
	GeoID    int
}
/*
State
0 UNKNOWN
1 NOT_INUSE
2 INUSE
3 BUSY
4 INVALID
5 UNAVAILABLE
6 RINGING
7 RINGINUSE
8 ONHOLD*/

var limitsMap = make(map[string]LimitsItem)//лимиты вычитываются 1 раз в секунду
var worklimitsMap = make(map[string]LimitsItem)//фактически занятые линии, заполняются в процессе работы и сравниваются с лимитами
var callEntry = make(map[string]CallEntryItem)//звонки онлайн

var mutex = &sync.RWMutex{}

//stats
type HangupStats map[int]int // Код отбоя -> Количество
type UIDStats map[string]HangupStats // UID -> Статистика отбоя

func UpdateStats(stats UIDStats, uid string, hangupCode int) {
	// Если статистики для этого UID еще нет, создаем новую мапу
	if _, ok := stats[uid]; !ok {
		stats[uid] = make(HangupStats)
	}

	stats[uid][hangupCode]++
}

func GetPercentages(stats UIDStats, uid string) map[int]float64 {
	hangupStats, ok := stats[uid]
	if !ok {
		return nil // или возвращать ошибку
	}

	total := 0
	for _, count := range hangupStats {
		total += count
	}

	percentages := make(map[int]float64)
	for code, count := range hangupStats {
		percentages[code] = (float64(count) / float64(total)) * 100
	}

	return percentages
}



func CallsAdd(targetMap map[string]CallEntryItem, key string, item CallEntryItem) {
    mutex.Lock()
    targetMap[key] = item
    mutex.Unlock()
}

func CallsFind(targetMap map[string]CallEntryItem, key string) (CallEntryItem, bool) {
    mutex.RLock()
    item, exists := targetMap[key]
    mutex.RUnlock()
    return item, exists
}

func CallsDelete(targetMap map[string]CallEntryItem, key string) {
    mutex.Lock()
    delete(targetMap, key)
    mutex.Unlock()
}


func LimitsAdd(targetMap map[string]LimitsItem, key string, item LimitsItem) {
    mutex.Lock()
    targetMap[key] = item
    mutex.Unlock()
}

func LimitsFind(targetMap map[string]LimitsItem, key string) (LimitsItem, bool) {
    mutex.RLock()
    item, exists := targetMap[key]
    mutex.RUnlock()
    return item, exists
}

func LimitsDelete(targetMap map[string]LimitsItem, key string) {
    mutex.Lock()
    delete(targetMap, key)
    mutex.Unlock()
}

func LimitsLineGet(targetMap map[string]LimitsItem, key string) {
    mutex.Lock()
    defer mutex.Unlock()

    item, exists := targetMap[key]
    if exists {
        item.NLines++
        targetMap[key] = item
    }
}

func LimitsLineRelease(targetMap map[string]LimitsItem, key string) {
    mutex.Lock()
    defer mutex.Unlock()

    item, exists := targetMap[key]
    if exists {
        item.NLines--
        targetMap[key] = item
    }
}


func LimitsGetOrDefault(targetMap map[string]LimitsItem, key string) LimitsItem {
	mutex.RLock()  
	item, exists := targetMap[key]
	mutex.RUnlock()

	if !exists {
		return defaultLimits
	}
	return item
}

func setupDatabase(config DatabaseConfig) {
	connectionString := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s",
		config.User, config.Password, config.Host, config.Port, config.Name,
	)

	var err error
	db, err = sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Database is not responsive: %v", err)
	}

	fmt.Println("Connected to the database successfully!")
}

func handleUpload(c *gin.Context) {
    // Получаем файл из запроса
    file, uploadErr := c.FormFile("file")
    if uploadErr != nil {
        c.JSON(400, gin.H{"message": "Error retrieving the file"})
        return
    }

    // Открываем файл для чтения
    fileContent, err := file.Open()
    if err != nil {
        c.JSON(500, gin.H{"message": "Error opening the file"})
        return
    }
    defer fileContent.Close()

    // Чтение содержимого файла
    bytes, readErr := ioutil.ReadAll(fileContent)
    if readErr != nil {
        c.JSON(500, gin.H{"message": "Error reading the file"})
        return
    }

    // Теперь 'phoneNumber' будет содержать данные из файла
    phoneNumber := string(bytes)
    
    parameter1 := c.PostForm("parameter1")
    parameter2 := c.PostForm("parameter2")
    ivr := c.PostForm("ivr")
    parameter3 := c.PostForm("parameter3")
    parameter4 := c.PostForm("parameter4")
//    parameter5 := c.PostForm("parameter5")
    parameter5, _ := strconv.Atoi(c.PostForm("parameter5"))
	parameter5 *= 1000

    startdate := c.PostForm("startdatetimeDate") + " " + c.PostForm("startdatetimeTime")
    stopdate := c.PostForm("stopdatetimeDate") + " " + c.PostForm("stopdatetimeTime")
    starttime := c.PostForm("daytimestart")+":00"
    stoptime := c.PostForm("daytimestop")+":00"
    itype := c.PostForm("type")
    geo_id := c.PostForm("geo")
    uniqueID := generateRandomID()

    // Вставка данных в БД
    log.Println("INSERT INTO tasks(phone_number, nlines, cps, ivr, retries, retry_time, call_timeout, start, stop, starttime, stoptime, type, uid, geo_id) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", phoneNumber, parameter1, parameter2, ivr, parameter3, parameter4, parameter5, startdate, stopdate, starttime, stoptime, itype, uniqueID, geo_id)

    stmt, err := db.Prepare("INSERT INTO tasks(phone_number, nlines, cps, ivr, retries, retry_time, call_timeout, start, stop, starttime, stoptime, type, uid, geo_id) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
    if err != nil {
   	    fmt.Println("Error during insertion:", err)
        c.JSON(500, gin.H{"message": "Error preparing statement"})
        return
    }
    defer stmt.Close()

    _, err = stmt.Exec(phoneNumber, parameter1, parameter2, ivr, parameter3, parameter4, parameter5, startdate, stopdate, starttime, stoptime, itype, uniqueID, geo_id)
    if err != nil {
   	    fmt.Println("Error during insertion:", err)
        c.JSON(500, gin.H{"message": "Error inserting data"})
        return
    }

    c.JSON(200, gin.H{"message": "Data uploaded successfully"})
}

func getIVRNames() ([]string, error) {
	rows, err := db.Query("SELECT name FROM tasks_ivr")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ivrNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		ivrNames = append(ivrNames, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ivrNames, nil
}

func processTasks(db *sql.DB) {
    ticker := time.NewTicker(1 * time.Second)
    for {
        select {
        case <-ticker.C:
//	   	    log.Println("preprocessing task tick")
            tasks, err := getReadyTasks(db)
            if err != nil {
                log.Println("Error fetching tasks:", err)
                continue
            }
            
            for _, task := range tasks {
           	    log.Println("preprocessing task branch:", task.UID)
                err := processTask(db, &task)
                if err != nil {
                    log.Println("Error processing task:", err)
                }
            }
        }
    }
}

func getReadyTasks(db *sql.DB) ([]Task, error) {
    rows, err := db.Query(`SELECT tasks.id, tasks.phone_number, tasks.ivr, tasks.uid, tasks.type, tasks.retries, tasks.retry_time, tasks.call_timeout, tasks.ready, geos.src, tasks.start, tasks.stop FROM tasks JOIN geos on tasks.geo_id=geos.id WHERE ready = 1`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tasks []Task
    for rows.Next() {
        var t Task
        err := rows.Scan(&t.ID, &t.PhoneNumber, &t.IVR, &t.UID, &t.Type, &t.Retries, &t.Retry_time, &t.Dial_time, &t.Ready, &t.Src, &t.StartTime, &t.StopTime)
        if err != nil {
            return nil, err
        }
        tasks = append(tasks, t)
    }

    return tasks, nil
}

func processTask(db *sql.DB, task *Task) error {
	var callcount int
    tx, err := db.Begin()
    if err != nil {
        return err
    }

    log.Println("processing task branch:", task.UID)

	r := csv.NewReader(strings.NewReader(task.PhoneNumber))
	r.Comma = ';' // Разделитель
	r.LazyQuotes = true // Нестрогое обращение с кавычками

		for {
		    record, err := r.Read()
		    if err == io.EOF {
		        break
		    }
		    if err != nil {
		        tx.Rollback()
		        return err
		    }
		    //dirty hack
		//    task.Retries = 3;
		//    log.Println("record is:",record[0],record[1],task.UID, task.IVR, task.Retries, task.Ready, task.Type)
		        lead, dst := record[0], record[1]
		        callcount++
		        _, err = tx.Exec(`INSERT INTO calls_queue (lead, src, dst, branch, ivr, retries_left) VALUES (?, ?, ?, ?, ?, ?)`, lead, task.Src, dst, task.UID, task.IVR, task.Retries)
		        if err != nil {
		            tx.Rollback()
		            log.Println(err)
		            return err
		        
		    	}
		}

    _, err = tx.Exec(`UPDATE tasks SET ready = 2 WHERE id = ?`, task.ID)
    if err != nil {
        tx.Rollback()
        return err
    }
    _, err = tx.Exec(`INSERT INTO dialer_branches (branch, start_time, stop_time, rows_total) VALUES (?,?,?,?)`, task.UID, task.StartTime, task.StopTime, callcount)
    if err != nil {
        tx.Rollback()
        return err
    }
    return tx.Commit()
}

//processing call queue table
func processLimits(db *sql.DB) {
    ticker := time.NewTicker(1 * time.Second)
    for {
        select {
        case <-ticker.C:
		    geos, err := fetchGeos()
		    if err != nil {
                log.Println("Error fetching geos to limits:", err)
                continue
            }

		    tasks, err := getTasks()
		    if err != nil {
                log.Println("Error fetching tasks to limits:", err)
                continue
            }

    		newLimit := make(map[string]LimitsItem)

           	mutex.Lock()
            
            for _, geo := range geos {
					newLimit[strconv.Itoa(geo.ID)] = LimitsItem{NLines: geo.NLines, CPS: geo.CPS}
//		         	limitsMap = newLimit
//		       		mutex.Unlock()
            }
//    		newLimit := make(map[string]LimitsItem)
            
            for _, task := range tasks {
					newLimit[task.UID.String] = LimitsItem{NLines: task.NLines, CPS: task.CPS}
//		         	mutex.Lock()
		         	limitsMap = newLimit
            }

      		mutex.Unlock()


//        log.Println("Limits: ",limitsMap)
        }
    }
}



//processing call queue table
func processCallQueueue(db *sql.DB, config GlobalConfig) {
    //todo db garbage collector from tasks and calls_queue

   	//sleep for init from db
   	time.Sleep(2 * time.Second)
  
	var pauseDuration time.Duration
	var cpspauseDuration time.Duration
    var i int

	if queueCPS, _ := strconv.Atoi(config.QueueCPS); queueCPS == 0 {
	    pauseDuration = 20 * time.Millisecond // CPS 50
	} else {
	    pauseTime := 1.0 / float64(queueCPS)
	    pauseDuration = time.Duration(pauseTime * 1e9) // 1e9 nanoseconds is 1 second
	}
//	log.Println("ticker:", pauseDuration)

	ticker := time.NewTicker(pauseDuration)
//    ticker := time.NewTicker(10 * time.Millisecond)
    for {
        select {
        case <-ticker.C:

        	  tasks, err := getActiveTasks(db)
            if err != nil {
                log.Println("Error fetching tasks:", err)
//                continue
            }
            for _, task := range tasks {

//           	    log.Println("preprocessing task branch:", task.UID.String)

//nice place to create go routine ...!!!


//init new if doesn't exist
           	    _, exists := LimitsFind(worklimitsMap, task.UID.String)
				if ! exists {
						LimitsAdd(worklimitsMap, task.UID.String, LimitsItem{NLines:0, CPS:limitsMap[task.UID.String].CPS})
           	    }
           	    _, exists = LimitsFind(worklimitsMap, strconv.Itoa(task.GeoID))
				if ! exists {
						LimitsAdd(worklimitsMap, strconv.Itoa(task.GeoID), LimitsItem{NLines:0, CPS: limitsMap[strconv.Itoa(task.GeoID)].CPS})
           	    }


//set cps limit
            	if (limitsMap[task.UID.String].CPS > limitsMap[strconv.Itoa(task.GeoID)].CPS && limitsMap[strconv.Itoa(task.GeoID)].CPS == 0) {
	            		item, exists := worklimitsMap[task.UID.String]
						if exists {
						    item.CPS = limitsMap[task.UID.String].CPS
//						    item.NLines = worklimitsMap[task.UID.String].NLines
						    worklimitsMap[task.UID.String] = item
						} else {
			           	    log.Println("OOPS workinglimits:", worklimitsMap)
						    //worklimitsMap[task.UID.String] = item
						}
//	            		worklimitsMap[task.UID.String].CPS=limitsMap[task.UID.String].CPS
            	}

//check/set lines limit
//						i++
//		           	    log.Println("preprocessing ", i, "calls")

          	    if (limitsMap[task.UID.String].NLines > worklimitsMap[task.UID.String].NLines && limitsMap[strconv.Itoa(task.GeoID)].NLines > worklimitsMap[strconv.Itoa(task.GeoID)].NLines){
		           	    _, exists := LimitsFind(worklimitsMap, task.UID.String)
		           	    if ! exists  {
							LimitsAdd(worklimitsMap, task.UID.String, LimitsItem{NLines:0, CPS: limitsMap[task.UID.String].CPS})
						}
		           	   	_, exists = LimitsFind(worklimitsMap, strconv.Itoa(task.GeoID))
		           	   	if ! exists {
							LimitsAdd(worklimitsMap, strconv.Itoa(task.GeoID), LimitsItem{NLines:0, CPS: limitsMap[strconv.Itoa(task.GeoID)].CPS})
						}
//						log.Println("calling!!!!")
						//nice place to make calls
			            calls, err := getReadyCalls(db, task.UID.String)
			            if err != nil {
			                log.Println("Error fetching calls:", err)
//			                continue
			            }

  //         	    log.Println("workinglimits after:", worklimitsMap)

						i++
		           	    log.Println("preprocessing ", i," calls:", calls)
			            for _, call := range calls {
//			           	    log.Println("preprocessing call branch:", call.UID)
							//cps limit
			           	    log.Println("workinglimits:", worklimitsMap)
							//need workaround for sleep XXXms for first calls!!!!
							if isCurrentTimeBetween(call.StartTime,call.StopTime){
								if worklimitsMap[task.UID.String].CPS == 0 {
								    cpspauseDuration = 20 * time.Millisecond // CPS 50
								} else {
								    pauseTime := 1.0 / float64(worklimitsMap[task.UID.String].CPS)
								    cpspauseDuration = time.Duration(pauseTime * 1e9) // 1e9 nanoseconds is 1 second
								}

                                //slow start needed for initial cps hold
								if len(calls) < 10 {
								    cpspauseDuration = 200 * time.Millisecond // CPS 5
								}

//								log.Println("ticker:", cpspauseDuration, worklimitsMap[task.UID.String].CPS)

//								time.Sleep(cpspauseDuration)
							 	//to be counted acd, avg calltime
//								log.Println("call type:", call.Type)

								switch call.Type {
								 case 3:
								 	//to be counted acd, avg calltime
									log.Println("dialer predictive:", call)
									    _, exists := QueueFind(queuesMap, call.QueueName)
						           	    if exists {
						           	    	if queuesMap[call.QueueName].Available>0 {
						           	    		log.Println("AVAIBLE:", queuesMap[call.QueueName].Available)
											 	LimitsLineGet(worklimitsMap,task.UID.String)
											 	LimitsLineGet(worklimitsMap,strconv.Itoa(task.GeoID))
								                go processCall(db, &call, cpspauseDuration)
						           	    	}
										}
								 case 2:
								 	//just pickup free queue member
									log.Println("dialer progressive:", call)
									    _, exists := QueueFind(queuesMap, call.QueueName)
						           	    if exists {
						           	    	if queuesMap[call.QueueName].Available>0 {
						           	    		log.Println("AVAIBLE:", queuesMap[call.QueueName].Available)
											 	LimitsLineGet(worklimitsMap,task.UID.String)
											 	LimitsLineGet(worklimitsMap,strconv.Itoa(task.GeoID))
								                go processCall(db, &call, cpspauseDuration)
						           	    	}
										}

								 case 1:
								 	//don't care dial em all
								 	LimitsLineGet(worklimitsMap,task.UID.String)
								 	LimitsLineGet(worklimitsMap,strconv.Itoa(task.GeoID))
								 	log.Println("dialer normal:", call)
					                go processCall(db, &call, cpspauseDuration)
//								 	log.Println("dialer normal:", call)
								 default:
								 	log.Println("dialer updredicted:", call)
								 }
					      	}      
					  	}

				}
            }
        }
    }
}

func getActiveTasks(db *sql.DB) ([]Task, error) {
    var tasks []Task

    rows, err := db.Query("SELECT id, uid, type, geo_id FROM tasks where tasks.stop >= now() and tasks.start <= now() and ready=2 order by id desc")//active now and processed
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var t Task
        if err := rows.Scan(&t.ID, &t.UID, &t.Type, &t.GeoID); err != nil {
            return nil, err
        }
        tasks = append(tasks, t)
    }
    if err = rows.Err(); err != nil {
        return nil, err
    }
    return tasks, nil
}


func getRandomSrc(input string) string {
	items := strings.Split(input, ",")
	if len(items) == 0 || (len(items) == 1 && items[0] == "") {
		items = []string{"000000"}
	}
	rand.Seed(time.Now().UnixNano())
	return strings.TrimSpace(items[rand.Intn(len(items))])
}


func getReadyCalls(db *sql.DB, branch string) ([]Call, error) {
    rows, err := db.Query(`SELECT calls_queue.id, calls_queue.lead, calls_queue.branch, calls_queue.dst, calls_queue.src, calls_queue.ivr, calls_queue.retries_left, tasks.type, tasks.starttime, tasks.stoptime, tasks.geo_id, tasks.retry_time, tasks_ivr.queue_name, tasks.call_timeout FROM calls_queue LEFT JOIN tasks ON calls_queue.branch=tasks.uid LEFT JOIN tasks_ivr ON tasks_ivr.name=tasks.ivr WHERE retries_left > 0 and calls_queue.retry_time <= now() and calls_queue.branch=?`, branch)
//    rows, err := db.Query(`SELECT calls_queue.id, calls_queue.lead, calls_queue.branch, calls_queue.dst, calls_queue.src, calls_queue.ivr, calls_queue.retries_left, tasks.starttime, tasks.stoptime, tasks.type FROM calls_queue  JOIN tasks ON calls_queue.branch=tasks.uid WHERE calls_queue.retries_left > 0 and tasks.stop >= now() and tasks.start <= now()`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var calls []Call
    for rows.Next() {
        var t Call
        err := rows.Scan(&t.ID, &t.Lead, &t.UID, &t.dst, &t.src, &t.IVR, &t.R_left, &t.Type, &t.StartTime, &t.StopTime, &t.GeoID, &t.RetryTime, &t.QueueName, &t.Timeout)
        // сканировать другие поля...
        if err != nil {
            return nil, err
        }
        calls = append(calls, t)
    }

    return calls, nil
}


func processCall(db *sql.DB, call *Call, cpsDuration time.Duration) {
    log.Println("processing call dst/branch!!!!:", call.dst, call.UID)
    _, exists := CallsFind(callEntry, call.dst)
    time.Sleep(cpsDuration)
    if ! exists {
        CallsAdd(callEntry, call.dst, CallEntryItem{ID: call.ID, Src: getRandomSrc(call.src), Dst: call.dst, Lead: call.Lead, State: 0, UID: call.UID, GeoID: call.GeoID})
		log.Println("calling: ",callEntry[call.dst].Dst)

	    err := processCallMinusRetry(db, call)
		if err != nil {
			log.Println("Error fetching calls:", err)
	    }
   		log.Println("callEntry: ",callEntry)
   		log.Println("call: ",call)
   		callChan <- *call
    } else {
		log.Println("call already exists: ", call.dst, callEntry[call.dst].State)
	}
//	log.Println("callEntry: ",callEntry)
//	CallsDelete(callEntry, call.dst)
//	log.Println("callEntry after delete: ",callEntry)
//    return errors.New("okay")
     return
}

func processCallMinusRetry(db *sql.DB, call *Call) error{
    // Выполняем запрос на обновление данных в базе
    _, err := db.Exec("UPDATE calls_queue set retries_left=retries_left-1, retry_time=date_add(now(), interval ? second) WHERE id = ?", call.RetryTime, call.ID)
    if err != nil {
        return err
    }
    return nil
}

func processCallRetryTime(db *sql.DB, call *Call) error{
    // Выполняем запрос на обновление данных в базе
    _, err := db.Exec("UPDATE calls_queue set retry_time=date_add(now() + interval ? second) WHERE id = ?", call.RetryTime, call.ID)
    if err != nil {
        return err
    }
    return nil
}


func isCurrentTimeBetween(startTimeStr, endTimeStr string) bool {
	parseToHourMin := func(input string) (int, int, error) {
		parts := strings.Split(input, ":")
		if len(parts) < 2 {
			return 0, 0, fmt.Errorf("invalid time format")
		}
		hour, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, err
		}
		minute, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, err
		}
		return hour, minute, nil
	}

	currentTime := time.Now()
	hourStart, minStart, errStart := parseToHourMin(startTimeStr)
	hourEnd, minEnd, errEnd := parseToHourMin(endTimeStr)

	// If parsing fails or start & end times are both 0, return true
	if errStart != nil || errEnd != nil || (hourStart == 0 && minStart == 0 && hourEnd == 0 && minEnd == 0) {
		return true
	}

	startTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), hourStart, minStart, 0, 0, currentTime.Location())
	endTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), hourEnd, minEnd, 0, 0, currentTime.Location())

	// Handle crossing midnight
	if startTime.After(endTime) {
		return currentTime.After(startTime) || currentTime.Before(endTime)
	}

	return currentTime.After(startTime) && currentTime.Before(endTime)
}

func hashPassword(password string) (string) {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes)
}


func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func isValidUser(username, password string) bool {	
	var storedPasswordHash string

	err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&storedPasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		panic(err)
	}

//	log.Println("check pass: ", password, storedPasswordHash)

	return checkPasswordHash(password, storedPasswordHash)
}

func countForUID(uid string, entries map[string]CallEntryItem) map[string]int {
    counts := map[string]int{"state1": 0, "state2": 0}

    for _, entry := range entries {
        if entry.UID == uid {
            if entry.State == 1 {
                counts["state1"]++
            } else if entry.State == 2 {
                counts["state2"]++
            }
        }
    }
    return counts
}


func GetReportData(startDate, endDate string) ([]map[string]interface{}, error) {
    query := `
        SELECT c.type, c.country, c.provider, TRUNCATE(ROUND(sum(c.billsec)/60,0),0) as total_min, sum(IF(c.billsec>0,1,0)) as answered_calls, count(*) as total_calls
        FROM cdr c
        WHERE c.calldate BETWEEN ? AND ? AND c.channel LIKE 'Local%' AND type='out-ivr' AND dst <> '1000'
        GROUP BY c.type, c.country, c.provider
    `

    rows, err := db.Query(query, startDate, endDate)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var results []map[string]interface{}
    
    cols, err := rows.Columns()
    if err != nil {
        return nil, err
    }

    for rows.Next() {
        columns := make([]interface{}, len(cols))
        columnPointers := make([]interface{}, len(cols))
        for i := range columns {
            columnPointers[i] = &columns[i]
        }

        if err := rows.Scan(columnPointers...); err != nil {
            return nil, err
        }

	m := make(map[string]interface{})
	for i, colName := range cols {
	    val := columnPointers[i].(*interface{})
	    byteValue, ok := (*val).([]byte)
	    if ok {
	        m[colName] = string(byteValue)
	    } else {
	        m[colName] = *val
	    }
	}
        results = append(results, m)
    }

    return results, nil
}

