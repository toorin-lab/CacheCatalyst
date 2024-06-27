package fileserver

import (
	"errors"
	"io"
	"os"
)

func getServiceWorkerFullPath(root string) string {
	return root + "/sw.js"
}

func getServiceWorkerRelativePath() string {
	return "modules/caddyhttp/fileserver/sw.js"
}

func loadCacheV2ServiceWorker(root string) error {
	return copyFile("sw.js", root+"sw.js")
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
			return err
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
