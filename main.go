package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type TimeUpdate struct {
	Number int    `json:"number"`
	Time   string `json:"time"`
}

type Update struct {
	Event string     `json:"event"`
	Data  TimeUpdate `json:"data"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/time", handleTime)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleTime(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	done := make(chan bool)
	go func() {
		<-r.Context().Done()
		close(done)
	}()

	for i := 0; i < 10; i++ {
		select {
		case <-done:
			return
		default:
			update := Update{Event: "update time", Data: TimeUpdate{Number: i, Time: time.Now().Format(time.RFC1123)}}
			data, err := json.Marshal(update)
			if err != nil {
				log.Printf("Error marshalling JSON: %v", err)
				continue
			}
			fmt.Fprintf(w, "%s\n\n", data)
			flusher.Flush()
			time.Sleep(2 * time.Second)
		}
	}

	close := Update{Event: "close time", Data: TimeUpdate{Number: -1, Time: time.Now().Format(time.RFC1123)}}
	data, _ := json.Marshal(close)
	fmt.Fprintf(w, "%s\n", data)
	flusher.Flush()
}
