// Copyright 2015 Kevin Bowrin All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"	
	"time"
)

type Agreement struct {
	ID int64
	Title string
	Description string
	Slug string
	Created time.Time
	Enabled bool
}

type Owner struct {
	ID int64
    OwnsAgreementID int64
    Username string
}

type AgreementText struct {
	ID int64
	BaseAgreementID int64
	Title sql.NullString
	Content string
	Created time.Time
	EnactmentDate time.Time
	ReplacesAgreementTextID int64
}

type UserType string

const (
	Student UserType = "Student"
	GraduateStudent UserType = "Graduate Student"
	Faculty UserType = "Faculty"
	Employee UserType = "Employee"
)

type Signature struct {
	ID int64
	SignedAgreementTextID int64
	Username string
	FirstName string
	LastName string
	UserType UserType
	Email string
	Department string
	BannerID int64
	SignedTimestampUTC time.Time
}  


