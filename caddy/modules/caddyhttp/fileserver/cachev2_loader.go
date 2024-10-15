package fileserver

import (
	"errors"
	"io"
	"os"

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
