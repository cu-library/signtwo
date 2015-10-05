// Copyright 2015 Carleton University Library All rights reserved.
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
	"github.com/cu-library/signtwo/db"
	"github.com/cu-library/signtwo/ldap"
	l "github.com/cu-library/signtwo/loglevel"
	"log"
	"os"
	"strings"
	"html/template"
	"net/http"
	"github.com/oxtoacart/bpool"
	"encoding/hex"
	jwt "github.com/dgrijalva/jwt-go"
	"time"

    "github.com/gorilla/mux"
	"github.com/gorilla/csrf"
)

const (
	// The prefix for all the environment variables
	EnvPrefix string = "SIGNTWO_"

	// The default address to serve from
	DefaultAddress string = ":8877"

	// The default log level
	DefaultLogLevel = "WARN"

	// The default ldap port
	DefaultLDAPPort = 636

	// The number of buffers to preallocate in the buffer pool
	DefaultBufferPoolSize = 128

	// The size of buffers to preallocate in the buffer pool
	DefaultBufferPoolAlloc = 100000

	// Default cookie name to store 
	DefaultCookieName = "session-token"

)

var (
	address  = flag.String("address", DefaultAddress, "Address the server will bind on.")
	logLevel = flag.String("loglevel", DefaultLogLevel, "The maximum log level which will be logged.\n"+
		                                                "        ERROR < WARN < INFO < DEBUG < TRACE\n"+
		                                                "        For example, TRACE will log everything,\n"+
		                                                "        INFO will log INFO, WARN, and ERROR messages.")
	databaseURL      = flag.String("dburl", "", "Database URL, eg: postgres://username:password@localhost/databasename")
	ldapServer       = flag.String("ldapserver", "", "The LDAP server address. Will always use LDAPS/SSL.")
	ldapPort         = flag.Int("ldapport", DefaultLDAPPort, "The LDAP server port.")
	ldapBindUsername = flag.String("ldapuser", "", "The username for the service account LDAP will use for the initial bind.")
	ldapBindPassword = flag.String("ldappass", "", "The password for the service account LDAP will use for the initial bind.")
	secretHex        = flag.String("secret", "", "A random string of hex characters, 192 characters long.\n"+
		                                         "       Generate using 'openssl rand -hex 160'" )		                      
	basepath         = flag.String("basepath", "", "A base bath that the application is served on. https://hostname.com/basepath/")

	// Templates inherit from base by cloning base and adding more content.
	baseTemplate = template.Must(template.ParseFiles("templates/base.tmpl"))
	homeTemplate = template.Must(template.Must(baseTemplate.Clone()).ParseFiles("templates/home.tmpl"))
	loginTemplate = template.Must(template.Must(baseTemplate.Clone()).ParseFiles("templates/login.tmpl"))
    fourOhFourTemplate = template.Must(template.Must(baseTemplate.Clone()).ParseFiles("templates/fourohfour.tmpl"))
	
    // Because templates need to be executed before we know if they'll cause an error, 
    // we store the output of the template execution in a buffer. This is the buffer
    // pool we grap buffers from. If there is no error, the data is then passed
    // to the client. 
	// https://godoc.org/github.com/oxtoacart/bpool#SizedBufferPool
	bufferPool = bpool.NewSizedBufferPool(DefaultBufferPoolSize, DefaultBufferPoolAlloc)

	jwtSecret []byte

)

func init() {

	// Set a custom usage message.
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

	// Parse the flags
	flag.Parse()

	// Parse and set the log level
	parsedLogLevel, err := l.ParseLogLevel(*logLevel)
	l.Set(parsedLogLevel)
	if err != nil {
		l.Log(l.WarnMessage, err)
	}

	// If there are any environment variables, they 
	// are used to set flags which were not passed in 
	// as a command line option.
	// Default < Environment variable < Command line option 
	overrideUnsetFlagsFromEnvironmentVariables()

	l.Log(l.InfoMessage, "Starting Signtwo")
	defer l.Log(l.InfoMessage, "Exiting, goodbye!")
	
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
	if *secretHex == "" {
		log.Fatal("FATAL: An secret is required. Generate using 'openssl rand -hex 160'")
	} 

	secret, err := hex.DecodeString(*secretHex)
	if err != nil || len(secret) != 160{ 
		log.Fatal("FATAL: An unable to parse secret. Generate using 'openssl rand -hex 160'")
	}

	csrfSecret := secret[:32]
	jwtSecret = secret[32:]

	CSRF := csrf.Protect(csrfSecret, csrf.Secure(false), csrf.FieldName("csrf-token"), csrf.CookieName("csrf-token")) //TODO REMOVE

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

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(fourOhFour) 
	r.Path("/").Methods("GET").HandlerFunc(homeHandler)	
	r.Path("/login").Methods("GET").HandlerFunc(loginGETHandler)
	r.Path("/login").Methods("POST").HandlerFunc(loginPOSTHandler)
	r.PathPrefix("/static/").Methods("GET").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
      
	log.Fatalf("FATAL: %v", http.ListenAndServe(*address, CSRF(r)))
	
}

func homeHandler(w http.ResponseWriter, r *http.Request) {		
	l.Log(l.TraceMessage, "Home Handler visited.")	


	renderTemplateOr500(w, homeTemplate, nil)
}

func loginGETHandler(w http.ResponseWriter, r *http.Request) {		
	l.Log(l.TraceMessage, "Login GET Handler visited.")	
	renderTemplateOr500(w, loginTemplate, map[string]interface{}{
        csrf.TemplateTag: customTokenField(r),
    })
}

func loginPOSTHandler(w http.ResponseWriter, r *http.Request) {		
	l.Log(l.TraceMessage, "Login POST Handler visited.")

	username := r.FormValue("username")

	// Create the token
    token := jwt.New(jwt.SigningMethodHS512)
    // Set some claims
    token.Claims["username"] = username
    token.Claims["exp"] = time.Now().Add(time.Minute * 5).Unix()
   
    // Sign and get the complete encoded token as a string
    tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "<html><head></head><body><pre>500 - Error while signing token</pre></body></html>")
		l.Log(l.ErrorMessage, "500!")
		return
	}

    l.Logf(l.TraceMessage, "JWT Token: %v", tokenString)
  
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func fourOhFour(w http.ResponseWriter, r *http.Request) {	
    l.Log(l.TraceMessage, "404 Handler visited.")		
	renderTemplateOr500CustomStatus(w, fourOhFourTemplate, nil, http.StatusNotFound)	
}

func renderTemplateOr500(w http.ResponseWriter, tmpl *template.Template, data map[string]interface{}) {
    renderTemplateOr500CustomStatus(w, tmpl, data, http.StatusOK)
}

func renderTemplateOr500CustomStatus(w http.ResponseWriter, tmpl *template.Template, data map[string]interface{}, status int) {
    // Create a buffer to temporarily write to and check if any errors were encounted.
    buf := bufferPool.Get()
    defer bufferPool.Put(buf)

    err := tmpl.Execute(buf, data)
    if err != nil {
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "<html><head></head><body><pre>500 - Template Error</pre></body></html>")
		l.Log(l.ErrorMessage, "500!")
		return
    }

    // Set the header and write the buffer to the http.ResponseWriter
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.WriteHeader(status)
    buf.WriteTo(w)
}

func customTokenField(r *http.Request) template.HTML {
	fragment := fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`,
		"csrf-token", csrf.Token(r))

	return template.HTML(fragment)
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
