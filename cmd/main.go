package main

import (
	"log"
	"m/handlers"
	"m/scheduler"
	"net/http"

	_ "modernc.org/sqlite"
)

func main() {
	webDir := "web"
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", handlers.GetNextDateHandler)
	http.HandleFunc("/api/task", handlers.PostTaskHandler)
	// http.HandleFunc("/api/tasks", handlers.GetTaskHandler)
	//r.Delete("...", Handlers.DeleteTask)
	scheduler.Build()
	scheduler.Open()
	log.Fatal(http.ListenAndServe(":7540", nil))
}
