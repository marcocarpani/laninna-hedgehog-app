// La Ninna Hedgehog App - Offline Functionality
class OfflineManager {
  constructor() {
    this.isOnline = navigator.onLine;
    this.pendingRequests = [];
    this.offlineIndicator = null;
    this.db = null;
    
    this.init();
  }
  
  async init() {
    // Register service worker
    this.registerServiceWorker();
    
    // Set up online/offline event listeners
    this.setupNetworkListeners();
    
    // Create offline UI elements
    this.createOfflineUI();
    
    // Initialize IndexedDB
    await this.initIndexedDB();
    
    // Check for pending requests on startup
    await this.checkPendingRequests();
  }
  
  registerServiceWorker() {
    if ('serviceWorker' in navigator) {
      window.addEventListener('load', () => {
        navigator.serviceWorker.register('/static/js/service-worker.js')
          .then(registration => {
            console.log('ServiceWorker registration successful with scope: ', registration.scope);
            
            // Set up message listener for service worker
            navigator.serviceWorker.addEventListener('message', event => {
              this.handleServiceWorkerMessage(event);
            });
          })
          .catch(error => {
            console.error('ServiceWorker registration failed: ', error);
          });
      });
    } else {
      console.warn('Service workers are not supported in this browser');
    }
  }
  
  setupNetworkListeners() {
    // Listen for online/offline events
    window.addEventListener('online', () => this.handleOnlineStatus(true));
    window.addEventListener('offline', () => this.handleOnlineStatus(false));
    
    // Initial status check
    this.handleOnlineStatus(navigator.onLine);
  }
  
  handleOnlineStatus(isOnline) {
    this.isOnline = isOnline;
    
    if (isOnline) {
      console.log('🌐 App is online');
      this.hideOfflineIndicator();
      this.syncPendingRequests();
    } else {
      console.log('📴 App is offline');
      this.showOfflineIndicator();
    }
    
    // Dispatch custom event for other components
    window.dispatchEvent(new CustomEvent('connectivityChange', { 
      detail: { isOnline } 
    }));
  }
  
  createOfflineUI() {
    // Create offline indicator
    this.offlineIndicator = document.createElement('div');
    this.offlineIndicator.className = 'offline-indicator hidden';
    this.offlineIndicator.innerHTML = `
      <div class="offline-content">
        <i class="fas fa-wifi-slash"></i>
        <span>Modalità offline</span>
      </div>
    `;
    document.body.appendChild(this.offlineIndicator);
    
    // Add CSS for offline indicator
    const style = document.createElement('style');
    style.textContent = `
      .offline-indicator {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        background-color: #f44336;
        color: white;
        text-align: center;
        padding: 8px;
        z-index: 9999;
        transition: transform 0.3s ease-in-out;
      }
      
      .offline-indicator.hidden {
        transform: translateY(-100%);
      }
      
      .offline-content {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
      }
      
      .offline-badge {
        display: inline-block;
        background-color: #f44336;
        color: white;
        border-radius: 4px;
        padding: 2px 6px;
        font-size: 0.75rem;
        margin-left: 8px;
      }
      
      .cached-item {
        border-left: 4px solid #ff9800 !important;
      }
      
      .pending-item {
        border-left: 4px solid #f44336 !important;
      }
    `;
    document.head.appendChild(style);
  }
  
  showOfflineIndicator() {
    if (this.offlineIndicator) {
      this.offlineIndicator.classList.remove('hidden');
    }
  }
  
  hideOfflineIndicator() {
    if (this.offlineIndicator) {
      this.offlineIndicator.classList.add('hidden');
    }
  }
  
  async initIndexedDB() {
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
        this.db = event.target.result;
        resolve();
      };
      
      request.onerror = event => {
        console.error('IndexedDB error:', event.target.error);
        reject(event.target.error);
      };
    });
  }
  
  async checkPendingRequests() {
    if (!this.db) return;
    
    try {
      const tx = this.db.transaction('pending-requests', 'readonly');
      const store = tx.objectStore('pending-requests');
      const requests = await store.getAll();
      
      if (requests.length > 0) {
        console.log(`Found ${requests.length} pending requests`);
        
        // Show notification if there are pending requests
        if (typeof showToast === 'function') {
          showToast(`Hai ${requests.length} richieste in attesa di sincronizzazione`, 'info');
        }
        
        // Try to sync if online
        if (this.isOnline) {
          this.syncPendingRequests();
        }
      }
    } catch (error) {
      console.error('Error checking pending requests:', error);
    }
  }
  
  async syncPendingRequests() {
    if (!this.isOnline || !navigator.serviceWorker.controller) return;
    
    // Notify service worker to sync pending requests
    navigator.serviceWorker.controller.postMessage({
      type: 'SYNC_PENDING_REQUESTS'
    });
    
    // Register background sync if supported
    if ('sync' in navigator.serviceWorker.controller) {
      try {
        await navigator.serviceWorker.ready;
        await navigator.serviceWorker.sync.register('sync-pending-requests');
      } catch (error) {
        console.error('Background sync registration failed:', error);
      }
    }
  }
  
  handleServiceWorkerMessage(event) {
    const { data } = event;
    
    if (data.type === 'GET_AUTH_TOKEN') {
      // Send auth token to service worker
      event.ports[0].postMessage({
        token: localStorage.getItem('token')
      });
    } else if (data.type === 'REQUEST_SYNCED') {
      // Handle synced request
      console.log('Request synced:', data);
      
      // Show notification
      if (typeof showToast === 'function') {
        showToast('Dati sincronizzati con successo', 'success');
      }
      
      // Refresh data if needed
      if (data.url.includes('/api/hedgehogs')) {
        if (typeof loadDashboardStats === 'function') {
          loadDashboardStats();
        }
      }
    }
  }
  
  // API wrapper for fetch that handles offline mode
  async fetchWithOfflineSupport(url, options = {}) {
    // Add auth token if not present
    if (!options.headers) {
      options.headers = {};
    }
    
    if (!options.headers.Authorization && localStorage.getItem('token')) {
      options.headers.Authorization = `Bearer ${localStorage.getItem('token')}`;
    }
    
    try {
      // Try regular fetch first
      const response = await fetch(url, options);
      return response;
    } catch (error) {
      console.log('Fetch error:', error);
      
      // Check if we're actually offline according to the browser
      if (!navigator.onLine) {
        this.isOnline = false;
        this.showOfflineIndicator();
        console.log('Browser reports offline status, handling accordingly');
        
        // Only store POST, PUT, DELETE requests
        if (['POST', 'PUT', 'DELETE'].includes(options.method)) {
          await this.storePendingRequest(url, options);
          
          // Show notification
          if (typeof showToast === 'function') {
            showToast('Richiesta salvata per la sincronizzazione', 'info');
          }
          
          // Return a fake successful response
          return new Response(JSON.stringify({
            success: true,
            offline: true,
            message: 'Request queued for sync when online'
          }), {
            status: 202,
            headers: { 'Content-Type': 'application/json' }
          });
        }
        
        // For GET requests, try to get from IndexedDB
        if (options.method === 'GET' || !options.method) {
          const data = await this.getApiDataFromIndexedDB(url);
          if (data) {
            return new Response(JSON.stringify(data), {
              status: 200,
              headers: { 'Content-Type': 'application/json' }
            });
          }
        }
        
        // If we get here, we're offline and couldn't find cached data
        throw new Error('Sei offline e questi dati non sono disponibili nella cache');
      } else {
        // We're online but still got an error - likely a server issue
        console.log('Browser reports online status, but request failed');
        
        // Try to get from cache as a fallback
        if (options.method === 'GET' || !options.method) {
          const data = await this.getApiDataFromIndexedDB(url);
          if (data) {
            console.log('Found cached data, using it as fallback');
            if (typeof showToast === 'function') {
              showToast('Usando dati in cache mentre il server è irraggiungibile', 'warning');
            }
            return new Response(JSON.stringify(data), {
              status: 200,
              headers: { 'Content-Type': 'application/json' }
            });
          }
        }
        
        // Create a more specific error message based on the error
        let errorMessage = 'Errore del server. Il server è irraggiungibile.';
        
        // Throw a more descriptive error
        throw new Error(errorMessage);
      }
    }
  }
  
  async storePendingRequest(url, options) {
    if (!this.db) return;
    
    try {
      const tx = this.db.transaction('pending-requests', 'readwrite');
      const store = tx.objectStore('pending-requests');
      
      // Parse request body if it's a string
      let body = options.body;
      if (typeof body === 'string') {
        try {
          body = JSON.parse(body);
        } catch (e) {
          // Keep as string if not valid JSON
        }
      }
      
      // Store the request
      await store.add({
        url,
        method: options.method || 'GET',
        body,
        headers: options.headers,
        token: localStorage.getItem('token'),
        timestamp: Date.now()
      });
      
      await tx.complete;
      console.log('Request stored for later sync');
    } catch (error) {
      console.error('Error storing pending request:', error);
    }
  }
  
  async getApiDataFromIndexedDB(url) {
    if (!this.db) return null;
    
    try {
      const tx = this.db.transaction('api-data', 'readonly');
      const store = tx.objectStore('api-data');
      
      // Extract the path from the URL
      const path = new URL(url, window.location.origin).pathname;
      
      // Get the data with the path as key
      const item = await store.get(path);
      
      await tx.complete;
      
      return item ? item.data : null;
    } catch (error) {
      console.error('Error getting data from IndexedDB:', error);
      return null;
    }
  }
  
  // Add offline badges to elements that display cached data
  addOfflineBadges() {
    if (!this.isOnline) {
      document.querySelectorAll('.card:not(.has-offline-badge)').forEach(card => {
        card.classList.add('has-offline-badge', 'cached-item');
        
        const badge = document.createElement('span');
        badge.className = 'offline-badge';
        badge.textContent = 'Offline';
        
        // Add badge to the card header or to the card itself
        const header = card.querySelector('.card-header') || card;
        header.appendChild(badge);
      });
    }
  }
}

// Initialize offline manager
let offlineManager;

document.addEventListener('DOMContentLoaded', () => {
  offlineManager = new OfflineManager();
  
  // Make fetchWithOfflineSupport available globally
  window.fetchWithOfflineSupport = (url, options) => {
    return offlineManager.fetchWithOfflineSupport(url, options);
  };
});

// Override fetch for API calls to use offline support
const originalFetch = window.fetch;
window.fetch = function(url, options = {}) {
  // Only intercept API calls
  if (url.includes('/api/') && offlineManager) {
    return offlineManager.fetchWithOfflineSupport(url, options);
  }
  
  // Use original fetch for other requests
  return originalFetch(url, options);
};