self.addEventListener('install', event => {
    console.log('V1 installing…');
    self.etags = new Map();
});

self.addEventListener('activate', event => {
    console.log("SW activated!")
})

const putInCache = async (request, response) => {
    const cache = await caches.open("v2");
    await cache.put(request, response);
};

const openDatabase = () => {
    return new Promise((resolve, reject) => {
        const request = indexedDB.open('myDatabase', 1);

        request.onupgradeneeded = event => {
            const db = event.target.result;
            db.createObjectStore('myStore', { keyPath: 'key' });
        };

        request.onsuccess = event => {
            resolve(event.target.result);
        };

        request.onerror = event => {
            reject('Database error: ', event.target.error);
        };
    });
};

const storeData = async (key, value) => {
    const db = await openDatabase();
    const transaction = db.transaction('myStore', 'readwrite');
    const objectStore = transaction.objectStore('myStore');

    objectStore.put({ key, value });

    return transaction.complete;
};

const loadData = async () => {
    const db = await openDatabase();
    const transaction = db.transaction('myStore', 'readonly');
    const objectStore = transaction.objectStore('myStore');

    return new Promise((resolve, reject) => {
        const request = objectStore.getAll();

        request.onsuccess = event => {
            const map = new Map();
            event.target.result.forEach(record => {
                map.set(record.key, record.value);
            });
            resolve(map);
        };

        request.onerror = event => {
            reject('Data retrieval error: ', event.target.error);
        };
    });
};


const getAllEtags = async () => {
    if (!self.etags || self.etags.size === 0) {
        self.etags = await loadData();
    }

    return self.etags
}

const storeEtag = async (url, etag) => {
    self.etags.set(url, etag);
    await storeData(url, etag);
}

self.addEventListener('fetch', event => {
    const requestUrl = new URL(event.request.url);
    const ownOrigin = self.location.origin;
    if (requestUrl.origin !== ownOrigin) {
        // Return a dummy response for requests to other hosts (temporary for load time testing)
        event.respondWith(
            new Response('This is a dummy response for requests to other hosts.', {
                status: 200,
                headers: { 'Content-Type': 'text/plain' }
            })
        );
    } else if (event.request.referrer === "") { // FIXME: better way to check index.html
        event.respondWith(
            handleIndexHtmlRequest(event.request)
        );
    } else {
        event.respondWith(
            handleResourceRequest(event.request)
        );
    }
});


async function handleIndexHtmlRequest(request) {
    etags = await getAllEtags();
    let requestWithHeaders = request;
    if (etags && request.referrer === "") {
        const jsonedEtags = JSON.stringify(Object.fromEntries(etags));
        requestWithHeaders = new Request(request, {
            headers: {
                ...request.headers,
                "xetags": jsonedEtags,
            }
        });
        console.log("etags", jsonedEtags);
    }

    const responseFromNetwork = await fetch(requestWithHeaders);
    self.changedFlagStorage = new Map(); // Initialize storage for current session
    processChangedFiles(responseFromNetwork, self.changedFlagStorage);

    return responseFromNetwork;
}

async function handleResourceRequest(request) {
    const url = new URL(request.url);
    const uriKey = url.pathname;
    const changed = self.changedFlagStorage.get(uriKey);
    console.log(uriKey, "has changed/not", changed);
    if (!changed) {
        const responseFromCache = await caches.match(request);
        if (responseFromCache) {
            console.log(uriKey, "response is cached");
            return responseFromCache;
        }
        console.log(uriKey, "response is not cached");
    }

    // reload set to false to use browser cache for second request. first request can't
    // manipulate service worker cache cause of
    // not registering sw yet.
    const response = await fetch(request, {reload: changed === true});
    putInCache(request, response.clone());
    const etag = response.headers.get('Etag');
    console.log("etag get for ", uriKey, etag);
    if (etag) {
        await storeEtag(uriKey, etag);
    }

    return response;
}

function processChangedFiles(response, changedFlagStorage) {
    const changedFiles = response.headers.get('X-Changed-Files');
    console.log("changed files", changedFiles);
    if (changedFiles) {
        const resources = changedFiles.split(', ').map(url => url.trim());
        resources.forEach(uri => {changedFlagStorage.set(uri, true)});
        console.log("changed storage: ", changedFlagStorage);
        return
    }
    console.log("no changed Files here!");
}
