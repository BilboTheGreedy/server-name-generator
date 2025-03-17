/**
 * Release service to handle reservation release operations
 */
function releaseReservation(reservationId) {
    console.log('Releasing reservation:', reservationId);
    
    fetch('/api/release', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ reservationId })
    })
    .then(response => {
        console.log('Release response status:', response.status);
        
        if (!response.ok) {
            return response.json().then(err => { 
                throw new Error(err.message || 'Failed to release reservation'); 
            });
        }
        return response.json();
    })
    .then(data => {
        console.log('Release success:', data);
        showAlert('Reservation released successfully', 'success');
        loadReservations(); // Refresh the reservations list
        loadDashboardData(); // Refresh dashboard data
    })
    .catch(error => {
        console.error('Release error:', error);
        showAlert(error.message, 'danger');
    });
}