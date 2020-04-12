package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/tlb_pretty"
)

func PrintAst(node *tlb_pretty.AstNode, spacesCount int) {
	spaces := strings.Repeat(" ", spacesCount)
	for k, v := range node.Fields {
		if vv, ok := v.(string); ok {
			if len(vv) > 40 {
				fmt.Print(spaces, k, ": ", string(vv[:10]), "...\n")
			} else {
				fmt.Print(spaces, k, ": ", string(vv), "\n")
			}
		} else if vv, ok := v.(*tlb_pretty.AstNode); ok {
			fmt.Print(spaces, k, ": ", "........", "\n")
			if k == "extra" {
				PrintAst(vv, spacesCount+2)
			}
			if spacesCount > 0 {
				PrintAst(vv, spacesCount+2)
			}

			//PrintAst(vv, spacesCount+2)
		}
	}
}

func main() {
	input := flag.String("in", "block.pretty", "Pretty blocks file or dir with such files.")
	out := flag.String("out", "", "Output directory where json files will be stored.")
	consoleMode := flag.Bool("console_mode", false, "If true, all blocks will be written to STDOUT instead of files.")
	silentMode := flag.Bool("silent_mode", false, "If true, all blocks will be just silently simplified and converted without writing them anywhere.")

	flag.Parse()

	var filenames []string
	var dirInput bool

	if items, err := ioutil.ReadDir(*input); err != nil {
		filenames = []string{*input}
	} else {
		dirInput = true
		for _, item := range items {
			if !item.IsDir() {
				filenames = append(filenames, item.Name())
			}
		}
	}

	start := time.Now()
	for _, file := range filenames {
		path := file

		if dirInput {
			path = *input + "/" + file
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			// Can easily happen if we couldn't ReadDir because of some reasons other than it being not a directory.
			log.Fatal(err)
		}

		// Parse
		since := time.Now()
		node := tlb_pretty.NewParser().Parse(data)
		fmt.Println("Parsed for:", time.Since(since))

		// Simplify
		since = time.Now()
		newNode, err := tlb_pretty.NewTreeSimplifier().Simplify(node)
		if err != nil {
			log.Fatal("Simplify err:", err)
		}
		fmt.Println("Simplified for:", time.Since(since))

		since = time.Now()
		block, err := tlb_pretty.NewAstTonConverter().ConvertToBlock(newNode)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Converted for:", time.Since(since))

		if *silentMode {
			continue
		}

		blockJson, err := json.MarshalIndent(block, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		if *consoleMode {
			if len(block.Transactions) > 0 {
				fmt.Println(string(blockJson))
			}
			continue
		}

		// adds _example after input filename and replaces extension to .json. Example: block.pretty -> block_example.json
		fileParts := strings.Split(file, ".")
		fileParts[0] = fileParts[0] + "_example"
		fileParts[len(fileParts) - 1] = "json"
		filename := *out + "/" + strings.Join(fileParts, ".")

		outFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		if _, err = outFile.Write(blockJson); err != nil {
			log.Fatal(err)
		}

		if err = outFile.Sync(); err != nil {
			log.Fatal(err)
		}

		if err = outFile.Close(); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Time total:", time.Since(start))
}
