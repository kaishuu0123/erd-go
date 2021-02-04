package main

import (
	"bytes"
	"strings"
	"testing"
	"text/template"
)

func TestErd_unquote(t *testing.T) {
	e := &Erd{}
	value := e.unquote("\"test\"")
	if value != "test" {
		t.Errorf("got: %v\nwant: %v", value, "test")
	}
}

func TestIntegration(t *testing.T) {
	contents := `
[Person]
*name
height
weight
+birth_location_id

[Location]
*id
city
state
country

Person *--1 Location
	`
	parser := &Parser{Buffer: contents}
	parser.Init()
	err := parser.Parse()
	if err != nil {
		t.Fatal(err)
	}

	parser.Execute()

	if parser.Erd.IsError {
		t.Fatal()
	}

	dot, _ := Asset("templates/dot.tmpl")
	tables, _ := Asset("templates/dot_tables.tmpl")
	relations, _ := Asset("templates/dot_relations.tmpl")
	templates := template.Must(template.New("").Funcs(template.FuncMap{"StringsJoin": strings.Join}).Parse(string(dot) + string(tables) + string(relations)))

	fd := bytes.NewBufferString("")
	if err := templates.ExecuteTemplate(fd, "dot", parser.Erd); err != nil {
		t.Fatal(err)
	}
}
