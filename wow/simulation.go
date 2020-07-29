package wow

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"
)

//curl -X POST -d "config=$(cat simple.txt)" 127.0.0.1:8080/api/sim

//Queue our queue of sims
var Queue = make(chan SimQueue, 5000)

//SimQueue for the storage of w and config
type SimQueue struct {
	Config   string
	Response chan string
}

//RunCalculation return a calculation throu websocket
func RunCalculation(configFile string, c *websocket.Conn) {
	var ret = make(chan string, 100)
	Queue <- SimQueue{Config: configFile, Response: ret}
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case t, ok := <-ret:
			if !ok {
				return
			}
			err := c.WriteMessage(websocket.TextMessage, []byte(t))
			if err != nil {
				fmt.Printf("Error : %s\n", err)
				return
			}
		case <-tick:
			err := c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Queue position : %d", len(Queue))))
			if err != nil {
				fmt.Printf("Error : %s\n", err)
				return
			}
		}
	}
}

//RunQueue to process queue
func RunQueue() {
	for {
		q := <-Queue
		sim(q)
	}
}

func sim(q SimQueue) {
	fmt.Printf("Processing file : %s\n", q.Config)
	cmd := exec.Command("../simc/engine/simc", q.Config)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		return
	}
	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			q.Response <- scanner.Text()
		}
		close(q.Response)
	}()

	err = cmd.Run()
	if err != nil {
		fmt.Println("Error : ", err.Error())
	}
}
