package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"wowsim/wow"
)

//----------- web handlers

func web() {
	log.Println("Running web server")
	r := mux.NewRouter()
	r.Use(cacheMiddleware)

	r.HandleFunc("/api/sim", runSim)
	r.HandleFunc("/api/wssim", wsSim)
	r.HandleFunc("/api/wow/character/{region:[a-z]+}/{realm:[a-z]+}/{character:[a-z]+}", webCharacter)
	r.HandleFunc("/api/wow/character/media/{region:[a-z]+}/{realm:[a-z]+}/{character:[a-z]+}", webCharacterMedia)
	r.HandleFunc("/api/wow/character/appearance/{region:[a-z]+}/{realm:[a-z]+}/{character:[a-z]+}", webCharacterAppearance)

	r.HandleFunc("/api/wow/item/{region:[a-z]+}/{id:[0-9]+}", webItem)
	r.HandleFunc("/api/wow/item/media/{region:[a-z]+}/{id:[0-9]+}", webItemMedia)
	http.ListenAndServe(":8080", r)
}

func webCharacter(w http.ResponseWriter, req *http.Request) {
	setupResponse(&w, req)
	vars := mux.Vars(req)
	ret, _ := json.Marshal(wow.GetCharacterEquipment(vars["region"], vars["realm"], vars["character"]))
	writeCache(req.URL.Path, ret)
	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}
func webCharacterAppearance(w http.ResponseWriter, req *http.Request) {
	setupResponse(&w, req)
	vars := mux.Vars(req)
	ret, _ := json.Marshal(wow.GetCharacterAppearance(vars["region"], vars["realm"], vars["character"]))
	writeCache(req.URL.Path, ret)
	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}
func webCharacterMedia(w http.ResponseWriter, req *http.Request) {
	setupResponse(&w, req)
	vars := mux.Vars(req)
	ret, _ := json.Marshal(wow.GetCharacterMedia(vars["region"], vars["realm"], vars["character"]))
	writeCache(req.URL.Path, ret)
	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}
func webItem(w http.ResponseWriter, req *http.Request) {
	setupResponse(&w, req)
	vars := mux.Vars(req)
	ret, _ := json.Marshal(wow.GetItem(vars["region"], vars["id"]))
	writeCache(req.URL.Path, ret)
	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}
func webItemMedia(w http.ResponseWriter, req *http.Request) {
	setupResponse(&w, req)
	vars := mux.Vars(req)
	ret, _ := json.Marshal(wow.GetItemMedia(vars["region"], vars["id"]))
	writeCache(req.URL.Path, ret)
	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

//----------- calculation webhandler

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	}, Subprotocols: []string{"v10.stomp", "v11.stomp", "v12.stomp"},
} // use default options

func wsSim(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	c, err := upgrader.Upgrade(w, r, nil)
	c.SetReadDeadline(time.Time{})
	c.SetWriteDeadline(time.Time{})
	if err != nil {
		fmt.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		messageType, tmpMsg, err := c.NextReader()
		if err != nil {
			c.Close()
			break
		}
		msg, _ := ioutil.ReadAll(tmpMsg)
		//messageType, msg, err := c.ReadMessage()
		configFile := fmt.Sprintf("configs/config-%s.txt", time.Now().Format("20060102.15.04.05.000000"))
		if messageType == websocket.TextMessage && err == nil {
			tmpConfig, _ := strconv.Unquote(string(msg))
			if len(tmpConfig) > 0 {
				ioutil.WriteFile(configFile, []byte(tmpConfig), 0644)
				go wow.RunCalculation(configFile, c)
			}
		}
	}
}

func runSim(w http.ResponseWriter, req *http.Request) {
	setupResponse(&w, req)
	switch req.Method {
	case "POST":
		if err := req.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		config := req.FormValue(`config`)
		if len(req.Form) == 0 || config == "" {
			fmt.Fprintf(w, "Missing configuration")
			return
		}

		configFile := fmt.Sprintf("configs/config-%s.txt", time.Now().Format("20060102.15.04.05.000000"))
		ioutil.WriteFile(configFile, []byte(config), 0644)

		var ret = make(chan string, 100)

		wow.Queue <- wow.SimQueue{Config: configFile, Response: ret}

		fmt.Fprintf(w, "Queue position : %d\n", len(wow.Queue))

		tick := time.Tick(5 * time.Second)
	loop:
		for {
			select {
			case t, ok := <-ret:
				if !ok {
					break loop
				}
				fmt.Fprintf(w, "%s\n", t)
				tick = nil //we are no longer in queue ...
			case <-tick:
				fmt.Fprintf(w, "Queue position : %d\n", len(wow.Queue))
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
		//TODO: check content ... only items
		//TODO: retreat data for json return
	}
}

//----------- utils
func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

//To cache some request and avoid requesting again
func cacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setupResponse(&w, r)
		cachedFile := "./cache" + r.URL.Path
		if strings.HasPrefix(r.URL.Path, "/api/wow") && fileExists(cachedFile) {
			w.Header().Set("Content-Type", "application/json")
			http.ServeFile(w, r, cachedFile)

			//file, err := os.Open(cachedFile)
			//if err != nil {
			//	log.Println("Problem with file : ", err.Error())
			//}
			//defer file.Close()
			//w.Header().Set("Content-Type", "application/json")
			//l, err := io.Copy(w, file)
			//if err != nil {
			//	log.Println("Problem writting response : ", err.Error())
			//}
			//log.Println("Hit cache : ", cachedFile, "Wrote : ", l)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeCache(path string, response []byte) {
	elems := strings.Split(path, "/")
	path = "./cache" + strings.Join(elems[:len(elems)-1], "/")
	file := path + "/" + elems[len(elems)-1]

	log.Printf("Creating path : %s and file %s\n", path, file)

	os.MkdirAll(path, os.ModePerm)
	ioutil.WriteFile(file, response, os.ModePerm)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func main() {
	go wow.RunQueue()
	web()
}
