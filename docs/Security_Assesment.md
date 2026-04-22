# Security Assessment

The following is a risk assessment of our Minitwit application.

## Risk Identification

### Assets

- Web frontend application
- Backend API
- PostgreSQL database
  - Database server
  - Schemas and tables
- Session tokens
- Docker
  - Docker images
  - Docker containers
  - Swarm manager node
- Monitoring system
  - Prometheus metrics system
  - Grafana dashboards and monitoring interface
- DigitalOcean droplets
- Secrets and Configuration data
  - .env files
  - API keys, database credentials
  - TLS certificates (not implemented yet)

## Table of threat sources and risk scenarios  

| Asset                     | Threat Source | Risk Scenario |
|---------------------------|---------------|---------------|
| Web frontend application  |               |               |
| Backend API               |               |               |
| Database server           |               |               |
| Database schemas and tables        |               |               |
| Session tokens            |               |               |
| Docker images             |               |               |
| Docker containers         |               |               |
| Docker swarm manager node        |               |               |
