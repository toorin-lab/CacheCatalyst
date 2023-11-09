# caddy-cache
A modified version of Caddy web server.
It includes a "last-modified" attribute in the tag of local objects.

## Generate sample html page

### Install Jinja2

```bash
pip install Jinja2
```

### Run generate command
```bash
python shell.py -media-dir YOUR_MEDIA_PATH -prefix PREFIX_INSTEAD_MEDIA_PATH -init-dataset DL_PIC_COUNT
```

this command (optionally) download DL_PIC_COUNT pictures from unsample.net website, export them in YOUR_MEDIA_PATH and generate a html page with DL_PIC_COUNT img tag with PREFIX_INSTEAD_MEDIA_PATH prefix and all images listed in YOUR_MEDIA_PATH.

example:
```bash
python shell.py -media-dir /home/divar/Desktop/x -prefix http://localhost -init-dataset 2
```

## Sample Caddyfile

You can use this sample configuration in `Caddyfile` file in `caddy` dir.
```
localhost:80 {
    root * /home/divar/Desktop/x
    file_server
}
```

Note: run caddy file with --config Caddyfile arguments.
