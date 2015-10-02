// Copyright 2015 Kevin Bowrin All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Provides database models and persistance 
// for the signtwo web application. 
package db

import (
	"database/sql"
	_ "github.com/lib/pq"
)

var db *sql.DB

func Connect(databaseURL string) error {
	var err error
	db, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return err
	}
	return nil
}

func Close() {
	db.Close()
}

func (agreement *Agreement) Store() int64 error {

	err := db.Ping()
	if err != nil {
		return 0, err
	}

	var agreementID int64


	// Create a new agreement
	if agreement.ID == 0 {


	} 
    // Update an existing agreement
	else {



	}
	
}