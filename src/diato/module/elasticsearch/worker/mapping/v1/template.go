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

var template = `{
   "template":"diato-httprequest-*-v1",
   "aliases":{
      "diato-httprequest":{

      },
      "diato-httprequest-v1":{

      }
   },
   "settings":{
      "analysis":{
         "analyzer":{
            "lowercase":{
               "type":"custom",
               "tokenizer":"keyword",
               "filter":[
                  "lowercase"
               ]
            }
         }
      }
   },
   "mappings":{
      "session":{
         "_all":{
            "enabled":false
         },
         "properties":{
            "Host":{
               "type":"string",
               "analyzer":"lowercase"
            },
            "SLD":{
               "type":"string",
               "analyzer":"lowercase"
            },
            "Url":{
               "type":"string",
               "index":"not_analyzed"
            },
            "Path":{
               "type":"string",
               "index":"not_analyzed"
            },
            "Query":{
               "type":"string",
               "index":"not_analyzed"
            },
            "Timestamp":{
               "type":"date"
            },
            "Duration":{
               "type":"float"
            },
            "RemoteIp":{
               "type":"ip"
            },
            "ResponseCode":{
               "type":"short"
            },
            "Method":{
               "type":"string",
               "analyzer":"lowercase"
            },
            "HttpVersion":{
               "type":"float"
            },
            "Referrer":{
               "type":"string",
               "index":"not_analyzed"
            },
            "UserAgent":{
               "properties":{
                  "Raw":{
                     "type":"string",
                     "index":"not_analyzed"
                  },
                  "Browser":{
                     "properties":{
                        "Name":{
                           "type":"string",
                           "analyzer":"lowercase"
                        },
                        "Version":{
                           "type":"string",
                           "analyzer":"lowercase"
                        },
                        "VersionMajor":{
                           "type":"integer"
                        },
                        "VersionMinor":{
                           "type":"integer"
                        },
                        "Engine":{
                           "properties":{
                              "Name":{
                                 "type":"string",
                                 "analyzer":"lowercase"
                              },
                              "Version":{
                                 "type":"string",
                                 "analyzer":"lowercase"
                              },
                              "VersionMajor":{
                                 "type":"integer"
                              },
                              "VersionMinor":{
                                 "type":"integer"
                              }

                           }
                        }
                     }
                  }
               }
            },
            "Diato":{
               "properties":{
                  "Hostname":{
                     "type":"string",
                     "index":"not_analyzed"
                  }
               }
            }
         }
      }
   }
}`
