package fileserver

import (
	"bufio"
	"bytes"
	"io/fs"
	"slices"
	"strings"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func findTags(root *html.Node, tags []atom.Atom) []*html.Node {
	current := root.FirstChild
	var foundTags []*html.Node
	for current != nil {
		if slices.Contains(tags, current.DataAtom) {
			foundTags = append(foundTags, current)
		} else {
			foundTags = append(foundTags, findTags(current, tags)...)
		}
		current = current.NextSibling
	}
	return foundTags
}

func injectLastModifiedToMediaTags(fileSystem fs.FS, rootDir, htmlString string) (string, error) {
	const LastModified = "last-modified"
	root, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		return "", err
	}

	imageTags := findTags(root, []atom.Atom{atom.Img})
	for _, imgNode := range imageTags {
		var src *html.Attribute
		oldLastModifiedIndex := -1
		for i, attr := range imgNode.Attr {
			attrCopy := attr
			if attr.Key == LastModified {
				oldLastModifiedIndex = i
			} else if attr.Key == "src" {
				src = &attrCopy
			}
		}
		if src == nil || !strings.HasPrefix(src.Val, "/") {
			continue
		}
		if oldLastModifiedIndex != -1 {
			imgNode.Attr = append(imgNode.Attr[:oldLastModifiedIndex], imgNode.Attr[oldLastModifiedIndex+1:]...)
		}

		fileName := strings.TrimSuffix(caddyhttp.SanitizedPathJoin(rootDir, src.Val), "/")
		stat, err := fs.Stat(fileSystem, fileName)
		if err != nil {
			continue
		}

		imgNode.Attr = append(imgNode.Attr, html.Attribute{
			Key: LastModified,
			Val: stat.ModTime().String(),
		})
	}
	buffer := new(bytes.Buffer)
	err = html.Render(bufio.NewWriter(buffer), root)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
