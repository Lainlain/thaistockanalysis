# Deployment Guide

## Overview

This guide covers various deployment options for ThaiStockAnalysis, from development to production environments.

## Quick Deployment Options

### Option 1: Direct Binary Deployment (Recommended)

```bash
# Build for target platform
GOOS=linux GOARCH=amd64 go build -o thaistockanalysis cmd/server/main.go

# Create directory structure
mkdir -p /opt/thaistockanalysis/{web,articles,data}

# Copy files
cp thaistockanalysis /opt/thaistockanalysis/
cp -r web/* /opt/thaistockanalysis/web/
cp -r articles/* /opt/thaistockanalysis/articles/

# Set permissions
chmod +x /opt/thaistockanalysis/thaistockanalysis
chown -R www-data:www-data /opt/thaistockanalysis
```

### Option 2: Docker Deployment

```bash
# Build and run with Docker
docker build -t thaistockanalysis .
docker run -p 7777:7777 -v ./data:/app/data thaistockanalysis
```

### Option 3: Systemd Service

```bash
# Install as systemd service
sudo cp scripts/thaistockanalysis.service /etc/systemd/system/
sudo systemctl enable thaistockanalysis
sudo systemctl start thaistockanalysis
```

## Production Deployment

### Server Requirements

#### Minimum Requirements
- **CPU**: 1 vCPU
- **RAM**: 512MB
- **Storage**: 1GB SSD
- **OS**: Linux (Ubuntu 20.04+ recommended)

#### Recommended Requirements
- **CPU**: 2 vCPU
- **RAM**: 2GB
- **Storage**: 10GB SSD
- **OS**: Linux (Ubuntu 22.04 LTS)

### Step-by-Step Production Setup

#### 1. Server Preparation

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install essential packages
sudo apt install -y nginx certbot python3-certbot-nginx ufw

# Create application user
sudo useradd -r -s /bin/false -d /opt/thaistockanalysis thaistockanalysis

# Create application directory
sudo mkdir -p /opt/thaistockanalysis/{bin,web,articles,data,logs}
sudo chown -R thaistockanalysis:thaistockanalysis /opt/thaistockanalysis
```

#### 2. Application Deployment

```bash
# Copy application files
sudo cp thaistockanalysis /opt/thaistockanalysis/bin/
sudo cp -r web/* /opt/thaistockanalysis/web/
sudo cp -r articles/* /opt/thaistockanalysis/articles/

# Set permissions
sudo chmod +x /opt/thaistockanalysis/bin/thaistockanalysis
sudo chown -R thaistockanalysis:thaistockanalysis /opt/thaistockanalysis
```

#### 3. Systemd Service Configuration

Create `/etc/systemd/system/thaistockanalysis.service`:

```ini
[Unit]
Description=ThaiStockAnalysis Server
Documentation=https://github.com/your-username/thaistockanalysis
After=network.target
Wants=network.target

[Service]
Type=simple
User=thaistockanalysis
Group=thaistockanalysis
WorkingDirectory=/opt/thaistockanalysis
ExecStart=/opt/thaistockanalysis/bin/thaistockanalysis
ExecReload=/bin/kill -HUP $MAINPID
KillMode=mixed
KillSignal=SIGTERM
Restart=always
RestartSec=5

# Security settings
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/thaistockanalysis/data /opt/thaistockanalysis/logs
PrivateTmp=true
ProtectKernelTunables=true
ProtectControlGroups=true
ProtectKernelModules=true
MemoryDenyWriteExecute=true

# Environment variables
Environment=PORT=7777
Environment=DATABASE_PATH=/opt/thaistockanalysis/data/admin.db
Environment=ARTICLES_DIR=/opt/thaistockanalysis/articles
Environment=TEMPLATE_DIR=/opt/thaistockanalysis/web/templates
Environment=STATIC_DIR=/opt/thaistockanalysis/web/static
Environment=DEBUG_MODE=false

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable thaistockanalysis
sudo systemctl start thaistockanalysis
sudo systemctl status thaistockanalysis
```

#### 4. Nginx Configuration

Create `/etc/nginx/sites-available/thaistockanalysis`:

```nginx
server {
    listen 80;
    server_name your-domain.com www.your-domain.com;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied expired no-cache no-store private must-revalidate auth;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/javascript
        application/xml+rss
        application/json;

    # Static files
    location /static/ {
        alias /opt/thaistockanalysis/web/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    # Main application
    location / {
        proxy_pass http://127.0.0.1:7777;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        proxy_read_timeout 86400;
    }

    # Logs
    access_log /var/log/nginx/thaistockanalysis.access.log;
    error_log /var/log/nginx/thaistockanalysis.error.log;
}
```

Enable the site:

```bash
sudo ln -s /etc/nginx/sites-available/thaistockanalysis /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

#### 5. SSL/TLS Setup

```bash
# Get SSL certificate
sudo certbot --nginx -d your-domain.com -d www.your-domain.com

# Test SSL renewal
sudo certbot renew --dry-run
```

#### 6. Firewall Configuration

```bash
# Configure UFW
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 'Nginx Full'
sudo ufw enable
```

## Docker Deployment

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.24.6-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Production stage
FROM alpine:latest

# Install dependencies
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/main .

# Copy web assets and articles
COPY --from=builder /app/web ./web
COPY --from=builder /app/articles ./articles

# Create data directory
RUN mkdir -p data

# Expose port
EXPOSE 7777

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:7777/ || exit 1

# Run the application
CMD ["./main"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  thaistockanalysis:
    build: .
    ports:
      - "7777:7777"
    volumes:
      - ./data:/root/data
      - ./articles:/root/articles
    environment:
      - DEBUG_MODE=false
      - PORT=7777
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:7777/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/ssl/certs
    depends_on:
      - thaistockanalysis
    restart: unless-stopped
```

Build and run:

```bash
# Build and start
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## Cloud Deployment

### AWS EC2

#### 1. Launch EC2 Instance

```bash
# Create security group
aws ec2 create-security-group \
  --group-name thaistockanalysis-sg \
  --description "Security group for ThaiStockAnalysis"

# Add rules
aws ec2 authorize-security-group-ingress \
  --group-name thaistockanalysis-sg \
  --protocol tcp \
  --port 22 \
  --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \
  --group-name thaistockanalysis-sg \
  --protocol tcp \
  --port 80 \
  --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \
  --group-name thaistockanalysis-sg \
  --protocol tcp \
  --port 443 \
  --cidr 0.0.0.0/0
```

#### 2. User Data Script

```bash
#!/bin/bash
yum update -y
yum install -y nginx

# Download and install application
cd /opt
wget https://github.com/your-username/thaistockanalysis/releases/latest/download/thaistockanalysis-linux
chmod +x thaistockanalysis-linux

# Setup systemd service
curl -o /etc/systemd/system/thaistockanalysis.service \
  https://raw.githubusercontent.com/your-username/thaistockanalysis/main/scripts/thaistockanalysis.service

systemctl enable thaistockanalysis
systemctl start thaistockanalysis
systemctl enable nginx
systemctl start nginx
```

### DigitalOcean Droplet

```bash
# Create droplet
doctl compute droplet create thaistockanalysis \
  --size s-1vcpu-1gb \
  --image ubuntu-22-04-x64 \
  --region nyc1 \
  --ssh-keys $SSH_KEY_ID

# Get droplet IP
doctl compute droplet list
```

### Google Cloud Platform

```bash
# Create VM instance
gcloud compute instances create thaistockanalysis \
  --machine-type=e2-micro \
  --zone=us-central1-a \
  --image-family=ubuntu-2204-lts \
  --image-project=ubuntu-os-cloud \
  --boot-disk-size=10GB \
  --tags=http-server,https-server

# Create firewall rules
gcloud compute firewall-rules create allow-http \
  --allow tcp:80 \
  --source-ranges 0.0.0.0/0 \
  --target-tags http-server

gcloud compute firewall-rules create allow-https \
  --allow tcp:443 \
  --source-ranges 0.0.0.0/0 \
  --target-tags https-server
```

## Monitoring & Maintenance

### Log Management

```bash
# Application logs
journalctl -u thaistockanalysis -f

# Nginx logs
tail -f /var/log/nginx/thaistockanalysis.access.log
tail -f /var/log/nginx/thaistockanalysis.error.log

# System resources
htop
df -h
free -h
```

### Backup Strategy

```bash
#!/bin/bash
# backup.sh - Database backup script

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/opt/backups"
DB_PATH="/opt/thaistockanalysis/data/admin.db"

mkdir -p $BACKUP_DIR

# Backup database
cp $DB_PATH $BACKUP_DIR/admin_${DATE}.db

# Backup articles
tar -czf $BACKUP_DIR/articles_${DATE}.tar.gz /opt/thaistockanalysis/articles/

# Cleanup old backups (keep 30 days)
find $BACKUP_DIR -name "*.db" -mtime +30 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete

echo "Backup completed: $DATE"
```

Setup cron job:

```bash
# Add to crontab
0 2 * * * /opt/scripts/backup.sh >> /var/log/backup.log 2>&1
```

### Performance Monitoring

```bash
# Install monitoring tools
sudo apt install -y htop iotop netstat-nat

# Monitor application
htop
sudo netstat -tlnp | grep :7777
sudo iotop

# Check disk usage
df -h
du -sh /opt/thaistockanalysis/*
```

### Health Checks

```bash
#!/bin/bash
# health-check.sh

# Check if service is running
if ! systemctl is-active --quiet thaistockanalysis; then
    echo "Service is down, restarting..."
    systemctl restart thaistockanalysis
fi

# Check if port is listening
if ! netstat -tlnp | grep -q ":7777"; then
    echo "Port 7777 not listening"
    systemctl restart thaistockanalysis
fi

# Check response
if ! curl -f http://localhost:7777 > /dev/null 2>&1; then
    echo "Application not responding"
    systemctl restart thaistockanalysis
fi
```

### Updates and Rollbacks

```bash
# Update application
sudo systemctl stop thaistockanalysis
sudo cp thaistockanalysis-new /opt/thaistockanalysis/bin/thaistockanalysis
sudo systemctl start thaistockanalysis

# Rollback
sudo systemctl stop thaistockanalysis
sudo cp thaistockanalysis-backup /opt/thaistockanalysis/bin/thaistockanalysis
sudo systemctl start thaistockanalysis
```

## Troubleshooting

### Common Issues

#### Service Won't Start

```bash
# Check service status
sudo systemctl status thaistockanalysis

# Check logs
sudo journalctl -u thaistockanalysis -n 50

# Check file permissions
ls -la /opt/thaistockanalysis/bin/thaistockanalysis
```

#### Database Issues

```bash
# Check database file
sqlite3 /opt/thaistockanalysis/data/admin.db ".tables"

# Recreate database
rm /opt/thaistockanalysis/data/admin.db
sudo systemctl restart thaistockanalysis
```

#### Performance Issues

```bash
# Check resource usage
top
free -h
df -h

# Check application logs
journalctl -u thaistockanalysis | grep ERROR
```

### Recovery Procedures

#### Restore from Backup

```bash
# Stop service
sudo systemctl stop thaistockanalysis

# Restore database
cp /opt/backups/admin_20240925_120000.db /opt/thaistockanalysis/data/admin.db

# Restore articles
tar -xzf /opt/backups/articles_20240925_120000.tar.gz -C /

# Fix permissions
sudo chown -R thaistockanalysis:thaistockanalysis /opt/thaistockanalysis

# Start service
sudo systemctl start thaistockanalysis
```

## Security Considerations

### Application Security

- Regular security updates
- Secure file permissions
- Input validation
- SQL injection prevention
- XSS protection

### Server Security

- Regular OS updates
- Firewall configuration
- SSH key authentication
- Fail2ban for brute force protection
- Regular security audits

### SSL/TLS

- Use strong cipher suites
- HSTS headers
- Regular certificate renewal
- Monitor certificate expiration

This comprehensive deployment guide should help you deploy ThaiStockAnalysis in any environment, from development to production.