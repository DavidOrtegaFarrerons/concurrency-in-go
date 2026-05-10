package pipeline

import (
	"io"
	"strings"
	"sync"
)

type Stage interface {
	Execute(<-chan string) <-chan string
}

type RemoveDebugStringsStage struct {
}

func (s *RemoveDebugStringsStage) Execute(data <-chan string) <-chan string {
	ch := make(chan string)
	go func() {
		for d := range data {
			if strings.Contains(d, "DEBUG") {
				continue
			}

			ch <- d
		}
		close(ch)
	}()

	return ch
}

type TransformTextToUppercaseStage struct {
}

func (s *TransformTextToUppercaseStage) Execute(data <-chan string) <-chan string {
	ch := make(chan string)
	go func() {
		for d := range data {
			ch <- strings.ToUpper(d)
		}

		close(ch)
	}()

	return ch
}

type Pipeline struct {
	wg     sync.WaitGroup
	stages []Stage
	output io.Writer
}

func New(stages []Stage, output io.Writer) *Pipeline {
	return &Pipeline{
		wg:     sync.WaitGroup{},
		stages: stages,
		output: output,
	}
}

func (p *Pipeline) Start(text []string) {
	readData := func(text []string) <-chan string {
		input := make(chan string)

		go func() {
			defer close(input)

			for _, t := range text {
				input <- t
			}
		}()

		return input
	}

	var ch = readData(text)

	for _, s := range p.stages {
		ch = s.Execute(ch)
	}

	for r := range ch {
		p.output.Write([]byte(r + "\n"))
	}
}
