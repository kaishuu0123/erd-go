package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"text/template"

	flags "github.com/jessevdk/go-flags"
	"golang.org/x/crypto/ssh/terminal"
)

// Options for the command line tool
type Options struct {
	OutFormat  string `short:"f" long:"fmt" description:"output format (dot only)"`
	InputFile  string `short:"i" long:"input" description:"input will be read from the given file."`
	OutputFile string `short:"o" long:"output" description:"output will be written to the given file."`
}

var opts Options

func main() {
	contents := ""
	logStderr := log.New(os.Stderr, "", 0)

	optsParser := flags.NewParser(&opts, flags.Default)
	optsParser.Name = filepath.Base(os.Args[0])
	optsParser.Usage = "[OPTIONS] PATTERN [PATH]"

	args, err := optsParser.Parse()
	if err != nil {
		logStderr.Println(err)
		os.Exit(1)
	}

	if terminal.IsTerminal(int(syscall.Stdin)) {
		if len(args) == 0 && opts.InputFile == "" {
			optsParser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
		buffer, err := ioutil.ReadFile(opts.InputFile)
		if err != nil {
			logStderr.Println(err)
			os.Exit(1)
		}
		contents = string(buffer)
	} else {
		body, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			logStderr.Println(err)
			os.Exit(1)
		}
		contents = string(body)
	}

	parser := &Parser{Buffer: contents}
	err = parser.Init()
	if err != nil {
		logStderr.Println(err)
		os.Exit(1)
	}
	err = parser.Parse()
	if err != nil {
		logStderr.Println(err)
		os.Exit(1)
	}

	parser.Execute()

	if parser.Erd.IsError {
		os.Exit(1)
	}
	parser.Erd.CalcIsolated()

	dot, _ := Asset("templates/dot.tmpl")
	tables, _ := Asset("templates/dot_tables.tmpl")
	relations, _ := Asset("templates/dot_relations.tmpl")
	templates := template.Must(
		template.New("").Funcs(template.FuncMap{"StringsJoin": strings.Join}).Parse(
			string(dot) +
				string(tables) +
				string(relations)))

	fd := os.Stdout
	if opts.OutputFile != "" {
		fd, err = os.Create(opts.OutputFile)
		if err != nil {
			logStderr.Println(err)
			os.Exit(1)
		}
	}

	var erdbuf bytes.Buffer
	err = templates.ExecuteTemplate(&erdbuf, "dot", parser.Erd)
	if err != nil {
		logStderr.Println(err)
		os.Exit(1)
	}

	// The OutFormat only works with Graphviz together
	if opts.OutFormat != "" {
		dotcmd := "dot"
		if runtime.GOOS == "windows" {
			dotcmd = "dot.exe"
		}
		cmd := exec.Command(dotcmd, fmt.Sprintf("-T%s", opts.OutFormat))
		cmd.Stdin = &erdbuf
		cmd.Stdout = fd
		cmd.Stderr = fd
		err = cmd.Run()
		if err != nil {
			logStderr.Println(err)
			os.Exit(1)
		}
	} else {
		n, err := io.Copy(fd, &erdbuf)
		if err != nil {
			logStderr.Printf("failed to copy buffer: err: %v, copied %d bytes\n", err, n)
			os.Exit(1)
		}
	}
}
