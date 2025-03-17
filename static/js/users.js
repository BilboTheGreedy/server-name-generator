/**
 * User management functionality for Server Name Generator
 */

// Global users data
let allUsers = [];
let isEditMode = false;

// Initialize user management
document.addEventListener('DOMContentLoaded', function() {
    // Add event listeners for user management templates
    document.addEventListener('templatesLoaded', function() {
        if (document.getElementById('usersTable')) {
            setupUserManagement();
        }
    });
});

// Setup user management
function setupUserManagement() {
    // Load users data
    loadUsers();
    
    // Setup event listeners
    document.getElementById('refreshUsers').addEventListener('click', loadUsers);
    document.getElementById('createUserBtn').addEventListener('click', showCreateUserModal);
    document.getElementById('saveUserBtn').addEventListener('click', saveUser);
    document.getElementById('confirmDeleteUserBtn').addEventListener('click', deleteUser);
    document.getElementById('savePasswordBtn').addEventListener('click', changePassword);
    
    // Setup filters
    document.getElementById('filter-user-role').addEventListener('change', applyUserFilters);
    document.getElementById('search-user').addEventListener('input', applyUserFilters);
    
    // Setup action listeners using event delegation
    document.addEventListener('click', function(e) {
        // Edit user button
        if (e.target.classList.contains('edit-user-btn') || e.target.closest('.edit-user-btn')) {
            const button = e.target.classList.contains('edit-user-btn') ? e.target : e.target.closest('.edit-user-btn');
            const userId = button.dataset.id;
            showEditUserModal(userId);
        }
        
        // Delete user button
        if (e.target.classList.contains('delete-user-btn') || e.target.closest('.delete-user-btn')) {
            const button = e.target.classList.contains('delete-user-btn') ? e.target : e.target.closest('.delete-user-btn');
            const userId = button.dataset.id;
            const username = button.dataset.username;
            showDeleteUserModal(userId, username);
        }
        
        // Change password button
        if (e.target.classList.contains('change-password-btn') || e.target.closest('.change-password-btn')) {
            const button = e.target.classList.contains('change-password-btn') ? e.target : e.target.closest('.change-password-btn');
            const userId = button.dataset.id;
            showChangePasswordModal(userId);
        }
    });
}

// Load all users
function loadUsers() {
    fetch('/api/users')
        .then(response => {
            if (!response.ok) {
                throw new Error(`Error: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            allUsers = data;
            applyUserFilters();
        })
        .catch(error => {
            console.error('Error loading users:', error);
            showAlert('Failed to load users: ' + error.message, 'danger');
        });
}

// Apply filters to users
function applyUserFilters() {
    const roleFilter = document.getElementById('filter-user-role').value;
    const searchTerm = document.getElementById('search-user').value.toLowerCase();
    
    const filteredUsers = allUsers.filter(user => {
        // Apply role filter
        if (roleFilter && user.role !== roleFilter) {
            return false;
        }
        
        // Apply search filter
        if (searchTerm) {
            const searchableText = `${user.username} ${user.email}`.toLowerCase();
            return searchableText.includes(searchTerm);
        }
        
        return true;
    });
    
    updateUsersTable(filteredUsers);
}

// Update users table
function updateUsersTable(users) {
    const tableBody = document.getElementById('usersTable');
    tableBody.innerHTML = '';
    
    if (users.length === 0) {
        const row = document.createElement('tr');
        const cell = document.createElement('td');
        cell.colSpan = 5;
        cell.textContent = 'No users found';
        cell.className = 'text-center';
        row.appendChild(cell);
        tableBody.appendChild(row);
        return;
    }
    
    users.forEach(user => {
        const row = document.createElement('tr');
        
        // Username cell
        const usernameCell = document.createElement('td');
        usernameCell.textContent = user.username;
        row.appendChild(usernameCell);
        
        // Email cell
        const emailCell = document.createElement('td');
        emailCell.textContent = user.email;
        row.appendChild(emailCell);
        
        // Role cell
        const roleCell = document.createElement('td');
        const roleBadge = document.createElement('span');
        roleBadge.classList.add('badge');
        
        if (user.role === 'admin') {
            roleBadge.classList.add('bg-danger');
            roleBadge.textContent = 'Admin';
        } else {
            roleBadge.classList.add('bg-secondary');
            roleBadge.textContent = 'User';
        }
        
        roleCell.appendChild(roleBadge);
        row.appendChild(roleCell);
        
        // Created at cell
        const createdAtCell = document.createElement('td');
        const date = new Date(user.createdAt);
        createdAtCell.textContent = date.toLocaleString();
        row.appendChild(createdAtCell);
        
        // Actions cell
        const actionsCell = document.createElement('td');
        
        // Edit button
        const editBtn = document.createElement('button');
        editBtn.classList.add('btn', 'btn-sm', 'btn-outline-primary', 'edit-user-btn', 'me-1');
        editBtn.innerHTML = '<i class="bi bi-pencil"></i>';
        editBtn.title = 'Edit';
        editBtn.dataset.id = user.id;
        actionsCell.appendChild(editBtn);
        
        // Change password button
        const passwordBtn = document.createElement('button');
        passwordBtn.classList.add('btn', 'btn-sm', 'btn-outline-warning', 'change-password-btn', 'me-1');
        passwordBtn.innerHTML = '<i class="bi bi-key"></i>';
        passwordBtn.title = 'Change Password';
        passwordBtn.dataset.id = user.id;
        actionsCell.appendChild(passwordBtn);
        
        // Delete button (don't allow deleting yourself)
        if (user.id !== getCurrentUserId()) {
            const deleteBtn = document.createElement('button');
            deleteBtn.classList.add('btn', 'btn-sm', 'btn-outline-danger', 'delete-user-btn');
            deleteBtn.innerHTML = '<i class="bi bi-trash"></i>';
            deleteBtn.title = 'Delete';
            deleteBtn.dataset.id = user.id;
            deleteBtn.dataset.username = user.username;
            actionsCell.appendChild(deleteBtn);
        }
        
        row.appendChild(actionsCell);
        tableBody.appendChild(row);
    });
}

// Show create user modal
function showCreateUserModal() {
    isEditMode = false;
    document.getElementById('userModalTitle').textContent = 'Create New User';
    document.getElementById('userForm').reset();
    document.getElementById('userId').value = '';
    document.getElementById('passwordField').style.display = 'block';
    document.getElementById('password').required = true;
    
    const userModal = new bootstrap.Modal(document.getElementById('userModal'));
    userModal.show();
}

// Show edit user modal
function showEditUserModal(userId) {
    isEditMode = true;
    const user = allUsers.find(u => u.id === userId);
    if (!user) {
        console.error('User not found:', userId);
        return;
    }
    
    document.getElementById('userModalTitle').textContent = 'Edit User';
    document.getElementById('userId').value = user.id;
    document.getElementById('username').value = user.username;
    document.getElementById('email').value = user.email;
    document.getElementById('role').value = user.role;
    document.getElementById('passwordField').style.display = 'none';
    document.getElementById('password').required = false;
    
    const userModal = new bootstrap.Modal(document.getElementById('userModal'));
    userModal.show();
}

// Show change password modal
function showChangePasswordModal(userId) {
    document.getElementById('passwordForm').reset();
    document.getElementById('passwordUserId').value = userId;
    document.getElementById('passwordError').classList.add('d-none');
    
    const passwordModal = new bootstrap.Modal(document.getElementById('passwordModal'));
    passwordModal.show();
}

// Show delete user modal
function showDeleteUserModal(userId, username) {
    document.getElementById('deleteUserId').value = userId;
    document.getElementById('deleteUsername').textContent = username;
    
    const deleteModal = new bootstrap.Modal(document.getElementById('deleteUserModal'));
    deleteModal.show();
}

// Save user (create or update)
function saveUser() {
    // Get form data
    const userId = document.getElementById('userId').value;
    const username = document.getElementById('username').value;
    const email = document.getElementById('email').value;
    const role = document.getElementById('role').value;
    const password = document.getElementById('password').value;
    
    // Validate form
    if (!username || !email || (!isEditMode && !password)) {
        showAlert('Please fill in all required fields', 'danger');
        return;
    }
    
    // Create request payload
    let payload;
    let url;
    let method;
    
    if (isEditMode) {
        // Update existing user
        payload = { username, email, role };
        url = `/api/users/${userId}`;
        method = 'PUT';
    } else {
        // Create new user
        payload = { username, email, role, password };
        url = '/api/users';
        method = 'POST';
    }
    
    // Send request
    fetch(url, {
        method: method,
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
        // Hide modal
        const userModal = bootstrap.Modal.getInstance(document.getElementById('userModal'));
        userModal.hide();
        
        // Show success message
        const action = isEditMode ? 'updated' : 'created';
        showAlert(`User ${action} successfully`, 'success');
        
        // Reload users
        loadUsers();
    })
    .catch(error => {
        console.error('Error saving user:', error);
        showAlert(error.message, 'danger');
    });
}

// Change user password
function changePassword() {
    // Get form data
    const userId = document.getElementById('passwordUserId').value;
    const newPassword = document.getElementById('newPassword').value;
    const confirmPassword = document.getElementById('confirmPassword').value;
    
    // Validate passwords match
    if (newPassword !== confirmPassword) {
        document.getElementById('passwordError').textContent = 'Passwords do not match';
        document.getElementById('passwordError').classList.remove('d-none');
        return;
    }
    
    // Validate password length
    if (newPassword.length < 8) {
        document.getElementById('passwordError').textContent = 'Password must be at least 8 characters';
        document.getElementById('passwordError').classList.remove('d-none');
        return;
    }
    
    // Send request
    fetch(`/api/users/${userId}/password`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ password: newPassword })
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
        const passwordModal = bootstrap.Modal.getInstance(document.getElementById('passwordModal'));
        passwordModal.hide();
        
        // Show success message
        showAlert('Password changed successfully', 'success');
    })
    .catch(error => {
        console.error('Error changing password:', error);
        document.getElementById('passwordError').textContent = error.message;
        document.getElementById('passwordError').classList.remove('d-none');
    });
}

// Delete user
function deleteUser() {
    const userId = document.getElementById('deleteUserId').value;
    
    fetch(`/api/users/${userId}`, {
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
        const deleteModal = bootstrap.Modal.getInstance(document.getElementById('deleteUserModal'));
        deleteModal.hide();
        
        // Show success message
        showAlert('User deleted successfully', 'success');
        
        // Reload users
        loadUsers();
    })
    .catch(error => {
        console.error('Error deleting user:', error);
        showAlert(error.message, 'danger');
    });
}

// Helper function to get current user ID
function getCurrentUserId() {
    return window.authService && window.authService.getCurrentUser ? 
        window.authService.getCurrentUser().id : null;
}