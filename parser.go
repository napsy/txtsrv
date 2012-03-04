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
    "strings"
    "os"
    "fmt"
    "io/ioutil"
)


type LinkInfo struct {
    Begin, End int
    Title, Link string
}

func IsSection(line []byte) (section bool, level int, text string) {
    level_begin := 0
    level_end := 0

    for _, ch := range(line) {
        if ch == '=' {
            if len(text) > 0 { // validate section
                level_end++
            } else {
                level_begin++
            }
        } else {
            if level_begin == 0 {
                section = false
                break
            }
            text += string(ch)
        }
    }
    if (level_begin != level_end) || level_begin == 0 || level_end == 0 {
        //panic("missing '=' characters at the beginning or in the end")
        section = false
    } else {
        section = true
    }
    level = level_begin
    return section, level, text
}
/*
   Klikni [[sem kajle|http://www.example.com]] za primer.
          ^                                  ^
          8                                  42
 */
func FindNextLink(line []byte) *LinkInfo {
    var (info LinkInfo
         link_stage int
         end_search bool)

    for i, ch := range(line) {
        if end_search {
            break
        }
        switch (ch) {
            case '[':
                if link_stage == 1 {
                    info.Begin = i - 1
                }
                link_stage++
            case ']':
                if link_stage == 4 {
                    end_search = true
                    info.End = i + 1
                    continue
                }
                link_stage++
            case '|':
                link_stage++
            default:
                if link_stage == 1 { // ignore single '[' char
                    link_stage = 0
                    continue
                }
                if link_stage == 4 {
                    fmt.Println("Missing ] char.")
                    end_search = true
                    link_stage = 0
                    continue
                }
                if link_stage == 2 { // appending to title
                    info.Title += string(ch)
                } else if link_stage == 3 { // appending to link
                    info.Link += string(ch)
                }
        }
    }
    if (link_stage != 4) {
        return nil
    }
    return &info
}

func TestFindNextLink() {
    var link *LinkInfo
    sample1 := `Th[is i]s [[a complete link|http://example.com]] to click on. And a [[second link|http://google.com]] for a search engine`
    sample1_slice := []byte(sample1)
    for {
        link = FindNextLink(sample1_slice)
        if link == nil {
            break
        }
        fmt.Printf("LINK[%d-%d] -- '%s' => '%s'\n", link.Begin, link.End, link.Title, link.Link)
        sample1_slice = sample1_slice[link.End:]
    }
    fmt.Println("TestFindNextLink() passed!")
}

func ProcessLinks(line []byte) (out string) {
    var link *LinkInfo
    var cur_pos int
    for {
        link = FindNextLink(line[cur_pos:])
        if link == nil {
            break
        }
        out += string(line[cur_pos:cur_pos+link.Begin])
        out += "<a href=\"" + link.Link + "\">" + link.Title + "</a>"
        cur_pos =cur_pos + link.End
    }
    out += string(line[cur_pos:])
    return out
}

func ProcessContext(ctx []byte) (out string) {
    lines := strings.Split(string(ctx), "\n")
    for _, line := range(lines) {
        s, l, text := IsSection([]byte(line))
        if s {
            out += "<h" + string(l+48) + ">" + text + "</h" + string(l+48) + ">"
        } else {
            links := ProcessLinks([]byte(line))
            out += links
        }
    }
    return out
}

func TestIsSection() {
    var (s bool
        l int
        text string)
    line1 := "=== title 1 ==="
    line2 := "== another title ==="
    line3 := "not a section ==="

    s, l, text = IsSection([]byte(line1))
    if s != true || l != 3 || text != " title 1 " {
        panic("TestIsSection() failed: line1")
    }
    s, l, text = IsSection([]byte(line2))
    if s != false {
        panic("TestIsSection() failed: line2")
    }
    s, l, text = IsSection([]byte(line3))
    if s != false {
        panic("TestIsSection() failed: line3")
    }
    fmt.Println("TestIsSection() passed!")
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Please specify input file!")
        return
    }
    // TODO: implement file caching.
    if ctx, err := ioutil.ReadFile(os.Args[1]); err == nil {
        out :=ProcessContext(ctx)
        fmt.Println(out)
    } else {
        fmt.Println("Error: " + err.Error())
    }
}

