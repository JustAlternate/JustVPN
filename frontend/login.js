document.getElementById('loginForm').addEventListener('submit', async function (e) {
    e.preventDefault(); // Prevent the form from submitting the traditional way

    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const messageElement = document.getElementById('message');
    
    // Clear previous messages
    messageElement.textContent = '';
    messageElement.className = '';
    
    // Disable form inputs and show loading state
    const submitButton = this.querySelector('button[type="submit"]');
    const originalButtonText = submitButton.innerHTML;
    submitButton.innerHTML = `
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" class="animate-spin">
            <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" stroke-opacity="0.25"/>
            <path d="M12 2C6.47715 2 2 6.47715 2 12C2 14.7255 3.1 17.1962 4.8 19" stroke="currentColor" stroke-width="4" stroke-linecap="round"/>
        </svg>
        Signing in...
    `;
    submitButton.disabled = true;
    
    // Disable form inputs
    document.getElementById('username').disabled = true;
    document.getElementById('password').disabled = true;

    try {
        // Send the login request to the API
        const response = await fetch(getApiUrl('login'), {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, password }),
        });

        const data = await response.json();

        if (response.ok) {
            // Store the JWT token in localStorage
            localStorage.setItem('token', data.token);
            
            // Show success message
            messageElement.textContent = 'Login successful! Redirecting...';
            messageElement.className = 'success';
            
            // Redirect after a short delay
            setTimeout(() => {
                window.location.href = './index.html';
            }, 1000);
        } else {
            // Show error message
            messageElement.textContent = data.error || 'Login failed. Please check your credentials.';
            messageElement.className = 'error';
            
            // Reset form state
            submitButton.innerHTML = originalButtonText;
            submitButton.disabled = false;
            document.getElementById('username').disabled = false;
            document.getElementById('password').disabled = false;
        }
    } catch (error) {
        // Show error message for network/server issues
        messageElement.textContent = 'Connection error. Please try again later.';
        messageElement.className = 'error';
        console.error('Error:', error);
        
        // Reset form state
        submitButton.innerHTML = originalButtonText;
        submitButton.disabled = false;
        document.getElementById('username').disabled = false;
        document.getElementById('password').disabled = false;
    }
});

// Add keypress event to password field to submit on Enter
document.getElementById('password').addEventListener('keypress', function(e) {
    if (e.key === 'Enter') {
        e.preventDefault();
        document.getElementById('loginForm').dispatchEvent(new Event('submit'));
    }
});

// Add animation for the SVG in the button
document.head.insertAdjacentHTML('beforeend', `
    <style>
        .animate-spin {
            animation: spin 1s linear infinite;
        }
        @keyframes spin {
            from {
                transform: rotate(0deg);
            }
            to {
                transform: rotate(360deg);
            }
        }
    </style>
`);
