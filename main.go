// Copyright 2015 Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

/*
Package signtwo provides a web application which allows
instutitions to handle user agreements, signatures, and protected
file access.
*/
package main

/*
	// Create a new slug by lowercasing/ascii-ing the title
	var slugre = regexp.MustCompile("[^a-z0-9]+")
	slug := strings.Trim(slugre.ReplaceAllString(strings.ToLower(title), "-"), "-")
*/

import (
	"flag"
	"fmt"
	"github.com/cu-library/signtwo/db"
	"github.com/cu-library/signtwo/ldap"
	l "github.com/cu-library/signtwo/loglevel"
	"log"
	"os"
	"strings"
	"html/template"
	"net/http"
	"github.com/oxtoacart/bpool"
)

const (
	//The prefix for all the environment variables
	EnvPrefix string = "SIGNTWO_"

	//The default address to serve from
	DefaultAddress string = ":8877"
)

var (
	address  = flag.String("address", DefaultAddress, "Address the server will bind on.")
	logLevel = flag.String("loglevel", "WARN", "The maximum log level which will be logged.\n"+
		"        ERROR < WARN < INFO < DEBUG < TRACE\n"+
		"        For example, TRACE will log everything,\n"+
		"        INFO will log INFO, WARN, and ERROR messages.")
	databaseURL      = flag.String("dburl", "", "Database URL, eg: postgres://username:password@localhost/databasename")
	ldapServer       = flag.String("ldapserver", "", "The LDAP server address. Will always use LDAPS/SSL.")
	ldapPort         = flag.Int("ldapport", 636, "The LDAP server port.")
	ldapBindUsername = flag.String("ldapuser", "", "The username for the service account LDAP will use for the initial bind.")
	ldapBindPassword = flag.String("ldappass", "", "The password for the service account LDAP will use for the initial bind.")

	baseTemplate = template.Must(template.ParseFiles("templates/base.tmpl"))
	homeTemplate = template.Must(template.Must(baseTemplate.Clone()).ParseFiles("templates/home.tmpl"))

	bufferPool = bpool.NewSizedBufferPool(128, 100000)
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
	parsedLogLevel, err := l.ParseLogLevel(*logLevel)
	l.Set(parsedLogLevel)
	if err != nil {
		l.Log(err, l.WarnMessage)
	}

	overrideUnsetFlagsFromEnvironmentVariables()

	l.Log("Starting Signtwo", l.InfoMessage)
	defer l.Log("Exiting, goodbye!", l.InfoMessage)
	
	if *databaseURL == "" {
		log.Fatal("FATAL: A database url is required.")
	}
	if *ldapServer == "" {
		log.Fatal("FATAL: An LDAP server address is required.")
	}
	if *ldapBindUsername == "" {
		log.Fatal("FATAL: An LDAP service account username is required.")
	}
	if *ldapBindPassword == "" {
		log.Fatal("FATAL: An LDAP service account password is required.")
	}
	
	err = db.Connect(*databaseURL)
	if err != nil {
		log.Fatalf("FATAL: Could not connect to a database using the provided database url: %v", err)
	}
	defer db.Close()
	
	err = ldap.Connect(*ldapServer, *ldapPort, *ldapBindUsername, *ldapBindPassword, parsedLogLevel)
	if err != nil {
		log.Fatalf("FATAL: Could not connect and bind to LDAP using the provided information: %v", err)
	}
	defer ldap.Close()	

	http.HandleFunc("/", homeHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Fatalf("FATAL: %v", http.ListenAndServe(*address, nil))
	
}

func homeHandler(w http.ResponseWriter, r *http.Request) {		
	if r.URL.Path != "/" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "<html><head></head><body><pre>404 - Not Found</pre></body></html>")
		l.Log("404 Handler visited.", l.TraceMessage)
		return
	}

	l.Log("Home Handler visited.", l.TraceMessage)	
	renderTemplateOr500(w, homeTemplate, nil)
}

func renderTemplateOr500(w http.ResponseWriter, tmpl *template.Template, data map[string]interface{}) {
    // Create a buffer to temporarily write to and check if any errors were encounted.
    buf := bufferPool.Get()
    defer bufferPool.Put(buf)

    err := tmpl.Execute(buf, data)
    if err != nil {
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "<html><head></head><body><pre>500 - Template Error</pre></body></html>")
		l.Log("500!", l.ErrorMessage)
		return
    }

    // Set the header and write the buffer to the http.ResponseWriter
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    buf.WriteTo(w)
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
