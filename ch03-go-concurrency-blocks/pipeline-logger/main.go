package main

import (
	"bufio"
	"main/pipeline"
	"os"
)

func main() {
	pipeline := pipeline.New(
		[]pipeline.Stage{
			&pipeline.RemoveDebugStringsStage{},
			&pipeline.TransformTextToUppercaseStage{},
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
