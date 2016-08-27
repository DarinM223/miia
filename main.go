package main

import (
	"fmt"
	"github.com/DarinM223/miia/graph"
	_ "github.com/davecgh/go-spew/spew"
	"io/ioutil"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Needs to specify the command to run (`compile` or `run`)")
		return
	}

	command := os.Args[1]
	switch command {
	case "compile":
		if len(os.Args) <= 3 {
			fmt.Println("`compile` needs to specify the path to the code file and the path to the output graph file")
			return
		}

		codePath := os.Args[2]
		graphPath := os.Args[3]

		absCodePath, err := filepath.Abs(codePath)
		if err != nil {
			fmt.Println("Error translating code path to absolute path: ", err)
			return
		}

		absGraphPath, err := filepath.Abs(graphPath)
		if err != nil {
			fmt.Println("Error translating graph path to absolute path: ", err)
			return
		}

		globals := graph.NewGlobals()

		code, err := ioutil.ReadFile(absCodePath)
		if err != nil {
			fmt.Println("Error reading file: ", err)
			return
		}

		parser := Parser{0, string(code)}

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

		fanout := 20
		graph.SetNodesFanOut(resultNode, fanout)
		globals.SetResultNodeID(resultNode.ID())

		outputFile, err := os.Create(absGraphPath)
		if err != nil {
			fmt.Println("Error opening output file: ", err)
			return
		}

		graph.WriteGlobals(outputFile, globals)
	case "run":
		if len(os.Args) <= 2 {
			fmt.Println("`run` needs to specify the path to the output graph file")
			return
		}

		graphPath := os.Args[2]
		absGraphPath, err := filepath.Abs(graphPath)
		if err != nil {
			fmt.Println("Error translating graph path to absolute path: ", err)
			return
		}

		file, err := os.Open(absGraphPath)
		if err != nil {
			fmt.Println("Error opening output file: ", err)
			return
		}

		globals, err := graph.ReadGlobals(file)
		if err != nil {
			fmt.Println("Error reading nodes from graph file: ", err)
			return
		}

		resultChan := make(chan graph.Msg, graph.InChanSize)
		globals.ResultNode().ParentChans()[69] = resultChan
		globals.Run()

		if msg, ok := <-resultChan; ok {
			fmt.Println("Result: ", msg)
			if streamMsg, ok := msg.(graph.StreamMsg); ok {
				for i := 0; i < streamMsg.Len.Len()-1; i++ {
					if nextMsg, ok := <-resultChan; ok {
						fmt.Println("Stream result: ", nextMsg)
					}
				}
			}
		}
	default:
		fmt.Println("miia only accepts `compile` and `run` as commands")
	}
}
