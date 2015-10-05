// Copyright 2015 Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Provides database models and persistance
// for the signtwo web application.
package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"errors"
	l "github.com/cu-library/signtwo/loglevel"
)

var db *sql.DB

func Connect(databaseURL string) error {

	l.Log(l.InfoMessage, "Connecting to database...")

	// This err ensures the db variable refers to the global one
	var err error
	db, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return err
	}

	// Can we access the database?
	err = db.Ping()
	if err != nil {
		return  err
	}

	// Check to see that the database has all the tables we need.
	
	rows, err := db.Query("SELECT table_name " +
                          "FROM information_schema.tables " +
                          "WHERE table_schema = 'webapp'; ")
	if err != nil {
	    return err
    }
    defer rows.Close()

    // Go doesn't have sets, per se. Fake with map.
    requiredTables := map[string]bool{"agreement":true, "owner":true, "agreement_text":true, "signature":true}

    for rows.Next() {
    	var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			return err
		}
		l.Logf(l.TraceMessage, "Found table %v", tableName)
		delete(requiredTables, tableName)	
	}

	if len(requiredTables) != 0{
		return errors.New("Unable to find all required tables in database, "+ 
			              "please check the database creation documentation.")
	}

	l.Log(l.InfoMessage, "Successful database connection.")
	return nil	
}

func Close() {
	l.Log(l.TraceMessage, "Closing database connection...")
	db.Close()
	l.Log(l.TraceMessage, "Successfully closed database connection.")
}

func (agreement *Agreement) Store() (int64, error) {

	// Can we access the database?
	err := db.Ping()
	if err != nil {
		return 0, err
	}

	// Begin a transaction.
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}

	var returnedAgreementID int64

	if agreement.ID == 0 {
		// Create a new agreement from a zero'd struct
		err = tx.QueryRow("INSERT INTO agreement(title,description,created,enabled) "+
			"VALUES($1,$2,$3,$4) "+
			"RETURNING id;",
			agreement.Title,
			agreement.Description,
			agreement.Created,
			agreement.Enabled).Scan(&returnedAgreementID)
	} else {
		// Update an existing agreement
		err = tx.QueryRow("UPDATE agreement "+
			"SET title = $1, description = $2, created = $3, enabled = $4 "+
			"WHERE id  = $5 "+
			"RETURNING id;",
			agreement.Title,
			agreement.Description,
			agreement.Created,
			agreement.Enabled,
			agreement.ID).Scan(&returnedAgreementID)
	}

	if err != nil {
		return 0, err
	}

	//Commit the transaction
	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return returnedAgreementID, nil
}
