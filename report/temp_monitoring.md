For monitoring we used these technologies:
- Grafana
- Prometheus
- Loki

### Dashboard structure
This is how our minitoring dashboard looks
![minitwit_dashboard.png](./images/minitwit_dashboard.png)

Data is gathered from the minitwit application by pulling with with Prometheus. Logs are gotten by using the loki logging driver on our minitwit containers, which push to an aggregator container on the monitoring machine. Grafana can then pull data from Prometheus and Loki to display in the dashboard. The dashboard shows:
- If the server is running or is down
- The 99'th and 99.9'th percentile of response time to see if we have requests that are taking longer than they should
- The average time requests take to know how the system is responding in general
- The amount of requests pr second too see the load on the system 
- The logs from minitwit too see what is happening on the system

### Alerting
We set up an alert that used Grafana's build in Discord web hook, such that a Discord bot would send a message in our Discord server in case the minitwit system went down or was unreachable by Prometheus. We tested this artificially, but never in practice since the system didn't go down after we added the alert.

### Monitoring issues
One large issue we ran into was that our Promethus setup wasn't collecting accurate infromation from the minitwit system, after we started using swarm to have multiple replicas. This was an unfortunate side effect of Prometheus being a pull system. Because, when Prometheus tried to pull data if would only get the monitoring data from one of the replicas. To solve this we needed to move over to a push system, where each of the replicas would need a system to push their monitoring data to Prometheus such that it would be able to aggregate data from all the replicas, or show data for each replica. However, we didn't have the time to implement this as there where other things that took higher priority, so this never reached the top of the priority list.