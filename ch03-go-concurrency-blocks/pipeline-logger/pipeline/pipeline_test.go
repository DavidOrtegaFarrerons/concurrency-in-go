package pipeline

import (
	"bytes"
	"io"
	"sync"
	"testing"
)

func TestPipelineStart(t *testing.T) {
	tests := []struct {
		name           string
		input          []string
		expectedOutput string
	}{
		{
			name:           "filters out DEBUG lines",
			input:          []string{"DEBUG something", "INFO hello"},
			expectedOutput: "INFO HELLO\n",
		},
		{
			name:           "uppercases lines",
			input:          []string{"INFO user logged in"},
			expectedOutput: "INFO USER LOGGED IN\n",
		},
		{
			name:           "multiple lines mixed",
			input:          []string{"INFO start", "DEBUG skip this", "ERROR failed"},
			expectedOutput: "INFO START\nERROR FAILED\n",
		},
		{
			name:           "all DEBUG lines filtered",
			input:          []string{"DEBUG a", "DEBUG b"},
			expectedOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			pipeline := Pipeline{
				wg:     sync.WaitGroup{},
				stages: []Stage{&RemoveDebugStringsStage{}, &TransformTextToUppercaseStage{}},
				output: buf,
			}

			pipeline.Start(tt.input)

			if buf.String() != tt.expectedOutput {
				t.Errorf("Got %+v\n want %+v", buf.String(), tt.expectedOutput)
			}
		})
	}
}

func BenchmarkPipelineStart(b *testing.B) {
	input := []string{
		"INFO hello",
		"DEBUG should disappear",
		"ERROR something failed",
		"INFO another line",
	}

	for b.Loop() {
		pipeline := Pipeline{
			stages: []Stage{
				&RemoveDebugStringsStage{},
				&TransformTextToUppercaseStage{},
			},
			output: io.Discard,
		}

		pipeline.Start(input)
	}
}
