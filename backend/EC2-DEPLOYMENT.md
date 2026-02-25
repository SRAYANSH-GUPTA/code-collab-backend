# EC2 Deployment Guide

## Prerequisites

- EC2 instance running (Ubuntu/Amazon Linux recommended)
- Docker and Docker Compose installed
- SSH access to your EC2 instance

## Step-by-Step Deployment

### 1. Configure AWS Security Group

**Go to AWS Console → EC2 → Security Groups → Select your instance's security group → Edit inbound rules**

Add these rules:

| Type | Protocol | Port Range | Source | Description |
|------|----------|------------|--------|-------------|
| Custom TCP | TCP | 8080 | 0.0.0.0/0 | Application API |
| Custom TCP | TCP | 3000 | 0.0.0.0/0 | Grafana Dashboard |
| Custom TCP | TCP | 9090 | 0.0.0.0/0 | Prometheus (Optional) |
| Custom TCP | TCP | 3100 | 0.0.0.0/0 | Loki (Optional) |
| SSH | TCP | 22 | Your IP | SSH Access |

**Security Notes:**
- For production, restrict source to your IP range instead of `0.0.0.0/0`
- Prometheus and Loki ports are optional - only needed for direct access
- Consider using an Application Load Balancer with SSL for production

### 2. SSH into Your EC2 Instance

```bash
ssh -i your-key.pem ec2-user@your-ec2-public-ip
```

Or for Ubuntu:
```bash
ssh -i your-key.pem ubuntu@your-ec2-public-ip
```

### 3. Install Docker and Docker Compose (if not installed)

**For Amazon Linux 2:**
```bash
sudo yum update -y
sudo yum install docker -y
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -a -G docker ec2-user

sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

newgrp docker
```

**For Ubuntu:**
```bash
sudo apt update
sudo apt install docker.io docker-compose -y
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker ubuntu

newgrp docker
```

### 4. Clone/Upload Your Code

**Option A - Using Git:**
```bash
git clone https://github.com/your-repo/code-collab-backend.git
cd code-collab-backend/backend
```

**Option B - Using SCP:**
```bash
scp -i your-key.pem -r /path/to/backend ec2-user@your-ec2-ip:~/
ssh -i your-key.pem ec2-user@your-ec2-ip
cd backend
```

### 5. Configure Environment Variables

```bash
cp .env.example .env
nano .env
```

Edit the `.env` file with your actual credentials:
```bash
PORT=8080
ENV=production
LOKI_URL=http://loki:3100

SUPABASE_URL=your-supabase-url
SUPABASE_ANON_KEY=your-supabase-key

AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key

LAMBDA_ARN_TYPESCRIPT=your-lambda-arn
LAMBDA_ARN_PYTHON=your-lambda-arn
LAMBDA_ARN_DART=your-lambda-arn
LAMBDA_ARN_GO=your-lambda-arn
LAMBDA_ARN_CPP=your-lambda-arn

USE_MOCK_LAMBDA=false
USE_MOCK_AUTH=false
```

**Save and exit:** Press `Ctrl+X`, then `Y`, then `Enter`

### 6. Update Grafana Root URL

Get your EC2 public IP:
```bash
EC2_IP=$(curl -s http://checkip.amazonaws.com)
echo "Your EC2 IP: $EC2_IP"
```

Update docker-compose.yml:
```bash
sed -i "s|GF_SERVER_ROOT_URL=.*|GF_SERVER_ROOT_URL=http://${EC2_IP}:3000|g" docker-compose.yml
```

### 7. Deploy with Docker Compose

```bash
docker-compose up -d --build
```

### 8. Verify Deployment

```bash
docker-compose ps

docker logs codecollab-backend

curl http://localhost:8080/health
```

### 9. Access Your Services

Get your EC2 public IP:
```bash
curl http://checkip.amazonaws.com
```

Then access from your browser:
- **Grafana**: `http://YOUR_EC2_IP:3000` (admin/admin123)
- **API**: `http://YOUR_EC2_IP:8080`
- **API Docs**: `http://YOUR_EC2_IP:8080/docs`
- **Prometheus**: `http://YOUR_EC2_IP:9090`
- **Metrics**: `http://YOUR_EC2_IP:8080/metrics`

## Quick Deploy Script

Or simply run the automated setup script:

```bash
chmod +x ec2-setup.sh
./ec2-setup.sh
```

## Troubleshooting

### Can't access Grafana

1. **Check security group:**
   ```bash
   curl -v http://localhost:3000
   ```
   If this works but external access doesn't, check your security group rules.

2. **Check Grafana is running:**
   ```bash
   docker logs grafana
   ```

3. **Check firewall (if using UFW on Ubuntu):**
   ```bash
   sudo ufw allow 3000
   sudo ufw allow 8080
   sudo ufw allow 9090
   sudo ufw allow 3100
   ```

### Services not starting

```bash
docker-compose logs
```

### Out of disk space

```bash
df -h

docker system prune -a
```

### Permission denied errors

```bash
sudo chown -R $USER:$USER .
```

## Production Recommendations

### 1. Use HTTPS with SSL Certificate

Install Nginx and Certbot:
```bash
sudo apt install nginx certbot python3-certbot-nginx -y
```

Configure Nginx reverse proxy:
```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /grafana/ {
        proxy_pass http://localhost:3000/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

Get SSL certificate:
```bash
sudo certbot --nginx -d your-domain.com
```

### 2. Change Default Passwords

Update in docker-compose.yml:
```yaml
- GF_SECURITY_ADMIN_PASSWORD=YOUR_STRONG_PASSWORD
```

Update in .env:
```bash
ADMIN_PASSWORD=your-strong-password
```

### 3. Set Up Automatic Backups

Create backup script:
```bash
#!/bin/bash
docker exec prometheus tar -czf - /prometheus > prometheus-backup-$(date +%Y%m%d).tar.gz
docker exec loki tar -czf - /loki > loki-backup-$(date +%Y%m%d).tar.gz
docker exec grafana tar -czf - /var/lib/grafana > grafana-backup-$(date +%Y%m%d).tar.gz
```

Add to crontab:
```bash
crontab -e

0 2 * * * /home/ec2-user/backup-monitoring.sh
```

### 4. Enable Auto-restart on Reboot

```bash
sudo systemctl enable docker

docker-compose up -d
```

All services have `restart: unless-stopped` configured, so they'll auto-start.

### 5. Monitor Disk Usage

```bash
docker system df

docker volume ls
```

Set up log rotation in docker-compose.yml:
```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

### 6. Restrict Access by IP

In Security Group, change source from `0.0.0.0/0` to specific IP ranges:
- Your office IP
- VPN IP range
- Specific trusted IPs

## Monitoring Your Monitoring Stack

### Check Service Health

```bash
curl http://localhost:8080/health

curl http://localhost:9090/-/healthy

curl http://localhost:3100/ready

curl http://localhost:3000/api/health
```

### View Logs

```bash
docker-compose logs -f app

docker-compose logs -f grafana

docker-compose logs --tail=100 prometheus
```

### Resource Usage

```bash
docker stats

docker system df
```

## Updating the Application

```bash
git pull origin main

docker-compose down

docker-compose up -d --build
```

## Environment Variables - Important Notes

### About the .env File

- **NO PORT NEEDED**: The `.env` file is a configuration file, NOT a service
- **NEVER COMMIT**: Add `.env` to `.gitignore` (security!)
- **STAYS ON SERVER**: Only exists on your EC2 instance
- **LOADED BY DOCKER**: Docker Compose reads it automatically

### Creating .env on EC2

```bash
cp .env.example .env
nano .env
```

Add your actual credentials (they're secret!)

### Checking .env is Loaded

```bash
docker exec codecollab-backend env | grep SUPABASE
```

## Common Issues

### 1. "Connection refused" from Grafana to Prometheus

**Problem**: Datasource URL is wrong
**Solution**: Use `http://prometheus:9090` NOT `http://localhost:9090`

### 2. "Permission denied" errors

**Problem**: File ownership issues
**Solution**:
```bash
sudo chown -R $USER:$USER .
docker-compose down
docker-compose up -d
```

### 3. Can't access from browser but works on server

**Problem**: Security group not configured
**Solution**: Add inbound rules for ports 3000, 8080, etc.

### 4. Services keep restarting

**Problem**: Check logs for errors
**Solution**:
```bash
docker-compose logs app
docker-compose logs grafana
```

## Support

For issues:
1. Check logs: `docker-compose logs`
2. Check service status: `docker-compose ps`
3. Check EC2 security group rules
4. Check .env file is configured correctly

## Useful Commands Reference

```bash
docker-compose up -d              # Start services
docker-compose down               # Stop services
docker-compose ps                 # Check status
docker-compose logs -f app        # Follow app logs
docker-compose restart app        # Restart app
docker system prune -a            # Clean up space
docker-compose pull               # Update images
docker stats                      # Resource usage
```
