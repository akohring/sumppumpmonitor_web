package main

import (
	"database/sql"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

func dbInit() error {
	db, err := dbConnect()
	if err != nil {
		return err
	}
	defer db.Close()

	exec(db, "CREATE TABLE IF NOT EXISTS pit (pitID INTEGER PRIMARY KEY, healthy INTEGER, lastUpdated timestamp default (strftime('%s', 'now')))")
	exec(db, "CREATE TABLE IF NOT EXISTS pitLevels (pitID INTEGER, dateCreated timestamp default (strftime('%s', 'now')), level REAL, PRIMARY KEY (pitID, dateCreated))")
	return nil
}

func dbConnect() (*sql.DB, error) {
	return sql.Open("sqlite3", scriptHome+"/sumppumpmonitor.db")
}

func exec(db *sql.DB, sql string) error {
	statement, err := db.Prepare(sql)
	if err != nil {
		return err
	}
	statement.Exec()
	return nil
}

func selectAllPits() []*Pit {
	var pits = []*Pit{}

	db, err := dbConnect()
	if err != nil {
		log.Error(err)
		return pits
	}
	defer db.Close()

	rows, err := db.Query("SELECT pitID, healthy, lastUpdated FROM pit")
	if err != nil {
		log.Error(err)
		return pits
	}

	for rows.Next() {
		var pitID int
		var healthy bool
		var lastUpdated time.Time
		rows.Scan(&pitID, &healthy, &lastUpdated)
		var pit = &Pit{PitID: pitID, Healthy: healthy, LastUpdated: lastUpdated}
		pits = append(pits, pit)
	}
	return pits
}

func selectAllPitData() []*Pit {
	var pits = []*Pit{}

	db, err := dbConnect()
	if err != nil {
		log.Error(err)
		return pits
	}
	defer db.Close()

	rows, err := db.Query("SELECT p.pitID, p.healthy, p.lastUpdated, pl.dateCreated, pl.level FROM pit p, pitLevels pl WHERE p.pitID = pl.pitID")
	if err != nil {
		log.Error(err)
		return pits
	}

	for rows.Next() {
		var pitID int
		var healthy bool
		var lastUpdated time.Time
		var dateCreated time.Time
		var level float64
		rows.Scan(&pitID, &healthy, &lastUpdated, &dateCreated, &level)

		var pit *Pit
		for _, p := range pits {
			if p.PitID == pitID {
				pit = p
			}
		}
		if pit == nil {
			pit = &Pit{PitID: pitID, Healthy: healthy, LastUpdated: lastUpdated}
			pits = append(pits, pit)
		}
		pit.PitLevels = append(pit.PitLevels, PitLevel{PitID: pitID, DateCreated: dateCreated, Level: level})
	}
	return pits
}

func deleteHistoricPitLevels() {
	db, err := dbConnect()
	if err != nil {
		log.Error(err)
		return
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM pitLevels WHERE date(dateCreated) < date('now','-7 days')")
	if err != nil {
		log.Error(err)
		return
	}
}

func insertPitLevel(pitLevel PitLevel) {
	err := createPitIfNotExists(pitLevel.PitID)
	if err != nil {
		log.Error(err)
		return
	}

	db, err := dbConnect()
	if err != nil {
		log.Error(err)
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO pitLevels(pitID, dateCreated, level) VALUES(?,?,?)")
	if err != nil {
		log.Error(err)
		return
	}

	_, err = stmt.Exec(pitLevel.PitID, pitLevel.DateCreated, pitLevel.Level)
	if err != nil {
		log.Error(err)
		return
	}
}

func updatePitHealth(pit Pit) {
	err := createPitIfNotExists(pit.PitID)
	if err != nil {
		log.Error(err)
		return
	}

	db, err := dbConnect()
	if err != nil {
		log.Error(err)
		return
	}
	defer db.Close()

	_, err = db.Exec("UPDATE pit SET healthy = ?, lastUpdated = ? WHERE pitID = ?", pit.Healthy, pit.LastUpdated, pit.PitID)
	if err != nil {
		log.Error(err)
		return
	}
}

func rowExists(query string, args ...interface{}) (bool, error) {
	db, err := dbConnect()
	if err != nil {
		return false, err
	}
	defer db.Close()

	var exists bool
	query = fmt.Sprintf("SELECT exists (%s)", query)
	err = db.QueryRow(query, args...).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}

func createPitIfNotExists(pitID int) error {
	rowExists, err := rowExists("SELECT pitID FROM pit WHERE pitID=?", pitID)
	if err != nil {
		return err
	}
	if !rowExists {
		_, err = createPit(pitID)
	}
	return err
}

func createPit(pitID int) (Pit, error) {
	db, err := dbConnect()
	defer db.Close()

	var pit = Pit{PitID: pitID, Healthy: false, LastUpdated: time.Now()}
	stmt, err := db.Prepare("INSERT INTO pit(pitID, healthy, lastUpdated) VALUES(?,?,?)")
	if err != nil {
		return pit, err
	}
	_, err = stmt.Exec(pit.PitID, pit.Healthy, pit.LastUpdated)
	return pit, err
}
