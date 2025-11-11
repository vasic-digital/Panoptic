# Panoptic Production Deployment Guide

**Version**: 1.0
**Last Updated**: 2025-11-11
**Target Audience**: DevOps Engineers, System Administrators, Release Managers

---

## Table of Contents

1. [Overview](#overview)
2. [System Requirements](#system-requirements)
3. [Pre-Deployment Checklist](#pre-deployment-checklist)
4. [Deployment Methods](#deployment-methods)
5. [Configuration Management](#configuration-management)
6. [Security Configuration](#security-configuration)
7. [Monitoring and Observability](#monitoring-and-observability)
8. [Backup and Recovery](#backup-and-recovery)
9. [Scaling Guidelines](#scaling-guidelines)
10. [Troubleshooting](#troubleshooting)

---

## Overview

Panoptic is a comprehensive automated testing framework designed for enterprise-scale deployment. This guide covers production deployment scenarios across various environments.

### Deployment Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Load Balancer / Proxy                 │
└──────────────┬──────────────────────────────────────────┘
               │
     ┌─────────┴──────────┬──────────────┬────────────────┐
     │                    │              │                │
┌────▼─────┐      ┌───────▼───┐   ┌─────▼────┐   ┌──────▼───┐
│ Panoptic │      │ Panoptic  │   │ Panoptic │   │ Panoptic │
│ Node 1   │      │ Node 2    │   │ Node 3   │   │ Node N   │
└────┬─────┘      └───────┬───┘   └─────┬────┘   └──────┬───┘
     │                    │              │                │
     └────────────────────┴──────────────┴────────────────┘
                          │
              ┌───────────┴───────────┐
              │                       │
         ┌────▼────┐            ┌─────▼─────┐
         │ Cloud   │            │ Enterprise│
         │ Storage │            │ Backend   │
         └─────────┘            └───────────┘
```

---

## System Requirements

### Minimum Requirements

**Hardware:**
- CPU: 4 cores (2.0 GHz or higher)
- RAM: 8 GB
- Storage: 50 GB available space
- Network: 100 Mbps

**Software:**
- Go 1.21 or later
- Supported OS: Linux (Ubuntu 20.04+, RHEL 8+), macOS 12+, Windows Server 2019+
- Browser: Chrome/Chromium 90+ (for web automation)

### Recommended Requirements (Production)

**Hardware:**
- CPU: 8+ cores (3.0 GHz or higher)
- RAM: 16-32 GB
- Storage: 200 GB SSD
- Network: 1 Gbps

**Software:**
- Go 1.22+
- Linux (Ubuntu 22.04 LTS or RHEL 9)
- Docker 24+ (for containerized deployments)
- Kubernetes 1.28+ (for orchestrated deployments)

### Browser Requirements (Web Automation)

- Chrome/Chromium 120+
- ChromeDriver matching Chrome version
- Xvfb (for headless Linux environments)

### Mobile Testing Requirements

**Android:**
- Android SDK Platform Tools 34+
- ADB (Android Debug Bridge)
- Android devices or emulators running Android 10+

**iOS:**
- macOS with Xcode 15+
- iOS Simulator or physical devices running iOS 15+
- libimobiledevice (for device communication)

---

## Pre-Deployment Checklist

### 1. Environment Preparation

- [ ] System meets minimum requirements
- [ ] All required software installed
- [ ] Network connectivity verified
- [ ] Firewall rules configured
- [ ] SSL/TLS certificates obtained
- [ ] DNS records configured

### 2. Security Setup

- [ ] Enterprise license obtained and validated
- [ ] User accounts and roles configured
- [ ] API keys generated
- [ ] Cloud storage credentials secured
- [ ] Audit logging enabled
- [ ] Compliance requirements documented

### 3. Storage Configuration

- [ ] Cloud storage provider selected (AWS S3, GCP, Azure, or Local)
- [ ] Storage buckets/containers created
- [ ] Access policies configured
- [ ] Backup location configured
- [ ] Retention policies defined

### 4. Testing

- [ ] Unit tests passing (go test ./internal/... ./cmd/...)
- [ ] Integration tests passing
- [ ] E2E tests passing
- [ ] Performance benchmarks run
- [ ] Security scan completed

### 5. Documentation

- [ ] Configuration templates prepared
- [ ] Runbooks documented
- [ ] Contact information updated
- [ ] Escalation procedures defined

---

## Deployment Methods

### Method 1: Binary Deployment (Standalone)

**Best for**: Single server, development, testing environments

#### 1. Build the Binary

```bash
# Clone repository
git clone https://github.com/yourusername/panoptic.git
cd panoptic

# Run tests
go test ./internal/... ./cmd/...

# Build for production
go build -ldflags="-s -w" -o panoptic main.go

# Verify binary
./panoptic --version
```

#### 2. Create Directory Structure

```bash
sudo mkdir -p /opt/panoptic/{bin,config,data,logs,output}
sudo mv panoptic /opt/panoptic/bin/
sudo chmod +x /opt/panoptic/bin/panoptic
```

#### 3. Create Configuration

```bash
# Copy example configs
cp examples/enterprise_config.yaml /opt/panoptic/config/
cp examples/test_config.yaml /opt/panoptic/config/

# Edit for your environment
sudo nano /opt/panoptic/config/enterprise_config.yaml
```

#### 4. Create Systemd Service (Linux)

```bash
sudo nano /etc/systemd/system/panoptic.service
```

```ini
[Unit]
Description=Panoptic Test Automation Framework
After=network.target

[Service]
Type=simple
User=panoptic
Group=panoptic
WorkingDirectory=/opt/panoptic
ExecStart=/opt/panoptic/bin/panoptic run /opt/panoptic/config/test_config.yaml
Restart=on-failure
RestartSec=10
StandardOutput=append:/opt/panoptic/logs/panoptic.log
StandardError=append:/opt/panoptic/logs/panoptic-error.log

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable panoptic
sudo systemctl start panoptic
sudo systemctl status panoptic
```

---

### Method 2: Docker Deployment

**Best for**: Development, testing, consistent environments

#### 1. Create Dockerfile

```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .

RUN apk add --no-cache git make
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o panoptic main.go

FROM alpine:latest

RUN apk add --no-cache \
    chromium \
    chromium-chromedriver \
    xvfb \
    ca-certificates \
    tzdata

WORKDIR /app

COPY --from=builder /app/panoptic /usr/local/bin/panoptic
COPY examples/ /app/config/

RUN addgroup -g 1000 panoptic && \
    adduser -D -u 1000 -G panoptic panoptic && \
    mkdir -p /app/output /app/data && \
    chown -R panoptic:panoptic /app

USER panoptic

EXPOSE 8080

ENV DISPLAY=:99

ENTRYPOINT ["panoptic"]
CMD ["run", "/app/config/test_config.yaml"]
```

#### 2. Build and Run

```bash
# Build image
docker build -t panoptic:latest .

# Run container
docker run -d \
  --name panoptic \
  -v /path/to/config:/app/config \
  -v /path/to/output:/app/output \
  -e PANOPTIC_ENV=production \
  panoptic:latest

# View logs
docker logs -f panoptic
```

#### 3. Docker Compose Setup

```yaml
version: '3.8'

services:
  panoptic:
    image: panoptic:latest
    container_name: panoptic
    restart: unless-stopped
    volumes:
      - ./config:/app/config:ro
      - ./output:/app/output
      - ./data:/app/data
    environment:
      - PANOPTIC_ENV=production
      - TZ=UTC
    networks:
      - panoptic-network
    ports:
      - "8080:8080"

  # Optional: Local storage for cloud provider
  minio:
    image: minio/minio:latest
    container_name: panoptic-storage
    command: server /data --console-address ":9001"
    volumes:
      - minio-data:/data
    environment:
      - MINIO_ROOT_USER=admin
      - MINIO_ROOT_PASSWORD=changeme
    ports:
      - "9000:9000"
      - "9001:9001"
    networks:
      - panoptic-network

networks:
  panoptic-network:
    driver: bridge

volumes:
  minio-data:
```

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f panoptic

# Stop services
docker-compose down
```

---

### Method 3: Kubernetes Deployment

**Best for**: Production, high availability, scalability

#### 1. Create Namespace

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: panoptic
```

#### 2. Create ConfigMap

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: panoptic-config
  namespace: panoptic
data:
  enterprise_config.yaml: |
    enabled: true
    organization:
      name: "Your Organization"
      id: "your-org-id"
    license:
      type: "enterprise"
      max_users: 1000
      expiration_date: "2030-12-31T23:59:59Z"
    storage:
      data_path: "/app/data"
```

#### 3. Create Secret

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: panoptic-secrets
  namespace: panoptic
type: Opaque
stringData:
  aws-access-key: "your-access-key"
  aws-secret-key: "your-secret-key"
  api-token: "your-api-token"
```

#### 4. Create Persistent Volume Claim

```yaml
# pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: panoptic-data
  namespace: panoptic
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 100Gi
  storageClassName: fast-ssd
```

#### 5. Create Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: panoptic
  namespace: panoptic
  labels:
    app: panoptic
spec:
  replicas: 3
  selector:
    matchLabels:
      app: panoptic
  template:
    metadata:
      labels:
        app: panoptic
    spec:
      containers:
      - name: panoptic
        image: panoptic:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: PANOPTIC_ENV
          value: "production"
        - name: AWS_ACCESS_KEY_ID
          valueFrom:
            secretKeyRef:
              name: panoptic-secrets
              key: aws-access-key
        - name: AWS_SECRET_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: panoptic-secrets
              key: aws-secret-key
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
        - name: data
          mountPath: /app/data
        - name: output
          mountPath: /app/output
        resources:
          requests:
            cpu: "2"
            memory: "4Gi"
          limits:
            cpu: "4"
            memory: "8Gi"
        livenessProbe:
          exec:
            command:
            - /usr/local/bin/panoptic
            - --version
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          exec:
            command:
            - /usr/local/bin/panoptic
            - --version
          initialDelaySeconds: 5
          periodSeconds: 10
      volumes:
      - name: config
        configMap:
          name: panoptic-config
      - name: data
        persistentVolumeClaim:
          claimName: panoptic-data
      - name: output
        emptyDir: {}
```

#### 6. Create Service

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: panoptic
  namespace: panoptic
  labels:
    app: panoptic
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: panoptic
```

#### 7. Deploy to Kubernetes

```bash
# Apply all configurations
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml
kubectl apply -f pvc.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml

# Verify deployment
kubectl get pods -n panoptic
kubectl get svc -n panoptic

# View logs
kubectl logs -f deployment/panoptic -n panoptic

# Scale deployment
kubectl scale deployment panoptic --replicas=5 -n panoptic
```

---

## Configuration Management

### Environment Variables

```bash
# Application
export PANOPTIC_ENV=production
export PANOPTIC_LOG_LEVEL=info
export PANOPTIC_OUTPUT_DIR=/opt/panoptic/output

# Cloud Storage
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export PANOPTIC_CLOUD_BUCKET=panoptic-artifacts

# Enterprise
export PANOPTIC_ENTERPRISE_CONFIG=/opt/panoptic/config/enterprise_config.yaml
export PANOPTIC_LICENSE_KEY=your-license-key
```

### Configuration Files

**Location**: `/opt/panoptic/config/` or `/app/config/` (Docker)

**Main Configuration**: `test_config.yaml`
```yaml
name: "Production Test Suite"
output: "/app/output"

apps:
  - name: "Production App"
    type: "web"
    url: "https://your-production-app.com"

actions:
  - name: "health_check"
    type: "navigate"
    value: "https://your-production-app.com/health"

settings:
  headless: true
  enable_metrics: true
  log_level: "info"

  cloud:
    provider: "aws"
    bucket: "panoptic-artifacts"
    enable_sync: true

  enterprise:
    config_path: "/app/config/enterprise_config.yaml"
```

---

## Security Configuration

### 1. SSL/TLS Configuration

```yaml
# Enable HTTPS for cloud uploads
settings:
  cloud:
    provider: "aws"
    bucket: "panoptic-artifacts"
    enable_encryption: true
    encryption_type: "AES256"
```

### 2. Credentials Management

**Never store credentials in configuration files!**

Use environment variables or secret management systems:

```bash
# AWS Credentials
export AWS_ACCESS_KEY_ID=$(aws ssm get-parameter --name /panoptic/aws-key --query Parameter.Value --output text)
export AWS_SECRET_ACCESS_KEY=$(aws ssm get-parameter --name /panoptic/aws-secret --with-decryption --query Parameter.Value --output text)

# Or use AWS IAM roles (recommended for EC2/ECS)
```

### 3. Network Security

**Firewall Rules:**
```bash
# Allow outbound HTTPS (cloud storage)
sudo ufw allow out 443/tcp

# Allow outbound HTTP (if needed)
sudo ufw allow out 80/tcp

# Restrict inbound (if API is exposed)
sudo ufw allow from 10.0.0.0/8 to any port 8080
```

### 4. File Permissions

```bash
# Restrict configuration files
sudo chmod 600 /opt/panoptic/config/enterprise_config.yaml
sudo chown panoptic:panoptic /opt/panoptic/config/enterprise_config.yaml

# Secure data directory
sudo chmod 700 /opt/panoptic/data
sudo chown -R panoptic:panoptic /opt/panoptic/data
```

---

## Monitoring and Observability

### 1. Log Management

**Log Locations:**
- Application logs: `/opt/panoptic/logs/panoptic.log`
- Error logs: `/opt/panoptic/logs/panoptic-error.log`
- Audit logs: `/opt/panoptic/data/audit.json`

**Log Rotation (logrotate):**
```bash
sudo nano /etc/logrotate.d/panoptic
```

```
/opt/panoptic/logs/*.log {
    daily
    rotate 30
    compress
    delaycompress
    notifempty
    create 0640 panoptic panoptic
    sharedscripts
    postrotate
        systemctl reload panoptic > /dev/null 2>&1 || true
    endscript
}
```

### 2. Metrics Collection

Monitor key metrics:
- Test execution time
- Success/failure rates
- Screenshot count
- Storage usage
- Memory usage
- CPU usage

### 3. Health Checks

```bash
# Systemd health check
systemctl status panoptic

# Docker health check
docker ps | grep panoptic

# Kubernetes health check
kubectl get pods -n panoptic

# Application health
/opt/panoptic/bin/panoptic --version
```

---

## Backup and Recovery

### 1. Backup Strategy

**What to Backup:**
- Configuration files (`/opt/panoptic/config/`)
- Enterprise data (`/opt/panoptic/data/`)
- Test results (if not in cloud storage)
- Audit logs

**Backup Script:**
```bash
#!/bin/bash
# /opt/panoptic/scripts/backup.sh

BACKUP_DIR="/backup/panoptic"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# Backup configuration
tar -czf $BACKUP_DIR/config_$DATE.tar.gz /opt/panoptic/config/

# Backup data
tar -czf $BACKUP_DIR/data_$DATE.tar.gz /opt/panoptic/data/

# Backup logs
tar -czf $BACKUP_DIR/logs_$DATE.tar.gz /opt/panoptic/logs/

# Remove backups older than 30 days
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete

echo "Backup completed: $DATE"
```

**Cron Schedule:**
```bash
# Daily backup at 2 AM
0 2 * * * /opt/panoptic/scripts/backup.sh >> /opt/panoptic/logs/backup.log 2>&1
```

### 2. Recovery Procedures

```bash
# Stop service
sudo systemctl stop panoptic

# Restore configuration
tar -xzf /backup/panoptic/config_YYYYMMDD_HHMMSS.tar.gz -C /

# Restore data
tar -xzf /backup/panoptic/data_YYYYMMDD_HHMMSS.tar.gz -C /

# Verify permissions
sudo chown -R panoptic:panoptic /opt/panoptic

# Start service
sudo systemctl start panoptic
sudo systemctl status panoptic
```

---

## Scaling Guidelines

### Horizontal Scaling

**Add More Nodes:**
```bash
# Kubernetes
kubectl scale deployment panoptic --replicas=10 -n panoptic

# Docker Swarm
docker service scale panoptic=10
```

**Distributed Testing:**
```yaml
settings:
  cloud:
    provider: "aws"
    distributed:
      enabled: true
      node_count: 5
      max_concurrent: 20
```

### Vertical Scaling

**Increase Resources:**
```yaml
# Kubernetes resources
resources:
  requests:
    cpu: "4"
    memory: "8Gi"
  limits:
    cpu: "8"
    memory: "16Gi"
```

### Performance Tuning

**Optimize Executor:**
- Use lazy initialization for heavy components
- Enable result caching
- Adjust browser pool size
- Configure parallel test execution

---

## Troubleshooting

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for detailed troubleshooting procedures.

**Quick Diagnostics:**

```bash
# Check service status
systemctl status panoptic

# View recent logs
tail -100 /opt/panoptic/logs/panoptic.log

# Check disk space
df -h /opt/panoptic

# Check browser availability
which chromium-browser
chromium-browser --version

# Test configuration
/opt/panoptic/bin/panoptic run --dry-run /opt/panoptic/config/test_config.yaml
```

---

## Support and Resources

- **Documentation**: https://github.com/yourusername/panoptic/docs
- **Issues**: https://github.com/yourusername/panoptic/issues
- **Security**: security@yourcompany.com
- **Enterprise Support**: support@yourcompany.com

---

**Document Version**: 1.0
**Effective Date**: 2025-11-11
**Review Date**: 2025-12-11
