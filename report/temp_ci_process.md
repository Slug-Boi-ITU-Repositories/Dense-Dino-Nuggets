<!--
This file is a temp file so we don't get merge conflicts on report.md will merge into that one later
 -->

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

#### The Test Pipeline

Below is a flowchart showing the Test CI pipeline. The start is on the left when a developer pushes code to a PR or a push happens on the branch Main (e.g. when code is merged or rebased onto it). The pipeline uses Dagger which is mentioned in our depenedency tables above. This allows us to create multiplatform workflows (that is workflows that will work on any platform that has some type of action runner) with minimal setup. All the platform specific workflow has to do is start the dagger engine and run the checks.

![test_CI_pipeline.png](img/test_CI_pipeline.png)

#### The Release Pipeline

Below is a flowchart showing the Release CI pipeline

![]()

### Monitoring of Minitwit



### Logging
