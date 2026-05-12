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
| PostgreSQL database       | External attackers, Malicious insiders | Stolen credentials or misconfiguration allows for direct database access |
| Session tokens| External attackers | Attacker exploits insecure cookies and the fact that we use http to steal session tokens and hijack user accounts |
| Docker Swarm infrastructure | Attackers, misconfiguration | An exposed swarm manager would allow for full cluster takeover |
| Monitoring system | External attackers | Unauthorized access to Grafana exposes internal service endpoints, system metrics, and logs, aiding further attacks |
| DigitalOcean droplets | External attackers, leaked SSH key, weak firewall | Unauthorized SSH access leads to full server compromise |
| Ingress/points managers | External attackers | Misconfigured ingress exposes internal services (e.g. database or monitoring endpoints) to the public internet |
| Secrets/config data| External attackers, repository leaks | Exposed API keys allow direct access to services |
| User credentials | External attackers | Weak password hashing allows attackers to crack stolen hashes and reuse credentials |

## Risk Analysis

Below we define the scale for our risk matrix for our analysis. Then we analyse the risk of the scenarios and describe possible ways of mitigating risk.

### Likelihood Scale

| Level | Description |
|-------|-------------|
| Low (1) | Difficult to exploit, requires special conditions |
| Medium (2) | Possible with some effort |
| High (3) | Easy to exploit, common attack |


### Impact Scale

| Level | Description |
|-------|-------------|
| Low (1) | Minor impact, no sensitive data |
| Medium (2) | Partial data exposure or service disruption |
| High (3) | Major data breach or full system compromise |

### Risk Matrix

Risk Score = Likelihood × Impact

| Score | Risk Level |
|-------|------------|
| 1–2 | Low |
| 3–4 | Medium |
| 6 | High |
| 9 | Critical |

### Risk Evaluation and Mitigation

| Asset | Scenario | Likelihood | Impact | Score | Risk Level | Mitigation Strategy |
|-------|----------|------------|--------|-------|------------|---------------------|
| Backend API | Injection vulnerability | 1 | 3 | 3 | Medium | Use GORM (ORM) with parameterized queries, input validation, and escaping html input |
| PostgreSQL database | Credential compromise | 2 | 3 | 6 | High | Restrict database access to internal network, use strong credentials, do not expose database publicly |
| Session tokens | Token theft | 2 | 3 | 6 | High | Use secure cookies, HTTPS, short expiry |
| Docker Swarm | Cluster takeover | 2 | 3 | 6 | High | Do not expose Swarm manager, restrict access via firewall, secure configuration |
| Monitoring system | Exposure of information | 2 | 2 | 4 | Medium | Require authentication for dashboards, limit access |
| DigitalOcean droplets | SSH compromise | 1 | 3 | 6 | High | Use SSH key-based authentication, apply firewall restrictions |
| Secrets/config | Credential leak | 1 | 3 | 9 | Critical | Store secrets securely - "Secrets belpng in vaults", Use enviroment variables |
| Ingress | Service exposure | 2 | 2 | 4 | High | Stop exposing ports that should not be exposed publically |
| User credentials | Weak hashing | 1 | 3 | 6 | High | Use strong hashing + salting |

### Additional points

- Monitoring is used to detect abnormal behavior and support incident response.
- Security should be implemented using a defense-in-depth approach, where multiple controls protect each critical asset.
- We should implement backups

## Residual Risk

After mitigation we still have to look out for

- New or unknown vulnerabilities
- Misconfigurations
- Human error

## Conclusion

Key findings:

- The most critical risks involve the backend API and sensitive credentials
- Infrastructure misconfiguration presents significant risk
- Proper security controls significantly reduce overall risk exposure
