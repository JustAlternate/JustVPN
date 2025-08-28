// Demo mode functions to simulate API calls

const demo = {
    // Simulate fetching public IP
    async getPublicIP() {
        // Simulate network delay
        await new Promise(resolve => setTimeout(resolve, 800));
        
        // Return a fake IP
        return { ip: '192.168.1.100' };
    },
    
    // Simulate initializing a session
    async initSession() {
        // Simulate network delay
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        // Return a fake session ID
        return { sessionID: 'demo-session-' + Math.random().toString(36).substr(2, 9) };
    },
    
    // Simulate starting VPN connection
    async startConnection(region, duration) {
        // Simulate network delay
        await new Promise(resolve => setTimeout(resolve, 2000));
        
        // Return fake connection data
        const fakeIP = '192.168.1.100';
        
        const publicKey = 'bm90LXJlYWwta2V5LWZvci1kZW1vLXB1cnBvc2VzCg==';
        
        return {
            host_endpoint: fakeIP,
            public_key: publicKey
        };
    },
    
    // Generate fake logs
    generateLogs(callback) {
        const sessionId = 'demo-session-' + Math.random().toString(36).substr(2, 9);
        const region = localStorage.getItem('demoRegion') || 'eu-central';
        const duration = localStorage.getItem('demoDuration') || '60';
        const date = new Date().toISOString().split('T')[0];
        const time1 = new Date(Date.now() - 300000).toTimeString().substr(0,8); // 5 minutes ago
        const time2 = new Date(Date.now() - 240000).toTimeString().substr(0,8); // 4 minutes ago
        const time3 = new Date(Date.now() - 180000).toTimeString().substr(0,8); // 3 minutes ago
        const time4 = new Date(Date.now() - 120000).toTimeString().substr(0,8); // 2 minutes ago
        const time5 = new Date(Date.now() - 60000).toTimeString().substr(0,8);  // 1 minute ago
        const time6 = new Date().toTimeString().substr(0,8);                   // now
        
        // Use a fake IP for the logs
        const fakeIP = '192.168.1.100';
        
        const logs = [
            "Initializing VPN connection...",
            "Creating new session...",
            `Session created with ID: ${sessionId}`,
            `Connecting to region: ${region}`,
            `Connection will be active for: ${duration} seconds`,
            "WebSocket connected successfully",
            `${date} ${time1} Creating TerraformService and Init...`,
            `${date} ${time2} Parsing Response information...`,
            `${date} ${time2} Terraform Apply for ${fakeIP} ${region}...`,
            `${date} ${time3} Getting hostIp for ${fakeIP} ${region}...`,
            `${date} ${time3} Getting PubKey for ${fakeIP} ${region}...`,
            `${date} ${time4} Attempt 1: Error occurred when fetching pubkey, retrying in 10 seconds...`,
            `${date} ${time5} Attempt 2: Error occurred when fetching pubkey, retrying in 10 seconds...`,
            `${date} ${time6} Creating the response for ${fakeIP} ${region}...`,
            `${date} ${time6} Launching timer before destroy for ${fakeIP} ${region}...`,
            `${date} ${time6} Finished handling request`,
            "VPN connection established successfully!",
            "WireGuard configuration is ready for download"
        ];
        
        let index = 0;
        const interval = setInterval(() => {
            if (index < logs.length) {
                callback(logs[index]);
                index++;
            } else {
                clearInterval(interval);
            }
        }, 500);
        
        return interval;
    }
};

// Check if we're in demo mode
function isDemoMode() {
    return localStorage.getItem('demoMode') === 'true';
}

// Override fetch for demo mode
const originalFetch = window.fetch;

window.fetch = function(url, options = {}) {
    // Check if we're in demo mode and the request is to our API
    if (isDemoMode() && url.includes(config.apiBaseUrl)) {
        // Extract endpoint from URL
        const endpoint = url.replace(config.apiBaseUrl, '').replace(/^\//, '');
        
        // Handle different endpoints
        switch(endpoint) {
            case 'init':
                return Promise.resolve({
                    ok: true,
                    json: () => demo.initSession()
                });
                
            case 'start':
                return Promise.resolve({
                    ok: true,
                    json: () => {
                        // Get region and duration from form data in request body
                        const formData = new URLSearchParams(options.body);
                        const region = formData.get('region');
                        const duration = formData.get('timeWantedBeforeDeletion');
                        
                        // Store for log generation
                        localStorage.setItem('demoRegion', region);
                        localStorage.setItem('demoDuration', duration);
                        
                        return demo.startConnection(region, duration);
                    }
                });
                
            default:
                // For other endpoints, return a generic success response
                return Promise.resolve({
                    ok: true,
                    json: () => ({ message: 'Demo mode: Request successful' })
                });
        }
    }
    
    // For non-API requests or when not in demo mode, use original fetch
    return originalFetch.apply(this, arguments);
};