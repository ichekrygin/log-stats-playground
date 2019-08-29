package monitor

import (
	"bufio"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Entry - minimalistic data structure for log Entry record
type Entry struct {
	TimeStamp time.Time
	Method    string
	Path      string
	Section   string
}

// NewEntry - parse the data to return Entry value
//"remotehost","rfc931","authuser","date","request","status","bytes"
//"10.0.0.2","-","apache",1549573860,"GET /api/user HTTP/1.0",200,1234
func NewEntry(data []string) (*Entry, error) {
	if len(data) < 7 {
		return nil, errors.Errorf("malformed input, expected 7 elements, got: %v", data)
	}
	ts, err := strconv.ParseInt(data[3], 10, 64)
	if err != nil {
		return nil, errors.Errorf("malformed input, invalid timestamp value: %s", data[3])
	}

	requestFields := strings.Split(data[4], " ")
	if len(requestFields) < 3 {
		return nil, errors.Errorf("malformed request data: %s", data[4])
	}

	pathFields := strings.Split(requestFields[1], "/")
	if len(pathFields) < 2 {
		return nil, errors.Errorf("malformed request Path: %s", requestFields[1])
	}

	return &Entry{
		TimeStamp: time.Unix(ts, 0),
		Method:    requestFields[0][1:],
		Path:      requestFields[1],
		Section:   pathFields[1],
	}, nil
}

// Segment - fixed range Segment window with start time, Total and per Section hits Count
type Segment struct {
	// Segment start time
	start time.Time
	// Segment Total hits
	total int
	// Segment hits Count per Section
	hits map[string]int
}

// NewSegment with start time
func NewSegment(t time.Time) *Segment {
	return &Segment{
		start: t,
		hits:  make(map[string]int),
		total: 0,
	}
}

// AddSection - increment counters
func (s *Segment) AddSection(section string) {
	s.hits[section]++
	s.total++
}

// SegmentCount tuple
type SegmentCount struct {
	Segment string
	Count   int
}

// TopSections returns top N segments in descending order by Count.
// If N > segments length, return all segments
func (s *Segment) TopSections(n int) []SegmentCount {
	var data []SegmentCount

	for k, v := range s.hits {
		data = append(data, SegmentCount{k, v})
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].Count < data[j].Count
	})

	// Reverse
	for i := len(data)/2 - 1; i >= 0; i-- {
		opp := len(data) - 1 - i
		data[i], data[opp] = data[opp], data[i]
	}

	if len(data) > n {
		return data[:n]
	}
	return data
}

// Alert - helper structure to keep track of threshold and current state
type Alert struct {
	currentState bool
	threshold    float64
}

// Check - sets and returns current state based on the incoming value
func (a *Alert) Check(value float64, ts time.Time) {
	oldState := a.currentState
	a.currentState = value > a.threshold
	if oldState != a.currentState {
		if a.currentState {
			logrus.Infof("Alert - hits: %f, triggered at: %s", value, ts.Format(time.RFC3339))
		} else {
			logrus.Infof("Alert - hits: %f, reset at: %s", value, ts.Format(time.RFC3339))
		}
	}
}

type Span struct {
	sums            []int
	segmentsInSpan  int
	segmentIndex    int
	segmentDuration time.Duration
	alert           *Alert
}

func NewSpan(segmentSeconds, spanSeconds int, alertThreshold float64) (*Span, error) {
	segmentsInSpan := spanSeconds / segmentSeconds
	if segmentsInSpan < 1 {
		return nil, errors.Errorf("Invalid Segment/Span seconds values combination: %d, %d", segmentSeconds, spanSeconds)
	}

	return &Span{
		sums:            make([]int, segmentsInSpan),
		segmentsInSpan:  segmentsInSpan,
		segmentIndex:    segmentsInSpan,
		segmentDuration: time.Duration(segmentSeconds) * time.Second,
		alert: &Alert{
			currentState: false,
			threshold:    alertThreshold,
		},
	}, nil
}

func (s *Span) curIndex() int {
	return (len(s.sums) + s.segmentIndex) % len(s.sums)
}

func (s *Span) prevIndex() int {
	return (len(s.sums) + s.segmentIndex - 1) % len(s.sums)
}

func (s *Span) Update(n int, ts time.Time) {
	tmp := -s.sums[s.curIndex()]
	s.sums[s.curIndex()] = n + s.sums[s.prevIndex()]
	s.segmentIndex++

	s.alert.Check(float64(tmp+s.sums[s.curIndex()])/float64(s.segmentsInSpan), ts)
}

func (s *Span) Total() int {
	return s.sums[s.curIndex()]
}

// Process data with provided Alert threshold level for average hits Count
func Process(s *bufio.Scanner, span *Span) error {
	var seg *Segment

	s.Scan() // skip the first line

	for s.Scan() {
		text := s.Text()

		rec, err := NewEntry(strings.Split(text, ","))
		if err != nil {
			// TODO: other considerations
			// 	- skip all bad records
			//  - skip bad records until the threshold
			return errors.Wrapf(err, "failed to parse record: %s", text)
		}

		// Fixed range window
		if seg == nil || rec.TimeStamp.Sub(seg.start) > span.segmentDuration {
			if seg != nil {
				span.Update(seg.total, rec.TimeStamp)
				// Report stats for last Segment
				logrus.Infof("%02d sec stats: %v, %v", span.segmentDuration/time.Second, seg.TopSections(3), span.Total())
			}
			// Starting a new Segment
			seg = NewSegment(rec.TimeStamp)
		}

		seg.AddSection(rec.Section)
	}

	// Report last Segment
	if seg != nil {
		logrus.Infof("%02d sec stats: %v, %v", span.segmentDuration/time.Second, seg.TopSections(3), span.Total())
	}

	// return scanner error (if any)
	return s.Err()
}
