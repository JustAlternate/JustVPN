// Frontend configuration file
const config = {
		// Replace with production API URL
		apiBaseUrl: 'https://vpn.justalternate.fr/api',
    
    // WebSocket protocol - automatically determined based on current protocol
    wsProtocol: window.location.protocol === 'https:' ? 'wss://' : 'ws://',
    
    // External services
    ipifyUrl: 'https://api.ipify.org/?format=json'
};

// Extract the hostname from the API base URL for WebSocket connections
const getApiHostname = () => {
    try {
        const url = new URL(config.apiBaseUrl);
        return url.host;
    } catch (error) {
        console.error('Invalid API base URL:', error);
        return window.location.host;
    }
};

// Get full API URL for a specific endpoint
const getApiUrl = (endpoint) => {
    // Remove leading slash if present
    const cleanEndpoint = endpoint.startsWith('/') ? endpoint.substring(1) : endpoint;
    
    // Ensure apiBaseUrl doesn't end with a slash
    const baseUrl = config.apiBaseUrl.endsWith('/') 
        ? config.apiBaseUrl.slice(0, -1) 
        : config.apiBaseUrl;
        
    return `${baseUrl}/${cleanEndpoint}`;
};

// Get WebSocket URL for a specific endpoint
const getWsUrl = (endpoint, params = {}) => {
    const queryParams = new URLSearchParams(params).toString();
    const queryString = queryParams ? `?${queryParams}` : '';
    
    // Remove leading slash if present
    const cleanEndpoint = endpoint.startsWith('/') ? endpoint.substring(1) : endpoint;
    
    return `${config.wsProtocol}${getApiHostname()}/${cleanEndpoint}${queryString}`;
};
