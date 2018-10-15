package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Relation struct {
	LeftTableName      string
	LeftCardinality    string
	RightTableName     string
	RightCardinality   string
	RelationAttributes map[string]string
}

type Index struct {
	Title    string
	Columns  []string
	IsUnique bool
}

type Column struct {
	Title            string
	ColumnAttributes map[string]string
}

type Table struct {
	Title           string
	TableAttributes map[string]string
	Columns         []Column
	CurrentColumnId int
	PrimaryKeys     []int
}

type Title struct {
	Title           string
	TitleAttributes map[string]string
}

type Erd struct {
	Title            Title
	Tables           map[string]*Table
	Relations        []Relation
	CurrentRelation  Relation
	key              string
	value            string
	CurrentTableName string
	IsError          bool
	line             int
}

func (e *Erd) addTableTitle(t string) {
	t = strings.Trim(t, "\"")
	e.Tables[e.CurrentTableName].Title = t
}

func (e *Erd) ClearTableAndColumn() {
	e.CurrentTableName = ""
}

func (e *Erd) AddTitleKeyValue() {
	if e.Title.TitleAttributes == nil {
		e.Title.TitleAttributes = map[string]string{}
	}
	e.Title.TitleAttributes[e.key] = e.value
}

func (e *Erd) AddTable(text string) {
	if e.Tables == nil {
		e.Tables = map[string]*Table{}
	}
	e.Tables[text] = &Table{Title: text, TableAttributes: map[string]string{}}
	e.CurrentTableName = text
}

func (e *Erd) AddTableKeyValue() {
	table := e.Tables[e.CurrentTableName]
	if table.TableAttributes == nil {
		table.TableAttributes = map[string]string{}
	}
	table.TableAttributes[e.key] = e.value
}

func (e *Erd) AddColumn(text string) {
	if e.CurrentTableName == "" {
		e.Error(errors.New("Invalid State"))
	}

	table := e.Tables[e.CurrentTableName]
	table.Columns = append(table.Columns, Column{Title: text, ColumnAttributes: map[string]string{}})
	table.CurrentColumnId = len(table.Columns) - 1
}

func (e *Erd) AddColumnKeyValue() {
	table := e.Tables[e.CurrentTableName]
	column := table.Columns[table.CurrentColumnId]
	if column.ColumnAttributes == nil {
		column.ColumnAttributes = map[string]string{}
	}
	column.ColumnAttributes[e.key] = e.value
	e.key = ""
	e.value = ""
}

func (e *Erd) SetKey(text string) {
	e.key = text
	if len(e.key) > 0 && e.key[0] == '"' {
		e.key = e.unquote(e.key)
	}
}

func (e *Erd) SetValue(text string) {
	e.value = text
	if len(e.value) > 0 && e.value[0] == '"' {
		e.value = e.unquote(e.value)
	}
}

func (e *Erd) AddRelation() {
	e.Relations = append(e.Relations, e.CurrentRelation)
	e.CurrentRelation = Relation{}
}

func (e *Erd) AddRelationKeyValue() {
	if e.CurrentRelation.RelationAttributes == nil {
		e.CurrentRelation.RelationAttributes = map[string]string{}
	}
	e.CurrentRelation.RelationAttributes[e.key] = e.value
}

func (e *Erd) SetRelationLeft(text string) {
	e.CurrentRelation.LeftTableName = text
}

func (e *Erd) SetCardinalityLeft(text string) {
	e.CurrentRelation.LeftCardinality = text
}

func (e *Erd) SetRelationRight(text string) {
	e.CurrentRelation.RightTableName = text
}

func (e *Erd) SetCardinalityRight(text string) {
	e.CurrentRelation.RightCardinality = text
}

func (e *Erd) unquote(str string) string {
	s, err := strconv.Unquote(str)
	if err != nil {
		e.Error(err)
	}
	return s
}

func (e *Erd) Error(err error) {
	panic(err)
}

func (c *Erd) Err(pos int, buffer string) {
	fmt.Println("")
	a := strings.Split(buffer[:pos], "\n")
	row := len(a) - 1
	column := len(a[row]) - 1

	lines := strings.Split(buffer, "\n")
	for i := row - 5; i <= row; i++ {
		if i < 0 {
			i = 0
		}

		fmt.Println(lines[i])
	}

	s := ""
	for i := 0; i <= column; i++ {
		s += " "
	}
	ln := len(strings.Trim(lines[row], " \r\n"))
	for i := column + 1; i < ln; i++ {
		s += "~"
	}
	fmt.Println(s)

	fmt.Println("error")
	c.IsError = true
}
