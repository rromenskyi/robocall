package main

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/heltonmarx/goami/ami"
)

type AMIResponse map[string]string

type Asterisk struct {
	socket *ami.Socket
	uuid   string

	events chan AMIResponse
	stop   chan struct{}
	wg     sync.WaitGroup
}

type QueueSummary struct {
	ActionID        string `map:"ActionID"`
	Available       int    `map:"Available"`
	Callers         int    `map:"Callers"`
	Event           string `map:"Event"`
	HoldTime        int    `map:"HoldTime"`
	LoggedIn        int    `map:"LoggedIn"`
	LongestHoldTime int    `map:"LongestHoldTime"`
	Queue           string `map:"Queue"`
	TalkTime        int    `map:"TalkTime"`
}

var queuesMap = make(map[string]QueueSummary) //мапа очередей

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

// NewAsterisk initializes an AMI socket with the requested event mask.
func NewAsterisk(ctx context.Context, host string, username string, secret string, eventMask string) (*Asterisk, error) {
	_ = ctx

	socket, err := ami.NewSocket(host)
	if err != nil {
		return nil, err
	}
	if _, err := ami.Connect(socket); err != nil {
		return nil, err
	}
	uuid, err := ami.GetUUID()
	if err != nil {
		return nil, err
	}
	ok, err := ami.Login(socket, username, secret, eventMask, uuid)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("ami login failed")
	}
	as := &Asterisk{
		socket: socket,
		uuid:   uuid,
		events: make(chan AMIResponse),
		stop:   make(chan struct{}),
	}
	return as, nil
}

// Logoff closes the current session with AMI.
func (as *Asterisk) Logoff(ctx context.Context) error {
	_ = ctx
	close(as.stop)
	as.wg.Wait()

	_, err := ami.Logoff(as.socket, as.uuid)
	return err
}

// Events returns an channel with events received from AMI.
func (as *Asterisk) Events() <-chan AMIResponse {
	return as.events
}

func sendAMICommand(socket *ami.Socket, command []string) error {
	for _, part := range command {
		if err := socket.Send("%s", part); err != nil {
			return err
		}
	}
	return nil
}

func decodeAMIMessage(socket *ami.Socket) (AMIResponse, error) {
	raw, err := socket.Recv()
	if err != nil {
		return nil, err
	}
	if raw == "" {
		return nil, errors.New("received empty AMI message")
	}

	message := make(AMIResponse)
	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		message[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	return message, scanner.Err()
}

func getAMIMessageList(socket *ami.Socket, command []string, actionID, event, complete string) ([]AMIResponse, error) {
	if err := sendAMICommand(socket, command); err != nil {
		return nil, err
	}

	list := make([]AMIResponse, 0)
	state := 0

	for {
		message, err := decodeAMIMessage(socket)
		if err != nil {
			return nil, err
		}
		if messageActionID := message["ActionID"]; messageActionID != "" && messageActionID != actionID {
			return nil, errors.New("invalid ActionID")
		}

		switch state {
		case 0:
			if message["Response"] != "Success" {
				return nil, errors.New(message["Message"])
			}
			state = 1
		case 1:
			if message["Event"] == complete {
				return list, nil
			}
			if message["Event"] == event {
				list = append(list, message)
			}
		}
	}
}

func (as *Asterisk) QueueSummary(ctx context.Context) ([]AMIResponse, error) {
	_ = ctx

	command := []string{
		"Action: QueueSummary",
		"\r\nActionID: ",
		as.uuid,
		"\r\n\r\n",
	}

	return getAMIMessageList(as.socket, command, as.uuid, "QueueSummary", "QueueSummaryComplete")
}

func (as *Asterisk) Originate(ctx context.Context, data ami.OriginateData) (AMIResponse, error) {
	_ = ctx
	return ami.Originate(as.socket, as.uuid, data)
}

func isTransientQueueSummaryError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.TrimSpace(strings.ToLower(err.Error()))
	return message == "" || strings.Contains(message, "broken pipe")
}

func queueSummaryWithRetry(config *AMIConfig, asterisk *Asterisk) (*Asterisk, []AMIResponse, error) {
	current := asterisk
	var lastErr error

	for attempt := 0; attempt < 2; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		summary, err := current.QueueSummary(ctx)
		cancel()
		if err == nil {
			return current, summary, nil
		}

		lastErr = err
		if !isTransientQueueSummaryError(err) || attempt == 1 {
			return current, nil, err
		}

		if config != nil && strings.Contains(strings.ToLower(err.Error()), "broken pipe") {
			reconnected, reconnectErr := NewAsterisk(context.Background(), config.Host, config.User, config.Password, "off")
			if reconnectErr != nil {
				return current, nil, reconnectErr
			}
			current = reconnected
		}

		time.Sleep(200 * time.Millisecond)
	}

	return current, nil, lastErr
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
			events, err := ami.GetEvents(as.socket)
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
			var summary []AMIResponse
			var err error

			asterisk, summary, err = queueSummaryWithRetry(nil, asterisk)
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

	asterisk, err := NewAsterisk(ctx, config.Host, config.User, config.Password, "off")
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
			summaryClient, summary, err := queueSummaryWithRetry(&config, asterisk)
			asterisk = summaryClient
			if err != nil {
				log.Println("Error getting Queue Summary:", err)
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
							err = UpdateQueueSummary(db, qs, "COM")

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

func handleEvents(eventsChan <-chan AMIResponse) {
	for {
		event, ok := <-eventsChan
		if !ok {
			// Канал закрыт
			return
		}

		// Обработка события здесь
		if eventType, exists := event["Event"]; exists && eventType != "" {
			switch eventType {
			case "Newchannel":
				callID := event["Uniqueid"]
				number := strings.Split(strings.Split(event["Channel"], "@")[0], "/")[1]
				log.Println("Набор номера ", number, " с ID", callID)
				entry, exists := CallsFind(callEntry, number)
				if exists {
					entry.State = 1
					CallsAdd(callEntry, number, entry)
					log.Println("Number: ", entry.Dst, " connect state: ", entry.State)
				}

			case "BridgeEnter":
				callID := event["Uniqueid"]
				number := strings.Split(strings.Split(event["Channel"], "@")[0], "/")[1]
				log.Println("Звонок на ", number, " с ID", callID, "в состоянии поднята трубка")
				entry, exists := CallsFind(callEntry, number)
				if exists {
					entry.State = 2
					CallsAdd(callEntry, number, entry)
					log.Println("Number: ", entry.Dst, " connect state: ", entry.State)
				}
			case "Hangup":
				callID := event["Uniqueid"]
				cause := event["Cause"]
				number := strings.Split(strings.Split(event["Channel"], "@")[0], "/")[1]
				log.Println("Звонок на ", number, " с ID", callID, "был завершен по причине", cause)
				entry, exists := CallsFind(callEntry, number)
				if exists {
					entry.State = 3
					CallsAdd(callEntry, number, entry)
					log.Println("Number: ", entry.Dst, " disconnect state: ", entry.State)
					CallsDelete(callEntry, number)
					if entry.UID != "" {
						LimitsLineRelease(worklimitsMap, entry.UID)
					}
					if entry.GeoID != 0 {
						LimitsLineRelease(worklimitsMap, strconv.Itoa(entry.GeoID))
					}
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

	asterisk, err := NewAsterisk(ctx, config.Host, config.User, config.Password, "off")
	if err != nil {
		log.Fatal(err)
	}
	defer asterisk.Logoff(ctx)
	log.Printf("connected with asterisk\n")

	eventAsterisk, err := NewAsterisk(ctx, config.Host, config.User, config.Password, "system,call,all,user")
	if err != nil {
		log.Fatal(err)
	}
	defer eventAsterisk.Logoff(ctx)
	eventAsterisk.wg.Add(1)
	go eventAsterisk.run(ctx)
	go handleEvents(eventAsterisk.events)

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

			log.Println("Originate:", call)

			orig := &ami.OriginateData{
				Channel:  "Local/" + call.dst + "@auto-dialer/nj",
				Exten:    "1000",
				Callerid: callEntry[call.dst].Src,
				Context:  call.IVR,
				Priority: 1,
				Timeout:  call.Timeout,
				Variable: strings.Join([]string{"BATCH=" + call.UID, "NUM=" + call.dst, "CONT=" + call.IVR}, ","),
				Account:  call.Lead,
				Async:    "true",
			}

			_, err := asterisk.Originate(ctx, *orig) //// get response at depending on response limitsget/release!
			if err != nil {
				log.Println("Error during call originate:", err)
			}
			//	        case <-ctx.Done():
			//				close(callChan)
			//				return

		}
	}

}

func MapToQueueSummary(data map[string]string) (*QueueSummary, error) {
	qs := &QueueSummary{}
	// Для каждого поля структуры заполняем значение из словаря
	if val, ok := data["ActionID"]; ok {
		qs.ActionID = val
	}

	if val, ok := data["Queue"]; ok {
		qs.Queue = val
	}

	if val, ok := data["Available"]; ok {
		qs.Available, _ = strconv.Atoi(val)
	}

	if val, ok := data["Callers"]; ok {
		qs.Callers, _ = strconv.Atoi(val)
	}

	if val, ok := data["HoldTime"]; ok {
		qs.HoldTime, _ = strconv.Atoi(val)
	}

	if val, ok := data["LoggedIn"]; ok {
		qs.LoggedIn, _ = strconv.Atoi(val)
	}

	if val, ok := data["LongestHoldTime"]; ok {
		qs.LongestHoldTime, _ = strconv.Atoi(val)
	}

	if val, ok := data["TalkTime"]; ok {
		qs.TalkTime, _ = strconv.Atoi(val)
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
		queuesMap[qs.Queue] = QueueSummary{Queue: qs.Queue, Available: qs.Available, Callers: qs.Callers, HoldTime: qs.HoldTime, LoggedIn: qs.LoggedIn, LongestHoldTime: qs.LongestHoldTime, TalkTime: qs.TalkTime}
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
