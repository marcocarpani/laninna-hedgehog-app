// La Ninna Hedgehog App - Service Worker
const CACHE_NAME = 'laninna-cache-v1';
const STATIC_CACHE_NAME = 'laninna-static-v1';
const DATA_CACHE_NAME = 'laninna-data-v1';

// Assets to cache on install
const STATIC_ASSETS = [
  '/',
  '/static/css/mobile.css',
  '/static/css/desktop.css',
  '/static/css/mobile-fixes.css',
  '/static/js/mobile.js',
  '/static/js/export.js',
  '/static/js/offline.js',
  '/login',
  '/hedgehogs',
  '/rooms',
  '/notifications',
  '/room-builder',
  '/tutorial',
  'https://cdn.tailwindcss.com',
  'https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css'
];

// API routes to cache
const API_ROUTES = [
  '/api/hedgehogs',
  '/api/rooms',
  '/api/areas',
  '/api/therapies',
  '/api/weight-records',
  '/api/notifications'
];

// Install event - cache static assets
self.addEventListener('install', event => {
  console.log('[Service Worker] Installing Service Worker...');
  
  // Skip waiting to ensure the new service worker activates immediately
  self.skipWaiting();
  
  event.waitUntil(
    caches.open(STATIC_CACHE_NAME)
      .then(cache => {
        console.log('[Service Worker] Caching static assets');
        return cache.addAll(STATIC_ASSETS);
      })
      .catch(error => {
        console.error('[Service Worker] Cache installation failed:', error);
      })
  );
});

// Activate event - clean up old caches
self.addEventListener('activate', event => {
  console.log('[Service Worker] Activating Service Worker...');
  
  // Claim clients to ensure the service worker controls all pages
  event.waitUntil(self.clients.claim());
  
  // Clean up old caches
  event.waitUntil(
    caches.keys().then(cacheNames => {
      return Promise.all(
        cacheNames.map(cacheName => {
          if (
            cacheName !== STATIC_CACHE_NAME &&
            cacheName !== DATA_CACHE_NAME
          ) {
            console.log('[Service Worker] Removing old cache:', cacheName);
            return caches.delete(cacheName);
          }
        })
      );
    })
  );
});

// Fetch event - handle requests
self.addEventListener('fetch', event => {
  const url = new URL(event.request.url);
  
  // Handle API requests
  if (API_ROUTES.some(route => url.pathname.startsWith(route))) {
    event.respondWith(handleApiRequest(event.request));
    return;
  }
  
  // Handle static assets and pages
  event.respondWith(
    caches.match(event.request)
      .then(response => {
        // Return cached response if found
        if (response) {
          return response;
        }
        
        // Otherwise fetch from network
        return fetch(event.request)
          .then(networkResponse => {
            // Don't cache non-GET requests
            if (event.request.method !== 'GET') {
              return networkResponse;
            }
            
            // Cache successful responses
            if (networkResponse.status === 200) {
              const clonedResponse = networkResponse.clone();
              caches.open(STATIC_CACHE_NAME)
                .then(cache => {
                  cache.put(event.request, clonedResponse);
                });
            }
            
            return networkResponse;
          })
          .catch(error => {
            console.error('[Service Worker] Fetch failed:', error);
            
            // For navigation requests, return the offline page
            if (event.request.mode === 'navigate') {
              return caches.match('/');
            }
            
            // Return a fallback for other resources
            return new Response('Network error occurred', {
              status: 503,
              statusText: 'Service Unavailable'
            });
          });
      })
  );
});

// Handle API requests with network-first strategy
async function handleApiRequest(request) {
  const token = await getAuthToken();
  
  // Create a new request with the auth token
  const authenticatedRequest = createAuthenticatedRequest(request, token);
  
  try {
    // Try network first
    const networkResponse = await fetch(authenticatedRequest);
    
    // Clone the response before caching it
    const responseToCache = networkResponse.clone();
    
    // Cache the response for offline use
    if (networkResponse.status === 200) {
      const cache = await caches.open(DATA_CACHE_NAME);
      await cache.put(request, responseToCache);
      
      // Also store in IndexedDB for offline data manipulation
      const data = await responseToCache.clone().json();
      await storeApiDataInIndexedDB(request.url, data);
    }
    
    return networkResponse;
  } catch (error) {
    console.log('[Service Worker] Network request failed, using cache', error);
    
    // Try to get from cache
    const cachedResponse = await caches.match(request);
    if (cachedResponse) {
      return cachedResponse;
    }
    
    // If not in cache, try to get from IndexedDB
    const data = await getApiDataFromIndexedDB(request.url);
    if (data) {
      return new Response(JSON.stringify(data), {
        headers: { 'Content-Type': 'application/json' }
      });
    }
    
    // Return empty data with offline indicator
    return new Response(JSON.stringify({ 
      offline: true, 
      message: 'You are offline and this data is not cached' 
    }), {
      headers: { 'Content-Type': 'application/json' }
    });
  }
}

// Create an authenticated request with the token
function createAuthenticatedRequest(request, token) {
  // Clone the request to modify headers
  const headers = new Headers(request.headers);
  
  // Add auth token if available
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }
  
  // Create a new request with the same properties but new headers
  return new Request(request.url, {
    method: request.method,
    headers: headers,
    body: request.body,
    mode: request.mode,
    credentials: request.credentials,
    cache: request.cache,
    redirect: request.redirect,
    referrer: request.referrer,
    integrity: request.integrity
  });
}

// Get auth token from IndexedDB
async function getAuthToken() {
  return new Promise((resolve) => {
    // Try to get the token from clients
    self.clients.matchAll().then(clients => {
      if (clients.length > 0) {
        // Send message to client to get token
        const client = clients[0];
        const messageChannel = new MessageChannel();
        
        messageChannel.port1.onmessage = event => {
          resolve(event.data.token);
        };
        
        client.postMessage({ type: 'GET_AUTH_TOKEN' }, [messageChannel.port2]);
        
        // Set a timeout in case client doesn't respond
        setTimeout(() => resolve(null), 500);
      } else {
        resolve(null);
      }
    });
  });
}

// Store API data in IndexedDB
async function storeApiDataInIndexedDB(url, data) {
  const db = await openDatabase();
  const tx = db.transaction('api-data', 'readwrite');
  const store = tx.objectStore('api-data');
  
  // Extract the path from the URL
  const path = new URL(url).pathname;
  
  // Store the data with the path as key
  await store.put({
    path: path,
    data: data,
    timestamp: Date.now()
  });
  
  await tx.complete;
}

// Get API data from IndexedDB
async function getApiDataFromIndexedDB(url) {
  try {
    const db = await openDatabase();
    const tx = db.transaction('api-data', 'readonly');
    const store = tx.objectStore('api-data');
    
    // Extract the path from the URL
    const path = new URL(url).pathname;
    
    // Get the data with the path as key
    const item = await store.get(path);
    
    await tx.complete;
    
    return item ? item.data : null;
  } catch (error) {
    console.error('[Service Worker] Error getting data from IndexedDB:', error);
    return null;
  }
}

// Open IndexedDB database
function openDatabase() {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('laninna-offline-db', 1);
    
    request.onupgradeneeded = event => {
      const db = event.target.result;
      
      // Create object stores if they don't exist
      if (!db.objectStoreNames.contains('api-data')) {
        const store = db.createObjectStore('api-data', { keyPath: 'path' });
        store.createIndex('timestamp', 'timestamp', { unique: false });
      }
      
      if (!db.objectStoreNames.contains('pending-requests')) {
        db.createObjectStore('pending-requests', { keyPath: 'id', autoIncrement: true });
      }
    };
    
    request.onsuccess = event => {
      resolve(event.target.result);
    };
    
    request.onerror = event => {
      console.error('[Service Worker] IndexedDB error:', event.target.error);
      reject(event.target.error);
    };
  });
}

// Listen for messages from the client
self.addEventListener('message', event => {
  if (event.data.type === 'SYNC_PENDING_REQUESTS') {
    syncPendingRequests();
  }
});

// Sync pending requests when online
async function syncPendingRequests() {
  try {
    const db = await openDatabase();
    const tx = db.transaction('pending-requests', 'readonly');
    const store = tx.objectStore('pending-requests');
    
    // Get all pending requests
    const requests = await store.getAll();
    
    await tx.complete;
    
    // Process each request
    for (const request of requests) {
      try {
        const response = await fetch(request.url, {
          method: request.method,
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${request.token}`
          },
          body: request.body ? JSON.stringify(request.body) : undefined
        });
        
        if (response.ok) {
          // Remove the request from pending requests
          const deleteTx = db.transaction('pending-requests', 'readwrite');
          const deleteStore = deleteTx.objectStore('pending-requests');
          await deleteStore.delete(request.id);
          await deleteTx.complete;
          
          console.log('[Service Worker] Synced pending request:', request.url);
          
          // Notify clients about successful sync
          self.clients.matchAll().then(clients => {
            clients.forEach(client => {
              client.postMessage({
                type: 'REQUEST_SYNCED',
                url: request.url,
                method: request.method,
                success: true
              });
            });
          });
        }
      } catch (error) {
        console.error('[Service Worker] Error syncing request:', error);
      }
    }
  } catch (error) {
    console.error('[Service Worker] Error syncing pending requests:', error);
  }
}

// Background sync for pending requests
self.addEventListener('sync', event => {
  if (event.tag === 'sync-pending-requests') {
    event.waitUntil(syncPendingRequests());
  }
});