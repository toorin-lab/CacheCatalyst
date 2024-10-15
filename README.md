# caddy-cache
A modified version of Caddy web server.
It includes a "last-modified" attribute in the tag of local objects.


## Cache V2
Customization is Added to caddy to behave more efficiently with HTTP cache.

A service worker can be accessible in `/sw.js` that registers to any index.html by appending a script to the html file if Header `X-CacheV2-Extension-Enabled` is set to `true`.
If you enable this option, DOM is being interpreted, and `link`, `img` and `script` that have `src` or `href` attributes are elicited to find etags If they have been placed in the current host.

In the end, Header `X-Etag-Config` is set by JSON etags calculated in the previous step.

## Test
To test new behavior you can see `web-benchmarking` project.
