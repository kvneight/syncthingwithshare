// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

// +build integration

package integration

import (
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReset(t *testing.T) {
	// Clean and start a syncthing instance

	log.Println("Cleaning...")
	err := removeAll("s1", "h1/index*")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("s1", 0755); err != nil {
		t.Fatal(err)
	}

	log.Println("Creating files...")
	size := createFiles(t)

	p := startInstance(t, 1)
	defer checkedStop(t, p)

	m, err := p.Model("default")
	if err != nil {
		t.Fatal(err)
	}
	expected := size
	if m.LocalFiles != expected {
		t.Fatalf("Incorrect number of files after initial scan, %d != %d", m.LocalFiles, expected)
	}

	// Clear all files but restore the folder marker
	log.Println("Cleaning...")
	err = removeAll("s1")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("s1", 0755); err != nil {
		t.Fatal(err)
	}
	if fd, err := os.Create("s1/.stfolder"); err != nil {
		t.Fatal(err)
	} else {
		fd.Close()
	}

	// Reset indexes of an invalid folder
	log.Println("Reset invalid folder")
	_, err = p.Post("/rest/system/reset?folder=invalid", nil)
	if err == nil {
		t.Fatalf("Cannot reset indexes of an invalid folder")
	}

	// Reset indexes of the default folder
	log.Println("Reset indexes of default folder")
	_, err = p.Post("/rest/system/reset?folder=default", nil)
	if err != nil {
		t.Fatal("Failed to reset indexes of the default folder:", err)
	}

	// Syncthing restarts on reset. But we set STNORESTART=1 for the tests. So
	// we wait for it to exit, then do a stop so the rc.Process is happy and
	// restart it again.
	time.Sleep(time.Second)
	checkedStop(t, p)
	p = startInstance(t, 1)

	m, err = p.Model("default")
	if err != nil {
		t.Fatal(err)
	}
	expected = 0
	if m.LocalFiles != expected {
		t.Fatalf("Incorrect number of files after initial scan, %d != %d", m.LocalFiles, expected)
	}

	// Recreate the files and scan
	log.Println("Creating files...")
	size = createFiles(t)

	if err := p.Rescan("default"); err != nil {
		t.Fatal(err)
	}

	// Verify that we see them
	m, err = p.Model("default")
	if err != nil {
		t.Fatal(err)
	}
	expected = size
	if m.LocalFiles != expected {
		t.Fatalf("Incorrect number of files after second creation phase, %d != %d", m.LocalFiles, expected)
	}

	// Reset all indexes
	log.Println("Reset DB...")
	_, err = p.Post("/rest/system/reset?folder=default", nil)
	if err != nil {
		t.Fatalf("Failed to reset indexes", err)
	}

	// Syncthing restarts on reset. But we set STNORESTART=1 for the tests. So
	// we wait for it to exit, then do a stop so the rc.Process is happy and
	// restart it again.
	time.Sleep(time.Second)
	checkedStop(t, p)

	p = startInstance(t, 1)
	defer checkedStop(t, p)

	m, err = p.Model("default")
	if err != nil {
		t.Fatal(err)
	}
	expected = size
	if m.LocalFiles != expected {
		t.Fatalf("Incorrect number of files after initial scan, %d != %d", m.LocalFiles, expected)
	}
}

func createFiles(t *testing.T) int {
	// Create eight empty files and directories
	files := []string{"f1", "f2", "f3", "f4", "f11", "f12", "f13", "f14"}
	dirs := []string{"d1", "d2", "d3", "d4", "d11", "d12", "d13", "d14"}
	all := append(files, dirs...)

	for _, file := range files {
		fd, err := os.Create(filepath.Join("s1", file))
		if err != nil {
			t.Fatal(err)
		}
		fd.Close()
	}

	for _, dir := range dirs {
		err := os.Mkdir(filepath.Join("s1", dir), 0755)
		if err != nil {
			t.Fatal(err)
		}
	}

	return len(all)
}
