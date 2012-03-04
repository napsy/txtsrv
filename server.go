/*
 * Copyright (c) 2012, Luka Napotnik <luka.napotnik@gmail.com>
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *     * Redistributions of source code must retain the above copyright
 *       notice, this list of conditions and the following disclaimer.
 *     * Redistributions in binary form must reproduce the above copyright
 *       notice, this list of conditions and the following disclaimer in the
 *       documentation and/or other materials provided with the distribution.
 *     * Neither the name of the <organization> nor the
 *       names of its contributors may be used to endorse or promote products
 *       derived from this software without specific prior written permission.
 * 
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
 * ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL <COPYRIGHT HOLDER> BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package main


import (
    "log"
    "html"
    "strings"
    "net/http"
    "io/ioutil"
    "os"
    "os/exec"
)


func GetContent(fileName string) []byte {
    cwd, _ := os.Getwd()
    cmd := exec.Command(cwd + "/parser", fileName)
    out, err := cmd.Output()
    if err != nil {
        out = []byte(err.Error())
    }
    return out
}

func PutOnCache(fileName, cacheFileName string) []byte {
    documentHtml := GetContent(fileName)
    err := ioutil.WriteFile(cacheFileName, documentHtml, 0644)
    if err != nil {
        log.Printf("Unable to cache '%s': %s", fileName, err.Error())
    }
    return documentHtml
}

func Index(w http.ResponseWriter, req *http.Request) {
    path := html.EscapeString(req.URL.Path[1:])
    if strings.Index(path, "/..") > -1 {
        log.Println("Invalid url ", path)
        return
    }
    arguments := strings.Split(html.EscapeString(req.URL.RawQuery), "?")
    if arguments[0] == "edit" {
        log.Println("Going into edit mode ...")
        w.Write([]byte("Editing document ..."))
        return
    }

    // First check if there's a cached HTML of the requested document
    cachedDocument := "cached/" + path + ".html"
    document := "data/" + path + ".txt"

    documentHtml := []byte{}
    documentHtml, err := ioutil.ReadFile(cachedDocument)
    if err != nil { // Doesn't exist, try to open unformatted text and generate HTML
        // from it.
        _, err := os.Stat(document)
        if err != nil {
            w.Header().Set("Content-Type", "text/html")
            w.Write([]byte("Document was not found - <a href=\"" + path + "?edit\">create</a>"))
            log.Printf("Document '%s' was not found", path)
        } else {
            log.Printf("Document '%s' not cached, generating HTML ...", path)
            documentHtml := PutOnCache(document, cachedDocument)
            w.Header().Set("Content-Type", "text/html")
            w.Write(documentHtml)
        }
    } else {
        statCached, _ := os.Stat(cachedDocument)
        statSource, err := os.Stat(document)
        if err != nil { // Cache exists but the source text file doesn't
            w.Header().Set("Content-Type", "text/html")
            w.Write([]byte("Document not found - <a href=\"" + path + "?edit\">create</a>"))
            log.Printf("Cache for '%s' exists but source text not found, removing cache ...", path)
            // Remove cached file
            err := os.Remove(cachedDocument)
            if err != nil {
                log.Printf("Error removing unused cache for '%s': %s", path, err.Error())
            }
        } else {
            timeCached := statCached.ModTime()
            timeSource := statSource.ModTime()
            if timeSource.Unix() > timeCached.Unix() {
                log.Printf("Source text for '%s' was modified, updating cache ...", path)
                documentHtml = PutOnCache(document, cachedDocument)
            }
            w.Header().Set("Content-Type", "text/html")
            w.Write(documentHtml)
        }
    }
}

func main() {
	http.HandleFunc("/", Index)
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		log.Fatal(err)
	}
}
