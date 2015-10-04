// Copyright 2015 Carleton University Library All rights reserved.
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
