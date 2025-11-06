#!/bin/bash

echo "======================================"
echo "Code Linter Backend - EC2 Setup Script"
echo "======================================"
echo ""

# Update system
echo "📦 Updating system packages..."
sudo apt update && sudo apt upgrade -y

# Install Go
echo "🔧 Installing Go..."
wget https:
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
rm go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
export PATH=$PATH:/usr/local/go/bin

# Install Node.js
echo "🔧 Installing Node.js..."
curl -fsSL https:
sudo apt install -y nodejs

# Install Git
echo "🔧 Installing Git..."
sudo apt install -y git

# Create app directory
echo "📁 Creating application directory..."
sudo mkdir -p /opt/code-linter
sudo chown $USER:$USER /opt/code-linter

echo ""
echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "1. Clone your repository to /opt/code-linter/"
echo "2. Copy .env file to backend directory"
echo "3. Build backend: cd backend && go build -o code-linter-server main.go"
echo "4. Setup systemd service (see AWS_MUMBAI_DEPLOYMENT.md)"
echo ""
echo "Repository structure should be:"
echo "/opt/code-linter/codecollab/backend/"
echo "/opt/code-linter/codecollab/lambda/"
echo ""
