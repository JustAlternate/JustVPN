# JustVPN

JustVPN is a self-hosted VPN solution that allows you to provision a Wireguard VPN server on a Linode Nanode instance with just a few clicks. It uses Terraform for infrastructure provisioning and provides a simple web interface for managing your VPN connections.

## Features

- One-click deployment of Wireguard VPN server on Linode
- User authentication system
- Web-based interface for managing VPN connections
- Automatic configuration file generation for Wireguard clients
- Live logs of the provisioning process
- Configurable connection duration with automatic cleanup
- Fully configurable through environment variables

## Preview

![](./assets/login.png)
![](./assets/final.png)

## Prerequisites

- [Go](https://golang.org/doc/install) (1.18 or later)
- [Terraform](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli) (1.9.x or later)
- [Node.js](https://nodejs.org/) (for frontend development, optional)
- A [Linode](https://www.linode.com/) account with API access

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/JustVPN.git
cd JustVPN
```

### 2. Configure Environment Variables

Copy the example environment file and edit it with your settings:

```bash
cp .env.example .env
```

Edit the `.env` file with your preferred text editor:

```
SSH_PASSWORD=your-ssh-password-for-vpn-servers
JWT_SECRET=your-secure-jwt-secret-for-authentication
API_PORT=8081
FRONTEND_PORT=3000
USERS_FILE_PATH=./src/users.json
TERRAFORM_WORKING_DIR=./src
IAC_DIR_PATH=../iac
API_BASE_URL=http://localhost:8081
FRONTEND_CORS_ORIGIN=*
```

### 3. Configure Terraform Variables

Copy the example Terraform variables file and edit it with your Linode API credentials:

```bash
cp iac/secrets.tfvars.example iac/secrets.tfvars
```

Edit the `iac/secrets.tfvars` file:

```
linode_token = "your-linode-api-token"
root_pass = "secure-root-password-for-deployed-servers"
```

### 4. Set Up User Authentication

Create a hashed password for your user(s):

```bash
# Install bcrypt tool if you don't have it
go install github.com/unixpickle/gobcrypt/bcrypt@latest

# Generate a hashed password
echo -n "your_password_here" | bcrypt
```

Edit the `src/users.json` file to include your username and the hashed password:

```json
{
  "users": [
    {
      "username": "YourUsername",
      "password": "hashed-password-from-bcrypt"
    }
  ]
}
```

### 5. Install Dependencies

```bash
go mod tidy
```

### 6. Initialize Terraform

```bash
cd src
terraform init
cd ..
```

## Running JustVPN

### Start the Backend Server

```bash
go run src/main.go
```

The server will start on port 8081 (or the port specified in your `.env` file).

### Access the Web Interface

Open your browser and navigate to:

```
http://localhost:8081
```

If you want to serve the frontend from a different location, you can copy the `frontend` directory to your web server.

### Using Docker

You can also run JustVPN using Docker:

```bash
# Build the Docker image
docker build -t justvpn:latest .

# Build and run the containers
docker-compose up -d
```

The frontend will be available at `http://localhost:3000` (or the port specified in your `.env` file).

## Configuration Options

JustVPN is highly configurable through environment variables:

| Variable | Description | Default Value |
|----------|-------------|---------------|
| SSH_PASSWORD | SSH password for VPN servers | *Required* |
| JWT_SECRET | Secret for JWT authentication | *Required* |
| API_PORT | Backend server port | 8081 |
| FRONTEND_PORT | Frontend server port (Docker only) | 3000 |
| USERS_FILE_PATH | Path to users JSON file | ./src/users.json |
| TERRAFORM_WORKING_DIR | Working directory for Terraform | ./src |
| IAC_DIR_PATH | Path to infrastructure as code files | ../iac |
| API_BASE_URL | Base URL for API endpoints | http://localhost:8081 |
| FRONTEND_CORS_ORIGIN | CORS origin configuration | * |

## Usage

1. **Log in** with your username and password
2. **Configure your VPN connection**:
   - Your public IP will be automatically detected
   - Select a region for your VPN server
   - Choose the connection duration
3. **Click "Create Secure Connection"** to provision your VPN server
4. **Download the Wireguard configuration file** when the server is ready
5. **Import the configuration** into your Wireguard client

## API Endpoints

- `POST /login`: Authenticate and get a JWT token
- `POST /init`: Initialize a session for WebSocket communication
- `POST /start`: Provision a VPN server (requires authentication)
- `GET /health`: Check if the API is running
- `GET /ws`: WebSocket endpoint for live logs

## Frontend Development

The frontend is built with vanilla JavaScript and can be customized as needed:

- All API URLs are configurable through the `config.js` file
- The API base URL is injected at build time from environment variables
- WebSocket connections are automatically configured based on the current protocol

To modify the frontend:

1. Edit the files in the `frontend` directory
2. If using Docker, rebuild the containers with `docker-compose up -d --build`

## Tips

- Linode offers nearly unlimited free data transfer on their Nanode instances, making it an ideal choice for VPN hosting.
- For security reasons, VPN servers are automatically destroyed after the specified duration.
- You can modify the `iac/variables.tf` file to customize the VPN server configuration.

## Troubleshooting

- If you encounter issues with authentication, check your JWT_SECRET in the .env file
- Make sure your Linode API token has the necessary permissions
- Check the logs for any error messages during the provisioning process
- For CORS issues, adjust the FRONTEND_CORS_ORIGIN environment variable

## License

[MIT License](LICENSE)
