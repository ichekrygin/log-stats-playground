package monitor

import (
	"bufio"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

const (
	dataSample = `"remotehost","rfc931","authuser","date","request","status","bytes"
"10.0.0.2","-","apache",1549573860,"GET /api/user HTTP/1.0",200,1234
"10.0.0.4","-","apache",1549573860,"GET /api/user HTTP/1.0",200,1234
"10.0.0.4","-","apache",1549573860,"GET /api/user HTTP/1.0",200,1234
"10.0.0.2","-","apache",1549573860,"GET /api/help HTTP/1.0",200,1234
"10.0.0.5","-","apache",1549573860,"GET /api/help HTTP/1.0",200,1234
"10.0.0.4","-","apache",1549573859,"GET /api/help HTTP/1.0",200,1234
`
)

func EquateErrors() cmp.Option {
	return cmp.Comparer(func(a, b error) bool {
		if a == nil || b == nil {
			return a == nil && b == nil
		}

		av := reflect.ValueOf(a)
		bv := reflect.ValueOf(b)
		if av.Type() != bv.Type() {
			return false
		}

		return a.Error() == b.Error()
	})
}

// Overall test - if I dont have time for more granular tests
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
				threshold: 10.0,
			},
			want: nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			span, err := NewSpan(10, 120, tt.args.threshold)
			if err != nil {
				log.Fatal(err.Error(), "failed setting up monitor")
			}
			if diff := cmp.Diff(Process(tt.args.scanner, span), tt.want); diff != "" {
				t.Errorf("Process() %s", diff)
			}
		})
	}
}

func Test_NewEntry(t *testing.T) {
	type want struct {
		entry *Entry
		err   error
	}
	tests := map[string]struct {
		args []string
		want want
	}{
		"Default": {
			args: []string{"\"10.0.0.2\"", "\"-\"", "\"apache\"", "1549573860", "\"GET /api/user HTTP/1.0\"", "200", "1234"},
			want: want{
				entry: &Entry{
					TimeStamp: time.Unix(1549573860, 0),
					Method:    "GET",
					Path:      "/api/user",
					Section:   "api",
				},
				err: nil,
			},
		},
		"EmpytData": {
			args: []string{},
			want: want{
				err: errors.Errorf("malformed input, expected 7 elements, got: %v", []string{}),
			},
		},
		"MalformedTimestamp": {
			args: []string{"\"10.0.0.2\"", "\"-\"", "\"apache\"", "154957x860", "\"GET /api/user HTTP/1.0\"", "200", "1234"},
			want: want{
				err: errors.Errorf("malformed input, invalid timestamp value: %v", "154957x860"),
			},
		},
		"MalformedRequest": {
			args: []string{"\"10.0.0.2\"", "\"-\"", "\"apache\"", "1549573860", "\"GET /api/user\"", "200", "1234"},
			want: want{
				err: errors.Errorf("malformed request data: %s", "\"GET /api/user\""),
			},
		},
		"MalformedRequestPath": {
			args: []string{"\"10.0.0.2\"", "\"-\"", "\"apache\"", "1549573860", "\"GET x HTTP/1.0\"", "200", "1234"},
			want: want{
				err: errors.Errorf("malformed request Path: %s", "x"),
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewEntry(tt.args)
			if diff := cmp.Diff(err, tt.want.err, EquateErrors()); diff != "" {
				t.Errorf("NewEntry() error %s", diff)
			}
			if diff := cmp.Diff(got, tt.want.entry); diff != "" {
				t.Errorf("NewEntry() %s", diff)
			}
		})
	}
}

func TestSegment_AddSection(t *testing.T) {
	type fields struct {
		start time.Time
		data  map[string]int
	}
	type args struct {
		section string
	}
	type want struct {
		hits map[string]int
	}
	tests := map[string]struct {
		fields fields
		args   args
		want   want
	}{
		"Default": {
			fields: fields{start: time.Now(), data: map[string]int{}},
			args:   args{"foo"},
			want:   want{hits: map[string]int{"foo": 1}},
		},
		"Update": {
			fields: fields{start: time.Now(), data: map[string]int{"foo": 9, "bar": 10}},
			args:   args{"foo"},
			want:   want{hits: map[string]int{"foo": 10, "bar": 10}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := NewSegment(tt.fields.start)
			s.hits = tt.fields.data

			s.AddSection(tt.args.section)

			if s.start != tt.fields.start {
				t.Errorf("Segment_AddSection() start: %v, expected: %v", s.start, tt.fields.start)
			}
			if diff := cmp.Diff(s.hits, tt.want.hits); diff != "" {
				t.Errorf("Segment_AddSEction() hits: %s", diff)
			}
		})
	}
}

func TestSegment_TopSections(t *testing.T) {
	type args struct {
		segment *Segment
		n       int
	}
	tests := map[string]struct {
		args args
		want []SegmentCount
	}{
		"empty": {
			args: args{
				segment: &Segment{},
				n:       0,
			},
			want: nil,
		},
		"lessThanN": {
			args: args{
				segment: &Segment{
					hits: map[string]int{
						"foo": 10,
						"bar": 20,
					},
				},
				n: 2,
			},
			want: []SegmentCount{
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
				segment: &Segment{
					hits: map[string]int{
						"foo": 10,
						"bar": 20,
						"baz": 30,
					},
				},
				n: 2,
			},
			want: []SegmentCount{
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
			got := tt.args.segment.TopSections(tt.args.n)
			cmp.Equal(got, tt.want)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("topSegments() %s", diff)
			}
		})
	}
}

func TestAlert_Check(t *testing.T) {
	type fields struct {
		currentState bool
		threshold    float64
	}
	type args struct {
		value float64
		ts    time.Time
	}
	tests := map[string]struct {
		fields fields
		args   args
		want   bool
	}{
		"AlertSet": {
			fields: fields{
				currentState: false,
				threshold:    10.0,
			},
			args: args{
				value: 10.01,
				ts:    time.Now(),
			},
			want: true,
		},
		"AlertSetNoChange": {
			fields: fields{
				currentState: true,
				threshold:    10.0,
			},
			args: args{
				value: 10.01,
				ts:    time.Now(),
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			a := &Alert{
				currentState: tt.fields.currentState,
				threshold:    tt.fields.threshold,
			}
			a.Check(tt.args.value, tt.args.ts)
			if a.currentState != tt.want {
				t.Errorf("Check() = %v, want %v", a.currentState, tt.want)
			}
		})
	}
}
