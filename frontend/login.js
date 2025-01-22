document.getElementById('loginForm').addEventListener('submit', async function (e) {
    e.preventDefault(); // Prevent the form from submitting the traditional way

    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    try {
        // Send the login request to the API
        const response = await fetch('https://vpn.justalternate.fr/api/login', {
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
            document.getElementById('message').textContent = 'Login successful!';
            window.location.href = './vpn.html';
        } else {
            document.getElementById('message').textContent = data.error || 'Login failed';
        }
    } catch (error) {
        document.getElementById('message').textContent = 'An error occurred. Please try again.';
        console.error('Error:', error);
    }
});
