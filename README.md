# CacheCatalyst
This is the source code of an optimized cache scheme for the Web that has been published in our paper entitled "Rethinking Web Caching: An Optimization for the Latency-Constrained Internet" in **HotNets'24**. In this paper, we discuss why the design of current web caching mechanisms is not optimal in the context of high-speed networks where latency, rather than bandwidth, is the primary bottleneck for web performance. Then, we propose an optimized caching approach in which web servers proactively provide clients with the latest validation tokens for resources during the initial step of page loading, allowing browsers to use unchanged cached content without unnecessary round trips.


## Cache V2
Customization is Added to caddy to behave more efficiently with HTTP cache.

A service worker can be accessible in `/sw.js` that registers to any index.html by appending a script to the html file if Header `X-CacheV2-Extension-Enabled` is set to `true`.
If you enable this option, DOM is being interpreted, and `link`, `img` and `script` that have `src` or `href` attributes are elicited to find etags If they have been placed in the current host.

In the end, Header `X-Etag-Config` is set by JSON etags calculated in the previous step.

## Test
To test new behavior you can see `web-benchmarking` project.
