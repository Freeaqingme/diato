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

package v1

import (
	"time"
)

// Initially derived from here, hoping to use compatible dash boards:
// https://www.elastic.co/blog/filebeat-modiles-access-logs-and-elasticsearch-storage-requirements
type httpRequest struct {
	Host      string
	SLD       string // Second Level Domain www.foo.bar.co.uk => bar.co.uk
	Url       string
	Path      string
	Query     string
	Timestamp time.Time
	Duration  float64 // In seconds

	//Body_sent struct {
	//	bytes int
	//}
	RemoteIp string
	//Local_ip      string
	ResponseCode int
	Method       string
	/*	Geoip        struct {
		Region_name    string
		CountryIsoCode string
		CityName       string
		Location       struct {
			Lat float32
			Lon float32
		}
		ContinentName string
	}*/
	HttpVersion float32
	Referrer    string
	UserAgent   userAgent

	Diato diatoInfo
}

type diatoInfo struct {
	Hostname string
	// TODO: Version
}

type userAgent struct {
	Raw      string
	Browser  browser
	Platform string
	Os       string
	Mobile   bool
	Bot      bool
}

type browser struct {
	Name         string
	Version      string
	VersionMajor int
	VersionMinor int
	Engine       browserEngine
}

type browserEngine struct {
	Name         string
	Version      string
	VersionMajor int
	VersionMinor int
}
