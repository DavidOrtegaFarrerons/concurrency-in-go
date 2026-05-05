package pipeline

import (
	"io"
	"strings"
	"sync"
)

type Step interface {
	Execute(<-chan string) <-chan string
}

type RemoveDebugStringsStep struct {
}

func (s *RemoveDebugStringsStep) Execute(data <-chan string) <-chan string {
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

type TransformTextToUppercaseStep struct {
}

func (s *TransformTextToUppercaseStep) Execute(data <-chan string) <-chan string {
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
	steps  []Step
	output io.Writer
}

func New(steps []Step, output io.Writer) *Pipeline {
	return &Pipeline{
		wg:     sync.WaitGroup{},
		steps:  steps,
		output: output,
	}
}

func (p *Pipeline) Start(text []string) {
	input := make(chan string)
	go func() {
		for _, t := range text {
			input <- t
		}
		close(input)
	}()

	var ch <-chan string = input
	for _, s := range p.steps {
		ch = s.Execute(ch)
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for r := range ch {
			p.output.Write([]byte(r + "\n"))
		}
	}()

	p.wg.Wait()
}
