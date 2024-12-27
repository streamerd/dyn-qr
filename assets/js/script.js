const ws = new WebSocket(`ws://${window.location.host}/ws`);
const qrCodeImg = document.getElementById('qrCode');
const stopNumberEl = document.getElementById('stopNumber');
const busTableBody = document.getElementById('busTableBody');
const lastUpdateEl = document.getElementById('lastUpdate');

function updateBusTable(data) {
    const parsedData = JSON.parse(data);
    
    // Update stop number
    stopNumberEl.textContent = parsedData.s;
    
    // Clear existing table rows
    busTableBody.innerHTML = '';
    
    // Add new rows for each bus
    parsedData.b.forEach(bus => {
        const row = document.createElement('tr');
        const arrivalText = bus.m === 0 ? 'Almost there!' : `${bus.m} min${bus.m === 1 ? '' : 's'}`;
        row.innerHTML = `
            <td>Bus ${bus.l}</td>
            <td class="minutes">${arrivalText}</td>
        `;
        busTableBody.appendChild(row);
    });

    lastUpdateEl.textContent = `Last updated: ${new Date().toLocaleTimeString()}`;
}

function updateQRCode(newSrc) {
    const img = document.getElementById('qrCode');
    img.src = newSrc;
}

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    updateQRCode(`/qr/${data.id}`);
    updateBusTable(data.data);
};

ws.onerror = function(error) {
    console.error('WebSocket Error:', error);
};

ws.onclose = function() {
    console.log('WebSocket Connection Closed');
    setTimeout(() => {
        window.location.reload();
    }, 5000);
}; 