//go:generate embed -local=false -embed=true -silent=false

package www

import (    
	"fmt"
	"os"
	"time"
    "bytes"   
    "compress/gzip"
    "encoding/base64"
    "io"
	"io/ioutil"
    "log"    
    "net/http"
	"sync"
	"errors"
)

var EmbedErrInvalid    = errors.New("invalid argument") // methods on embedFile will return this error when the receiver is nil
var EmbedErrPermission = errors.New("permission denied")
var EmbedErrExist      = errors.New("file already exists")
var EmbedErrNotExist   = errors.New("file does not exist")
var EmbedErrNotAFile   = errors.New("path is not a file")

// EmbedSetLocal tells embed to switch to local filesystem lookup by specifying 
// a filesystem path to act as your embed root. To set back to embeded mode, simply
// pass in a empty string
// example: mypkg.EmbedSetLocal("/path/to/folder") //local mode enabled
// example: mypkg.EmbedSetLocal("") //local mode disabled
var EmbedSetLocal = func(path string) {
	if path == "" {
		embedLocalMode = false
		embedLocalDir = ""
		return
	}
	
	embedLocalMode = true
	embedLocalDir = path
}

// EmbedHttpFS is the package-level generated http.FileSystem consumers can use 
// to serve the embeded content generated in this file.
var EmbedHttpFS = &embedFileSystem{}

// EmbedReadFile reads the file named by filename and returns the contents. A 
// successful call returns err == nil, not err == EOF. Because EmbedReadFile reads 
// the whole file, it does not treat an EOF from Read as an error to be reported. 
var EmbedReadFile = func(path string) ([]byte, error) {
	if embedLocalMode {
		return ioutil.ReadFile(path)
	}

	fi, ok := embedFilemap[path]

	if !ok {
		return []byte(""), EmbedErrNotExist
	}

	if fi.isDir {
		return []byte(""), EmbedErrNotAFile
	}

	if fi.decoded {
        return fi.data, nil
    }

	var data bytes.Buffer		

	b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBuffer(fi.data))
	gz, err := gzip.NewReader(b64)

	if err != nil {
		return []byte(""), err
	}

	if _, err = io.Copy(&data, gz); err != nil {
		return []byte(""), err
	}

	return data.Bytes(), nil	
}

// EmbedWriteFile writes data to a file named by filename. If the file does not 
// exist, EmbedWriteFile creates it with permissions perm; otherwise EmbedWriteFile 
// truncates it before writing.
var EmbedWriteFile = func(filename string, data []byte) error {	
	if embedLocalMode {
		return ioutil.WriteFile(filename, data, os.FileMode(0744))
	}

	fi, ok := embedFilemap[filename]

	if !ok {
		return EmbedErrNotExist
	}

	if fi.isDir {		
		return EmbedErrNotAFile
	}

	fi.mu.Lock()
	defer fi.mu.Unlock()

	if fi.decoded {
		fi.data = data
		return nil
	}

	var buf bytes.Buffer
	
	b64 := base64.NewEncoder(base64.StdEncoding, &buf)
	gz := gzip.NewWriter(b64)

	if _, err := io.Copy(gz, bytes.NewBuffer(data)); err != nil {
		return err
	}

	gz.Close()
	b64.Close()

	fi.data = buf.Bytes()
	fi.decoded = false

	return nil
}

// embedLocalMode indicates if this generated go file operates on the existing
// local filesystem or uses the embedded filesystem
var embedLocalMode = false

// embedLocalDir holds the real filesystem path to use when in local mode.
var embedLocalDir = ""

// embedFile impliments http.File interface
// see: https://golang.org/pkg/net/http/#File
type embedFile struct {
    reader   *bytes.Reader    
    info     *embedFileInfo
}

// Close ...
func (t *embedFile) Close() error {    
    return nil
}

// Read ...
func (t *embedFile) Read(p []byte) (int, error) {
    return t.reader.Read(p)
}

// Seek ...
// TODO: see if this is possible to impliment cleanly
func (t *embedFile) Seek(offset int64, whence int) (int64, error) {
    return t.reader.Seek(offset, whence)
}

// Readdir ...
// TODO: see if this is possible to impliment cleanly
func (t *embedFile) Readdir(count int) ([]os.FileInfo, error) {
    panic("Not Implimented")
    return nil, nil
}

// Stat ...
func (t *embedFile) Stat() (os.FileInfo, error) {
    return t.info, nil
}

// embedFileInfo impliments os.FileInfo
// see: https://golang.org/pkg/os/#FileInfo
type embedFileInfo struct {	
    path    string
    data    []byte
    decoded bool
	mu      *sync.Mutex

    name     string
    size     int64
    mode     uint32
	isDir    bool
}

// Name ...
func (t *embedFileInfo) Name() string {
    return t.name
}

// Size ...
func (t *embedFileInfo) Size() int64 {
    return t.size
}

// Mode ...
func (t *embedFileInfo) Mode() os.FileMode {    
    return os.FileMode(t.mode)
}

// ModTime ...
// TODO: pickup and provide actual mod time of the file during generation
func (t *embedFileInfo) ModTime() time.Time {
    return time.Now()
}

// IsDir ...
func (t *embedFileInfo) IsDir() bool {
    return t.isDir
}

// Sys ...
func (t *embedFileInfo) Sys() interface{} {    
    return nil
}

// embedFileSystem impliments http.FileSystem interface
// see: https://golang.org/pkg/net/http/#FileSystem
type embedFileSystem struct {
}

// Open ...
func (t *embedFileSystem) Open(name string) (http.File, error) {	
	if embedLocalMode {
		return os.Open(fmt.Sprintf("%v/%v", embedLocalDir, name))
	}

    f, ok := embedFilemap[name]

    if !ok {
        return nil, os.ErrNotExist
    }            

    if !f.decoded && !f.isDir { 
        var data bytes.Buffer		

        b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBuffer(f.data))
        gz, err := gzip.NewReader(b64)

        if err != nil {
            log.Println(err)
            return nil, err
        }

        if _, err := io.Copy(&data, gz); err != nil {
            log.Println(err)
            return nil, err
        }
		
        f.data = data.Bytes()
        f.decoded = true        
    }    

    file := &embedFile{
        info: f,
        reader: bytes.NewReader(f.data),
    }

    return file, nil		
}

// filemap holds the actual embeded file data
var embedFilemap = map[string]*embedFileInfo{
"/": &embedFileInfo{     
    data: []byte(""),
    name: "www",    
    isDir: true,
    size: 4096,
    mode: 2147484159,
	mu: &sync.Mutex{},
},
"/index.html": &embedFileInfo{     
    data: []byte("H4sIAAAJbogA/7LJKMnNsePlsslITUwB0fowRlJ+SqUdL5eCgoKCTYahnUdqTk6+jkJ4flFOio1+hiFYLVSNjT7YFEAAAAD//0wVRhhMAAAA"),
    name: "index.html",    
    isDir: false,
    size: 76,
    mode: 438,
	mu: &sync.Mutex{},
},
"/subfolder": &embedFileInfo{     
    data: []byte(""),
    name: "subfolder",    
    isDir: true,
    size: 8192,
    mode: 2147484159,
	mu: &sync.Mutex{},
},
"/subfolder/index.html": &embedFileInfo{     
    data: []byte("H4sIAAAJbogA/7LJMLQrLk1Ky89JSS2y0c8wtAMEAAD//yb2jSISAAAA"),
    name: "index.html",    
    isDir: false,
    size: 18,
    mode: 438,
	mu: &sync.Mutex{},
},
}