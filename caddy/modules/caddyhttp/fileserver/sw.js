self.addEventListener('install', event => {
    console.log('V1 installing…');
});

self.addEventListener('activate', event => {
    console.log("SW activated!")
})

const putInCache = async (request, response) => {
    const cache = await caches.open("v2");
    await cache.put(request, response);
};

const cacheFirst = async ({request, fallbackUrl}) => {
    // First try to get the resource from the cache
    var reloadCache = false;
    if (request.url.includes(request.referrer)) {
        const responseFromCache = await caches.match(request);
        if (responseFromCache) {
            var etag = responseFromCache.headers.get('Etag');

            var key = request.url.replace(request.referrer, "");
            var etagsMap = self.etags || {}
            var cachedEtag = etagsMap[key];
            if (cachedEtag) {
                if (etag == cachedEtag) {
                    console.log("Respond from cache", key);
                    return responseFromCache;
                } else {
                    reloadCache = true;
                    console.log("Reload cache", key);
                }
            }
        }
    }

    // Next try to get the resource from the network
    try {
        var options = {};
        if (reloadCache) {
            options = {
                cache: "reload"
            };
        }
        const responseFromNetwork = await fetch(request, options);
        const etagsJson = responseFromNetwork.headers.get('X-Etag-Config');
        // Modified version of caddy sends X-Etag-Config for every html.
        // Because this indicated first page request, we don't cache it.
        if (etagsJson != null) {
            self.etags = JSON.parse(etagsJson);
            return responseFromNetwork
        }

        // response may be used only once
        // we need to save clone to put one copy in cache
        // and serve second one
        putInCache(request, responseFromNetwork.clone());
        return responseFromNetwork;
    } catch (error) {
        const fallbackResponse = await caches.match(fallbackUrl);
        if (fallbackResponse) {
            return fallbackResponse;
        }
        // when even the fallback response is not available,
        // there is nothing we can do, but we must always
        // return a Response object
        return new Response("Network error happened", {
            status: 408,
            headers: {"Content-Type": "text/plain"},
        });
    }
};

self.addEventListener("fetch", async (event) => {
    if (!event.request.url.includes(event.request.referrer)) {
        // silent requests to other hosts
        // TODO: remove this for production
        event.respondWith(new Response());
    } else {
        event.respondWith(
            cacheFirst({
                request: event.request,
                fallbackUrl: "/404.png",
            })
        );
    }
});
