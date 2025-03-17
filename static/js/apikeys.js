/**
 * API Key management functionality for Server Name Generator
 */

// Global API key data
let allApiKeys = [];
let isAdminUser = false;

// Initialize API key management
document.addEventListener('DOMContentLoaded', function() {
    // Add event listeners for API key management templates
    document.addEventListener('templatesLoaded', function() {
        if (document.getElementById('apiKeysTable')) {
            setupApiKeyManagement();
        }
    });
});

// Setup API key management
function setupApiKeyManagement() {
    // Check if current user is admin
    isAdminUser = window.authService && window.authService.hasRole ? 
        window.authService.hasRole('admin') : false;
    
    // Load API keys data
    loadApiKeys();
    
    // Setup event listeners
    document.getElementById('refreshApiKeys').addEventListener('click', loadApiKeys);
    document.getElementById('createApiKeyBtn').addEventListener('click', showCreateApiKeyModal);
    document.getElementById('generateApiKeyBtn').addEventListener('click', generateApiKey);
    document.getElementById('confirmRevokeKeyBtn').addEventListener('click', revokeApiKey);
    
    // Copy API key button
    document.getElementById('copyApiKeyBtn').addEventListener('click', function() {
        const keyInput = document.getElementById('generatedApiKey');
        keyInput.select();
        document.execCommand('copy');
        showAlert('API key copied to clipboard', 'success');
    });
    
    // Setup action listeners using event delegation
    document.addEventListener('click', function(e) {
        // Revoke API key button
        if (e.target.classList.contains('revoke-key-btn') || e.target.closest('.revoke-key-btn')) {
            const button = e.target.classList.contains('revoke-key-btn') ? e.target : e.target.closest('.revoke-key-btn');
            const keyId = button.dataset.id;
            const keyName = button.dataset.name;
            showRevokeKeyModal(keyId, keyName);
        }
    });
}

// Load API keys
function loadApiKeys() {
    // Determine which endpoint to use based on user role
    const endpoint = isAdminUser ? '/api/api-keys' : '/api/api-keys';
    
    fetch(endpoint)
        .then(response => {
            if (!response.ok) {
                throw new Error(`Error: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            allApiKeys = data;
            updateApiKeysTable(allApiKeys);
        })
        .catch(error => {
            console.error('Error loading API keys:', error);
            showAlert('Failed to load API keys: ' + error.message, 'danger');
        });
}

// Update API keys table
function updateApiKeysTable(keys) {
    const tableBody = document.getElementById('apiKeysTable');
    tableBody.innerHTML = '';
    
    if (keys.length === 0) {
        const row = document.createElement('tr');
        const cell = document.createElement('td');
        cell.colSpan = 6;
        cell.textContent = 'No API keys found';
        cell.className = 'text-center';
        row.appendChild(cell);
        tableBody.appendChild(row);
        return;
    }
    
    keys.forEach(key => {
        const row = document.createElement('tr');
        
        // Name cell
        const nameCell = document.createElement('td');
        nameCell.textContent = key.name;
        if (key.description) {
            const descriptionEl = document.createElement('div');
            descriptionEl.className = 'small text-muted';
            descriptionEl.textContent = key.description;
            nameCell.appendChild(descriptionEl);
        }
        row.appendChild(nameCell);
        
        // Created at cell
        const createdAtCell = document.createElement('td');
        const createdDate = new Date(key.createdAt);
        createdAtCell.textContent = createdDate.toLocaleString();
        row.appendChild(createdAtCell);
        
        // Expires at cell
        const expiresAtCell = document.createElement('td');
        if (key.expiresAt) {
            const expiryDate = new Date(key.expiresAt);
            expiresAtCell.textContent = expiryDate.toLocaleString();
            
            // Add warning if expires soon
            const now = new Date();
            const daysUntilExpiry = Math.ceil((expiryDate - now) / (1000 * 60 * 60 * 24));
            if (daysUntilExpiry <= 7 && daysUntilExpiry > 0) {
                expiresAtCell.innerHTML += ` <span class="badge bg-warning text-dark">Expires in ${daysUntilExpiry} day${daysUntilExpiry !== 1 ? 's' : ''}</span>`;
            } else if (daysUntilExpiry <= 0) {
                expiresAtCell.innerHTML += ' <span class="badge bg-danger">Expired</span>';
            }
        } else {
            expiresAtCell.textContent = 'Never';
        }
        row.appendChild(expiresAtCell);
        
        // Last used cell
        const lastUsedCell = document.createElement('td');
        if (key.lastUsed) {
            const lastUsedDate = new Date(key.lastUsed);
            lastUsedCell.textContent = lastUsedDate.toLocaleString();
        } else {
            lastUsedCell.textContent = 'Never';
        }
        row.appendChild(lastUsedCell);
        
        // Status cell
        const statusCell = document.createElement('td');
        const statusBadge = document.createElement('span');
        statusBadge.classList.add('badge');
        
        if (key.isActive) {
            statusBadge.classList.add('bg-success');
            statusBadge.textContent = 'Active';
        } else {
            statusBadge.classList.add('bg-secondary');
            statusBadge.textContent = 'Revoked';
        }
        
        statusCell.appendChild(statusBadge);
        row.appendChild(statusCell);
        
        // Actions cell
        const actionsCell = document.createElement('td');
        
        if (key.isActive) {
            // Revoke button
            const revokeBtn = document.createElement('button');
            revokeBtn.classList.add('btn', 'btn-sm', 'btn-outline-danger', 'revoke-key-btn');
            revokeBtn.innerHTML = '<i class="bi bi-x-circle"></i> Revoke';
            revokeBtn.title = 'Revoke API Key';
            revokeBtn.dataset.id = key.id;
            revokeBtn.dataset.name = key.name;
            actionsCell.appendChild(revokeBtn);
        }
        
        row.appendChild(actionsCell);
        tableBody.appendChild(row);
    });
}

// Show create API key modal
function showCreateApiKeyModal() {
    document.getElementById('apiKeyForm').reset();
    
    // Set default checked state for scopes
    document.getElementById('scopeRead').checked = true;
    document.getElementById('scopeReserve').checked = true;
    document.getElementById('scopeCommit').checked = false;
    document.getElementById('scopeRelease').checked = false;
    
    const apiKeyModal = new bootstrap.Modal(document.getElementById('apiKeyModal'));
    apiKeyModal.show();
}

// Show revoke API key modal
function showRevokeKeyModal(keyId, keyName) {
    document.getElementById('revokeKeyId').value = keyId;
    document.getElementById('revokeKeyName').textContent = keyName;
    
    const revokeModal = new bootstrap.Modal(document.getElementById('revokeKeyModal'));
    revokeModal.show();
}

// Generate API key
function generateApiKey() {
    // Get form data
    const name = document.getElementById('keyName').value;
    const description = document.getElementById('keyDescription').value;
    const expiresIn = parseInt(document.getElementById('keyExpiry').value);
    
    // Get selected scopes
    const scopes = [];
    if (document.getElementById('scopeRead').checked) scopes.push('read');
    if (document.getElementById('scopeReserve').checked) scopes.push('reserve');
    if (document.getElementById('scopeCommit').checked) scopes.push('commit');
    if (document.getElementById('scopeRelease').checked) scopes.push('release');
    
    // Validate form
    if (!name) {
        showAlert('Please enter a name for your API key', 'danger');
        return;
    }
    
    // Create request payload
    const payload = {
        name,
        description,
        expiresIn,
        scopes
    };
    
    // Send request
    fetch('/api/api-keys', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
    })
    .then(response => {
        if (!response.ok) {
            return response.json().then(data => {
                throw new Error(data.message || `Error: ${response.status}`);
            });
        }
        return response.json();
    })
    .then(data => {
        // Hide create modal
        const apiKeyModal = bootstrap.Modal.getInstance(document.getElementById('apiKeyModal'));
        apiKeyModal.hide();
        
        // Show the API key in the result modal
        document.getElementById('generatedApiKey').value = data.key;
        
        // Show the API key created modal
        const apiKeyCreatedModal = new bootstrap.Modal(document.getElementById('apiKeyCreatedModal'));
        apiKeyCreatedModal.show();
        
        // Reload API keys
        loadApiKeys();
    })
    .catch(error => {
        console.error('Error generating API key:', error);
        showAlert(error.message, 'danger');
    });
}

// Revoke API key
function revokeApiKey() {
    const keyId = document.getElementById('revokeKeyId').value;
    
    fetch(`/api/api-keys/${keyId}`, {
        method: 'DELETE'
    })
    .then(response => {
        if (!response.ok) {
            return response.json().then(data => {
                throw new Error(data.message || `Error: ${response.status}`);
            });
        }
        return response.json();
    })
    .then(data => {
        // Hide modal
        const revokeModal = bootstrap.Modal.getInstance(document.getElementById('revokeKeyModal'));
        revokeModal.hide();
        
        // Show success message
        showAlert('API key revoked successfully', 'success');
        
        // Reload API keys
        loadApiKeys();
    })
    .catch(error => {
        console.error('Error revoking API key:', error);
        showAlert(error.message, 'danger');
    });
}