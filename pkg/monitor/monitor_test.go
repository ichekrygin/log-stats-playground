package monitor

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	dataSample = `"remotehost","rfc931","authuser","date","request","status","bytes"
10.0.0.2","-","apache",1549573860,"GET /api/user HTTP/1.0",200,1234
"10.0.0.4","-","apache",1549573860,"GET /api/user HTTP/1.0",200,1234
"10.0.0.4","-","apache",1549573860,"GET /api/user HTTP/1.0",200,1234
"10.0.0.2","-","apache",1549573860,"GET /api/help HTTP/1.0",200,1234
"10.0.0.5","-","apache",1549573860,"GET /api/help HTTP/1.0",200,1234
"10.0.0.4","-","apache",1549573859,"GET /api/help HTTP/1.0",200,1234
`
)

//  Overall test - if I dont have time for more granular tests
func TestProcess(t *testing.T) {
	f, err := os.Open("../../sample_csv.txt")
	if err != nil {
		t.Errorf("failed to open test data file: %v", err)
	}

	type args struct {
		scanner   *bufio.Scanner
		threshold float64
	}

	tests := map[string]struct {
		args args
		want error
	}{
		"Sample": {
			args: args{
				scanner:   bufio.NewScanner(strings.NewReader(dataSample)),
				threshold: 10.0,
			},
			want: nil,
		},
		"File": {
			args: args{
				scanner:   bufio.NewScanner(f),
				threshold: 100.0,
			},
			want: nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if diff := cmp.Diff(Process(tt.args.scanner, tt.args.threshold), tt.want); diff != "" {
				t.Errorf("Process() %s", diff)
			}
		})
	}
}
