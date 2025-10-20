# Quick Start Guide for Testing Security Fixes

## Prerequisites

- Go 1.25+
- Bun (latest version)
- Playwright
- Docker (for PostgreSQL with SSL)
- Running EduHub server and client

## 1. Backend Security Tests

### Setup
```bash
cd server

# Ensure all dependencies are installed
go mod download

# Set environment for testing
export DB_SKIP_CONNECT=1
export APP_ENV=test
```

### Run Security Tests
```bash
# Run all security tests
go test ./tests/security_test.go -v

# Run specific test
go test ./tests/security_test.go -v -run TestMultiTenantIsolation

# Run with coverage
go test ./tests/security_test.go -v -cover
```

## 2. Frontend E2E Security Tests

### Setup
```bash
cd client

# Install dependencies
bun install

# Install Playwright browsers
bun run playwright install
```

### Run E2E Tests
```bash
# Run all E2E security tests
bun run playwright test tests/e2e/security.spec.ts

# Run with UI mode (recommended for debugging)
bun run playwright test tests/e2e/security.spec.ts --ui

# Run specific test
bun run playwright test tests/e2e/security.spec.ts -g "Multi-Tenant"

# Generate HTML report
bun run playwright test tests/e2e/security.spec.ts --reporter=html
```

## 3. Manual Security Verification

### Test Multi-Tenant Isolation

#### Step 1: Create Two Test Colleges
```bash
# Start your server
cd server
go run main.go

# In another terminal, create test data
curl -X POST http://localhost:8080/api/colleges \
  -H "Content-Type: application/json" \
  -d '{"name": "Test College 1", "city": "City1"}'

curl -X POST http://localhost:8080/api/colleges \
  -H "Content-Type: application/json" \
  -d '{"name": "Test College 2", "city": "City2"}'
```

#### Step 2: Create Users in Each College
```bash
# User in College 1
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user1@college1.com",
    "password": "Test123!",
    "college_id": "1",
    "role": "student"
  }'

# User in College 2
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user2@college2.com",
    "password": "Test123!",
    "college_id": "2",
    "role": "student"
  }'
```

#### Step 3: Test Isolation
```bash
# Login as College 1 user
TOKEN1=$(curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user1@college1.com","password":"Test123!"}' \
  | jq -r '.token')

# Try to access College 2 data (should fail with 403)
curl -X GET http://localhost:8080/api/students \
  -H "Authorization: Bearer $TOKEN1" \
  -H "X-College-ID: 2"

# Expected: 403 Forbidden or empty result
```

### Test JWT Token Rotation

```bash
# Login and get initial token
TOKEN=$(curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@college.com","password":"Test123!"}' \
  | jq -r '.token')

echo "Initial Token: $TOKEN"

# Refresh token
NEW_TOKEN=$(curl -X POST http://localhost:8080/auth/refresh \
  -H "Authorization: Bearer $TOKEN" \
  | jq -r '.token')

echo "New Token: $NEW_TOKEN"

# Tokens should be different
test "$TOKEN" != "$NEW_TOKEN" && echo "✅ Token rotation working" || echo "❌ Token rotation failed"
```

### Test Error Sanitization

```bash
# Set production mode
export APP_ENV=production
export APP_DEBUG=false

# Trigger a database error
curl -X GET http://localhost:8080/api/students/99999999 \
  -H "Authorization: Bearer $TOKEN"

# Response should NOT contain:
# - SQL errors
# - File paths (.go files)
# - Stack traces (goroutine)
# - Database connection strings
```

### Test QR Code Security

```bash
# Generate QR code (as faculty)
FACULTY_TOKEN=$(curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"faculty@college.com","password":"Test123!"}' \
  | jq -r '.token')

QR_CODE=$(curl -X GET "http://localhost:8080/api/attendance/course/1/lecture/1/qrcode" \
  -H "Authorization: Bearer $FACULTY_TOKEN" \
  | jq -r '.qr_code')

# Try to use QR code from different college (should fail)
STUDENT_TOKEN_COLLEGE2=$(curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"student@college2.com","password":"Test123!"}' \
  | jq -r '.token')

curl -X POST http://localhost:8080/api/attendance/process-qr \
  -H "Authorization: Bearer $STUDENT_TOKEN_COLLEGE2" \
  -H "Content-Type: application/json" \
  -d "{\"qrcode_data\": \"$QR_CODE\"}"

# Expected: 403 Forbidden - "qr code belongs to different institution"
```

## 4. Database SSL Verification

### Setup PostgreSQL with SSL

```bash
# Create SSL certificates
openssl req -new -x509 -days 365 -nodes -text \
  -out server.crt \
  -keyout server.key \
  -subj "/CN=localhost"

# Update docker-compose or pg_hba.conf
# Then update .env.local:
```

```bash
# .env.local
DB_SSLMODE=require
DB_SSL_ROOT_CERT=/path/to/server.crt
```

### Test Connection
```bash
# Start server - should connect successfully
go run main.go

# Check logs for SSL connection
# Should NOT see: "WARNING: Database SSL is disabled in production environment"
```

## 5. WebSocket Testing

### Test with wscat
```bash
# Install wscat
npm install -g wscat

# Connect to WebSocket (requires valid token)
wscat -c "ws://localhost:8080/api/notifications/ws" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Should receive:
# {"type":"connected","data":{"message":"Connected to EduHub notifications"},"timestamp":"..."}

# Send ping
> {"type":"ping"}

# Should receive pong
< {"type":"pong","timestamp":"..."}
```

### Test Multi-Tenant Isolation
```bash
# Connect two clients from different colleges
# Broadcast notification to college 1
# Verify college 2 client doesn't receive it
```

## 6. Performance Testing

### Load Test Authentication
```bash
# Install hey (HTTP load generator)
go install github.com/rakyll/hey@latest

# Load test login endpoint
hey -n 1000 -c 10 -m POST \
  -H "Content-Type: application/json" \
  -d '{"email":"test@college.com","password":"Test123!"}' \
  http://localhost:8080/auth/login

# Check response times and success rate
```

### Load Test Protected Endpoints
```bash
# Get auth token first
TOKEN="your_token_here"

# Load test dashboard
hey -n 1000 -c 10 \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/dashboard
```

## 7. Integration Testing Checklist

### Pre-Deployment Checklist

- [ ] Multi-tenant isolation verified with two different colleges
- [ ] JWT tokens expire correctly
- [ ] Token refresh works
- [ ] Database connects with SSL in production
- [ ] Errors are sanitized in production mode
- [ ] QR codes expire after 15 minutes
- [ ] QR codes reject different college attempts
- [ ] WebSocket requires authentication
- [ ] WebSocket isolates by college
- [ ] Quiz auto-grading works for all question types
- [ ] Assignment late penalties calculate correctly
- [ ] Report generation produces valid PDFs

### Test Data Cleanup
```bash
# After testing, clean up test data
# Truncate test tables or drop test database
psql -h localhost -U postgres -d eduhub_test -c "TRUNCATE TABLE students, colleges, courses CASCADE;"
```

## 8. Troubleshooting

### Common Issues

#### Test Database Connection Fails
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check connection string
echo $DB_HOST $DB_PORT $DB_USER $DB_NAME

# Test direct connection
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME
```

#### Playwright Tests Fail
```bash
# Ensure browsers are installed
bun run playwright install

# Check if server is running
curl http://localhost:8080/health

# Run with debug mode
DEBUG=pw:api bun run playwright test
```

#### WebSocket Connection Fails
```bash
# Check if WebSocket endpoint is accessible
curl http://localhost:8080/api/notifications/ws

# Verify token is valid
curl http://localhost:8080/api/dashboard \
  -H "Authorization: Bearer $TOKEN"
```

## 9. Continuous Integration Setup

### GitHub Actions Example
```yaml
name: Security Tests

on: [push, pull_request]

jobs:
  security-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      - name: Run Security Tests
        run: |
          cd server
          go test ./tests/security_test.go -v
      
      - name: Setup Bun
        uses: oven-sh/setup-bun@v1
      
      - name: Run E2E Tests
        run: |
          cd client
          bun install
          bun run playwright install --with-deps
          bun run playwright test tests/e2e/security.spec.ts
```

## 10. Monitoring in Production

### Key Metrics to Monitor

```bash
# Failed authentication attempts
grep "authentication failed" /var/log/eduhub/app.log | wc -l

# Multi-tenant isolation violations (should be 0)
grep "Invalid college" /var/log/eduhub/app.log

# Token refresh rate
grep "RefreshToken" /var/log/eduhub/app.log | wc -l

# WebSocket connections
curl http://localhost:8080/api/websocket/stats
```

---

## Quick Commands Reference

```bash
# Backend tests
go test ./tests/security_test.go -v

# E2E tests
bun run playwright test tests/e2e/security.spec.ts --ui

# Start services
docker-compose up -d
go run main.go

# Test multi-tenant
./scripts/test-multi-tenant.sh

# Load test
hey -n 1000 -c 10 http://localhost:8080/api/dashboard
```

---

## Support

If you encounter any issues:
1. Check server logs: `tail -f /var/log/eduhub/app.log`
2. Verify environment variables: `env | grep -E "DB_|APP_|JWT_"`
3. Test database connection: `psql -h $DB_HOST -U $DB_USER -d $DB_NAME`
4. Review SECURITY_AND_FEATURES.md for implementation details
