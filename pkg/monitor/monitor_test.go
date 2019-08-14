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

func Test_segment_topSegments(t *testing.T) {
	type args struct {
		segment *segment
		n       int
	}
	tests := map[string]struct {
		args args
		want []segmentCount
	}{
		"empty": {
			args: args{
				segment: &segment{},
				n:       0,
			},
			want: nil,
		},
		"lessThanN": {
			args: args{
				segment: &segment{
					hits: map[string]int{
						"foo": 10,
						"bar": 20,
					},
				},
				n: 2,
			},
			want: []segmentCount{
				{
					Segment: "bar",
					Count:   20,
				},
				{
					Segment: "foo",
					Count:   10,
				},
			},
		},
		"moreThanN": {
			args: args{
				segment: &segment{
					hits: map[string]int{
						"foo": 10,
						"bar": 20,
						"baz": 30,
					},
				},
				n: 2,
			},
			want: []segmentCount{
				{
					Segment: "baz",
					Count:   30,
				},
				{
					Segment: "bar",
					Count:   20,
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.args.segment.topSections(tt.args.n)
			cmp.Equal(got, tt.want)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("topSegments() %s", diff)
			}
		})
	}
}

func Test_alert_check(t *testing.T) {
	tests := map[string]struct {
		alert *alert
		args  float64
		want  bool
	}{
		"CheckSet": {
			alert: &alert{
				currentState: false,
				threshold:    10.0,
			},
			args: 11.0,
			want: true,
		},
		"CheckNoChangeSet": {
			alert: &alert{
				currentState: true,
				threshold:    10.0,
			},
			args: 11.0,
			want: true,
		},
		"CheckUnset": {
			alert: &alert{
				currentState: true,
				threshold:    10.0,
			},
			args: 8.0,
			want: false,
		},
		"CheckNoChangeUnset": {
			alert: &alert{
				currentState: false,
				threshold:    10.0,
			},
			args: 8.0,
			want: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.alert.check(tt.args); got != tt.want {
				t.Errorf("check() = %v, want %v", got, tt.want)
			}
		})
	}
}
