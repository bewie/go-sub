package downloader

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"reflect"
	xmlrpc "github.com/sqp/go-xmlrpc"
)

const (
	openSubtiltleEndpoint = "http://api.opensubtitles.org/xml-rpc"
	chunkSize             = 65536
)

// Query ..
type Query struct {
	ListArgs  []interface{}
	Items     []*SubInfo
	Hash      uint64
	UserAgent string
	Token     string
	File      string
}

// SubInfo ...
type SubInfo struct {
	SubFromTrusted   string
	IDMovie          string
	SubDownloadLink  string
	MovieReleaseName string
	MatchedBy        string
	SubFileName      string
	LanguageName     string
	SubLanguageID    string
	IDSubtitleFile   string
	MovieHash        string
	SubFormat        string
	MovieKind        string
	SubHD            string
	UserRank         string
	SubAddDate       string
	SubDownloadsCnt  string
	IDMovieImdb      string
	UserNickName     string
}

// NewDL ...
func NewDL(dir, filename string) *Query {
	h, _ := Hash(fmt.Sprintf("%s/%s", dir, filename))

	return &Query{
		Hash:      h,
		UserAgent: "test",
		File:      filename,
	}
}

// CleanListArgs ..
func (q *Query) CleanListArgs() *Query {
	var c []interface{}
	q.ListArgs = c

	return q
}

// Process a xmlrpc call on OpenSubtitles.org server.
func call(name string, args ...interface{}) (xmlrpc.Struct, error) {
	res, e := xmlrpc.Call(openSubtiltleEndpoint, name, args...)
	if e == nil {
		if data, ok := res.(xmlrpc.Struct); ok {
			return data, e
		}
	}
	return nil, e
}

// GetLangMap ...
func GetLangMap(code string) string {
	// LangMapping See http://www.opensubtitles.org/addons/export_languages.php
	langMapping := map[string]string{
		"en": "eng",
		"fr": "fre",
		"de": "ger",
		"ca": "cat",
	}
	return langMapping[code]
}

// Connect : Initiate connection to OpenSubtitles.org to get a valid token.
func (q *Query) Connect() error {
	res, e := call("LogIn", "", "", "en", q.UserAgent)
	switch {
	case e != nil:
		return e
	case res == nil || len(res) == 0:
		return errors.New("connection problem")
	}

	if token, ok := res["token"].(string); ok {
		q.Token = token
		return nil
	}
	return errors.New("OpenSubtitles Token problem")
}

// Search ...
func (q *Query) Search() (items []*SubInfo, err error) {
	if q.Token == "" {
		err = q.Connect()
	}
	switch {
	case err != nil:
		return nil, err
	case q.Token == "":
		return nil, errors.New("invalid token")
	}

	searchData, e := call("SearchSubtitles", q.Token, q.ListArgs)
	if e != nil {
		return nil, e
	}
	for k, v := range searchData {
		if k == "data" {
			for _, data := range v.(xmlrpc.Array) { // Array of data
				if vMap, ok := data.(xmlrpc.Struct); ok {
					items = append(items, parseSubMap(vMap))
					// parseSubMap(vMap)
				}
			}
		}
	}
	return items, nil
}

func parseSubMap(parseMap map[string]interface{}) *SubInfo {
	typ := reflect.TypeOf(SubInfo{})
	n := typ.NumField()

	item := &SubInfo{}
	elem := reflect.ValueOf(item).Elem()

	for i := 0; i < n; i++ { // Parsing all fields in SubInfo type
		field := typ.Field(i)
		if v, ok := parseMap[field.Name]; ok { // Got matching row in map
			if elem.Field(i).Kind() == reflect.TypeOf(v).Kind() { // Types are compatible.
				elem.Field(i).Set(reflect.ValueOf(v))
			} else {
				fmt.Println("XML Import Field mismatch", field.Name, elem.Field(i).Kind(), reflect.TypeOf(v).Kind())
			}
		}
	}
	return item
}

//  - Import from : https://github.com/oz/osdb/blob/master/osdb.go

// HashFile generates an OSDB hash for an *os.File.
func HashFile(file *os.File) (hash uint64, err error) {
	fi, err := file.Stat()
	if err != nil {
		return
	}
	if fi.Size() < chunkSize {
		return 0, fmt.Errorf("File is too small")
	}

	// Read head and tail blocks.
	buf := make([]byte, chunkSize*2)
	err = readChunk(file, 0, buf[:chunkSize])
	if err != nil {
		return
	}
	err = readChunk(file, fi.Size()-chunkSize, buf[chunkSize:])
	if err != nil {
		return
	}

	// Convert to uint64, and sum.
	var nums [(chunkSize * 2) / 8]uint64
	reader := bytes.NewReader(buf)
	err = binary.Read(reader, binary.LittleEndian, &nums)
	if err != nil {
		return 0, err
	}
	for _, num := range nums {
		hash += num
	}

	return hash + uint64(fi.Size()), nil
}

// Hash generates an OSDB hash for a file.
func Hash(path string) (uint64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return HashFile(file)
}

// Read a chunk of a file at `offset` so as to fill `buf`.
func readChunk(file *os.File, offset int64, buf []byte) (err error) {
	n, err := file.ReadAt(buf, offset)
	if err != nil {
		return
	}
	if n != chunkSize {
		return fmt.Errorf("Invalid read %v", n)
	}
	return
}
