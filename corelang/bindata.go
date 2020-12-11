// Code generated by go-bindata. (@generated) DO NOT EDIT.

// Package corelang generated by go-bindata.// sources:
// lua/hostlang.lua
package corelang

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// ModTime return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _luaHostlangLua = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x56\x4d\x8f\xa3\x46\x10\xbd\xf3\x2b\x4a\x5c\xc0\x12\x94\xb3\x4a\x4e\xa3\xf1\x28\xb9\xe5\x90\x53\xb4\xca\x65\xb5\xb2\x7a\xed\xc2\x6e\x2d\x2e\x50\x77\x01\x9e\x5d\xcd\x7f\x8f\xba\x9b\xc6\x30\x83\x67\xa2\xf8\x04\x5d\xaf\xeb\xbd\xfa\xc4\x49\x59\xc2\x59\xa4\x7d\xd8\x6e\xeb\x4e\x95\x9d\x25\x63\xb1\x31\xa7\xed\xa0\xbf\xeb\xed\x1f\x9d\x34\x17\x75\xd2\x87\xcf\xea\x5b\x4d\x36\x82\xed\xc3\x76\x3b\x0c\x03\xd6\x9d\xf2\xd8\x56\xd7\xdb\x4f\xbf\xe2\x6f\xf8\x09\xcf\x72\xa9\xdf\x83\x1d\x1a\x16\x62\xb1\x01\x98\xfc\xd9\x58\xf9\x4b\xf1\x09\x76\xf0\xf3\x05\xfc\xaf\x2c\xe1\xf3\x99\x80\xbb\x0b\x19\x7d\x00\x56\x17\xb2\xad\x3a\x50\x00\xff\xa3\xcc\xdf\x54\x79\xf8\xec\x1d\x5b\xd3\x48\x23\xcf\x2d\x39\x4b\xe2\xdc\xec\x5b\x65\x88\x05\x76\xc0\xba\x2e\xc2\x91\xa8\xd3\xe2\xdd\x76\x55\xa5\xaf\xb0\x83\xec\x31\x3c\x3e\x65\xa3\xa5\x57\x75\x47\x73\xec\xe8\x3b\x1d\x65\xa5\xc9\xcb\x5c\x0f\x5e\x24\x46\x50\x96\x70\x21\x51\xe2\x12\x06\x55\x63\x62\x20\x49\x52\x75\x7c\x10\xdd\x30\xc4\xa0\x71\xdf\x2b\xa3\x1d\x30\x17\x75\xda\xb8\xab\x07\x43\x4a\x08\x14\x30\x0d\x10\xad\x5e\x40\xdd\x1c\x54\x0d\xbd\xa3\x99\x07\x32\x0b\xc2\x1d\xbe\x78\x6c\x8f\x6f\xe4\xba\x63\x4b\x32\x49\xcb\xfb\x02\x16\xfa\x37\x1e\x62\x48\x3a\xc3\xd0\x27\xc4\xc7\x35\xc5\xa3\xbf\x75\xbd\xb1\x64\x2b\xba\x19\x76\xf7\xc2\xf6\x28\x5e\x57\x3c\xca\xe1\x7b\x72\x5a\xa5\xcd\xba\x16\x67\xf9\x3f\x42\x5a\x33\x29\x71\x2e\x3e\x94\x31\xe5\x0f\xf7\x7b\xcd\x47\xba\xe6\x5c\x40\x28\xc9\x66\x46\xab\xe4\xd6\x6b\xbf\x67\x88\xe1\x6d\x06\x08\x07\x4e\x10\xec\xc0\xa8\xe1\x44\xe2\x3c\xc5\x7b\xc1\x97\xae\xe6\x38\x39\x13\x27\x30\xfe\x46\x89\x37\xb3\xb7\x38\xbd\x33\xeb\xda\xc0\x7c\x09\x57\xbe\x7e\x10\x9b\x34\x56\x8c\xe6\x53\xce\x8b\x46\x49\x1f\x63\xd1\x53\x44\x7e\xa8\xba\xba\x76\x03\x9b\x6f\x10\xd3\x5d\x8a\x98\x33\x8e\xb3\xd4\x18\x48\x1f\x3b\xfe\xce\xcd\xc0\x4f\xa9\xb3\x3f\xa5\x1f\x70\x32\x0d\xaf\x52\x5a\x40\x1f\xe8\xa7\x6c\x8e\x0f\x8d\x81\xcc\xf3\x64\x77\x93\x0e\x88\xf0\x26\xed\x21\xdf\x2b\x0d\x31\x4f\xbb\x51\x83\x25\xc9\x7b\x65\x0a\x48\xdd\xec\xa5\x05\x30\xfa\x21\x74\xf3\xbd\x06\x0a\xcb\xc7\xe1\xde\x5a\xb3\x90\x91\x6c\x0a\x66\x34\xce\xca\x5d\x38\x61\x9b\x95\xde\x5b\xab\x20\x86\x6d\x10\xd3\x18\x2b\xa4\x2b\xe0\x46\xa2\xd0\xb5\x66\xe1\xb5\x1e\x09\xf8\x77\xd8\xb4\xf5\x35\xbc\xc3\x18\xcb\xbd\xf3\xeb\x69\x95\xb6\x52\xb5\xa5\x35\x6a\x31\x1d\xbd\x43\x1c\x77\xf2\x8d\xb6\x80\xdb\x60\x5c\x97\x5c\x37\x21\x10\xea\x4d\x91\x74\x19\xab\x07\x4d\x62\xee\x93\xc7\xce\x5e\x09\xdb\xba\x65\x11\x17\x45\x4b\x4a\x26\x1a\x5d\xf9\xef\x86\x1b\x82\xb1\x9f\x5c\x5e\xdc\x8a\xfb\x46\x26\x5d\x0a\x8e\x9e\xb2\x2f\x19\xe2\x74\x01\x31\xfb\x1a\x3a\x77\x02\x2e\x02\x89\xb7\xa6\x0b\x4b\xec\xac\x0f\x9c\x88\xf8\x49\xf4\xb5\xd9\xbc\xe5\x9f\xa2\xc1\x14\x71\xe9\x25\x96\xea\xf5\x33\x07\xf2\xe0\xd9\x9f\x76\x2c\xba\x0e\xad\x37\x2f\xae\x0d\xc9\x2d\x4b\x28\xcb\x12\x5a\x43\x47\xaa\x34\xd3\x71\x4a\xa7\x75\x86\xff\xfe\x4b\x92\x5e\x99\x63\xfc\x2b\x70\x5b\x22\xe1\x14\x7f\x8c\x13\x4c\x36\x14\x49\x8c\x3a\x50\xfe\x4b\x01\xe9\x8f\x7c\x03\x83\x96\x33\x44\x80\x5b\x5d\x13\xda\xef\xa9\x74\x13\x2e\xb9\xcd\xd1\x2b\x63\xa8\x42\x43\x15\x99\xbd\x34\x79\x7a\xbd\x0b\x7f\x5e\x83\x3f\xdf\x85\xb7\xb0\xf3\x9f\x29\x64\x1a\x7e\x5e\x0b\x78\x7e\x99\xe7\xab\x0d\xf9\x9a\x6d\x6e\xb7\xa1\x92\x7f\x03\x00\x00\xff\xff\x4d\xbb\x8d\x52\xb0\x09\x00\x00")

func luaHostlangLuaBytes() ([]byte, error) {
	return bindataRead(
		_luaHostlangLua,
		"lua/hostlang.lua",
	)
}

func luaHostlangLua() (*asset, error) {
	bytes, err := luaHostlangLuaBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "lua/hostlang.lua", size: 2480, mode: os.FileMode(420), modTime: time.Unix(1517258895, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"lua/hostlang.lua": luaHostlangLua,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("nonexistent") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		canonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(canonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"lua": &bintree{nil, map[string]*bintree{
		"hostlang.lua": &bintree{luaHostlangLua, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(canonicalName, "/")...)...)
}
