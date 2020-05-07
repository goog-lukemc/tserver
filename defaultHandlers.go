// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tserver

import (
	"encoding/json"
	"log"
	"net/http"
)

func DefaultHandlers(s *ServerControl) {

	// Hides hidden files on the linux file system
	fs := dotFileHidingFileSystem{http.Dir(s.CFG.StaticDir)}

	// Local files a served under the /site/ request path by default.
	s.MUX.Handle("/", http.FileServer(fs))
}

func Respond(w http.ResponseWriter, com interface{}) {
	var bts []byte
	var err error
	if _, ok := com.([]byte); !ok {
		bts, err = json.Marshal(com)
		if err != nil {
			log.Fatalf("errMarshal:%s", err)
		}
	} else {
		bts = com.([]byte)
	}

	contentType := http.DetectContentType(bts)
	w.Header().Set("content-type", contentType)
	if _, err := w.Write(bts); err != nil {
		log.Fatalf("errHTTPWrite:%s", err)
	}
	return
}
