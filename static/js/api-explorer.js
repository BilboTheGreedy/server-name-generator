/**
 * API Explorer functionality
 */
document.addEventListener('templatesLoaded', function() {
    // Try out buttons
    document.getElementById('tryReserve').addEventListener('click', function() {
        toggleElement('reserveTestCard');
    });
    
    document.getElementById('tryCommit').addEventListener('click', function() {
        toggleElement('commitTestCard');
    });
    
    document.getElementById('tryGetReservations').addEventListener('click', function() {
        document.getElementById('getReservationsTestResult').classList.remove('d-none');
        fetch('/api/reservations')
            .then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP error ${response.status}`);
                }
                return response.json();
            })
            .then(data => {
                document.getElementById('getReservationsTestResponse').textContent = 
                    JSON.stringify(data, null, 2);
            })
            .catch(error => {
                document.getElementById('getReservationsTestResponse').textContent = 
                    `Error: ${error.message}`;
            });
    });
    
    document.getElementById('tryDeleteReservation').addEventListener('click', function() {
        toggleElement('deleteReservationTestCard');
    });
    
    document.getElementById('tryStats').addEventListener('click', function() {
        document.getElementById('statsTestResult').classList.remove('d-none');
        fetch('/api/stats')
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            document.getElementById('statsTestResponse').textContent = 
                JSON.stringify(data, null, 2);
        })
        .catch(error => {
            document.getElementById('statsTestResponse').textContent = 
                `Error: ${error.message}`;
        });
    });
    
    // Form submissions
    document.getElementById('reserveTestForm').addEventListener('submit', function(e) {
        e.preventDefault();
        const formData = new FormData(e.target);
        const payload = {};
        
        formData.forEach((value, key) => {
            if (value) payload[key] = value;
        });
        
        fetch('/api/reserve', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
        })
        .then(response => {
            return response.json().then(data => ({
                status: response.status,
                ok: response.ok,
                data
            }));
        })
        .then(result => {
            document.getElementById('reserveTestResult').classList.remove('d-none');
            
            if (result.ok) {
                document.getElementById('reserveTestResponse').textContent = 
                    JSON.stringify(result.data, null, 2);
            } else {
                document.getElementById('reserveTestResponse').textContent = 
                    `Error ${result.status}: ${JSON.stringify(result.data, null, 2)}`;
            }
        })
        .catch(error => {
            document.getElementById('reserveTestResult').classList.remove('d-none');
            document.getElementById('reserveTestResponse').textContent = 
                `Error: ${error.message}`;
        });
    });
    
    document.getElementById('commitTestForm').addEventListener('submit', function(e) {
        e.preventDefault();
        const reservationId = e.target.elements.reservationId.value;
        
        fetch('/api/commit', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ reservationId })
        })
        .then(response => {
            return response.json().then(data => ({
                status: response.status,
                ok: response.ok,
                data
            }));
        })
        .then(result => {
            document.getElementById('commitTestResult').classList.remove('d-none');
            
            if (result.ok) {
                document.getElementById('commitTestResponse').textContent = 
                    JSON.stringify(result.data, null, 2);
            } else {
                document.getElementById('commitTestResponse').textContent = 
                    `Error ${result.status}: ${JSON.stringify(result.data, null, 2)}`;
            }
        })
        .catch(error => {
            document.getElementById('commitTestResult').classList.remove('d-none');
            document.getElementById('commitTestResponse').textContent = 
                `Error: ${error.message}`;
        });
    });
    
    document.getElementById('deleteReservationTestForm').addEventListener('submit', function(e) {
        e.preventDefault();
        const reservationId = e.target.elements.reservationId.value;
        
        fetch(`/api/reservations/${reservationId}`, {
            method: 'DELETE'
        })
        .then(response => {
            return response.json().then(data => ({
                status: response.status,
                ok: response.ok,
                data
            }));
        })
        .then(result => {
            document.getElementById('deleteReservationTestResult').classList.remove('d-none');
            
            if (result.ok) {
                document.getElementById('deleteReservationTestResponse').textContent = 
                    JSON.stringify(result.data, null, 2);
            } else {
                document.getElementById('deleteReservationTestResponse').textContent = 
                    `Error ${result.status}: ${JSON.stringify(result.data, null, 2)}`;
            }
        })
        .catch(error => {
            document.getElementById('deleteReservationTestResult').classList.remove('d-none');
            document.getElementById('deleteReservationTestResponse').textContent = 
                `Error: ${error.message}`;
        });
    });
});