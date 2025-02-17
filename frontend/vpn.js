document.addEventListener('DOMContentLoaded', () => {
    const token = localStorage.getItem('token');
    if (!token) {
        window.location.href = './login.html';
    }
});

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
    responseBox.innerHTML = '<div class="logs"></div>';
    const logsDiv = responseBox.querySelector('.logs');

    try {
        const eventSource = new EventSource(`https://vpn.justalternate.fr/api/start?${data.toString()}`);

        eventSource.onmessage = function(event) {
            const data = JSON.parse(event.data);
            
            if (data.type === 'log') {
                // Create a new log entry
                const logEntry = document.createElement('div');
                logEntry.className = 'log-entry';
                logEntry.textContent = data.message;
                logsDiv.appendChild(logEntry);
                logEntry.scrollIntoView({ behavior: 'smooth' });
            } else if (data.type === 'result') {
                // Handle the final result
                eventSource.close();
                const result = data.data;
                responseBox.innerHTML = `
                    <div class="logs">${logsDiv.innerHTML}</div>
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
            }
        };

        eventSource.onerror = function(error) {
            console.error('EventSource failed:', error);
            eventSource.close();
            responseBox.innerHTML += `<p style="color: red;">Error: Connection failed</p>`;
            spinner.style.display = 'none';
        };
    } catch (error) {
        responseBox.innerHTML = `<p style="color: red;">Error: ${error.message}</p>`;
        spinner.style.display = 'none';
    } finally {
        spinner.style.display = 'none';
    }
});
