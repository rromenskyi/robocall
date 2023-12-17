package main

import (
	"context"
	"time"
//	"errors"
	"log"
	"sync"
	"strconv"
	"database/sql"
	"strings"

	"github.com/heltonmarx/goami/ami"
)

type Asterisk struct {
	socket *ami.Socket
	uuid   string

	events chan ami.Response
	stop   chan struct{}
	wg     sync.WaitGroup
}

type QueueSummary struct {
    ActionID         string `map:"ActionID"`
    Available        int    `map:"Available"`
    Callers          int    `map:"Callers"`
    Event            string `map:"Event"`
    HoldTime         int    `map:"HoldTime"`
    LoggedIn         int    `map:"LoggedIn"`
    LongestHoldTime  int    `map:"LongestHoldTime"`
    Queue            string `map:"Queue"`
    TalkTime         int    `map:"TalkTime"`
}

var queuesMap = make(map[string]QueueSummary)//мапа очередей

var callChan = make(chan Call)


func QueueAdd(targetMap map[string]QueueSummary, key string, item QueueSummary) {
    mutex.Lock()
    targetMap[key] = item
    mutex.Unlock()
}

func QueueFind(targetMap map[string]QueueSummary, key string) (QueueSummary, bool) {
    mutex.RLock()
    item, exists := targetMap[key]
    mutex.RUnlock()
    return item, exists
}

func QueueDelete(targetMap map[string]QueueSummary, key string) {
    mutex.Lock()
    delete(targetMap, key)
    mutex.Unlock()
}


// NewAsterisk initializes the AMI socket with a login and capturing the events.
func NewAsterisk(ctx context.Context, host string, username string, secret string) (*Asterisk, error) {
	socket, err := ami.NewSocket(ctx, host)
	if err != nil {
		return nil, err
	}
	uuid, err := ami.GetUUID()
	if err != nil {
		return nil, err
	}
	const events = "system,call,all,user"
	err = ami.Login(ctx, socket, username, secret, events, uuid)
	if err != nil {
		return nil, err
	}
	as := &Asterisk{
		socket: socket,
		uuid:   uuid,
		events: make(chan ami.Response),
		stop:   make(chan struct{}),
	}
	as.wg.Add(1)
	go as.run(ctx)
	return as, nil
}

// Logoff closes the current session with AMI.
func (as *Asterisk) Logoff(ctx context.Context) error {
	close(as.stop)
	as.wg.Wait()

	return ami.Logoff(ctx, as.socket, as.uuid)
}

// Events returns an channel with events received from AMI.
func (as *Asterisk) Events() <-chan ami.Response {
	return as.events
}

func (as *Asterisk) QueueSummary(ctx context.Context) ([]ami.Response, error) {
    // Передайте имя очереди, если необходимо, или просто используйте пустую строку для сводки по всем очередям
    return ami.QueueSummary(ctx, as.socket, as.uuid, "")
}


func (as *Asterisk) Originate(ctx context.Context, data ami.OriginateData) (ami.Response, error) {
    return ami.Originate(ctx, as.socket, as.uuid, data)
}


func (as *Asterisk) run(ctx context.Context) {
	defer as.wg.Done()
	for {
		select {
		case <-as.stop:
			return
		case <-ctx.Done():
			return
		default:
			events, err := ami.Events(ctx, as.socket)
			if err != nil {
				log.Printf("AMI events failed: %v\n", err)
				return
			}
			as.events <- events
		}
	}
}

func MonitorAsteriskQueue(asterisk *Asterisk) {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
            summary, err := asterisk.QueueSummary(ctx)
            cancel()  // Не забудьте завершить контекст после его использования
            if err != nil {
                log.Println("Error getting Queue Summary:", err)
            } else {
                for _, s := range summary {
                    log.Println(s)
                }
            }
        }
    }
}

func initAMI(config AMIConfig) {

	var qs *QueueSummary

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	asterisk, err := NewAsterisk(ctx, config.Host, config.User, config.Password)
	if err != nil {
		log.Fatal(err)
	}
	defer asterisk.Logoff(ctx)
	log.Printf("connected with asterisk\n")

ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
            summary, err := asterisk.QueueSummary(ctx)
            cancel()  // Не забудьте завершить контекст после его использования
				if err != nil {
				    log.Println("Error getting Queue Summary:", err)
				    //not working properly
				    if strings.Contains(err.Error(), "broken pipe") {
				        asterisk, err = NewAsterisk(ctx, config.Host, config.User, config.Password)
				        if err != nil {
				            log.Fatal(err)
				        }
				        summary, err = asterisk.QueueSummary(ctx)
				        if err != nil {
				            log.Println("Error getting Queue Summary after reconnect:", err)
				        }
				    } 
				}
            if summary != nil {
                for _, s := range summary {
//                    log.Println(s)
					qs, err = MapToQueueSummary(s)
					if err != nil {
						log.Fatal(err)
					} else {
//			                    log.Println(qs)
							if qs.Queue != "" {
							err = UpdateQueueSummary(db,qs,"COM")

							if err != nil {
								log.Fatal(err)
							}
						}
					}
                }
            }
        }
    }
}

func handleEvents(eventsChan <-chan ami.Response) {
    for {
        event, ok := <-eventsChan
        if !ok {
            // Канал закрыт
            return
        }

        // Обработка события здесь
        if events, exists := event["Event"]; exists && len(events) > 0 {
	        switch event["Event"][0] {
	        case "Newchannel":
	            callID := event["Uniqueid"][0]
	            number := strings.Split(strings.Split(event["Channel"][0], "@")[0], "/")[1]
	            log.Println("Набор номера ",number ," с ID", callID)
	            entry, exists := CallsFind(callEntry, number)
               if exists {
                    entry.State=1
                    callEntry[number] = entry
            		log.Println("Number: ",callEntry[number].Dst, " connect state: ", callEntry[number].State)
            	}

	        case "BridgeEnter":
	            callID := event["Uniqueid"][0]
	            number := strings.Split(strings.Split(event["Channel"][0], "@")[0], "/")[1]
	            log.Println("Звонок на ",number ," с ID", callID,"в состоянии поднята трубка")
	            entry, exists := CallsFind(callEntry, number)
               if exists {
                    entry.State=2
                    callEntry[number] = entry
            		log.Println("Number: ",callEntry[number].Dst, " connect state: ", callEntry[number].State)
            	}
	        case "Hangup":
	            callID := event["Uniqueid"][0]
	            cause := event["Cause"][0]
	            number := strings.Split(strings.Split(event["Channel"][0], "@")[0], "/")[1]
	            log.Println("Звонок на ",number ," с ID", callID, "был завершен по причине", cause)
	            entry, exists := CallsFind(callEntry, number)
               if exists {
                    entry.State=3
                    callEntry[number] = entry
             		log.Println("Number: ",callEntry[number].Dst, " disconnect state: ", callEntry[number].State)
            		CallsDelete(callEntry, number)
            		LimitsLineRelease(worklimitsMap,callEntry[number].UID)
					LimitsLineRelease(worklimitsMap,strconv.Itoa(callEntry[number].GeoID))
            	}

	        default:
	            // Обработка других типов событий
//	            log.Println("Received event:", event["Event"][0])
	        }
        }
    }
}

func initOAMI(config AMIConfig) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	asterisk, err := NewAsterisk(ctx, config.Host, config.User, config.Password)
	if err != nil {
		log.Fatal(err)
	}
	defer asterisk.Logoff(ctx)
	log.Printf("connected with asterisk\n")

    go handleEvents(asterisk.events)

//var callChan = make(chan Call)

//ticker := time.NewTicker(1 * time.Second)
//    defer ticker.Stop()

	    for {
	        select {
	        case call, ok := <-callChan:
	            if !ok {
	                // callChan был закрыт, завершаем горутину
	                return
	            }
	            
	            log.Println("Originate:",call)

	            orig := &ami.OriginateData{
	                Channel:  "Local/" + call.dst + "@auto-dialer/nj",
	                Exten:    "1000",
	                CallerID: callEntry[call.dst].Src,
	                Context:  call.IVR,
	                Priority: 1,
	                Timeout:  call.Timeout,
	                Variable: []string{"BATCH=" + call.UID, "NUM=" + call.dst, "CONT=" + call.IVR},
	                Account:  call.Lead,
	                Async:    "true",
	            }

	            _, err := asterisk.Originate(ctx, *orig)//// get response at depending on response limitsget/release!
	            if err != nil {
	                log.Println("Error during call originate:", err)
	            }
//	        case <-ctx.Done(): 
//				close(callChan)
//				return

	        }
	    }

}


func MapToQueueSummary(data map[string][]string) (*QueueSummary, error) {
    qs := &QueueSummary{}
    // Для каждого поля структуры заполняем значение из словаря
    if val, ok := data["ActionID"]; ok {
        qs.ActionID = val[0]
    }

    if val, ok := data["Queue"]; ok {
        qs.Queue = val[0]
    }


    if val, ok := data["Available"]; ok {
        qs.Available, _ = strconv.Atoi(val[0])
    }

    if val, ok := data["Callers"]; ok {
        qs.Callers, _ = strconv.Atoi(val[0])
    }

    if val, ok := data["HoldTime"]; ok {
        qs.HoldTime, _ = strconv.Atoi(val[0])
    }

    if val, ok := data["LoggedIn"]; ok {
        qs.LoggedIn, _ = strconv.Atoi(val[0])
    }

    if val, ok := data["LongestHoldTime"]; ok {
        qs.LongestHoldTime, _ = strconv.Atoi(val[0])
    }

    if val, ok := data["TalkTime"]; ok {
        qs.TalkTime, _ = strconv.Atoi(val[0])
    }




    return qs, nil
}

func UpdateQueueSummary(db *sql.DB, qs *QueueSummary, asteriskName string) error {
    // Проверка существует ли запись
    var id int
    err := db.QueryRow("SELECT id FROM queues WHERE Queue = ?", qs.Queue).Scan(&id)
    if err != nil {
        if err == sql.ErrNoRows { // Если записи нет, вставляем новую
            _, err := db.Exec(`INSERT INTO queues 
                                (Queue, Available, Callers, HoldTime, LoggedIn, LongestHoldTime, TalkTime, astrerisk) 
                                VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, 
                                qs.Queue, qs.Available, qs.Callers, qs.HoldTime, qs.LoggedIn, qs.LongestHoldTime, qs.TalkTime, asteriskName)
            return err
        }
        return err
    }
    _, exists := QueueFind(queuesMap, qs.Queue)
		           	    if exists {
		           	    	queuesMap[qs.Queue]=QueueSummary{Queue: qs.Queue, Available: qs.Available, Callers: qs.Callers, HoldTime: qs.HoldTime, LoggedIn: qs.LoggedIn, LongestHoldTime: qs.LongestHoldTime, TalkTime: qs.TalkTime}
						} else {
							QueueAdd(queuesMap, qs.Queue, QueueSummary{Queue: qs.Queue, Available: qs.Available, Callers: qs.Callers, HoldTime: qs.HoldTime, LoggedIn: qs.LoggedIn, LongestHoldTime: qs.LongestHoldTime, TalkTime: qs.TalkTime})
						}

    // Если запись существует, обновляем ее
    _, err = db.Exec(`UPDATE queues SET 
                      Available = ?, Callers = ?, HoldTime = ?, LoggedIn = ?, LongestHoldTime = ?, TalkTime = ?, astrerisk = ?
                      WHERE id = ?`, 
                      qs.Available, qs.Callers, qs.HoldTime, qs.LoggedIn, qs.LongestHoldTime, qs.TalkTime, asteriskName, id)
    return err
}