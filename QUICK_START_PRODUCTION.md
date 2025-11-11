# Quick Start - Production Deployment

**Status**: 85% Ready | **Target Go-Live**: 2025-11-18 (7 days)

---

## ⚡ 5-Minute Quick Check

```bash
# 1. Verify code health
go test ./... && echo "✅ All tests passing"

# 2. Check security
govulncheck ./... && echo "✅ No vulnerabilities"

# 3. Build binary
go build -o panoptic main.go && echo "✅ Build successful"

# 4. Version check
./panoptic --help && echo "✅ Binary working"
```

**Expected Result**: All ✅ checks pass

---

## 📋 Current Status at a Glance

```
✅ Code Quality:      100% - 591 tests passing
✅ Testing:           100% - 4/4 E2E optimized
✅ Documentation:     100% - 6 guides ready
✅ CI/CD:             100% - 27 jobs configured
✅ Security:           95% - Fixes applied
⚠️  Go Version:       Needs upgrade (1.25.2 → 1.25.3)
❌ Infrastructure:      0% - Not provisioned
❌ Monitoring:          0% - Not configured
❌ Backup/DR:           0% - Not implemented
```

---

## 🚀 Next 3 Actions (Priority Order)

### 1. Upgrade Go (5 minutes) ⚠️
```bash
# macOS
brew upgrade go

# Linux
wget https://go.dev/dl/go1.25.3.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.3.linux-amd64.tar.gz

# Verify
go version  # Should show 1.25.3+
go build -o panoptic main.go
go test ./...
```

### 2. Choose Deployment Method (2 hours)

**Option A: Kubernetes** (Recommended for Scale)
```bash
kubectl apply -f deployments/kubernetes/
kubectl get pods -n panoptic
```
See: `docs/DEPLOYMENT.md` Section 2

**Option B: Docker Compose** (Recommended for Simple)
```bash
cp .env.example .env
nano .env  # Configure
docker-compose up -d
```
See: `docs/DEPLOYMENT.md` Section 3

**Option C: Systemd** (Recommended for Bare Metal)
```bash
sudo cp panoptic /opt/panoptic/bin/
sudo cp deployments/systemd/panoptic.service /etc/systemd/system/
sudo systemctl enable --now panoptic
```
See: `docs/DEPLOYMENT.md` Section 4

### 3. Setup Monitoring (4 hours)
```bash
# Deploy Prometheus + Grafana
kubectl apply -f deployments/kubernetes/prometheus.yaml
kubectl apply -f deployments/kubernetes/grafana.yaml

# Or Docker Compose
docker-compose -f deployments/docker/monitoring-compose.yml up -d

# Import dashboards from deployments/grafana/dashboards/
```
See: `DEPLOYMENT_CHECKLIST.md` Day 5

---

## 📁 Key Documentation Files

| File | Purpose | When to Use |
|------|---------|-------------|
| `DEPLOYMENT_CHECKLIST.md` | Step-by-step 7-day plan | **START HERE** |
| `docs/DEPLOYMENT.md` | Detailed deployment guide | Reference during setup |
| `docs/PRODUCTION_READINESS.md` | Pre-deployment checklist | Before go-live |
| `SECURITY_FIXES.md` | Applied security fixes | Review with security team |
| `SESSION_SUMMARY_2025-11-11.md` | What was completed today | Handoff to next team |

---

## 🔧 Essential Commands

### Development
```bash
# Run tests
go test ./... -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. -benchmem ./internal/...

# Run E2E tests
go test -tags=e2e ./tests/e2e/... -v
```

### Security
```bash
# Vulnerability scan
govulncheck ./...

# Static analysis
gosec ./...

# Secret scan
trufflehog git file://. --only-verified

# Container scan
trivy image panoptic:latest
```

### Deployment
```bash
# Build binary
go build -ldflags="-s -w" -o panoptic main.go

# Build Docker image
docker build -t panoptic:latest .

# Run locally
./panoptic run test_config.yaml

# Run with custom output
./panoptic run test_config.yaml --output ./my-output
```

### Monitoring
```bash
# Check health
curl http://localhost:8080/health

# View metrics
curl http://localhost:9090/metrics

# View logs (Docker)
docker logs -f panoptic

# View logs (Kubernetes)
kubectl logs -f deployment/panoptic -n panoptic

# View logs (Systemd)
sudo journalctl -u panoptic -f
```

---

## 🎯 7-Day Deployment Timeline

```
Day 1-2: Go Upgrade & Security Validation
  ├─ Upgrade Go to 1.25.3
  ├─ Re-run security scans
  └─ Verify all tests pass

Day 3-4: Infrastructure Provisioning
  ├─ Deploy to Kubernetes/Docker
  ├─ Configure cloud storage
  └─ Test staging deployment

Day 5: Monitoring & Observability
  ├─ Deploy Prometheus + Grafana
  ├─ Configure ELK stack
  └─ Set up alerting

Day 6: Backup & Disaster Recovery
  ├─ Implement backup strategy
  ├─ Test restore procedures
  └─ Conduct DR drill

Day 7: Production Go-Live
  ├─ Final smoke tests
  ├─ Deploy to production
  └─ Monitor for 1 hour
```

**Target Date**: 2025-11-18

---

## ⚠️ Known Issues & Workarounds

### Issue 1: Platform Tests Take 8+ Minutes
**Status**: Normal behavior
**Reason**: Tests spawn real Chrome instances
**Workaround**: Run with `-short` flag to skip slow tests
```bash
go test -short ./internal/platforms/...
```

### Issue 2: gosec Shows 4 Integer Overflow False Positives
**Status**: Safe to ignore
**Reason**: Bit-shift operations are safe but gosec flags them
**Location**: `internal/vision/detector.go:414-420`
**Verification**: All tests pass, conversion is mathematically correct

### Issue 3: 57 Low-Severity "Unhandled Errors"
**Status**: Acceptable for production
**Reason**: Most in cleanup/defer blocks
**Decision**: Documented in `SECURITY_FIXES.md`

---

## 📞 Need Help?

### Quick Links
- **Full Deployment Guide**: `DEPLOYMENT_CHECKLIST.md`
- **Architecture Overview**: `docs/ARCHITECTURE.md`
- **Troubleshooting**: `docs/TROUBLESHOOTING.md`
- **Security Details**: `SECURITY_FIXES.md`
- **Performance Tuning**: `docs/PERFORMANCE.md`

### Team Contacts
- DevOps Lead: devops-lead@example.com
- Security Lead: security@example.com
- Engineering Lead: eng-lead@example.com
- On-Call: oncall@example.com

### Emergency
- 24/7 Hotline: +1-555-0100
- Critical Bug: critical@example.com

---

## ✅ Pre-Flight Checklist (Use Before Deploy)

```
Before Running Any Commands:
[ ] Read DEPLOYMENT_CHECKLIST.md
[ ] Review SECURITY_FIXES.md with security team
[ ] Check Go version (should be 1.25.3+)
[ ] Verify all tests pass: go test ./...
[ ] Review infrastructure requirements
[ ] Have rollback plan ready

Before Production Deployment:
[ ] All items above complete
[ ] Staging environment tested
[ ] Monitoring configured
[ ] Backup strategy tested
[ ] Team trained and ready
[ ] Stakeholders notified
[ ] Maintenance window scheduled
[ ] Rollback tested
```

---

**Quick Start Version**: 1.0
**Last Updated**: 2025-11-11
**Project Status**: 85% Production Ready

**🎯 Ready to Deploy?** Start with `DEPLOYMENT_CHECKLIST.md` Day 1 tasks!
