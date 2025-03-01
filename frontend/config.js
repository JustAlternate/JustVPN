// Frontend configuration file
const config = {
		// Replace with production API URL
		apiBaseUrl: 'https://vpn.justalternate.fr/api',
      
    // External services
    ipifyUrl: 'https://api.ipify.org/?format=json'
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
