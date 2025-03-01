document.addEventListener('DOMContentLoaded', async () => {
    const token = localStorage.getItem('token');
    if (!token) {
        window.location.href = './login.html';
        return;
    }
    
    // Show loading state for IP address
    const ipAddressInput = document.getElementById('ip_address');
    ipAddressInput.disabled = true;
    
    // Fetch user's public IP and populate the input field
    try {
        const ipResponse = await fetch(config.ipifyUrl);
        if (ipResponse.ok) {
            const ipData = await ipResponse.json();
            ipAddressInput.value = ipData.ip;
        } else {
            console.error('Failed to fetch public IP address:', ipResponse.statusText);
            ipAddressInput.placeholder = 'Failed to detect IP';
        }
    } catch (error) {
        console.error('Error fetching public IP address:', error);
        ipAddressInput.placeholder = 'Failed to detect IP';
    } finally {
        ipAddressInput.disabled = false;
    }
    
    // Add status indicator to the response box
    const responseBox = document.getElementById('response-box');
    responseBox.innerHTML = `
        <div class="status-badge disconnected">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M6 6L18 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            Disconnected
        </div>
        <span class="spinner" id="spinner"></span>
    `;
});

let currentSessionId = null;
let socket = null;

function clearLogs() {
    const logMessages = document.getElementById('log-messages');
    logMessages.innerHTML = `
        <h3>Live Logs</h3>
    `;
}

function addLogEntry(message) {
    const logMessages = document.getElementById('log-messages');
    const logEntry = document.createElement('div');
    logEntry.className = 'log-entry';
    logEntry.textContent = message;
    logMessages.appendChild(logEntry);
    logMessages.scrollTop = logMessages.scrollHeight;
}

function setupWebSocket(sessionId) {
    if (socket) {
        socket.close();
    }

    const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
    const wsUrl = `${protocol}localhost:8081/ws?session=${sessionId}`;
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
        const response = await fetch(getApiUrl('init'), {
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
        showNotification('You are not logged in.', 'error');
        setTimeout(() => {
            window.location.href = './login.html';
        }, 2000);
        return;
    }

    // Update UI to show connecting state
    const spinner = document.getElementById('spinner');
    spinner.style.display = 'inline-block';
    
    const responseBox = document.getElementById('response-box');
    responseBox.innerHTML = `
        <div class="status-badge connecting">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 2V6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M12 18V22" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M4.93 4.93L7.76 7.76" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M16.24 16.24L19.07 19.07" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M2 12H6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M18 12H22" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M4.93 19.07L7.76 16.24" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M16.24 7.76L19.07 4.93" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            Connecting...
        </div>
        <span class="spinner" id="spinner"></span>
    `;
    
    clearLogs(); // Clear previous logs
    addLogEntry('Initializing VPN connection...');

    try {
        // Initialize session if not already done
        if (!currentSessionId) {
            addLogEntry('Creating new session...');
            await initSession();
            addLogEntry(`Session created with ID: ${currentSessionId}`);
        }

        const formData = new FormData(event.target);
        formData.append('sessionID', currentSessionId);
        const data = new URLSearchParams(formData);

        addLogEntry(`Connecting to region: ${formData.get('region')}`);
        addLogEntry(`Connection will be active for: ${formData.get('timeWantedBeforeDeletion')} seconds`);
        
        const response = await fetch(getApiUrl('start'), {
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
                showNotification('Your session has expired. Please log in again.', 'error');
                setTimeout(() => {
                    window.location.href = './login.html';
                }, 2000);
                return;
            }
            throw new Error(error.error || 'Failed to start VPN process');
        }

        const result = await response.json();
        
        // Update UI to show connected state
        responseBox.innerHTML = `
            <div class="status-badge connected">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M20 6L9 17L4 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                Connected
            </div>
            <div class="response-item">
                <p><strong>Host Endpoint</strong></p>
                <div class="response-data">${result.host_endpoint}</div>
                
                <p><strong>Public Key</strong></p>
                <div class="response-data">${result.public_key}</div>
                
                <button id="downloadBtn" class="download-button">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M21 15V19C21 19.5304 20.7893 20.0391 20.4142 20.4142C20.0391 20.7893 19.5304 21 19 21H5C4.46957 21 3.96086 20.7893 3.58579 20.4142C3.21071 20.0391 3 19.5304 3 19V15" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                        <path d="M7 10L12 15L17 10" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                        <path d="M12 15V3" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                    Download WireGuard Config
                </button>
            </div>
        `;

        addLogEntry('VPN connection established successfully!');
        addLogEntry('WireGuard configuration is ready for download');

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
            
            addLogEntry('WireGuard configuration downloaded');
        });
    } catch (error) {
        console.error('Error:', error);
        
        // Update UI to show error state
        responseBox.innerHTML = `
            <div class="status-badge disconnected">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                    <path d="M6 6L18 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                Connection Failed
            </div>
            <div class="error-message">
                ${error.message}
            </div>
        `;
        
        addLogEntry(`Error: ${error.message}`);
    } finally {
        spinner.style.display = 'none';
    }
});

function showNotification(message, type = 'info') {
    // Create notification element if it doesn't exist
    let notification = document.getElementById('notification');
    if (!notification) {
        notification = document.createElement('div');
        notification.id = 'notification';
        notification.style.position = 'fixed';
        notification.style.top = '20px';
        notification.style.right = '20px';
        notification.style.padding = '12px 20px';
        notification.style.borderRadius = 'var(--radius)';
        notification.style.backgroundColor = type === 'error' ? 'var(--error-color)' : 'var(--primary-color)';
        notification.style.color = 'white';
        notification.style.boxShadow = 'var(--shadow)';
        notification.style.zIndex = '1000';
        notification.style.transition = 'transform 0.3s ease, opacity 0.3s ease';
        notification.style.transform = 'translateY(-20px)';
        notification.style.opacity = '0';
        document.body.appendChild(notification);
    }
    
    // Set notification content and show it
    notification.textContent = message;
    notification.style.backgroundColor = type === 'error' ? 'var(--error-color)' : 'var(--primary-color)';
    
    // Animate in
    setTimeout(() => {
        notification.style.transform = 'translateY(0)';
        notification.style.opacity = '1';
    }, 10);
    
    // Automatically hide after 5 seconds
    setTimeout(() => {
        notification.style.transform = 'translateY(-20px)';
        notification.style.opacity = '0';
        
        // Remove from DOM after animation completes
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 300);
    }, 5000);
}
