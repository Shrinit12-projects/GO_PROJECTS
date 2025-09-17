# Auth Service - Monitoring Stack

Complete monitoring setup for the Auth Service including Grafana dashboards, Prometheus configuration, and metrics collection.

## Directory Structure

```
monitoring/
├── grafana/
│   ├── dashboards/           # Grafana dashboard JSON files
│   │   ├── auth-overview.json
│   │   ├── auth-complete-monitoring.json
│   │   ├── advanced-dashboard.json
│   │   ├── security-dashboard.json
│   │   └── README.md
│   └── provisioning/         # Grafana provisioning configs
│       ├── dashboards/
│       └── datasources/
├── prometheus/
│   └── prometheus.yml        # Prometheus scraping config
└── README.md
```

## Components

### 🎯 **Grafana Dashboards**
- **Overview**: High-level service health and KPIs
- **Complete Monitoring**: 14-panel comprehensive dashboard
- **Advanced**: Detailed performance metrics
- **Security**: Authentication and security monitoring

### 📊 **Prometheus Configuration**
- Service discovery for auth-service:8080
- Metrics scraping every 10 seconds
- Custom auth service metrics collection

### 📈 **Metrics Collected**
- HTTP request metrics (rate, duration, status codes)
- Authentication metrics (login success/failure)
- JWT token operations (validate, refresh, blacklist)
- Database operations (MongoDB, Redis)
- System metrics (memory, goroutines, GC)
- Business metrics (user registrations)

## Usage

### Local Development
```bash
# Start with monitoring
docker-compose up --build

# Access dashboards
open http://localhost:3000  # Grafana (admin/admin)
open http://localhost:9090  # Prometheus
```

### Production Deployment
1. Copy monitoring configs to deployment environment
2. Update Prometheus targets for production endpoints
3. Configure Grafana datasources
4. Import dashboard JSON files

## Dashboard Import

1. Open Grafana → Dashboards → Import
2. Upload JSON files from `grafana/dashboards/`
3. Configure Prometheus datasource: `http://prometheus:9090`
4. Verify metrics are flowing

## Metrics Endpoints

- **Auth Service**: `http://auth-service:8080/metrics`
- **Prometheus**: `http://prometheus:9090/metrics`
- **Grafana**: `http://grafana:3000/metrics`