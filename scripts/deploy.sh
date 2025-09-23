#!/bin/bash
# Deployment script for ThaiStockAnalysis

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="thaistockanalysis"
APP_USER="thaistockanalysis"
APP_DIR="/opt/thaistockanalysis"
SERVICE_NAME="thaistockanalysis"
BINARY_NAME="thaistockanalysis-linux"

echo -e "${BLUE}🚀 Deploying ThaiStockAnalysis${NC}"

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}❌ This script must be run as root${NC}"
   exit 1
fi

# Check if binary exists
if [ ! -f "bin/${BINARY_NAME}" ]; then
    echo -e "${RED}❌ Binary not found: bin/${BINARY_NAME}${NC}"
    echo -e "${YELLOW}💡 Run ./scripts/build.sh first${NC}"
    exit 1
fi

# Create application user
echo -e "${YELLOW}👤 Creating application user...${NC}"
if ! id "${APP_USER}" &>/dev/null; then
    useradd -r -s /bin/false -d ${APP_DIR} ${APP_USER}
    echo -e "${GREEN}✅ User ${APP_USER} created${NC}"
else
    echo -e "${GREEN}✅ User ${APP_USER} already exists${NC}"
fi

# Create directory structure
echo -e "${YELLOW}📁 Creating directory structure...${NC}"
mkdir -p ${APP_DIR}/{bin,web,articles,data,logs}
chown -R ${APP_USER}:${APP_USER} ${APP_DIR}

# Stop service if running
echo -e "${YELLOW}🛑 Stopping service if running...${NC}"
if systemctl is-active --quiet ${SERVICE_NAME}; then
    systemctl stop ${SERVICE_NAME}
    echo -e "${GREEN}✅ Service stopped${NC}"
fi

# Copy application files
echo -e "${YELLOW}📋 Copying application files...${NC}"
cp bin/${BINARY_NAME} ${APP_DIR}/bin/${APP_NAME}
cp -r web/* ${APP_DIR}/web/ 2>/dev/null || echo "No web directory found"
cp -r articles/* ${APP_DIR}/articles/ 2>/dev/null || echo "No articles directory found"

# Set permissions
chmod +x ${APP_DIR}/bin/${APP_NAME}
chown -R ${APP_USER}:${APP_USER} ${APP_DIR}

# Create systemd service
echo -e "${YELLOW}⚙️  Creating systemd service...${NC}"
cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=ThaiStockAnalysis Server
Documentation=https://github.com/your-username/thaistockanalysis
After=network.target
Wants=network.target

[Service]
Type=simple
User=${APP_USER}
Group=${APP_USER}
WorkingDirectory=${APP_DIR}
ExecStart=${APP_DIR}/bin/${APP_NAME}
ExecReload=/bin/kill -HUP \$MAINPID
KillMode=mixed
KillSignal=SIGTERM
Restart=always
RestartSec=5

# Security settings
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${APP_DIR}/data ${APP_DIR}/logs
PrivateTmp=true
ProtectKernelTunables=true
ProtectControlGroups=true
ProtectKernelModules=true
MemoryDenyWriteExecute=true

# Environment variables
Environment=PORT=7777
Environment=DATABASE_PATH=${APP_DIR}/data/admin.db
Environment=ARTICLES_DIR=${APP_DIR}/articles
Environment=TEMPLATE_DIR=${APP_DIR}/web/templates
Environment=STATIC_DIR=${APP_DIR}/web/static
Environment=DEBUG_MODE=false

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and enable service
echo -e "${YELLOW}🔄 Reloading systemd...${NC}"
systemctl daemon-reload
systemctl enable ${SERVICE_NAME}

# Start service
echo -e "${YELLOW}▶️  Starting service...${NC}"
systemctl start ${SERVICE_NAME}

# Check service status
sleep 2
if systemctl is-active --quiet ${SERVICE_NAME}; then
    echo -e "${GREEN}✅ Service started successfully${NC}"
    echo -e "${GREEN}🌐 Application is running on http://localhost:7777${NC}"
else
    echo -e "${RED}❌ Service failed to start${NC}"
    echo -e "${YELLOW}📋 Check logs: journalctl -u ${SERVICE_NAME}${NC}"
    exit 1
fi

# Install nginx if not present
if ! command -v nginx &> /dev/null; then
    echo -e "${YELLOW}📦 Installing nginx...${NC}"
    apt update
    apt install -y nginx
fi

# Create nginx configuration
echo -e "${YELLOW}🌐 Creating nginx configuration...${NC}"
cat > /etc/nginx/sites-available/${APP_NAME} << EOF
server {
    listen 80;
    server_name localhost;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;

    # Static files
    location /static/ {
        alias ${APP_DIR}/web/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    # Main application
    location / {
        proxy_pass http://127.0.0.1:7777;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    # Logs
    access_log /var/log/nginx/${APP_NAME}.access.log;
    error_log /var/log/nginx/${APP_NAME}.error.log;
}
EOF

# Enable nginx site
if [ ! -L "/etc/nginx/sites-enabled/${APP_NAME}" ]; then
    ln -s /etc/nginx/sites-available/${APP_NAME} /etc/nginx/sites-enabled/
fi

# Test nginx configuration
if nginx -t; then
    echo -e "${GREEN}✅ Nginx configuration is valid${NC}"
    systemctl reload nginx
else
    echo -e "${RED}❌ Nginx configuration error${NC}"
fi

# Configure firewall if ufw is available
if command -v ufw &> /dev/null; then
    echo -e "${YELLOW}🔥 Configuring firewall...${NC}"
    ufw allow 22/tcp
    ufw allow 80/tcp
    ufw allow 443/tcp
    echo -e "${GREEN}✅ Firewall configured${NC}"
fi

echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"
echo -e "${BLUE}📊 Service status:${NC}"
systemctl status ${SERVICE_NAME} --no-pager -l

echo -e "${BLUE}🌐 Your application is available at:${NC}"
echo -e "${GREEN}  • http://localhost${NC}"
echo -e "${GREEN}  • http://localhost/admin${NC}"

echo -e "${YELLOW}💡 Useful commands:${NC}"
echo -e "  • Check logs: journalctl -u ${SERVICE_NAME} -f"
echo -e "  • Restart service: systemctl restart ${SERVICE_NAME}"
echo -e "  • Check status: systemctl status ${SERVICE_NAME}"
