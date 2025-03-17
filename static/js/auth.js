/**
 * Authentication functionality for Server Name Generator
 */

// Global auth state
const auth = {
    isAuthenticated: false,
    token: null,
    user: null,
    tokenExpiry: null
};

// Initialize auth from localStorage
document.addEventListener('DOMContentLoaded', function() {
    // Try to load authentication from localStorage
    const storedToken = localStorage.getItem('auth_token');
    const storedUser = localStorage.getItem('auth_user');
    const storedExpiry = localStorage.getItem('auth_expiry');
    
    if (storedToken && storedUser && storedExpiry) {
        // Check if token is expired
        const expiryDate = new Date(storedExpiry);
        if (expiryDate > new Date()) {
            // Token is still valid
            auth.isAuthenticated = true;
            auth.token = storedToken;
            auth.user = JSON.parse(storedUser);
            auth.tokenExpiry = expiryDate;
            
            // Set up authorization header for future requests
            setupAuthHeader(storedToken);
            showAuthenticatedUI();
        } else {
            // Token is expired, clear auth data
            clearAuth();
            showLoginForm();
        }
    } else {
        // No stored auth data
        clearAuth();
        showLoginForm();
    }
    
    // Set up login form submission handler
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }
    
    // Set up logout button handler
    document.addEventListener('click', function(e) {
        if (e.target.id === 'logoutBtn' || e.target.closest('#logoutBtn')) {
            e.preventDefault();
            handleLogout();
        }
    });
});

// Handle login form submission
function handleLogin(e) {
    e.preventDefault();
    
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    
    // Clear previous error messages
    hideLoginError();
    
    // Perform login request
    fetch('/api/auth/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ username, password })
    })
    .then(response => {
        if (!response.ok) {
            return response.json().then(data => {
                throw new Error(data.message || 'Invalid username or password');
            });
        }
        return response.json();
    })
    .then(data => {
        // Save authentication data
        auth.isAuthenticated = true;
        auth.token = data.token;
        auth.user = data.user;
        auth.tokenExpiry = new Date(data.expiresAt);
        
        // Store auth data in localStorage
        localStorage.setItem('auth_token', data.token);
        localStorage.setItem('auth_user', JSON.stringify(data.user));
        localStorage.setItem('auth_expiry', data.expiresAt);
        
        // Set up authorization header for future requests
        setupAuthHeader(data.token);
        
        // Show authenticated UI
        showAuthenticatedUI();
        
        // Show welcome message
        showAlert(`Welcome, ${data.user.username}!`, 'success');
    })
    .catch(error => {
        showLoginError(error.message);
    });
}

// Handle logout
function handleLogout() {
    // Clear auth data
    clearAuth();
    
    // Redirect to login
    window.location.href = '/admin';
}

// Clear authentication data
function clearAuth() {
    auth.isAuthenticated = false;
    auth.token = null;
    auth.user = null;
    auth.tokenExpiry = null;
    
    // Clear localStorage
    localStorage.removeItem('auth_token');
    localStorage.removeItem('auth_user');
    localStorage.removeItem('auth_expiry');
}

// Setup authorization header for all future fetch requests
function setupAuthHeader(token) {
    // Create a fetch proxy that adds the authorization header
    const originalFetch = window.fetch;
    window.fetch = function(url, options = {}) {
        // Only add auth header for API requests
        if (url.toString().includes('/api/')) {
            // Create headers if they don't exist
            options.headers = options.headers || {};
            
            // Don't override if Authorization is already set
            if (!options.headers.Authorization && !options.headers.authorization) {
                options.headers.Authorization = `Bearer ${token}`;
            }
        }
        
        return originalFetch(url, options);
    };
}

// Show login error message
function showLoginError(message) {
    const errorContainer = document.getElementById('login-error');
    const errorMessage = document.getElementById('login-error-message');
    
    if (errorContainer && errorMessage) {
        errorMessage.textContent = message;
        errorContainer.classList.remove('d-none');
    }
}

// Hide login error message
function hideLoginError() {
    const errorContainer = document.getElementById('login-error');
    
    if (errorContainer) {
        errorContainer.classList.add('d-none');
    }
}

// Show authenticated UI elements
function showAuthenticatedUI() {
    // Hide login form if present
    const loginForm = document.querySelector('.container:has(#loginForm)');
    if (loginForm) {
        loginForm.style.display = 'none';
    }
    
    // Show the main content
    const mainContent = document.querySelector('main');
    if (mainContent) {
        mainContent.classList.remove('d-none');
    }
    
    // Update user info in the UI
    updateUserInfo();
    
    // Load dashboard data
    if (typeof loadDashboardData === 'function') {
        loadDashboardData();
    }
    
    // Load reservations
    if (typeof loadReservations === 'function') {
        loadReservations();
    }
}

// Show login form
function showLoginForm() {
    // Load the login template if it doesn't exist
    if (!document.getElementById('loginForm')) {
        fetch('/static/templates/login.html')
            .then(response => response.text())
            .then(html => {
                // Hide main content
                const mainContent = document.querySelector('main');
                if (mainContent) {
                    mainContent.classList.add('d-none');
                }
                
                // Create login container
                const loginContainer = document.createElement('div');
                loginContainer.innerHTML = html;
                document.body.appendChild(loginContainer);
                
                // Add submit handler to login form
                const loginForm = document.getElementById('loginForm');
                if (loginForm) {
                    loginForm.addEventListener('submit', handleLogin);
                }
            })
            .catch(error => {
                console.error('Failed to load login template:', error);
            });
    } else {
        // Login form already exists, just show it
        const loginContainer = document.querySelector('.container:has(#loginForm)');
        if (loginContainer) {
            loginContainer.style.display = 'block';
        }
        
        // Hide main content
        const mainContent = document.querySelector('main');
        if (mainContent) {
            mainContent.classList.add('d-none');
        }
    }
}

// Update user info in the UI
function updateUserInfo() {
    // Add user info to navbar if not already present
    const navbar = document.querySelector('.navbar-nav');
    if (navbar && auth.user) {
        // Check if user info already exists
        if (!document.getElementById('userInfo')) {
            const userInfoHTML = `
                <li class="nav-item dropdown ms-md-auto">
                    <a class="nav-link dropdown-toggle" href="#" id="userInfo" role="button" data-bs-toggle="dropdown" aria-expanded="false">
                        <i class="bi bi-person-circle me-1"></i>${auth.user.username}
                    </a>
                    <ul class="dropdown-menu dropdown-menu-end" aria-labelledby="userInfo">
                        <li><span class="dropdown-item-text text-muted small">${auth.user.email}</span></li>
                        <li><span class="dropdown-item-text text-muted small">Role: ${auth.user.role}</span></li>
                        <li><hr class="dropdown-divider"></li>
                        <li><a class="dropdown-item" href="#" id="logoutBtn"><i class="bi bi-box-arrow-right me-1"></i>Logout</a></li>
                    </ul>
                </li>
            `;
            navbar.insertAdjacentHTML('beforeend', userInfoHTML);
        }
    }
}

// Check if user is authenticated
function isAuthenticated() {
    return auth.isAuthenticated;
}

// Get current user
function getCurrentUser() {
    return auth.user;
}

// Check if current user has a specific role
function hasRole(role) {
    return auth.user && auth.user.role === role;
}

// Export functions for use in other scripts
window.authService = {
    isAuthenticated,
    getCurrentUser,
    hasRole,
    login: handleLogin,
    logout: handleLogout
};


// Show authenticated UI elements
function showAuthenticatedUI() {
    // Hide login form if present
    const loginForm = document.querySelector('.container:has(#loginForm)');
    if (loginForm) {
        loginForm.style.display = 'none';
    }
    
    // Show the main content
    const mainContent = document.querySelector('main');
    if (mainContent) {
        mainContent.classList.remove('d-none');
    }
    
    // Update user info in the UI
    updateUserInfo();
    
    // Show/hide admin features based on user role
    if (typeof showHideAdminFeatures === 'function') {
        showHideAdminFeatures();
    }
    
    // Load dashboard data
    if (typeof loadDashboardData === 'function') {
        loadDashboardData();
    }
    
    // Load reservations
    if (typeof loadReservations === 'function') {
        loadReservations();
    }
}