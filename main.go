package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/bewie/go-sub/downloader"
)

func main() {
	// CLI and default options
	var lang = flag.String("lang", "fr", "Lang for subtitles")
	var path = flag.String("p", ".", "Relative or absolute path to video file(s)")
	var debug = flag.Bool("d", false, "Enable debug output")
	flag.Parse()
	downloadSubtitles(*lang, *path, *debug)
}

func downloadSubtitles(lang, p string, debug bool) {
	var files []os.FileInfo
	var dir string
	f, err := os.Stat(p)
	if err != nil {
		log.Fatal(err)
	}
	re := regexp.MustCompile("(video/)")
	if f.IsDir() {
		files, err = ioutil.ReadDir(p)
		dir = p
		if err != nil {
			log.Fatal(err)
		}
	} else {
		files = append(files, f)
		dir = path.Dir(p)
	}

	for _, f := range files {
		if f.IsDir() {
			fmt.Printf("%s/%s \n", dir, f.Name())
			downloadSubtitles(lang, fmt.Sprintf("%s/%s", dir, f.Name()), debug)
		} else {
			// test if a subtitle already exist for this file
			ext := mime.TypeByExtension(path.Ext(f.Name()))
			//  If is a video file
			if re.MatchString(ext) {
				if debug {
					fmt.Printf(" -> %s \n", f.Name())
				}
				strFileName := fmt.Sprintf("%s/%s.srt", dir, basename(f.Name()))
				if _, err := os.Stat(strFileName); os.IsNotExist(err) {
					downloadSubtitles4File(dir, f.Name(), lang, debug)
				} else {
					fmt.Printf("    Skipping: %s \n", f.Name())
				}
			}
		}
	}
}

func downloadSubtitles4File(dir, filename, lang string, debug bool) {
	var episode, season, show string
	movieShowPattern := `^(?P<movie>.*)[.\[( ](?P<year>(?:19|20)\d{2})`
	tvShowPatterns := [...]string{
		"^(?P<show>.+?)[ \\._\\-][Ss](?P<season>[0-9]{2})[ \\.\\-]?(?P<episode>[0-9]{2})[^0-9]*$",                                  // 0 - # foo.s0101
		"^(?P<show>.+?)[ ._-]\\[?(?P<season>[0-9]+)[xX](?P<episode>[0-9]+)\\]?[^/]*$",                                              // 1 - # foo.1x09*
		"^(?P<show>.+?)[ \\._\\-]?\\[(?P<season>[0-9]+?)[.](?P<episode>[0-9]+?)\\][ \\._\\-]?[^/]*$",                               // 2 - # foo - [01.09]
		"^(?P<show>.+?)[ ]?[ \\._\\-][ ]?[Ss](?P<season>[0-9]+)[\\.\\- ]?[Ee]?[ ]?(?P<episode>[0-9]+)[^\\/]*$",                     // 3 - # Foo - S2 E 02
		"(?P<show>.+)[ ]-[ ][Ee]pisode[ ]\\d+[ ]\\[[sS][ ]?(?P<season>\\d+)([ ]|[ ]-[ ]|-)([eE]|[eE]p)[ ]?(?P<episode>\\d+)\\].*$", // 4 - # Show - Episode 9999 [S 12 - Ep 131]
		"^(?P<show>.+)[ \\._\\-](?P<season>[0-9]{1})(?P<episode>[0-9]{2})[\\._ -][^\\/]*$",                                         // 5 - # foo.103*
		"^(?P<show>.+?)[ \\._\\-]\\[?[Ss](?P<season>[0-9]+)[\\. _-]?[Ee]?(?P<episode>[0-9]+)\\]?[^\\/]*$"}                          // 6 - # foo.s01.e01, foo.s01_e01

	dl := downloader.NewDL(dir, filename)
	err := dl.Connect()
	if err != nil {
		fmt.Printf("ERROR login: %s\n", err)
	}

	if re := regexp.MustCompile(movieShowPattern); re.MatchString(filename) {
		if debug {
			fmt.Printf("Match Movie\n")
		}
	} else {
		for _, pattern := range tvShowPatterns {
			re := regexp.MustCompile(pattern)
			if re.MatchString(filename) {
				groupNames := re.SubexpNames()
				for i, match := range re.FindStringSubmatch(filename) {
					switch {
					case groupNames[i] == "show":
						show = match
					case groupNames[i] == "season":
						season = match
					case groupNames[i] == "episode":
						episode = match
					}
				}
				break // match tv pattern
			}
		}
	}

	stat, _ := os.Stat(fmt.Sprintf("%s/%s", dir, filename))
	fileSize := fmt.Sprint(stat.Size())
	fileHash := fmt.Sprint(dl.Hash)
	dl.ListArgs = append(dl.ListArgs, map[string]string{"moviehash": fileHash, "moviebytesize": fileSize})
	if debug {
		fmt.Printf("- Search by Hash `%s` [%s]\n", filename, fileHash)
	}

	items, err := dl.Search()
	if err != nil {
		log.Fatal("On search ", err)
	}
	if len(items) == 0 {
		dl.CleanListArgs()
		if episode != "" && season != "" && show != "" {
			dl.ListArgs = append(dl.ListArgs, map[string]string{"query": show, "season": season, "episode": episode, "sublanguageid": downloader.GetLangMap(lang)})
		} else {
			dl.ListArgs = append(dl.ListArgs, map[string]string{"query": filename, "sublanguageid": downloader.GetLangMap(lang)})
		}
		if debug {
			fmt.Printf("- Not found, search by name `%s`\n", filename)
		}
		items, err = dl.Search()
		if err != nil {
			log.Fatal("On search ", err)
		}
	}

	if len(items) > 0 {
		strFileName := fmt.Sprintf("%s/%s.srt", dir, basename(filename))
		err := dl.Get(items[0].SubDownloadLink, strFileName) // Get the first one... probably could be smarter
		if err != nil {
			log.Fatal("On Get ", err)
		}
		fmt.Printf("- %s : OK\n", filename)
	} else {
		fmt.Printf("- Subtitles not found for `%s`\n", filename)
	}
}
func basename(s string) string {
	n := strings.LastIndexByte(s, '.')
	if n >= 0 {
		return s[:n]
	}
	return s
}
