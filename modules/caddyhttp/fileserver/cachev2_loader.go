package fileserver

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func getServiceWorkerFullPath(root string) string {
	return caddyhttp.SanitizedPathJoin(root, "/sw.js")
}

func getServiceWorkerRelativePath() string {
	return "modules/caddyhttp/fileserver/sw.js"
}

func loadCacheV2ServiceWorker(root string) error {
	return copyFile("modules/caddyhttp/fileserver/sw.js", root+"sw.js")
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}

	buf := make([]byte, 1024)

	for {
		n, err := source.Read(buf)
		if err != nil && errors.Is(err, io.EOF) {
			return nil
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}

	return nil
}

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

func registerServiceWorker(htmlString string, swFileName string) (string, error) {
	root, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		return "", err
	}

	jsCode := fmt.Sprintf(`
if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/%s').then(function() {
        return navigator.serviceWorker.ready;
    }).catch(function(error) {
        console.log('Error : ', error);
    });
}
`, swFileName)
	body := findTags(root, []atom.Atom{atom.Body})[0]
	script := &html.Node{
		Type:     html.ElementNode,
		Data:     "script",
		DataAtom: atom.Script,
	}
	script.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: jsCode,
	})
	if body.FirstChild != nil {
		body.InsertBefore(script, body.FirstChild)
	} else {
		body.AppendChild(script)
	}

	buffer := new(bytes.Buffer)
	w := bufio.NewWriter(buffer)
	err = html.Render(w, root)
	err = w.Flush()
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
