# Auth Service - Grafana Dashboards

This folder contains all Grafana dashboards for the Auth Service monitoring.

## Dashboard Files

### Main Dashboards
- `auth-overview.json` - High-level service overview with key metrics
- `auth-complete-monitoring.json` - Comprehensive monitoring with 14 panels
- `final-dashboard.json` - Complete metrics dashboard with real data

### Specialized Dashboards  
- `advanced-dashboard.json` - Advanced monitoring with 14 detailed panels
- `security-dashboard.json` - Security and alerts focused dashboard
- `working-dashboard.json` - Basic working dashboard with core metrics

### Development Dashboards
- `comprehensive-dashboard.json` - Development comprehensive dashboard
- `create-dashboard.json` - Basic dashboard template

## Dashboard Categories

### 📊 **Overview Dashboard**
- Service health status
- Key performance indicators
- Request volume trends
- Authentication activity

### 🔍 **Detailed Monitoring**
- HTTP request metrics by endpoint
- Response time percentiles
- JWT token operations
- Database performance
- Cache performance
- Memory and runtime metrics

### 🔒 **Security Monitoring**
- Failed login attempts
- Authentication success rates
- Token security operations
- Error rate monitoring
- SLA compliance tracking

## Usage

Import these dashboards into Grafana:
1. Open Grafana UI
2. Go to Dashboards → Import
3. Upload the JSON files
4. Configure Prometheus datasource

## Metrics Covered

- **HTTP Metrics**: Request rate, response time, status codes
- **Authentication**: Login success/failure, user registrations
- **JWT Operations**: Token validation, refresh, blacklist
- **Infrastructure**: Redis cache, MongoDB operations
- **System**: Memory usage, goroutines, GC performance
- **Security**: Failed attempts, error rates, availability