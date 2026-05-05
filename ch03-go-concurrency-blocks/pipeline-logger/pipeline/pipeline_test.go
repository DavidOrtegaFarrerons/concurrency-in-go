package pipeline

import (
	"bytes"
	"strings"
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
				steps:  []Step{&RemoveDebugStringsStep{}, &TransformTextToUppercaseStep{}},
				output: buf,
			}

			pipeline.Start(tt.input)

			if strings.Compare(buf.String(), tt.expectedOutput) != 0 {
				t.Errorf("Got %+v\n want %+v", buf.String(), tt.expectedOutput)
			}
		})
	}
}
