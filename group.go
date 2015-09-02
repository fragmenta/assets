package assets

import (
	"bytes"
	"fmt"
	"github.com/fragmenta/assets/internal/cssmin"
	"github.com/fragmenta/assets/internal/jsmin"
	"io/ioutil"
	"path"
)

// A sortable file array
type fileArray []*File

func (a fileArray) Len() int           { return len(a) }
func (a fileArray) Less(b, c int) bool { return a[b].name < a[c].name }
func (a fileArray) Swap(b, c int)      { a[b], a[c] = a[c], a[b] }

// Group holds a name and a list of files (images, scripts, styles)
type Group struct {
	name       string
	files      fileArray
	stylehash  string // the hash of the compiled group css file (if any)
	scripthash string // the hash of the compiled group js file (if any)
}

// Styles returns an array of file names for styles
func (g *Group) Styles() []*File {
	var styles []*File

	for _, f := range g.files {
		if f.Style() {
			styles = append(styles, f)
		}
	}

	return styles
}

// Scripts returns an array of file names for styles
func (g *Group) Scripts() []*File {
	var scripts []*File

	for _, f := range g.files {
		if f.Script() {
			scripts = append(scripts, f)
		}
	}

	return scripts
}

// CalculateAssetHashes calculates latest hashes given our files
// This is called after compiling individual assets
func (g *Group) Compile(dst string) error {
	var scriptHashes, styleHashes string
	var scriptWriter, styleWriter bytes.Buffer

	for _, f := range g.files {
		if f.Script() {
			scriptHashes += f.hash
			scriptWriter.Write(f.bytes)
			scriptWriter.WriteString("\n\n")
		} else if f.Style() {
			styleHashes += f.hash
			styleWriter.Write(f.bytes)
			styleWriter.WriteString("\n\n")
		}
	}
	// Generate hashes for the files concatted using our existing file hashes as input
	// NB this is not the hash of the minified file
	g.scripthash = bytesHash([]byte(scriptHashes))
	g.stylehash = bytesHash([]byte(styleHashes))

	// Write out this group's minified concatted files
	err := g.writeFiles(dst, scriptWriter, styleWriter)

	// Reset the buffers on our files, which we no longer need
	for _, f := range g.files {
		f.bytes = nil
	}

	return err
}

// writeScript
func (g *Group) writeFiles(dst string, scriptWriter, styleWriter bytes.Buffer) error {
	var err error

	// Minify CSS
	miniCSS := cssmin.Minify(styleWriter.Bytes())
	err = ioutil.WriteFile(g.StylePath(dst), miniCSS, permissions)
	if err != nil {
		return err
	}

	// Minify JS
	minijs, err := jsmin.Minify(scriptWriter.Bytes())
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(g.ScriptPath(dst), minijs, permissions)
	if err != nil {
		return err
	}

	// Now reset our bytes buffers
	scriptWriter.Reset()
	styleWriter.Reset()

	return nil
}

func (g *Group) AddAsset(n, h string) {
	file := &File{name: n, hash: h}
	g.files = append(g.files, file)
}

// ParseFile adds this asset to our list of files, along with a fingerprint based on the content
func (g *Group) ParseFile(p string, dst string) error {

	// Create the file
	file, err := NewFile(p)
	if err != nil {
		return err
	}
	g.files = append(g.files, file)

	dstf := file.AssetPath(dst)
	if file.Newer(dstf) {

		// Copy file over to assets folder in dst
		err = file.Copy(dstf)
		if err != nil {
			return err
		}

	}

	return nil
}

func (g *Group) String() string {
	return fmt.Sprintf("%s:%d", g.name, len(g.files))
}

func (g *Group) StyleName() string {
	return fmt.Sprintf("%s-%s.min.css", g.name, g.stylehash)
}

func (g *Group) StylePath(dst string) string {
	return path.Join(dst, "assets", "styles", g.StyleName())
}

func (g *Group) ScriptName() string {
	return fmt.Sprintf("%s-%s.min.js", g.name, g.scripthash)
}

func (g *Group) ScriptPath(dst string) string {
	return path.Join(dst, "assets", "scripts", g.ScriptName())
}

// MarshalJSON generates json for this collection, of the form {group:{file:hash}}
func (g *Group) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(fmt.Sprintf(`"%s":{"scripts":"%s","styles":"%s","files":{`,
		g.name, g.scripthash, g.stylehash))

	for i, f := range g.files {
		fb, err := f.MarshalJSON()
		if err != nil {
			return nil, err
		}
		b.Write(fb)
		if i+1 < len(g.files) {
			b.WriteString(",")
		}
	}

	b.WriteString("}}")

	return b.Bytes(), nil
}
