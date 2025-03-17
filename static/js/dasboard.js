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