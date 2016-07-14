package main

import (
	"fmt"
	"github.com/DarinM223/http-scraper/graph"
	"io/ioutil"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Need to specify path to code file")
		return
	}

	path := os.Args[1]
	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Println("Error translating to absolute path: ", err)
		return
	}

	code, err := ioutil.ReadFile(absPath)
	if err != nil {
		fmt.Println("Error reading file: ", err)
		return
	}

	parser := Parser{0, string(code)}
	globals := graph.NewGlobals()

	expr, err := parser.parseExpr()
	if err != nil {
		fmt.Println("Error parsing expression: ", err)
		return
	}

	resultNode, err := CompileExpr(globals, expr, NewScope(nil))
	if err != nil {
		fmt.Println("Error compiling expression: ", err)
		return
	}

	resultChan := make(chan graph.Msg, graph.InChanSize)
	resultNode.ParentChans()[69] = resultChan

	globals.Run()

	if msg, ok := <-resultChan; ok {
		fmt.Println("Result: ", msg)
		if streamMsg, ok := msg.(*graph.StreamMsg); ok {
			for i := 0; i < streamMsg.Len-1; i++ {
				if nextMsg, ok := <-resultChan; ok {
					fmt.Println("Stream result: ", nextMsg)
				}
			}
		}
	}
}
