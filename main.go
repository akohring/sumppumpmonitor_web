package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

const port = ":8080"

var scriptHome = ""

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, scriptHome+"/static/index.html")
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, scriptHome+"/static/favicon.ico")
}

func getAllPits(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(selectAllPits())
}

func getAllPitData(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(selectAllPitData())
}

func purgeHistoricPitLevels(w http.ResponseWriter, r *http.Request) {
	deleteHistoricPitLevels()
}

func (h *Hub) postPitLevel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	pitID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Error(err)
		return
	}

	level, err := strconv.ParseFloat(vars["level"], 64)
	if err != nil {
		log.Error(err)
		return
	}

	pitLevel := PitLevel{PitID: pitID, DateCreated: time.Now(), Level: level}
	insertPitLevel(pitLevel)

	select {
	case h.broadcast <- pitLevel:
	default:
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Hub) postPitHealth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pitID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Error(err)
		return
	}

	ok, err := strconv.ParseBool(vars["ok"])
	if err != nil {
		log.Error(err)
		return
	}

	pit := Pit{PitID: pitID, Healthy: ok, LastUpdated: time.Now()}
	updatePitHealth(pit)

	select {
	case h.broadcast <- pit:
	default:
	}
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(io.MultiWriter(os.Stdout, &lumberjack.Logger{
		Filename:   "/opt/sumppumpmonitor/log/web.log",
		MaxSize:    100, // megabytes
		MaxAge:     7,   //days
		MaxBackups: 7,
		Compress:   true, // disabled by default
	}))
	log.SetLevel(log.InfoLevel)

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	scriptHome = dir

	dbInit()
}

func main() {
	log.WithFields(log.Fields{
		"port": port,
	}).Info("Application Started")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	hub := newHub()
	go hub.run()

	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(scriptHome+"/static")))).Methods("GET")
	r.HandleFunc("/favicon.ico", faviconHandler).Methods("GET")
	r.HandleFunc("/pits", getAllPits).Methods("GET")
	r.HandleFunc("/pitlevels", getAllPitData).Methods("GET")
	r.HandleFunc("/pitlevels", purgeHistoricPitLevels).Methods("DELETE")
	r.HandleFunc("/pitlevel/{id}/{level}", hub.postPitLevel).Methods("POST")
	r.HandleFunc("/pithealth/{id}/{ok}", hub.postPitHealth).Methods("POST")
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) { serveWs(hub, w, r) })
	r.HandleFunc("/", rootHandler)

	http.Handle("/", r)

	go func() {
		if err := http.ListenAndServe(port, nil); err != nil {
			log.Fatal(err)
		}
	}()

	<-stop
	log.Info("Application stopped")
}
