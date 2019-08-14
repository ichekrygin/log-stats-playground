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

// entry - minimalistic data structure for log entry record
type entry struct {
	ts      time.Time
	method  string
	path    string
	section string
}

// newEntry - parse the data to return entry value
//"remotehost","rfc931","authuser","date","request","status","bytes"
//"10.0.0.2","-","apache",1549573860,"GET /api/user HTTP/1.0",200,1234
func newEntry(data []string) (*entry, error) {
	ts, err := strconv.ParseInt(data[3], 10, 64)
	if err != nil {
		return nil, err
	}

	requestFields := strings.Split(data[4], " ")
	if len(requestFields) < 3 {
		return nil, errors.Errorf("malformed request data: %s", data[4])
	}

	pathFields := strings.Split(requestFields[1], "/")
	if len(pathFields) < 2 {
		return nil, errors.Errorf("malform request path: %s", requestFields)
	}

	return &entry{
		ts:      time.Unix(ts, 0),
		method:  requestFields[0],
		path:    requestFields[1],
		section: pathFields[1],
	}, nil
}

// segment - fixed range segment window with start time, total and per section hits count
type segment struct {
	// segment start time
	start time.Time
	// segment total tits
	total int
	// segment hits count per section
	hits map[string]int
}

// newSegment with start time
func newSegment(t time.Time) *segment {
	return &segment{
		start: t,
		hits:  make(map[string]int),
		total: 0,
	}
}

// addSection - increment counters
func (s *segment) addSection(section string) {
	s.hits[section]++
	s.total++
}

// segmentCount tuple
type segmentCount struct {
	segment string
	count   int
}

// topSegments returns top N segments in descending order by count.
// If N > segments length, return all segments
func (s *segment) topSegments(n int) []segmentCount {
	var data []segmentCount

	for k, v := range s.hits {
		data = append(data, segmentCount{k, v})
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].count < data[j].count
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

// alert - helper structure to keep track of threshold and current state
type alert struct {
	currentState bool
	threshold    float64
}

// check - sets and returns current state based on the incoming value
func (a *alert) check(value float64) bool {
	a.currentState = value > a.threshold
	return a.currentState
}

// Process data with provided alert threshold level for average hits count
func Process(s *bufio.Scanner, alertThreshold float64) error {
	alert := &alert{
		threshold: alertThreshold,
	}
	// TODO: parameterize this
	segmentDuration := 10 * time.Second

	// how many segments in a given alert window
	// 2 minutes / 10 = 12 segments
	segmentsPerAlertWindow := 2 * 60 / 10

	var seg *segment

	// Running total of all requests per segment to retrieve the total of number
	// of request for a given time window
	// TODO: consider alternative data structure in terms of memory efficiency
	var runningSum []int

	s.Scan() // skip the first line

	for s.Scan() {
		text := s.Text()

		rec, err := newEntry(strings.Split(text, ","))
		if err != nil {
			// TODO: other considerations
			// 	- skip all bad records
			//  - skip bad records until the threshold
			return errors.Wrapf(err, "failed to parse record: %s", text)
		}

		// Fixed range window
		// TODO: other considerations - use sliding window
		if seg == nil || rec.ts.Sub(seg.start) > segmentDuration {
			if seg != nil {
				// Report stats for last segment
				logrus.Infof("%02d sec stats: %v", segmentDuration/time.Second, seg.topSegments(3))

				prevSum := 0
				if sumLen := len(runningSum); sumLen > 1 {
					prevSum = runningSum[sumLen-1]
				}
				runningSum = append(runningSum, seg.total+prevSum)

				if sumLen := len(runningSum); sumLen > segmentsPerAlertWindow-1 {
					totalRequests := runningSum[sumLen-1]
					if sumLen > segmentsPerAlertWindow-1 {
						totalRequests -= runningSum[sumLen-segmentsPerAlertWindow]
					}

					// Check alert state change and report if needed
					if avg, cur := float64(totalRequests)/float64(segmentsPerAlertWindow), alert.currentState; cur != alert.check(avg) {
						if alert.currentState {
							logrus.Infof("alert - hits: %f, triggered at: %s", avg, rec.ts.Format(time.RFC3339))
						} else {
							logrus.Infof("alert - hits: %f, reset at: %s", avg, rec.ts.Format(time.RFC3339))
						}
					}

				}
			}
			// Starting a new segment
			seg = newSegment(rec.ts)
		}

		seg.addSection(rec.section)
	}

	// Report last segment
	if seg != nil {
		logrus.Infof("10 sec stats: %v", seg.topSegments(3))
	}

	// return scanner error (if any)
	return s.Err()
}
