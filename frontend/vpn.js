document.addEventListener('DOMContentLoaded', async () => {
    const token = localStorage.getItem('token');
    if (!token) {
        window.location.href = './login.html';
        return;
    }
    // Fetch user's public IP and populate the input field
    try {
        const ipResponse = await fetch('https://api.ipify.org/?format=json');
        if (ipResponse.ok) {
            const ipData = await ipResponse.json();
            document.getElementById('ip_address').value = ipData.ip;
        } else {
            console.error('Failed to fetch public IP address:', ipResponse.statusText);
        }
    } catch (error) {
        console.error('Error fetching public IP address:', error);
    }
});

let currentSessionId = null;
let socket = null;

function clearLogs() {
    const logMessages = document.getElementById('log-messages');
    const responseBox = document.getElementById('response-box');
    
    logMessages.innerHTML = '<h3>Live Logs</h3>';
    responseBox.innerHTML = '<span class="spinner" id="spinner" style="display: none;"></span>';
}

function setupWebSocket(sessionId) {
    if (socket) {
        socket.close();
    }

    const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
    const wsUrl = `${protocol}vpn.justalternate.fr/api/ws?session=${sessionId}`;
    socket = new WebSocket(wsUrl);

    socket.onopen = () => {
        console.log('WebSocket connection established');
    };

    socket.onmessage = (event) => {
        const logMessages = document.getElementById('log-messages');
        const logMessage = document.createElement('div');
        logMessage.textContent = event.data;
        logMessages.appendChild(logMessage);
        logMessages.scrollTop = logMessages.scrollHeight;
    };

    socket.onerror = (error) => {
        console.error('WebSocket error:', error);
    };

    socket.onclose = () => {
        console.log('WebSocket connection closed');
    };

    return socket;
}

async function initSession() {
    try {
        const response = await fetch('https://vpn.justalternate.fr/api/init', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        const data = await response.json();
        currentSessionId = data.sessionID;

        // Setup WebSocket connection
        setupWebSocket(currentSessionId);

        return currentSessionId;
    } catch (error) {
        console.error('Error initializing session:', error);
        throw error;
    }
}

document.getElementById('apiForm').addEventListener('submit', async function(event) {
    event.preventDefault();

    const token = localStorage.getItem('token');
    if (!token) {
        alert('You are not logged in.');
        return;
    }

    const spinner = document.getElementById('spinner');
    spinner.style.display = 'inline-block';
    clearLogs(); // Clear previous logs

    try {
        // Initialize session if not already done
        if (!currentSessionId) {
            await initSession();
        }

        const formData = new FormData(event.target);
        formData.append('sessionID', currentSessionId);
        const data = new URLSearchParams(formData);

        const response = await fetch('https://vpn.justalternate.fr/api/start', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: data.toString()
        });

        if (!response.ok) {
            const error = await response.json();
            if (error.error && error.error.includes('Token expired')) {
                localStorage.removeItem('token');
                window.location.href = './login.html';
                return;
            }
            throw new Error('Failed to start VPN process');
        }

        const result = await response.json();
        const responseBox = document.getElementById('response-box');
        responseBox.innerHTML += `
            <div class="response-item">
                <p><strong>Host Endpoint:</strong></p>
                <p>${result.host_endpoint}</p>
                <p><strong>Public Key:</strong></p>
                <p>${result.public_key}</p>
                <button id="downloadBtn" class="download-button">Download wireguard.conf</button>
            </div>
        `;

        // Add download configuration functionality
        document.getElementById('downloadBtn').addEventListener('click', () => {
            const content = `[Interface]
Address = 10.0.1.2/24
DNS = 1.1.1.1
ListenPort = 51820
PrivateKey = wAO6Deuy2gllo4H8IYp0Twra7MmmJYHPaYaWTj9irXE=

[Peer]
AllowedIPs = 0.0.0.0/0
Endpoint = ${result.host_endpoint}:51820
PublicKey = ${result.public_key}
`;
            const blob = new Blob([content], { type: 'text/plain' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'wireguard.conf';
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
        });
    } catch (error) {
        console.error('Error:', error);
        const responseBox = document.getElementById('response-box');
        responseBox.innerHTML += `
            <div class="error-message">
                Error: ${error.message}
            </div>
        `;
    } finally {
        spinner.style.display = 'none';
    }
});
