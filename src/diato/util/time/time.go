// Diato - Reverse Proxying for Hipsters
//
// Copyright 2016-2017 Dolf Schimmel
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package time

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

const DURATION_REGEX_STRING = `^P(?P<years>\d+Y)?(?P<months>\d+M)?(?P<days>\d+D)?T?(?P<hours>\d+H)?(?P<minutes>\d+M)?(?P<seconds>\d+S)?$`

var durationRegex *regexp.Regexp

// Parse an ISO8601 duration
//
// See: http://stackoverflow.com/questions/28125963/golang-parse-time-duration
// Licensed under cc by-sa 3.0 originally provided by Régis B.
func ParseDuration(str string) (time.Duration, error) {
	if durationRegex == nil {
		durationRegex = regexp.MustCompile(DURATION_REGEX_STRING)
	}
	matches := durationRegex.FindStringSubmatch(str)
	if len(matches) != 7 {
		return 0, fmt.Errorf("time: invalid ISO8601 duration %s", str)
	}

	years := ParseInt64(matches[1])
	months := ParseInt64(matches[2])
	days := ParseInt64(matches[3])
	hours := ParseInt64(matches[4])
	minutes := ParseInt64(matches[5])
	seconds := ParseInt64(matches[6])

	hour := int64(time.Hour)
	minute := int64(time.Minute)
	second := int64(time.Second)

	duration := years*24*365*hour + months*30*24*hour + days*24*hour + hours*hour + minutes*minute + seconds*second
	return time.Duration(duration), nil
}

// See: http://stackoverflow.com/questions/28125963/golang-parse-time-duration
// Licensed under cc by-sa 3.0 originally provided by Régis B.
func ParseInt64(value string) int64 {
	if len(value) == 0 {
		return 0
	}
	parsed, err := strconv.Atoi(value[:len(value)-1])
	if err != nil {
		return 0
	}
	return int64(parsed)
}
