// 🦔 La Ninna - Mobile Navigation & UX

// TouchHandler class for improving touch interactions
class TouchHandler {
    constructor() {
        this.touchStartX = 0;
        this.touchStartY = 0;
        this.lastTap = 0;
        this.touchTimeout = null;
        this.longPressThreshold = 500; // ms
        this.doubleTapThreshold = 300; // ms
        this.touchMoveThreshold = 10; // pixels
        this.init();
    }

    init() {
        // Add touch event polyfills for canvas elements
        this.addCanvasTouchSupport();
        
        // Add general touch improvements
        this.addGeneralTouchImprovements();
    }

    addCanvasTouchSupport() {
        // Find all canvas elements and add touch event handlers
        document.addEventListener('DOMContentLoaded', () => {
            const canvases = document.querySelectorAll('canvas');
            canvases.forEach(canvas => {
                this.addTouchEventsToCanvas(canvas);
            });
        });
    }

    addTouchEventsToCanvas(canvas) {
        if (!canvas) return;
        
        // Store original mouse event handlers if they exist
        const originalMouseDown = canvas.onmousedown;
        const originalMouseMove = canvas.onmousemove;
        const originalMouseUp = canvas.onmouseup;
        
        // Add touch event handlers that translate to mouse events
        canvas.addEventListener('touchstart', (e) => {
            e.preventDefault(); // Prevent scrolling when touching canvas
            
            const touch = e.touches[0];
            this.touchStartX = touch.clientX;
            this.touchStartY = touch.clientY;
            
            // Create a simulated mouse event
            const mouseEvent = new MouseEvent('mousedown', {
                clientX: touch.clientX,
                clientY: touch.clientY,
                bubbles: true,
                cancelable: true,
                view: window
            });
            
            // Call the original mousedown handler if it exists
            if (typeof originalMouseDown === 'function') {
                canvas.dispatchEvent(mouseEvent);
            }
            
            // Set up long press detection
            this.touchTimeout = setTimeout(() => {
                // Simulate right-click or context menu for long press
                const contextEvent = new MouseEvent('contextmenu', {
                    clientX: touch.clientX,
                    clientY: touch.clientY,
                    bubbles: true,
                    cancelable: true,
                    view: window
                });
                canvas.dispatchEvent(contextEvent);
            }, this.longPressThreshold);
        });
        
        canvas.addEventListener('touchmove', (e) => {
            e.preventDefault();
            
            // Clear long press timeout if user is moving
            if (this.touchTimeout) {
                clearTimeout(this.touchTimeout);
                this.touchTimeout = null;
            }
            
            const touch = e.touches[0];
            
            // Create a simulated mouse event
            const mouseEvent = new MouseEvent('mousemove', {
                clientX: touch.clientX,
                clientY: touch.clientY,
                bubbles: true,
                cancelable: true,
                view: window
            });
            
            // Call the original mousemove handler if it exists
            if (typeof originalMouseMove === 'function') {
                canvas.dispatchEvent(mouseEvent);
            }
        });
        
        canvas.addEventListener('touchend', (e) => {
            e.preventDefault();
            
            // Clear long press timeout
            if (this.touchTimeout) {
                clearTimeout(this.touchTimeout);
                this.touchTimeout = null;
            }
            
            // Get the last touch position
            let x, y;
            if (e.changedTouches && e.changedTouches.length > 0) {
                const touch = e.changedTouches[0];
                x = touch.clientX;
                y = touch.clientY;
            } else {
                x = this.touchStartX;
                y = this.touchStartY;
            }
            
            // Create a simulated mouse event
            const mouseEvent = new MouseEvent('mouseup', {
                clientX: x,
                clientY: y,
                bubbles: true,
                cancelable: true,
                view: window
            });
            
            // Call the original mouseup handler if it exists
            if (typeof originalMouseUp === 'function') {
                canvas.dispatchEvent(mouseEvent);
            }
            
            // Handle double tap
            const now = Date.now();
            const timeDiff = now - this.lastTap;
            
            if (timeDiff < this.doubleTapThreshold) {
                // Double tap detected
                const dblClickEvent = new MouseEvent('dblclick', {
                    clientX: x,
                    clientY: y,
                    bubbles: true,
                    cancelable: true,
                    view: window
                });
                canvas.dispatchEvent(dblClickEvent);
            }
            
            this.lastTap = now;
        });
        
        // Prevent default touch actions on canvas
        canvas.addEventListener('touchcancel', (e) => {
            e.preventDefault();
            
            if (this.touchTimeout) {
                clearTimeout(this.touchTimeout);
                this.touchTimeout = null;
            }
        });
    }

    addGeneralTouchImprovements() {
        // Improve touch feedback for buttons and interactive elements
        document.addEventListener('DOMContentLoaded', () => {
            // Add active state to buttons when touched
            const buttons = document.querySelectorAll('button, .btn-primary, .btn-secondary, .btn-danger');
            buttons.forEach(button => {
                button.addEventListener('touchstart', () => {
                    button.classList.add('touch-active');
                });
                
                button.addEventListener('touchend', () => {
                    button.classList.remove('touch-active');
                });
                
                button.addEventListener('touchcancel', () => {
                    button.classList.remove('touch-active');
                });
            });
            
            // Improve range inputs for touch
            const rangeInputs = document.querySelectorAll('input[type="range"]');
            rangeInputs.forEach(input => {
                // Make range inputs larger on touch devices
                input.classList.add('touch-friendly-range');
            });
        });
    }
}



class MobileNavigation {
    constructor() {
        this.sidebar = null;
        this.overlay = null;
        this.bottomNav = null;
        this.isOpen = false;
        this.init();
    }

    init() {
        this.createSidebar();
        this.createBottomNav();
        this.setupEventListeners();
        this.setupEdgeDetection();
        this.handleResize();
    }

    createSidebar() {
        // Create overlay for mobile
        this.overlay = document.createElement('div');
        this.overlay.className = 'fixed inset-0 bg-black bg-opacity-50 z-40 hidden';
        this.overlay.onclick = () => this.closeSidebar();
        document.body.appendChild(this.overlay);

        // Create responsive sidebar
        this.sidebar = document.createElement('div');
        this.sidebar.className = 'fixed top-0 left-0 h-screen bg-white border-r border-gray-200 z-50 w-16 lg:w-32 lg:shadow-lg lg:border-r-0 transform -translate-x-full lg:translate-x-0 transition-transform duration-300 ease-in-out';
        this.sidebar.innerHTML = this.getSidebarContent();
        document.body.appendChild(this.sidebar);
        
        // Add responsive margin to main content
        const mainContent = document.querySelector('.min-h-screen');
        if (mainContent) {
            mainContent.classList.add('lg:ml-32');
        }
    }
    
    createBottomNav() {
        const currentPath = window.location.pathname;
        
        // Create bottom navigation bar for mobile
        this.bottomNav = document.createElement('div');
        this.bottomNav.className = 'fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 z-40 lg:hidden';
        this.bottomNav.innerHTML = `
            <div class="flex justify-around items-center h-16 px-2 safe-bottom">
                <a href="/" class="bottom-nav-item ${currentPath === '/' ? 'active' : ''}">
                    <i class="fas fa-home text-xl"></i>
                    <span class="text-xs mt-1">Home</span>
                </a>
                <a href="/hedgehogs" class="bottom-nav-item ${currentPath === '/hedgehogs' ? 'active' : ''}">
                    <i class="fas fa-paw text-xl"></i>
                    <span class="text-xs mt-1">Ricci</span>
                </a>
                <a href="/rooms" class="bottom-nav-item ${currentPath === '/rooms' ? 'active' : ''}">
                    <i class="fas fa-door-open text-xl"></i>
                    <span class="text-xs mt-1">Stanze</span>
                </a>
                <a href="/notifications" class="bottom-nav-item ${currentPath === '/notifications' ? 'active' : ''}">
                    <i class="fas fa-bell text-xl"></i>
                    <span class="text-xs mt-1">Notifiche</span>
                </a>
                <button class="bottom-nav-item" onclick="mobileNav.openSidebar()">
                    <i class="fas fa-ellipsis-h text-xl"></i>
                    <span class="text-xs mt-1">Menu</span>
                </button>
            </div>
        `;
        document.body.appendChild(this.bottomNav);
        
        // Add padding to the bottom of the page to account for the bottom nav
        document.body.classList.add('pb-16', 'lg:pb-0');
    }

    getSidebarContent() {
        const currentPath = window.location.pathname;
        
        return `
            <div class="flex flex-col h-full">
                <!-- Header -->
                <div class="flex items-center justify-center p-4 border-b border-gray-200 lg:py-4">
                    <div class="text-2xl">🦔</div>
                    <div class="hidden lg:block ml-3">
                        <h1 class="text-xl font-bold text-hedgehog-brown">La Ninna</h1>
                        <p class="text-xs text-gray-600">Centro Recupero</p>
                    </div>
                </div>

                <!-- Navigation -->
                <nav class="flex-1 p-2 lg:p-4 space-y-2">
                    <a href="/" onclick="this.href='/'; mobileNav.closeSidebar();" class="sidebar-icon-mobile lg:nav-item-vertical ${currentPath === '/' ? 'active' : ''}" title="Dashboard">
                        <i class="fas fa-chart-pie text-xl"></i>
                    </a>
                    <div class="border-t border-gray-200 mx-2 lg:mx-4"></div>
                    <a href="/hedgehogs" onclick="this.href='/hedgehogs'; mobileNav.closeSidebar();" class="sidebar-icon-mobile lg:nav-item-vertical ${currentPath === '/hedgehogs' ? 'active' : ''}" title="Gestione Ricci">
                        <i class="fas fa-paw text-xl"></i>
                    </a>
                    <div class="border-t border-gray-200 mx-2 lg:mx-4"></div>
                    <a href="/rooms" onclick="this.href='/rooms'; mobileNav.closeSidebar();" class="sidebar-icon-mobile lg:nav-item-vertical ${currentPath === '/rooms' ? 'active' : ''}" title="Gestione Stanze">
                        <i class="fas fa-door-open text-xl"></i>
                    </a>
                    <div class="border-t border-gray-200 mx-2 lg:mx-4"></div>
                    <a href="/room-builder" onclick="this.href='/room-builder'; mobileNav.closeSidebar();" class="sidebar-icon-mobile lg:nav-item-vertical ${currentPath === '/room-builder' ? 'active' : ''}" title="Room Builder">
                        <i class="fas fa-drafting-compass text-xl"></i>
                    </a>
                    <div class="border-t border-gray-200 mx-2 lg:mx-4"></div>
                    <a href="/notifications" onclick="this.href='/notifications'; mobileNav.closeSidebar();" class="sidebar-icon-mobile lg:nav-item-vertical ${currentPath === '/notifications' ? 'active' : ''}" title="Notifiche">
                        <i class="fas fa-bell text-xl"></i>
                    </a>
                    <div class="border-t border-gray-200 mx-2 lg:mx-4"></div>
                    <a href="/tutorial" onclick="window.location.href='/docs'; mobileNav.closeSidebar();" class="sidebar-icon-mobile lg:nav-item-vertical ${currentPath === '/docs' ? 'active' : ''}" title="Tutorial">
                        <i class="fas fa-book text-xl"></i>
                    </a>
                </nav>

                <!-- Footer -->
                <div class="p-2 lg:p-4 mt-auto">
                    <div class="border-t border-gray-200 mx-2 lg:mx-4 mb-2 lg:mb-4"></div>
                    <button onclick="logout(); mobileNav.closeSidebar();" class="sidebar-icon-mobile lg:nav-item-vertical text-red-600 hover:bg-red-50 w-full" title="">
                        <i class="fas fa-sign-out-alt text-xl"></i>
                        <span class="block text-sm ml-3 lg:text-xs lg:mt-1">Logout</span>
                    </button>
                </div>
            </div>
        `;
    }

    setupEventListeners() {
        // Mobile menu button
        const menuBtn = document.getElementById('mobile-menu-btn');
        if (menuBtn) {
            menuBtn.onclick = () => this.openSidebar();
        }

        // Handle window resize
        window.addEventListener('resize', () => this.handleResize());
    }

    openSidebar() {
        if (window.innerWidth >= 1024) return; // Don't open on desktop
        this.isOpen = true;
        this.sidebar.classList.remove('-translate-x-full');
        this.overlay.classList.remove('hidden');
        document.body.classList.add('overflow-hidden');
    }

    closeSidebar() {
        this.isOpen = false;
        this.sidebar.classList.add('-translate-x-full');
        this.overlay.classList.add('hidden');
        document.body.classList.remove('overflow-hidden');
    }

    setupEdgeDetection() {
        // Create invisible left edge trigger
        const edgeTrigger = document.createElement('div');
        edgeTrigger.className = 'fixed top-0 left-0 w-4 h-full z-20 lg:hidden';
        edgeTrigger.style.backgroundColor = 'transparent';
        document.body.appendChild(edgeTrigger);

        // Mouse events for mobile hover
        edgeTrigger.addEventListener('mouseenter', () => {
            if (window.innerWidth < 1024 && !this.isOpen) {
                this.openSidebar();
            }
        });

        // Touch events for mobile
        let touchStartX = 0;
        document.addEventListener('touchstart', (e) => {
            touchStartX = e.touches[0].clientX;
        });

        document.addEventListener('touchmove', (e) => {
            if (window.innerWidth < 1024 && !this.isOpen && touchStartX < 20) {
                const touchX = e.touches[0].clientX;
                if (touchX > touchStartX + 30) {
                    this.openSidebar();
                }
            }
        });
    }

    handleResize() {
        if (window.innerWidth >= 1024) { // lg breakpoint
            this.closeSidebar();
        }
    }
}

// Toast notification system
class ToastManager {
    constructor() {
        this.container = this.createContainer();
    }

    createContainer() {
        const container = document.createElement('div');
        container.className = 'fixed top-4 right-4 z-50 space-y-2';
        container.id = 'toast-container';
        document.body.appendChild(container);
        return container;
    }

    show(message, type = 'info', duration = 5000) {
        const toast = document.createElement('div');
        const colors = {
            success: 'bg-green-500',
            error: 'bg-red-500',
            warning: 'bg-yellow-500',
            info: 'bg-blue-500'
        };

        const icons = {
            success: 'fas fa-check-circle',
            error: 'fas fa-exclamation-circle',
            warning: 'fas fa-exclamation-triangle',
            info: 'fas fa-info-circle'
        };

        toast.className = `${colors[type]} text-white px-4 py-3 rounded-lg shadow-lg transform translate-x-full transition-transform duration-300 flex items-center space-x-3 max-w-sm`;
        toast.innerHTML = `
            <i class="${icons[type]}"></i>
            <span class="flex-1">${message}</span>
            <button onclick="this.parentElement.remove()" class="text-white hover:text-gray-200">
                <i class="fas fa-times"></i>
            </button>
        `;

        this.container.appendChild(toast);
        
        // Animate in
        setTimeout(() => toast.classList.remove('translate-x-full'), 100);
        
        // Auto remove
        setTimeout(() => {
            toast.classList.add('translate-x-full');
            setTimeout(() => toast.remove(), 300);
        }, duration);
    }
}

// Loading spinner
class LoadingManager {
    constructor() {
        this.overlay = this.createOverlay();
    }

    createOverlay() {
        const overlay = document.createElement('div');
        overlay.className = 'fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center hidden';
        overlay.id = 'loading-overlay';
        overlay.innerHTML = `
            <div class="bg-white rounded-lg p-6 flex flex-col items-center space-y-4">
                <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-hedgehog-brown"></div>
                <p class="text-gray-600">Caricamento...</p>
            </div>
        `;
        document.body.appendChild(overlay);
        return overlay;
    }

    show() {
        this.overlay.classList.remove('hidden');
    }

    hide() {
        this.overlay.classList.add('hidden');
    }
}

// SwipeHandler class for page navigation
class SwipeHandler {
    constructor() {
        this.touchStartX = 0;
        this.touchEndX = 0;
        this.minSwipeDistance = 100; // Minimum distance for a swipe to be registered
        this.swipeThreshold = 0.3; // Percentage of screen width required for a swipe
        this.navigationPaths = [
            '/',
            '/hedgehogs',
            '/rooms',
            '/notifications'
        ];
        this.init();
    }

    init() {
        // Only initialize on mobile devices
        if (window.innerWidth >= 1024) return;
        
        // Add swipe detection to the document
        document.addEventListener('touchstart', (e) => this.handleTouchStart(e), false);
        document.addEventListener('touchend', (e) => this.handleTouchEnd(e), false);
        
        // Add swipe indicators if we're on a page that supports swipe navigation
        this.addSwipeIndicators();
    }

    handleTouchStart(e) {
        this.touchStartX = e.touches[0].clientX;
    }

    handleTouchEnd(e) {
        // Don't handle swipes if the target is an interactive element
        if (this.isInteractiveElement(e.target)) return;
        
        this.touchEndX = e.changedTouches[0].clientX;
        this.handleSwipe();
    }

    isInteractiveElement(element) {
        // Check if the element or its parents are interactive elements
        const interactiveElements = ['A', 'BUTTON', 'INPUT', 'SELECT', 'TEXTAREA', 'CANVAS'];
        let currentElement = element;
        
        while (currentElement) {
            if (interactiveElements.includes(currentElement.tagName)) {
                return true;
            }
            
            // Check for interactive roles
            const role = currentElement.getAttribute('role');
            if (role && ['button', 'link', 'checkbox', 'radio', 'switch', 'tab'].includes(role)) {
                return true;
            }
            
            // Check for common interactive classes
            if (currentElement.classList) {
                if (
                    currentElement.classList.contains('btn-primary') ||
                    currentElement.classList.contains('btn-secondary') ||
                    currentElement.classList.contains('card') ||
                    currentElement.classList.contains('bottom-nav-item')
                ) {
                    return true;
                }
            }
            
            currentElement = currentElement.parentElement;
        }
        
        return false;
    }

    handleSwipe() {
        const swipeDistance = this.touchEndX - this.touchStartX;
        const screenWidth = window.innerWidth;
        
        // Check if the swipe distance is significant enough
        if (Math.abs(swipeDistance) < this.minSwipeDistance || 
            Math.abs(swipeDistance) / screenWidth < this.swipeThreshold) {
            return;
        }
        
        const currentPath = window.location.pathname;
        const currentIndex = this.navigationPaths.indexOf(currentPath);
        
        // If current page is not in the navigation paths, don't navigate
        if (currentIndex === -1) return;
        
        if (swipeDistance > 0) {
            // Swipe right (go to previous page)
            if (currentIndex > 0) {
                this.navigateWithTransition(this.navigationPaths[currentIndex - 1], 'slide-right');
            }
        } else {
            // Swipe left (go to next page)
            if (currentIndex < this.navigationPaths.length - 1) {
                this.navigateWithTransition(this.navigationPaths[currentIndex + 1], 'slide-left');
            }
        }
    }

    navigateWithTransition(path, direction) {
        // Add transition class to the body
        document.body.classList.add('page-transition', direction);
        
        // Navigate after a short delay to allow the transition to start
        setTimeout(() => {
            window.location.href = path;
        }, 50);
    }

    addSwipeIndicators() {
        const currentPath = window.location.pathname;
        const currentIndex = this.navigationPaths.indexOf(currentPath);
        
        // If current page is not in the navigation paths, don't add indicators
        if (currentIndex === -1) return;
        
        const container = document.createElement('div');
        container.className = 'swipe-indicators';
        
        // Create left indicator if not on the first page
        if (currentIndex > 0) {
            const leftIndicator = document.createElement('div');
            leftIndicator.className = 'swipe-indicator left';
            leftIndicator.innerHTML = `
                <i class="fas fa-chevron-left"></i>
                <span>${this.getPageName(this.navigationPaths[currentIndex - 1])}</span>
            `;
            container.appendChild(leftIndicator);
        }
        
        // Create right indicator if not on the last page
        if (currentIndex < this.navigationPaths.length - 1) {
            const rightIndicator = document.createElement('div');
            rightIndicator.className = 'swipe-indicator right';
            rightIndicator.innerHTML = `
                <span>${this.getPageName(this.navigationPaths[currentIndex + 1])}</span>
                <i class="fas fa-chevron-right"></i>
            `;
            container.appendChild(rightIndicator);
        }
        
        // Add indicators to the page
        document.body.appendChild(container);
    }

    getPageName(path) {
        switch (path) {
            case '/':
                return 'Home';
            case '/hedgehogs':
                return 'Ricci';
            case '/rooms':
                return 'Stanze';
            case '/notifications':
                return 'Notifiche';
            default:
                return 'Page';
        }
    }
}

// Initialize mobile navigation
let mobileNav, toastManager, loadingManager, touchHandler, swipeHandler;

document.addEventListener('DOMContentLoaded', function() {
    // Initialize components
    mobileNav = new MobileNavigation();
    toastManager = new ToastManager();
    loadingManager = new LoadingManager();
    touchHandler = new TouchHandler();
    swipeHandler = new SwipeHandler();
    
    // Add no-select class to prevent text selection on touch interactions
    document.querySelectorAll('.card, button, .btn-primary, .btn-secondary, .btn-danger, .sidebar-icon-mobile, .nav-item')
        .forEach(el => el.classList.add('no-select'));
    
    // Add touch-spacing class to form groups and button groups
    document.querySelectorAll('.form-group, .space-y-1, .space-y-2, .space-y-3')
        .forEach(el => el.classList.add('touch-spacing'));
    
    // Add touch-target class to small interactive elements
    document.querySelectorAll('button:not(.touch-target), .sidebar-icon-mobile:not(.touch-target)')
        .forEach(el => el.classList.add('touch-target'));
});

// Global toast function
function showToast(message, type = 'info') {
    if (toastManager) {
        toastManager.show(message, type);
    }
}

// Global loading functions
function showLoading() {
    if (loadingManager) {
        loadingManager.show();
    }
}

function hideLoading() {
    if (loadingManager) {
        loadingManager.hide();
    }
}

// Global delete confirmation modal
function showDeleteModal(title, message, onConfirm) {
    const modal = document.createElement('div');
    modal.className = 'fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4';
    modal.innerHTML = `
        <div class="bg-white rounded-2xl shadow-2xl max-w-md w-full p-6">
            <div class="flex items-center space-x-3 mb-4">
                <div class="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center">
                    <i class="fas fa-trash text-red-600 text-xl"></i>
                </div>
                <div>
                    <h3 class="text-lg font-bold text-gray-900">${title}</h3>
                    <p class="text-gray-600 text-sm">${message}</p>
                </div>
            </div>
            <div class="flex justify-end space-x-3 mt-6">
                <button onclick="this.closest('.fixed').remove()" 
                        class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors">
                    Annulla
                </button>
                <button onclick="this.closest('.fixed').remove(); (${onConfirm.toString()})()"
                        class="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors">
                    <i class="fas fa-trash mr-2"></i>Elimina
                </button>
            </div>
        </div>
    `;
    document.body.appendChild(modal);
}