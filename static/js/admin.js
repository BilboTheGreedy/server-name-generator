/**
 * Main admin dashboard functionality
 */

// Global variables for charts
let activityChart = null;
let environmentChart = null;
let regionsChart = null;
let functionsChart = null;
let allReservations = [];
let confirmModalAction = null;
let serverNameToDelete = null;
// Wait for DOM and templates to be loaded
document.addEventListener('DOMContentLoaded', function() {
    // Wait for templates to be loaded
    if (document.querySelectorAll('.tab-pane:empty').length === 0) {
        initializeApp();
    } else {
        document.addEventListener('templatesLoaded', initializeApp);
    }
});

// Initialize the app
function initializeApp() {
    // Tab Navigation
    const tabs = {
        dashboard: document.getElementById('dashboard'),
        generate: document.getElementById('generate'),
        manage: document.getElementById('manage'),
        statistics: document.getElementById('statistics'),
        apiExplorer: document.getElementById('apiExplorer')
    };
    
    const navLinks = {
        dashboard: [document.getElementById('nav-dashboard'), document.getElementById('side-dashboard')],
        generate: [document.getElementById('nav-generate'), document.getElementById('side-generate')],
        manage: [document.getElementById('nav-manage'), document.getElementById('side-manage')],
        statistics: [document.getElementById('side-stats')],
        apiExplorer: [document.getElementById('side-api')]
    };
    
    function showTab(tabName) {
        // Hide all tabs
        Object.values(tabs).forEach(tab => {
            if (tab) tab.classList.remove('show', 'active');
        });
        
        // Remove active class from all nav links
        Object.values(navLinks).flat().forEach(link => {
            if (link) link.classList.remove('active');
        });
        
        // Show selected tab
        if (tabs[tabName]) {
            tabs[tabName].classList.add('show', 'active');
        }
        
        // Set active class on relevant nav links
        if (navLinks[tabName]) {
            navLinks[tabName].forEach(link => {
                if (link) link.classList.add('active');
            });
        }
    }
    
    // Add click event listeners to navigation links
    Object.entries(navLinks).forEach(([tabName, links]) => {
        links.forEach(link => {
            if (link) {
                link.addEventListener('click', (e) => {
                    e.preventDefault();
                    showTab(tabName);
                });
            }
        });
    });
    
    // Setup input handlers for character counting and live preview
    setupInputHandlers();
    
    // Initial data load
    loadDashboardData();
    loadReservations();
    
    // Handle reservation form submission
    document.getElementById('reservationForm').addEventListener('submit', function(e) {
        e.preventDefault();
        
        const formData = {
            unitCode: document.getElementById('unitCode').value,
            type: document.getElementById('type').value,
            provider: document.getElementById('provider').value,
            region: document.getElementById('region').value,
            environment: document.getElementById('environment').value,
            function: document.getElementById('function').value
        };
        
        fetch('/api/reserve', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(formData)
        })
        .then(response => {
            if (!response.ok) {
                return response.json().then(err => { throw new Error(err.message || 'Failed to reserve name'); });
            }
            return response.json();
        })
        .then(data => {
            document.getElementById('serverName').textContent = data.serverName;
            document.getElementById('reservationId').textContent = data.reservationId;
            document.getElementById('reservationResult').classList.remove('d-none');
            showAlert('Name reserved successfully: ' + data.serverName, 'success');
            loadReservations(); // Refresh the reservations list
            loadDashboardData(); // Refresh dashboard data
        })
        .catch(error => {
            showAlert(error.message, 'danger');
        });
    });
    
    // Handle form reset
    document.getElementById('reservationForm').addEventListener('reset', function() {
        // Hide result div when form is reset
        document.getElementById('reservationResult').classList.add('d-none');
        
        // Reset preview to defaults (need to use setTimeout due to timing of reset event)
        setTimeout(() => {
            document.querySelectorAll('input[data-preview]').forEach(input => {
                const previewId = input.dataset.preview;
                const defaultValue = input.dataset.default;
                
                document.getElementById(previewId).textContent = defaultValue;
                document.getElementById(previewId).classList.remove('text-primary');
                document.getElementById(previewId).classList.add('text-secondary');
                
                // Reset counters
                const counterId = input.id + 'Count';
                document.getElementById(counterId).textContent = '0';
            });
        }, 50);
    });

    // Handle commit button click
    document.getElementById('commitBtn').addEventListener('click', function() {
        const reservationId = document.getElementById('reservationId').textContent;
        
        if (!reservationId) {
            showAlert('No reservation to commit', 'warning');
            return;
        }
        
        commitReservation(reservationId);
    });

    // Refresh buttons
    document.getElementById('refresh-dashboard').addEventListener('click', function() {
        loadDashboardData();
    });
    
    document.getElementById('refreshReservations').addEventListener('click', function() {
        loadReservations();
    });
    
    // Setup filters
    document.getElementById('filter-status').addEventListener('change', applyFilters);
    document.getElementById('filter-environment').addEventListener('change', applyFilters);
    document.getElementById('filter-region').addEventListener('change', applyFilters);
    document.getElementById('search').addEventListener('input', applyFilters);
    
    // Update the confirmAction event listener
    document.getElementById('confirmAction').addEventListener('click', function() {
        if (confirmModalAction) {
            const success = confirmModalAction();
            
            // Only close modal if the action was successful
            if (success !== false) {
                const confirmModal = bootstrap.Modal.getInstance(document.getElementById('confirmModal'));
                confirmModal.hide();
            }
        }
    });
    
        // Add input event listener to reset validation state
    document.getElementById('confirmServerName').addEventListener('input', function() {
        this.classList.remove('is-invalid');
    });

    // Action handlers (commit and delete) using event delegation
    document.addEventListener('click', function(e) {
        // Commit button clicked
        if (e.target.classList.contains('commit-btn')) {
            const reservationId = e.target.dataset.id;
            commitReservation(reservationId);
        }
        
    // Delete button clicked
    if (e.target.classList.contains('delete-btn')) {
        const reservationId = e.target.dataset.id;
        const serverName = e.target.dataset.name;
        const isCommitted = e.target.closest('tr').querySelector('.badge').textContent.trim() === 'Committed';
        
        // Store server name for validation
        serverNameToDelete = serverName;
        
        // Set up confirmation modal with appropriate warning
        let message = `Are you sure you want to delete the reservation for "${serverName}"?`;
        
        if (isCommitted) {
            message = `WARNING: "${serverName}" is COMMITTED. Deleting this reservation could cause conflicts if the server name is already in use.`;
            
            // Show name confirmation section for committed reservations
            document.getElementById('nameConfirmationSection').classList.remove('d-none');
            document.getElementById('confirmServerName').value = '';
            document.getElementById('confirmServerName').classList.remove('is-invalid');
        } else {
            // Hide name confirmation for non-committed reservations
            document.getElementById('nameConfirmationSection').classList.add('d-none');
        }
        
        document.getElementById('confirmMessage').innerHTML = message;
        
        document.getElementById('confirmAction').classList.add('btn-danger');
        document.getElementById('confirmAction').textContent = 'Delete';
        
        const confirmModal = new bootstrap.Modal(document.getElementById('confirmModal'));
        confirmModalAction = () => {
            // For committed reservations, verify the name matches
            if (isCommitted) {
                const enteredName = document.getElementById('confirmServerName').value;
                if (enteredName !== serverName) {
                    document.getElementById('confirmServerName').classList.add('is-invalid');
                    return false; // Prevent modal from closing
                }
            }
            
            // If validation passes, delete the reservation
            deleteReservation(reservationId);
            return true; // Allow modal to close
        };
        
        confirmModal.show();
    }

    });
}

// Function to set up input handlers for character counting and live preview
function setupInputHandlers() {
    const inputs = document.querySelectorAll('input[data-preview]');
    
    inputs.forEach(input => {
        const counterId = input.id + 'Count';
        const previewId = input.dataset.preview;
        const defaultValue = input.dataset.default;
        
        // Initial count
        document.getElementById(counterId).textContent = input.value.length;
        
        // Initial preview
        if (input.value) {
            document.getElementById(previewId).textContent = input.value.toUpperCase();
            document.getElementById(previewId).classList.remove('text-secondary');
            document.getElementById(previewId).classList.add('text-primary');
        } else {
            document.getElementById(previewId).textContent = defaultValue;
            document.getElementById(previewId).classList.remove('text-primary');
            document.getElementById(previewId).classList.add('text-secondary');
        }
        
        // Update on input
        input.addEventListener('input', function() {
            document.getElementById(counterId).textContent = this.value.length;
            
            if (this.value) {
                document.getElementById(previewId).textContent = this.value.toUpperCase();
                document.getElementById(previewId).classList.remove('text-secondary');
                document.getElementById(previewId).classList.add('text-primary');
            } else {
                document.getElementById(previewId).textContent = defaultValue;
                document.getElementById(previewId).classList.remove('text-primary');
                document.getElementById(previewId).classList.add('text-secondary');
            }
        });
    });
}

// Function to load dashboard data
function loadDashboardData() {
    fetch('/api/stats')
    .then(response => {
        if (!response.ok) {
            throw new Error('Failed to load dashboard data');
        }
        return response.json();
    })
    .then(data => {
        // Update stats
        document.getElementById('total-reservations').textContent = data.totalReservations;
        document.getElementById('committed-reservations').textContent = data.committedCount;
        document.getElementById('reserved-reservations').textContent = data.reservedCount;
        
        // Update recent reservations
        updateRecentReservations(data.recentReservations);
        
        // Update charts
        updateActivityChart(data.dailyActivity);
        updateEnvironmentChart(data.topEnvironments);
    })
    .catch(error => {
        showAlert(error.message, 'danger');
    });
}

// Function to update the activity chart
function updateActivityChart(activities) {
    const ctx = document.getElementById('activityChart');
    
    // Extract data for chart
    const labels = activities.map(activity => activity.date);
    const reservedData = activities.map(activity => activity.reserved);
    const committedData = activities.map(activity => activity.committed);
    
    // Create or update chart
    if (activityChart) {
        activityChart.data.labels = labels;
        activityChart.data.datasets[0].data = reservedData;
        activityChart.data.datasets[1].data = committedData;
        activityChart.update();
    } else {
        activityChart = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [
                    {
                        label: 'Reserved',
                        data: reservedData,
                        backgroundColor: '#ffc107',
                        borderColor: '#ffc107',
                        borderWidth: 1
                    },
                    {
                        label: 'Committed',
                        data: committedData,
                        backgroundColor: '#28a745',
                        borderColor: '#28a745',
                        borderWidth: 1
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            precision: 0
                        }
                    }
                }
            }
        });
    }
}

// Function to update the environment chart
function updateEnvironmentChart(environments) {
    const ctx = document.getElementById('environmentChart');
    
    // Extract data for chart
    const labels = environments.map(env => env.environment);
    const data = environments.map(env => env.count);
    const backgroundColors = [
        '#4e73df', '#1cc88a', '#36b9cc', '#f6c23e', '#e74a3b'
    ];
    
    // Create or update chart
    if (environmentChart) {
        environmentChart.data.labels = labels;
        environmentChart.data.datasets[0].data = data;
        environmentChart.update();
    } else {
        environmentChart = new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: labels,
                datasets: [{
                    data: data,
                    backgroundColor: backgroundColors,
                    hoverBackgroundColor: backgroundColors,
                    hoverBorderColor: "rgba(234, 236, 244, 1)",
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                cutout: '70%',
                plugins: {
                    legend: {
                        position: 'right'
                    }
                }
            }
        });
    }
}

// Function to update recent reservations table
function updateRecentReservations(reservations) {
    const tableBody = document.getElementById('recent-reservations');
    tableBody.innerHTML = '';
    
    if (reservations.length === 0) {
        const row = document.createElement('tr');
        const cell = document.createElement('td');
        cell.colSpan = 4;
        cell.textContent = 'No reservations found';
        cell.className = 'text-center';
        row.appendChild(cell);
        tableBody.appendChild(row);
        return;
    }
    
    reservations.forEach(reservation => {
        const row = document.createElement('tr');
        
        // Server name cell
        const nameCell = document.createElement('td');
        nameCell.className = 'font-monospace';
        nameCell.textContent = reservation.serverName;
        row.appendChild(nameCell);
        
        // Status cell
        const statusCell = document.createElement('td');
        const statusBadge = document.createElement('span');
        statusBadge.classList.add('badge');
        
        if (reservation.status === 'reserved') {
            statusBadge.classList.add('bg-warning', 'text-dark');
            statusBadge.textContent = 'Reserved';
        } else {
            statusBadge.classList.add('bg-success');
            statusBadge.textContent = 'Committed';
        }
        
        statusCell.appendChild(statusBadge);
        row.appendChild(statusCell);
        
        // Created at cell
        const createdAtCell = document.createElement('td');
        const date = new Date(reservation.createdAt);
        createdAtCell.textContent = date.toLocaleString();
        row.appendChild(createdAtCell);
        
        // Actions cell
        const actionsCell = document.createElement('td');
        
        if (reservation.status === 'reserved') {
            // Commit button
            const commitBtn = document.createElement('button');
            commitBtn.classList.add('btn', 'btn-sm', 'btn-outline-success', 'commit-btn', 'me-2');
            commitBtn.innerHTML = '<i class="bi bi-check-circle"></i>';
            commitBtn.title = 'Commit';
            commitBtn.dataset.id = reservation.id;
            actionsCell.appendChild(commitBtn);
        }
        
        // Delete button (now for both reserved and committed)
        const deleteBtn = document.createElement('button');
        deleteBtn.classList.add('btn', 'btn-sm', 'btn-outline-danger', 'delete-btn');
        deleteBtn.innerHTML = '<i class="bi bi-trash"></i>';
        deleteBtn.title = 'Delete';
        deleteBtn.dataset.id = reservation.id;
        deleteBtn.dataset.name = reservation.serverName;
        actionsCell.appendChild(deleteBtn);
        
        row.appendChild(actionsCell);
        tableBody.appendChild(row);
    });
}

// Function to load all reservations
function loadReservations() {
    fetch('/api/reservations')
    .then(response => {
        if (!response.ok) {
            throw new Error('Failed to load reservations');
        }
        return response.json();
    })
    .then(data => {
        allReservations = data;
        
        // Populate filter dropdowns
        populateFilterOptions(data);
        
        // Apply current filters
        applyFilters();
    })
    .catch(error => {
        showAlert(error.message, 'danger');
    });
}

// Function to populate filter options
function populateFilterOptions(reservations) {
    // Get unique environments
    const environments = [...new Set(reservations.map(r => r.environment))];
    const envSelect = document.getElementById('filter-environment');
    
    // Clear existing options except the first one
    while (envSelect.options.length > 1) {
        envSelect.options.remove(1);
    }
    
    // Add options
    environments.forEach(env => {
        const option = document.createElement('option');
        option.value = env;
        option.textContent = env;
        envSelect.appendChild(option);
    });
    
    // Get unique regions
    const regions = [...new Set(reservations.map(r => r.region))];
    const regionSelect = document.getElementById('filter-region');
    
    // Clear existing options except the first one
    while (regionSelect.options.length > 1) {
        regionSelect.options.remove(1);
    }
    
    // Add options
    regions.forEach(region => {
        const option = document.createElement('option');
        option.value = region;
        option.textContent = region;
        regionSelect.appendChild(option);
    });
}

// Function to apply filters
function applyFilters() {
    const status = document.getElementById('filter-status').value;
    const environment = document.getElementById('filter-environment').value;
    const region = document.getElementById('filter-region').value;
    const search = document.getElementById('search').value.toLowerCase();
    
    // Filter the reservations
    const filteredReservations = allReservations.filter(reservation => {
        // Status filter
        if (status && reservation.status !== status) {
            return false;
        }
        
        // Environment filter
        if (environment && reservation.environment !== environment) {
            return false;
        }
        
        // Region filter
        if (region && reservation.region !== region) {
            return false;
        }
        
        // Search filter
        if (search) {
            const searchableText = `${reservation.serverName} ${reservation.unitCode} ${reservation.type} ${reservation.provider} ${reservation.region} ${reservation.environment} ${reservation.function}`.toLowerCase();
            return searchableText.includes(search);
        }
        
        return true;
    });
    
    // Update the table
    updateReservationsTable(filteredReservations);
}

// Function to update the reservations table
function updateReservationsTable(reservations) {
    const tableBody = document.getElementById('reservationsTable');
    tableBody.innerHTML = '';
    
    if (reservations.length === 0) {
        const row = document.createElement('tr');
        const cell = document.createElement('td');
        cell.colSpan = 7;
        cell.textContent = 'No reservations found';
        cell.className = 'text-center';
        row.appendChild(cell);
        tableBody.appendChild(row);
        return;
    }
    
    reservations.forEach(reservation => {
        const row = document.createElement('tr');
        
        // Server name cell
        const nameCell = document.createElement('td');
        nameCell.className = 'font-monospace';
        nameCell.textContent = reservation.serverName;
        row.appendChild(nameCell);
        
        // Status cell
        const statusCell = document.createElement('td');
        const statusBadge = document.createElement('span');
        statusBadge.classList.add('badge');
        
        if (reservation.status === 'reserved') {
            statusBadge.classList.add('bg-warning', 'text-dark');
            statusBadge.textContent = 'Reserved';
        } else {
            statusBadge.classList.add('bg-success');
            statusBadge.textContent = 'Committed';
        }
        
        statusCell.appendChild(statusBadge);
        row.appendChild(statusCell);
        
        // UnitCode cell
        const unitCodeCell = document.createElement('td');
        unitCodeCell.textContent = reservation.unitCode;
        row.appendChild(unitCodeCell);
        
        // Type cell
        const typeCell = document.createElement('td');
        typeCell.textContent = reservation.type;
        row.appendChild(typeCell);
        
        // Environment cell
        const envCell = document.createElement('td');
        envCell.textContent = reservation.environment;
        row.appendChild(envCell);
        
        // Created at cell
        const createdAtCell = document.createElement('td');
        const date = new Date(reservation.createdAt);
        createdAtCell.textContent = date.toLocaleString();
        row.appendChild(createdAtCell);
        
        // Actions cell
        const actionsCell = document.createElement('td');
        
        if (reservation.status === 'reserved') {
            // Commit button
            const commitBtn = document.createElement('button');
            commitBtn.classList.add('btn', 'btn-sm', 'btn-outline-success', 'commit-btn', 'me-2');
            commitBtn.innerHTML = '<i class="bi bi-check-circle"></i>';
            commitBtn.title = 'Commit';
            commitBtn.dataset.id = reservation.id;
            actionsCell.appendChild(commitBtn);
        }
        
        // Delete button (now for both reserved and committed)
        const deleteBtn = document.createElement('button');
        deleteBtn.classList.add('btn', 'btn-sm', 'btn-outline-danger', 'delete-btn');
        deleteBtn.innerHTML = '<i class="bi bi-trash"></i>';
        deleteBtn.title = 'Delete';
        deleteBtn.dataset.id = reservation.id;
        deleteBtn.dataset.name = reservation.serverName;
        actionsCell.appendChild(deleteBtn);
        
        row.appendChild(actionsCell);
        tableBody.appendChild(row);
    });
}

// Function to commit a reservation
function commitReservation(reservationId) {
    fetch('/api/commit', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ reservationId })
    })
    .then(response => {
        if (!response.ok) {
            return response.json().then(err => { throw new Error(err.message || 'Failed to commit reservation'); });
        }
        return response.json();
    })
    .then(data => {
        showAlert('Reservation committed successfully', 'success');
        document.getElementById('reservationResult').classList.add('d-none');
        loadReservations(); // Refresh the reservations list
        loadDashboardData(); // Refresh dashboard data
    })
    .catch(error => {
        showAlert(error.message, 'danger');
    });
}

// Function to delete a reservation
function deleteReservation(reservationId) {
    fetch(`/api/reservations/${reservationId}`, {
        method: 'DELETE'
    })
    .then(response => {
        if (!response.ok) {
            return response.json().then(err => { throw new Error(err.message || 'Failed to delete reservation'); });
        }
        return response.json();
    })
    .then(data => {
        showAlert('Reservation deleted successfully', 'success');
        loadReservations(); // Refresh the reservations list
        loadDashboardData(); // Refresh dashboard data
    })
    .catch(error => {
        showAlert(error.message, 'danger');
    });
}

// Helper function to toggle visibility of elements
function toggleElement(elementId) {
    const element = document.getElementById(elementId);
    if (element.classList.contains('d-none')) {
        element.classList.remove('d-none');
    } else {
        element.classList.add('d-none');
    }
}

// Function to show alerts
function showAlert(message, type) {
    const alertDiv = document.createElement('div');
    alertDiv.className = `alert alert-${type} alert-dismissible fade show`;
    alertDiv.role = 'alert';
    alertDiv.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
    `;
    
    document.getElementById('alertContainer').appendChild(alertDiv);
    
    // Auto dismiss after 5 seconds
    setTimeout(() => {
        const bsAlert = new bootstrap.Alert(alertDiv);
        bsAlert.close();
    }, 5000);
}