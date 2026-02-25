# AWS Mumbai Server Deployment Guide

## Overview

Deploy your Code Linting Platform to AWS Mumbai (ap-south-1) region.

## Architecture

```
Users → Vercel (Frontend) → AWS Mumbai EC2 (Backend) → AWS Lambda (Linters)
                                     ↓
                               Supabase Auth
```

---

## Prerequisites

1. **AWS Account** with billing enabled
2. **AWS CLI** installed and configured
3. **Domain name** (optional, for HTTPS)
4. **SSH key pair** for EC2 access

---

## Part 1: EC2 Instance Setup (Mumbai Region)

### Step 1: Create EC2 Instance

**Via AWS Console:**

1. Go to AWS Console → EC2
2. **Region**: Select **Asia Pacific (Mumbai) ap-south-1**
3. Click "Launch Instance"

**Instance Configuration:**
- **Name**: `code-linter-backend`
- **AMI**: Ubuntu Server 22.04 LTS
- **Instance Type**: `t2.micro` (free tier) or `t3.small` (recommended)
- **Key Pair**: Create new or select existing
- **Network Settings**:
  - Auto-assign public IP: **Enable**
  - Security Group: Create new
    - **Rule 1**: SSH (port 22) - Your IP
    - **Rule 2**: Custom TCP (port 8080) - Anywhere (0.0.0.0/0)
    - **Rule 3**: HTTPS (port 443) - Anywhere (optional)
- **Storage**: 20 GB gp3

4. Click "Launch Instance"

**Via AWS CLI:**

```bash
# Configure AWS CLI for Mumbai region
aws configure set region ap-south-1

# Create security group
aws ec2 create-security-group \
  --group-name code-linter-sg \
  --description "Security group for code linter backend"

# Add SSH rule (replace YOUR_IP with your IP)
aws ec2 authorize-security-group-ingress \
  --group-name code-linter-sg \
  --protocol tcp \
  --port 22 \
  --cidr YOUR_IP/32

# Add WebSocket rule
aws ec2 authorize-security-group-ingress \
  --group-name code-linter-sg \
  --protocol tcp \
  --port 8080 \
  --cidr 0.0.0.0/0

# Launch instance
aws ec2 run-instances \
  --image-id ami-0f5ee92e2d63afc18 \
  --instance-type t3.small \
  --key-name YOUR_KEY_NAME \
  --security-groups code-linter-sg \
  --region ap-south-1
```

### Step 2: Connect to EC2

```bash
# Get instance public IP from AWS console
ssh -i your-key.pem ubuntu@YOUR_INSTANCE_IP
```

---

## Part 2: Server Setup

### Step 1: Install Dependencies

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Go
wget https:
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version

# Install Node.js (for Lambda functions)
curl -fsSL https:
sudo apt install -y nodejs
node --version

# Install Git
sudo apt install -y git
```

### Step 2: Clone and Setup Backend

```bash
# Create app directory
sudo mkdir -p /opt/code-linter
sudo chown ubuntu:ubuntu /opt/code-linter
cd /opt/code-linter

# Clone your repository (replace with your repo)
git clone https:
cd codecollab/backend

# Install Go dependencies
go mod download

# Create production .env file
nano .env
```

**Production .env file:**
```bash
SUPABASE_URL=https:
SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Im1ueWd0YW9laXBjaHppdnNlaW5nIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NTY4ODAxNDUsImV4cCI6MjA3MjQ1NjE0NX0.P-PjlTerqqvPR7hnPJ-CqJci_ulqq1yQSxDH4ZFSL0w

AWS_REGION=ap-south-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key

LAMBDA_ARN_TYPESCRIPT=arn:aws:lambda:ap-south-1:YOUR_ACCOUNT_ID:function:typescript-linter
LAMBDA_ARN_PYTHON=arn:aws:lambda:ap-south-1:YOUR_ACCOUNT_ID:function:python-linter
LAMBDA_ARN_DART=arn:aws:lambda:ap-south-1:YOUR_ACCOUNT_ID:function:dart-linter
LAMBDA_ARN_GO=arn:aws:lambda:ap-south-1:YOUR_ACCOUNT_ID:function:go-linter
LAMBDA_ARN_CPP=arn:aws:lambda:ap-south-1:YOUR_ACCOUNT_ID:function:cpp-linter

PORT=8080
ENV=production
USE_MOCK_LAMBDA=false
USE_MOCK_AUTH=false
```

### Step 3: Build Backend

```bash
# Build the Go binary
go build -o code-linter-server main.go

# Test run
./code-linter-server
# Press Ctrl+C to stop
```

---

## Part 3: Deploy Lambda Functions (Mumbai)

### Step 1: Create IAM Role for Lambda

```bash
# Create trust policy
cat > trust-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

# Create role
aws iam create-role \
  --role-name code-linter-lambda-role \
  --assume-role-policy-document file:

# Attach basic execution policy
aws iam attach-role-policy \
  --role-name code-linter-lambda-role \
  --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
```

### Step 2: Deploy TypeScript Linter Lambda

```bash
# On your local machine
cd lambda/typescript-linter

# Install dependencies
npm install --production

# Create deployment package
zip -r typescript-linter.zip index.js package.json node_modules/

# Deploy to Mumbai region
aws lambda create-function \
  --function-name typescript-linter \
  --runtime nodejs18.x \
  --role arn:aws:iam::YOUR_ACCOUNT_ID:role/code-linter-lambda-role \
  --handler index.handler \
  --zip-file fileb:
  --timeout 10 \
  --memory-size 256 \
  --region ap-south-1

# Test the function
aws lambda invoke \
  --function-name typescript-linter \
  --region ap-south-1 \
  --payload '{"language":"typescript","code":"const x: number = \"hello\";"}' \
  response.json

cat response.json
```

### Step 3: Get Lambda ARN

```bash
# Get the Lambda ARN
aws lambda get-function \
  --function-name typescript-linter \
  --region ap-south-1 \
  --query 'Configuration.FunctionArn'

# Update your backend .env with this ARN
```

---

## Part 4: Setup SystemD Service

Create a systemd service to run the backend automatically:

```bash
# Create service file
sudo nano /etc/systemd/system/code-linter.service
```

**Service file content:**
```ini
[Unit]
Description=Code Linter Backend Service
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/opt/code-linter/codecollab/backend
ExecStart=/opt/code-linter/codecollab/backend/code-linter-server
Restart=always
RestartSec=10
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=code-linter

[Install]
WantedBy=multi-user.target
```

**Enable and start service:**
```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service (start on boot)
sudo systemctl enable code-linter

# Start service
sudo systemctl start code-linter

# Check status
sudo systemctl status code-linter

# View logs
sudo journalctl -u code-linter -f
```

---

## Part 5: Setup Nginx Reverse Proxy (Optional but Recommended)

For HTTPS and better security:

```bash
# Install Nginx
sudo apt install -y nginx

# Create Nginx config
sudo nano /etc/nginx/sites-available/code-linter
```

**Nginx config:**
```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http:
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Enable site:**
```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/code-linter /etc/nginx/sites-enabled/

# Test config
sudo nginx -t

# Restart Nginx
sudo systemctl restart nginx
```

**Setup SSL with Let's Encrypt:**
```bash
# Install Certbot
sudo apt install -y certbot python3-certbot-nginx

# Get SSL certificate
sudo certbot --nginx -d your-domain.com

# Auto-renewal is set up automatically
```

---

## Part 6: Frontend Deployment (Vercel)

### Step 1: Update Frontend Environment

Update `frontend/.env.local`:
```bash
NEXT_PUBLIC_SUPABASE_URL=https:
NEXT_PUBLIC_SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Im1ueWd0YW9laXBjaHppdnNlaW5nIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NTY4ODAxNDUsImV4cCI6MjA3MjQ1NjE0NX0.P-PjlTerqqvPR7hnPJ-CqJci_ulqq1yQSxDH4ZFSL0w
NEXT_PUBLIC_WS_URL=ws:
# Or with domain: wss:
```

### Step 2: Deploy to Vercel

```bash
# Install Vercel CLI
npm install -g vercel

# Login
vercel login

# Deploy from frontend directory
cd frontend
vercel

# Set environment variables in Vercel dashboard
# Project Settings → Environment Variables
# Add all NEXT_PUBLIC_* variables
```

---

## Part 7: Security Configuration

### 1. EC2 Security Group Rules

**Minimum required:**
- SSH (22) - Your IP only
- WebSocket (8080) - Anywhere (or Vercel IPs)
- HTTPS (443) - Anywhere (if using Nginx)

### 2. Supabase Configuration

In Supabase dashboard:
- Add your EC2 IP to allowed origins
- Add your Vercel domain to allowed origins

### 3. Environment Variables

**NEVER commit:**
- `.env` files
- AWS credentials
- Supabase keys

---

## Part 8: Monitoring and Logs

### View Backend Logs
```bash
# Real-time logs
sudo journalctl -u code-linter -f

# Recent logs
sudo journalctl -u code-linter -n 100

# Logs from specific time
sudo journalctl -u code-linter --since "1 hour ago"
```

### Monitor Lambda
```bash
# View Lambda logs
aws logs tail /aws/lambda/typescript-linter --follow --region ap-south-1
```

### CloudWatch Metrics
- Go to AWS CloudWatch in Mumbai region
- Monitor Lambda invocations, errors, duration
- Set up alarms for errors

---

## Part 9: Updating the Application

### Update Backend

```bash
# SSH to EC2
ssh -i your-key.pem ubuntu@YOUR_INSTANCE_IP

# Navigate to directory
cd /opt/code-linter/codecollab

# Pull latest changes
git pull

# Rebuild
cd backend
go build -o code-linter-server main.go

# Restart service
sudo systemctl restart code-linter
```

### Update Lambda

```bash
# On local machine
cd lambda/typescript-linter
npm install --production
zip -r typescript-linter.zip index.js package.json node_modules/

# Update function
aws lambda update-function-code \
  --function-name typescript-linter \
  --region ap-south-1 \
  --zip-file fileb:
```

---

## Cost Estimation (Mumbai Region)

**Monthly costs:**
- EC2 t3.small: ~$15/month
- Lambda (1M requests): ~$0.20
- Data transfer: ~$5/month
- **Total: ~$20-25/month**

**Free Tier Eligible:**
- EC2 t2.micro: 750 hours/month free (first 12 months)
- Lambda: 1M requests/month free (always)

---

## Quick Commands Reference

```bash
# Check backend status
sudo systemctl status code-linter

# Restart backend
sudo systemctl restart code-linter

# View logs
sudo journalctl -u code-linter -f

# Check if port is open
sudo netstat -tulpn | grep 8080

# Test WebSocket
wscat -c ws:

# Update and restart
cd /opt/code-linter/codecollab && git pull && cd backend && go build -o code-linter-server main.go && sudo systemctl restart code-linter
```

---

## Troubleshooting

### Backend not starting?
```bash
# Check logs
sudo journalctl -u code-linter -n 50

# Check if port is in use
sudo lsof -i:8080

# Test manually
cd /opt/code-linter/codecollab/backend
./code-linter-server
```

### WebSocket connection failed?
- Check security group allows port 8080
- Verify EC2 public IP is correct
- Check if backend is running: `sudo systemctl status code-linter`

### Lambda timeout?
- Increase Lambda timeout: `aws lambda update-function-configuration --function-name typescript-linter --timeout 30 --region ap-south-1`

---

## Next Steps

1. ✅ Launch EC2 in Mumbai
2. ✅ Deploy backend
3. ✅ Deploy Lambda functions
4. ✅ Setup systemd service
5. ✅ Deploy frontend to Vercel
6. ✅ Test end-to-end
7. Optional: Setup domain and SSL
8. Optional: Setup CloudWatch monitoring
9. Optional: Setup auto-scaling

---

## Support

For issues:
- Check AWS CloudWatch logs
- Review systemd logs
- Test Lambda functions individually
- Verify Supabase auth is working

**Your Mumbai-based production deployment is ready!** 🚀🇮🇳
