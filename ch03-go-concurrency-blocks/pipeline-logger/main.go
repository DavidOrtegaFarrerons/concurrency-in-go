package main

import (
	"bufio"
	"main/pipeline"
	"os"
)

func main() {
	pipeline := pipeline.New(
		[]pipeline.Step{
			&pipeline.RemoveDebugStringsStep{},
			&pipeline.TransformTextToUppercaseStep{},
		},
		os.Stdout,
	)

	scanner := bufio.NewScanner(os.Stdin)
	var text []string

	for scanner.Scan() {
		text = append(text, scanner.Text())
	}

	pipeline.Start(text)
}
