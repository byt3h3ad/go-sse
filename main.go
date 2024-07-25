package main

import (
	"context"
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

	timeChannel := make(chan []byte)
	go sendTimeUpdates(r.Context(), timeChannel)

	for t := range timeChannel {
		fmt.Fprintf(w, "%s\n\n", t)
		flusher.Flush()
	}
}

func sendTimeUpdates(ctx context.Context, timeChannel chan<- []byte) {
	ticker := time.NewTicker(2 * time.Second)

	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			update := Update{Event: "update time", Data: TimeUpdate{Number: i + 1, Time: time.Now().Format(time.RFC1123)}}
			data, err := json.MarshalIndent(update, "", "  ")
			if err != nil {
				log.Printf("Error marshalling JSON: %v", err)
				continue
			}
			timeChannel <- data
		}
	}

	ticker.Stop()
	close(timeChannel)
	fmt.Println("Finished")
}
