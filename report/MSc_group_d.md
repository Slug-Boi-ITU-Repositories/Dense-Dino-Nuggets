# DevOps Group D

| Name    | Email |
| -------- | ------- |
| August Kofoed Brandt  | aubr@itu.dk |
| Emilia Victoria Helsted | ehel@itu.dk |
| Niklas Zeeberg Hessner Christensen | nizc@itu.dk |
| Philip Guozhi Han Pedersen | phgp@itu.dk |
| Theis Per Holm | thph@itu.dk |

## System's Perspective
<!---
Here we need:
A description and illustration of the:

- Design and architecture of your ITU-MiniTwit systems.

- All dependencies of your ITU-MiniTwit systems on all levels of abstraction and development stages. That is, list and briefly describe all technologies and tools you applied and depend on.

- Describe the current state of your systems, for example using results of static analysis and quality assessments.
-->
### Design and Architecture

### Dependencies

Our Minitwit uses the following dependencies:

#### Core dependencies

| Dependency   | Description |
| -------- | ------- |
| Golang  |  |
| PostgreSQL |  |
| Nginx |  |
| Docker |  |
| DigitalOcean |  |
| Vagrant | (Deprecated?) |
| OpenTofu ||
| Certbot ||


#### CI and Quality tooling

| Dependency   | Description |
| -------- | ------- |
| Dagger |  |
| Playwright ||
| misspell ||
| golangci-lint | |
| CodeQL ||
| SonarCloud ||
| Codacy ||

#### Monitoring (and logging?)

| Dependency   | Description |
| -------- | ------- |
| Grafana |  |
| Loki ||
| Prometheus ||

### Current State of our Minitwit

## Process' perspective
 <!---
This perspective should clarify how code or other artifacts come from idea into the running system and everything that happens on the way.

In particular, the following descriptions should be included:

- A complete description and illustration of stages and tools included in the CI/CD pipelines, including deployment and release of your systems.

- How do you monitor your systems and what precisely do you monitor?

- What do you log in your systems and how do you aggregate logs?

- Brief description of how you security hardened your systems.

- How do you handle availability and scaling in your systems?
-->

### CI/CD Pipelines

### Monitoring of Minitwit

### Logging

### Hardening of Minitwit

### Availability and Scaling

## Reflection Perspective
<!---
Describe the biggest issues, how you solved them, and which are major lessons learned with regards to:

- evolution and refactoring
- operation
- maintenance
of your ITU-MiniTwit systems. Link back to respective commit messages, issues, tickets, etc. to illustrate these.

Also reflect and describe what was the "DevOps" style of your work. For example, what did you do differently to previous development projects and how did it work?
-->

## Use of Generative AI
<!---
ITU's rules on the use of generative AI apply for this report too. They are described https://itustudent.itu.dk/Study%20Administration/Generative%20AI#Guidelines and in detail https://itustudent.itu.dk/-/media/ITU-Student/Study-Administration/GAI/Generative-AI-guidelines-for-students-Spring-2026-pdf.pdf. Please follow them. For your report that means that you have to state which generative AI tools have been used for which task(s) in your projects. Additionally, describe how generative AI tools have been used and briefly reflect and discuss how they supported or hindered your work process.

From the guidelines:
• State which generative AI technology has been used.
• Describe how generative AI technology has been used.
-->
We have used the following models while working on our Minitwit:

- ChatGPT
- DeepSeek
- Claude

<!---
Just some flowy thoughts:
I belive we accidently set up CodePilot reviews for some pull requests?

In general we have used generative AI to help explain topics and technologies so we could better understand them. 

When refactoring the tests from the original Minitwit, ChatGPT was used for debugging and identifying where the issue was, when there were issues with flashes not being set correctly in main.

For tasks such as translating our vagrant file of over 500 lines of code to an ansible it seemed smart to use generative AI. We did however run into problems with ChatGPT directly translating the code rather that using the perks of ansible to write "good" code. We had to spend a lot of time on fixing the generated code and while there was some learning in it we could have gotten more out of doing it from scratch
-->
