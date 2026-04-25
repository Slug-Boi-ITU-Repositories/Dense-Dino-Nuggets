# Security Assessment

The following is a risk assessment of our Minitwit application. 

## Risk Identification

### Assets

- Web frontend application
- Backend API
- PostgreSQL database
  - Database server
  - Schemas and tables
- Session tokens (Removed? -> Cookies)
- Docker
  - Docker images
  - Docker containers
  - Swarm manager node
- Monitoring system
  - Prometheus metrics system
  - Grafana dashboards and monitoring interface
  - Loki
- DigitalOcean droplets
- Ingress points/managers
- Secrets and Configuration data
  - .env files
  - API keys, database credentials
  - TLS certificates
- User credentials
  - passwords  

### Table of threat sources and risk scenarios  

| Asset                     | Threat Source | Risk Scenario |
|---------------------------|---------------|---------------|
| Backend API               | External attackers, Malicious users | Attacker exploits an injection vulnerability in an API endpoint to access or modify user data in the database. |
| PostgreSQL databas        | External attackers, Malicious insiders | Stolen credentials or misconfiguration allows for direct database access |
| Session tokens| External attackers | Attacker exploits XSS or insecure cookies to steal session tokens and hijack user accounts |
| Docker Swarm infrastructure | Attackers, misconfiguration | An exposed swarm manager would allow for full cluster takeover |
| Monitoring system | External attackers | Unauthorized access to Grafana exposes internal service endpoints, system metrics, and logs, aiding further attacks |
| DigitalOcean droplets | External attackers, leaked SSH key, weak firewall | Unauthorized SSH access leads to full server compromise |
| Ingress/points managers | External attackers | Misconfigured ingress exposes internal services (e.g. database or monitoring endpoints) to the public internet |
| Secrets/config data| External attackers, repository leaks | Exposed API keys allow direct access to services |
| User credentials | External attackers | Weak password hashing allows attackers to crack stolen hashes and reuse credentials |

## Risk Analysis

Below we define the scale for our risk matrix for our analysis.

### Likelihood Scale

| Level | Description |
|------|------------|
| Low (1) | Difficult to exploit, requires special conditions |
| Medium (2) | Possible with some effort |
| High (3) | Easy to exploit, common attack |


### Impact Scale

| Level | Description |
|------|------------|
| Low (1) | Minor impact, no sensitive data |
| Medium (2) | Partial data exposure or service disruption |
| High (3) | Major data breach or full system compromise |

### Risk Matrix

Risk Score = Likelihood × Impact

| Score | Risk Level |
|------|------------|
| 1–2 | Low |
| 3–4 | Medium |
| 6 | High |
| 9 | Critical |

