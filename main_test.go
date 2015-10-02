// Copyright 2015 Kevin Bowrin All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	l "github.com/cu-library/signtwo/loglevel"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func init() {
	l.Set(l.ErrorMessage)
	log.SetOutput(ioutil.Discard)
}

func TestOverrideUnsetFlagsFromEnvVars(t *testing.T) {

	oldAddress := os.Getenv(EnvPrefix + "ADDRESS")
	defer os.Setenv(EnvPrefix+"ADDRESS", oldAddress)
	newAddress := ":7654"
	os.Setenv(EnvPrefix+"ADDRESS", newAddress)

	overrideUnsetFlagsFromEnvironmentVariables()

	if *address != ":7654" {
		t.Error("Didn't override unset flags.")
	}

}
