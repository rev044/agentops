---
name: security
description: >
  Use when: "security", "penetration test", "pentest", "vulnerability", "OWASP",
  "network", "firewall", "SSL", "TLS", "DNS", "load balancer", "CDN", "encryption",
  "authentication", "authorization", "secrets".
version: 1.0.0
author: "AgentOps Team"
license: "MIT"
---

# Security Skill

Security testing and network infrastructure patterns.

## Quick Reference

| Area | Key Patterns | When to Use |
|------|--------------|-------------|
| **Penetration Testing** | Vulnerability assessment, exploitation | Security audits |
| **Network Engineering** | Load balancers, DNS, SSL/TLS, CDN | Infrastructure |

---

## Penetration Testing

### Assessment Methodology

| Phase | Activities |
|-------|------------|
| **Reconnaissance** | Information gathering, footprinting |
| **Scanning** | Port scanning, vulnerability scanning |
| **Exploitation** | Attempting to exploit vulnerabilities |
| **Post-Exploitation** | Privilege escalation, lateral movement |
| **Reporting** | Documenting findings and recommendations |

### OWASP Top 10 Testing

| Vulnerability | Test Method |
|---------------|-------------|
| **Injection** | SQLi, command injection payloads |
| **Broken Auth** | Session testing, credential stuffing |
| **Sensitive Data** | HTTPS checks, data exposure |
| **XXE** | XML entity injection |
| **Access Control** | IDOR, privilege escalation |
| **Misconfig** | Default creds, unnecessary services |
| **XSS** | Script injection, DOM manipulation |
| **Deserialization** | Object injection |
| **Components** | CVE scanning |
| **Logging** | Log injection, audit trail |

### Common Tools

```bash
# Network scanning
nmap -sV -sC -oA scan target.com

# Web vulnerability scanning
nikto -h https://target.com
nuclei -u https://target.com

# SSL/TLS testing
testssl.sh target.com:443

# Directory enumeration
gobuster dir -u https://target.com -w wordlist.txt
```

### Security Assessment Report Template

```markdown
# Security Assessment Report

**Target**: [System/Application]
**Date**: [Date]
**Assessor**: [Name]

## Executive Summary
[High-level findings and risk summary]

## Findings

### Critical
| ID | Finding | CVSS | Affected |
|----|---------|------|----------|
| C-01 | SQL Injection | 9.8 | /api/users |

### High
| ID | Finding | CVSS | Affected |
|----|---------|------|----------|
| H-01 | Weak Auth | 7.5 | /login |

### Medium/Low
[Similar format]

## Detailed Findings

### C-01: SQL Injection
**Severity**: Critical (CVSS 9.8)
**Location**: /api/users?id=
**Evidence**:
```
GET /api/users?id=1' OR '1'='1
Response: All user data returned
```
**Impact**: Full database access
**Remediation**: Use parameterized queries

## Recommendations
1. [Priority 1 action]
2. [Priority 2 action]
```

---

## Network Engineering

### Load Balancer Configuration

```yaml
# Example: HAProxy config
frontend http_front
  bind *:80
  bind *:443 ssl crt /etc/ssl/certs/
  redirect scheme https if !{ ssl_fc }
  default_backend servers

backend servers
  balance roundrobin
  option httpchk GET /health
  server server1 10.0.0.1:8080 check
  server server2 10.0.0.2:8080 check
```

### Load Balancing Algorithms

| Algorithm | Use Case |
|-----------|----------|
| **Round Robin** | Equal capacity servers |
| **Least Connections** | Varying request duration |
| **IP Hash** | Session persistence |
| **Weighted** | Different server capacities |

### DNS Configuration

```yaml
# Example DNS records
example.com.        A       203.0.113.10
www.example.com.    CNAME   example.com.
api.example.com.    A       203.0.113.20
                    AAAA    2001:db8::20
mail.example.com.   MX 10   mail1.example.com.
                    MX 20   mail2.example.com.
```

### DNS Best Practices

| Practice | Why |
|----------|-----|
| Low TTL for changes | Fast propagation |
| Multiple NS records | Redundancy |
| DNSSEC enabled | Prevent spoofing |
| CAA records | Control cert issuance |

---

## SSL/TLS Configuration

### Modern TLS Setup

```nginx
# Nginx SSL configuration
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
ssl_prefer_server_ciphers off;
ssl_session_timeout 1d;
ssl_session_cache shared:SSL:10m;
ssl_session_tickets off;

# HSTS
add_header Strict-Transport-Security "max-age=63072000" always;

# OCSP Stapling
ssl_stapling on;
ssl_stapling_verify on;
```

### Certificate Management

```bash
# Check certificate expiry
openssl s_client -connect example.com:443 2>/dev/null | openssl x509 -noout -dates

# Test SSL configuration
curl -vI https://example.com 2>&1 | grep -E "SSL|TLS|certificate"

# Generate CSR
openssl req -new -newkey rsa:2048 -nodes -keyout server.key -out server.csr
```

### SSL/TLS Checklist

- [ ] TLS 1.2+ only (no SSLv3, TLS 1.0/1.1)
- [ ] Strong cipher suites
- [ ] HSTS enabled
- [ ] Certificate chain complete
- [ ] OCSP stapling enabled
- [ ] Certificate not expiring soon
- [ ] No mixed content

---

## CDN Configuration

### CDN Benefits

| Benefit | Description |
|---------|-------------|
| **Performance** | Edge caching, lower latency |
| **Availability** | Global distribution |
| **Security** | DDoS protection, WAF |
| **Cost** | Reduced origin traffic |

### Cache Configuration

```yaml
# Example: CloudFront cache behavior
CacheBehavior:
  PathPattern: /static/*
  TTL:
    DefaultTTL: 86400
    MaxTTL: 31536000
  Compress: true
  CachePolicy: CachingOptimized

CacheBehavior:
  PathPattern: /api/*
  TTL:
    DefaultTTL: 0
    MaxTTL: 0
  CachePolicy: CachingDisabled
  OriginRequestPolicy: AllViewer
```

### Cache Headers

```
# Cache static assets
Cache-Control: public, max-age=31536000, immutable

# No cache for API
Cache-Control: no-store, no-cache, must-revalidate

# Conditional caching
Cache-Control: public, max-age=3600
ETag: "abc123"
```

---

## Security Headers

### Essential Headers

```nginx
# Security headers
add_header X-Content-Type-Options "nosniff" always;
add_header X-Frame-Options "DENY" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
add_header Content-Security-Policy "default-src 'self'; script-src 'self'" always;
add_header Permissions-Policy "geolocation=(), microphone=()" always;
```

### Content Security Policy

```
# Strict CSP example
Content-Security-Policy:
  default-src 'self';
  script-src 'self' 'nonce-abc123';
  style-src 'self' 'unsafe-inline';
  img-src 'self' data: https:;
  connect-src 'self' https://api.example.com;
  frame-ancestors 'none';
  base-uri 'self';
  form-action 'self';
```

---

## Secrets Management

### Best Practices

| Do | Don't |
|----|-------|
| Use secret managers | Hardcode secrets |
| Rotate regularly | Share secrets |
| Audit access | Log secrets |
| Encrypt at rest | Store in git |

### Secret Manager Integration

```python
# Example: AWS Secrets Manager
import boto3
import json

def get_secret(secret_name):
    client = boto3.client('secretsmanager')
    response = client.get_secret_value(SecretId=secret_name)
    return json.loads(response['SecretString'])

# Usage
db_creds = get_secret('prod/database')
connection_string = f"postgresql://{db_creds['username']}:{db_creds['password']}@..."
```

### Secrets Checklist

- [ ] No secrets in code
- [ ] No secrets in git history
- [ ] Secrets encrypted at rest
- [ ] Access audited
- [ ] Rotation policy defined
- [ ] Least privilege access
