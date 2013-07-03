package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

var newline = []uint8("\n")
var help = flag.Bool("help", false, "help")

func main() {
	flag.Parse()
	if *help {
		cmd := os.Args[0]
		if cmd[0:2] == "./" {
			cmd = cmd[2:]
		}
		fmt.Fprintf(os.Stderr, "Usage: "+cmd+" [file]"+"\n")
		fmt.Fprintf(os.Stderr, "   or: $COMMAND | "+cmd+"\n")
		os.Exit(0)
	}

	bufIn := bufio.NewReader(fileFromArguments())
	arr := make([]byte, 0, 1024*1024)
	buf := bytes.NewBuffer(arr)
	lineNum := int64(1)
	for {
		lastLine, err := bufIn.ReadBytes('\n')
		if err != nil && err != io.EOF {
			genericError(err)
		}

		if err == io.EOF && len(lastLine) == 0 {
			break
		}

		indentAndPrint(buf, lastLine, lineNum)
		buf.Reset()
		lineNum += 1

		if err == io.EOF {
			break
		}
	}
}

func indentAndPrint(buf *bytes.Buffer, js []byte, lineNum int64) {
	jsErr := json.Indent(buf, js, "", "  ")
	if jsErr != nil {
		malformedJSON(jsErr, js, lineNum)
	}
	os.Stdout.Write(buf.Bytes())
}

func malformedJSON(jsErr error, js []uint8, lineNum int64) {
	os.Stdout.Sync()

	synErr, isSynError := (jsErr).(*json.SyntaxError)
	if isSynError {
		fmt.Fprintf(os.Stderr, "ERROR: Broken json on line %d, char %d: %s\n", lineNum, synErr.Offset, jsErr)
		prefix := ""
		suffix := ""

		begin := 0
		if synErr.Offset > 15 {
			begin = int(synErr.Offset - 15)
			prefix = "..."
		}

		end := begin + 30
		if end > len(js) {
			end = len(js)
		} else {
			suffix = "..."
		}
		b := bytes.TrimRight(js[begin:end], "\r\n")
		fmt.Fprintf(os.Stderr, "  Context: %s%s%s\n", prefix, b, suffix)
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: Broken json on line %d: %s\n", lineNum, jsErr)
	}

	os.Exit(1)
}

func fileFromArguments() *os.File {
	args := flag.Args()
	if len(args) > 1 {
		msg := fmt.Sprintf("only specify 0 or 1 files in the arguments, not %d\n", len(args))
		genericError(errors.New(msg))
	}
	if len(args) == 0 {
		return os.Stdin
	}

	file, err := os.OpenFile(args[0], os.O_RDONLY, 0)
	if err != nil {
		genericError(err)
	}
	return file
}

func genericError(err error) {
	os.Stdout.Sync()
	fmt.Fprintf(os.Stderr, "ERROR: %s", err)
	os.Exit(2)
}
