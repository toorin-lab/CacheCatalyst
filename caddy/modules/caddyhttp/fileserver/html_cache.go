package fileserver

import (
	"bufio"
	"bytes"
	"encoding/json"
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

func fileLocatedInThisHost(url string) bool {
	return strings.Contains(url, "localhost") || !strings.Contains(strings.ToLower(url), "http")
}

func getNormalizedFileName(fileName string) string {
	qMarkIndex := strings.Index(fileName, "?")
	if qMarkIndex == -1 {
		return fileName
	}

	return fileName[:qMarkIndex]
}

func getEtagJson(fileSystem fs.FS, rootDir, htmlString string) (string, string, error) {
	etagMap := make(map[string]string)
	root, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		return "", "", err
	}

	imageTags := findTags(root, []atom.Atom{atom.Img, atom.Link})
	for _, imgNode := range imageTags {
		var src *html.Attribute
		for _, attr := range imgNode.Attr {
			attrCopy := attr
			if attr.Key == "src" {
				src = &attrCopy
				break
			} else if attr.Key == "href" {
				src = &attrCopy
				break
			}
		}

		if src == nil || !fileLocatedInThisHost(src.Val) {
			continue
		}

		fileName := strings.TrimSuffix(caddyhttp.SanitizedPathJoin(rootDir, getNormalizedFileName(src.Val)), "/")
		stat, err := fs.Stat(fileSystem, fileName)
		if err != nil {
			continue
		}

		etagMap[src.Val] = calculateEtag(stat)
	}

	jsCode := `
if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('sw.js').then(function() {
        return navigator.serviceWorker.ready;
    }).then(
		(reg) => {
			const renewFunc = function(event) {
				console.log("unload event.");
				reg.active.postMessage({
						type: 'renew',
				});
    		};
			window.onbeforeunload = renewFunc;
		}
	).catch(function(error) {
        console.log('Error : ', error);
    });

	
}
`
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
		return "", "", err
	}

	etagJson, err := json.Marshal(etagMap)
	if err != nil {
		return "", "", err
	}

	return buffer.String(), string(etagJson), nil
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
			Val: stat.ModTime().UTC().Format("2006-01-02 15:04:05"),
		})
	}
	buffer := new(bytes.Buffer)
	err = html.Render(bufio.NewWriter(buffer), root)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
