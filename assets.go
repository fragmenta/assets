// Compilation of assets in asset groups for web servers
package assets

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Collection holds the complete list of groups
type Collection struct {
	production bool
	path       string
	groups     []*Group
}

// New returns a new assets.Collection
func New(production bool) *Collection {
	c := &Collection{
		production: production,
		path:       "secrets/assets.json",
	}
	return c
}

func (c *Collection) Group(name string) *Group {
	for _, g := range c.groups {
		if g.name == name {
			return g
		}
	}
	return &Group{name: name}
}

// Group returns the named group if it exists, or creates it if not
func (c *Collection) FetchOrCreateGroup(name string) *Group {
	for _, g := range c.groups {
		if g.name == name {
			return g
		}
	}
	g := &Group{name: name}
	c.groups = append(c.groups, g)
	return g
}

// MarshalJSON generates json for this collection, of the form {group:{file:hash}}
func (c *Collection) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	b.WriteString("{")

	for i, g := range c.groups {
		gb, err := g.MarshalJSON()
		if err != nil {
			return nil, err
		}
		b.Write(gb)
		if i+1 < len(c.groups) {
			b.WriteString(",")
		}
	}

	b.WriteString("}")

	//	fmt.Printf("ASSETS:%s\n", b.String())

	return b.Bytes(), nil
}

// Save the assets to a file after compilation
func (c *Collection) Save() error {

	// Get a representation of each file and group as json
	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return fmt.Errorf("Error marshalling assets file %s %v", c.path, err)
	}

	// Write our assets json file to the path
	err = ioutil.WriteFile(c.path, data, 0644)
	if err != nil {
		return fmt.Errorf("Error writing assets file %s %v", c.path, err)
	}

	return nil
}

// Load the asset groups from the assets json file
// Call this on startup from your app to read the asset details after assets are compiled
func (c *Collection) Load() error {

	// Make sure we reset groups, in case we compiled
	c.groups = make([]*Group, 0)

	// Read our assets json file from the path
	file, err := ioutil.ReadFile(c.path)
	if err != nil {
		return fmt.Errorf("Error opening assets file %s %v", c.path, err)
	}

	// Unmarshal json Groups/sections/Files
	var data map[string]map[string]interface{}
	err = json.Unmarshal(file, &data)
	if err != nil {
		return fmt.Errorf("Error reading assets %s %v", c.path, err)
	}

	//	fmt.Printf("Loaded assets :%s\n", data)

	// Walk through data groups, creating our groups from it
	// or fetching existing ones
	for d, dv := range data {
		g := c.FetchOrCreateGroup(d)
		for k, v := range dv {

			switch k {
			case "scripts":
				g.scripthash = v.(string)
			case "styles":
				g.stylehash = v.(string)
			case "files":
				for n, h := range v.(map[string]interface{}) {
					g.AddAsset(n, h.(string))
				}
			}

		}

	}

	// Sort files by name after load
	for _, g := range c.groups {
		sort.Sort(g.files)
	}

	return nil
}

// Compile images, styles and scripts asset folders from src into dst (minifying and amalgamating)
func (c *Collection) Compile(src string, dst string) error {

	// First scan the directory for files we're interested in
	files, err := collectAssets(filepath.Clean(src), []string{"js", "css", ".jpg", ".png"})
	if err != nil {
		return err
	}

	// Handle each asset by adding it to a group
	// For now we only handle one group - the app group
	// later we might create groups for any folders with assets/images/xxx etc
	for _, f := range files {
		g := c.FetchOrCreateGroup("app")

		// Load the file bytes and generate a hash
		// copying it out to dst if require
		g.ParseFile(f, dst)

	}

	// For all our groups, compile them to one file, calculate global hash
	for _, g := range c.groups {
		err := g.Compile(dst)
		if err != nil {
			return err
		}
	}

	// Now save a representation of the groups/files to our json file
	err = c.Save()
	if err != nil {
		return err
	}

	return nil
}

// Collect the assets with this extension under src
func collectAssets(src string, extensions []string) ([]string, error) {

	assets := []string{}

	// filepath.Glob doesn't appear to support ** or {}
	// this should catch
	// src/app/images/img.png
	// src/app/assets/images/img.png
	// src/app/assets/images/group/img.png
	for _, e := range extensions {
		pattern := path.Join(src, "*/*/*."+e)
		files, err := filepath.Glob(pattern)
		if err != nil {
			return assets, err
		}
		assets = append(assets, files...)
		pattern = path.Join(src, "*/*/*/*."+e)
		files, err = filepath.Glob(pattern)
		if err != nil {
			return assets, err
		}
		assets = append(assets, files...)
		pattern = path.Join(src, "*/*/*/*/*."+e)
		files, err = filepath.Glob(pattern)
		if err != nil {
			return assets, err
		}
		assets = append(assets, files...)
	}

	return assets, nil

}

// bytesHash returns the sha hash of some bytes
func bytesHash(bytes []byte) string {
	//sum := sha256.Sum256(bytes)
	sum := sha1.Sum(bytes)

	// In golang 1.5, we can use RawURLEncoding, for now we trim padding
	e := base64.URLEncoding
	s := e.EncodeToString([]byte(sum[:]))
	return strings.TrimRight(s, "=")
}
