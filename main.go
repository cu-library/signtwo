// Copyright 2015 Kevin Bowrin All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

/*
Package signtwo provides a web application which allows
instutitions to handle user agreements, signatures, and protected
file access.
*/
package main

import (
	"flag"
	"fmt"
	l "github.com/cu-library/signtwo/loglevel"
	db "github.com/cu-library/signtwo/db"
	"log"
	"os"
	"strings"
)

const (
	//The prefix for all the curator environment variables
	EnvPrefix string = "SIGNTWO_"

	//The default address to serve from
	DefaultAddress string = ":8877"
)

var (
	address  = 	flag.String("address", DefaultAddress, "Address the server will bind on.")
	logLevel = 	flag.String("loglevel", "WARN", "The maximum log level which will be logged.\n"+
		"        ERROR < WARN < INFO < DEBUG < TRACE\n"+
		"        For example, TRACE will log everything,\n"+
		"        INFO will log INFO, WARN, and ERROR messages.")
	databaseURL = flag.String("dburl", "", "Database URL, eg: postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full")
)

func init() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "\nSigntwo: A user agreement web application.\nVersion: 0.0.1\n\n")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\n  The possible environment variables:")

		flag.VisitAll(func(f *flag.Flag) {
			uppercaseName := strings.ToUpper(f.Name)
			fmt.Fprintf(os.Stderr, "  %v%v\n", EnvPrefix, uppercaseName)
		})
	}
}

func main() {
	flag.Parse()
	l.Set(l.ParseLogLevel(*logLevel))
	overrideUnsetFlagsFromEnvironmentVariables()

	l.Log("Starting Signtwo", l.InfoMessage)
	if *databaseURL == "" {
		log.Fatal("FATAL: A database url is required.")
	} 

	err := db.Connect(*databaseURL)
	if err != nil {
		log.Fatal("FATAL: Could not connect to a database using the provided database url.")
	}
}

func overrideUnsetFlagsFromEnvironmentVariables() {
	listOfUnsetFlags := make(map[*flag.Flag]bool)

	//Ugly, but only way to get list of unset flags.
	flag.VisitAll(func(f *flag.Flag) { listOfUnsetFlags[f] = true })
	flag.Visit(func(f *flag.Flag) { delete(listOfUnsetFlags, f) })

	for k, _ := range listOfUnsetFlags {
		uppercaseName := strings.ToUpper(k.Name)
		environmentVariableName := fmt.Sprintf("%v%v", EnvPrefix, uppercaseName)
		environmentVariableValue := os.Getenv(environmentVariableName)
		if environmentVariableValue != "" {
			err := k.Value.Set(environmentVariableValue)
			if err != nil {
				log.Fatalf("FATAL: Unable to set configuration option %v from environment variable %v, "+
					"which has a value of \"%v\"",
					k.Name, environmentVariableName, environmentVariableValue)
			}
		}
	}
}
