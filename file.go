package assets

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const permissions = 0744

// File stores a filename and hash fingerprint for the asset file
type File struct {
	name string
	hash string

	// Used during compile phase
	path  string
	bytes []byte
}

func NewFile(p string) (*File, error) {

	// Load file from path to get bytes
	bytes, err := ioutil.ReadFile(p)
	if err != nil {
		return &File{}, err
	}

	// Calculate hash and save it
	file := &File{
		path:  p,
		name:  path.Base(p),
		hash:  bytesHash(bytes),
		bytes: bytes,
	}
	return file, nil
}

func (f *File) Style() bool {
	return strings.HasSuffix(f.name, ".css")
}

func (f *File) Script() bool {
	return strings.HasSuffix(f.name, ".js")
}

// MarshalJSON generates json for this collection, of the form {group:{file:hash}}
func (f *File) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	s := fmt.Sprintf("\"%s\":\"%s\"", f.name, f.hash)
	b.WriteString(s)

	return b.Bytes(), nil
}

// Newer returns true if file exists at path
func (f *File) Newer(dst string) bool {

	// Check mtimes
	stat, err := os.Stat(f.path)
	if err != nil {
		return false
	}
	srcM := stat.ModTime()
	stat, err = os.Stat(dst)

	// If the file doesn't exist, return true
	if os.IsNotExist(err) {
		return true
	}

	// Else check for other errors
	if err != nil {
		return false
	}

	dstM := stat.ModTime()

	return srcM.After(dstM)

}

// Copy our bytes to dstpath
func (f *File) Copy(dst string) error {
	err := ioutil.WriteFile(dst, f.bytes, permissions)
	if err != nil {
		return err
	}
	return nil
}

func (f *File) AssetPath(dst string) string {
	folder := "styles"
	if f.Script() {
		folder = "scripts"
	}
	return path.Join(dst, "assets", folder, f.name)
}

func (f *File) String() string {
	return fmt.Sprintf("%s:%s", f.name, f.hash)
}
