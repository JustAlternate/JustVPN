document.addEventListener('DOMContentLoaded', async () => {
    const token = localStorage.getItem('token');
    if (!token) {
        window.location.href = './login.html';
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

// Establish a WebSocket connection to the server for live logs
(function() {
    var protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
    var wsUrl = protocol + window.location.host + '/api/ws';
    var socket = new WebSocket(wsUrl);

    socket.onopen = function() {
        console.log('WebSocket connection established');
    };

    socket.onmessage = function(event) {
        var logContainer = document.getElementById('log-messages');
        if (!logContainer) return;
        var messageElem = document.createElement('p');
        messageElem.textContent = event.data;
        logContainer.appendChild(messageElem);
    };

    socket.onerror = function(error) {
        console.error('WebSocket Error:', error);
    };

    socket.onclose = function() {
        console.log('WebSocket connection closed');
    };
})();

document.getElementById('apiForm').addEventListener('submit', async function(event) {
    event.preventDefault();

    const token = localStorage.getItem('token');
    if (!token) {
        alert('You are not logged in.');
        return;
    }

    const formData = new FormData(event.target);
    const data = new URLSearchParams(formData);

    const spinner = document.getElementById('spinner');
    const responseBox = document.getElementById('response-box');
    spinner.style.display = 'inline-block';

    try {
        const response = await fetch('https://vpn.justalternate.fr/api/start', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: data.toString()
        });

        if (response.ok) {
            const result = await response.json();
            responseBox.innerHTML = `
                <div class="response-item">
                    <p><strong>Host Endpoint:</strong></p>
                    <p>${result.host_endpoint}</p>
                </div>
                <div class="response-item">
                    <p><strong>Public Key:</strong></p>
                    <p>${result.public_key}</p>
                </div>
                <button id="downloadBtn">Download wireguard.conf</button>
            `;

            document.getElementById('downloadBtn').addEventListener('click', () => {
                const content = `
                    [Interface]
                    Address = 10.0.1.2/24
                    DNS = 1.1.1.1
                    ListenPort = 51820
                    PrivateKey = wAO6Deuy2gllo4H8IYp0Twra7MmmJYHPaYaWTj9irXE=
                    [Peer]
                    AllowedIPs = 0.0.0.0/0
                    Endpoint = ${result.host_endpoint}:51820
                    PublicKey = ${result.public_key}
                `;
                const blob = new Blob([content], { type: 'application/octet-stream' });
                const url = URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = 'wireguard.conf';
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                URL.revokeObjectURL(url);
            });
        } else {
            const errorData = await response.json();
            if (errorData.error === "Unauthorized: Token expired") {
                localStorage.removeItem('token');
                window.location.href = '/login.html';
            } else {
                console.error('Failed to fetch protected data:', errorData.error);
            }
        }
    } catch (error) {
        responseBox.innerHTML = `<p style="color: red;">Error: ${error.message}</p>`;
    } finally {
        spinner.style.display = 'none';
    }
});
