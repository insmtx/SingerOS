# SingerOS

## Enterprise Digital Workforce Operating System

> Build, orchestrate and govern AI-powered digital employees for enterprise.

---

## 🚀 What is SingerOS?

**SingerOS** is an enterprise-grade Multi-Agent Operating System designed to power the next generation of digital workforce.

It is not a chatbot framework.
It is not a simple workflow engine.

SingerOS is:

> A distributed, governance-first AI execution system for enterprise digital transformation.

SingerOS enables organizations to:

* Design AI-powered digital employees
* Orchestrate multi-agent workflows
* Govern skills, models, and permissions
* Run intelligent task execution pipelines
* Operate in both private enterprise environments and SaaS sandbox mode

---

## 🧠 Why SingerOS?

Traditional workflow systems focus on deterministic task automation.

Modern enterprises require:

* Intelligent decision-making
* Cross-system reasoning
* Multi-agent collaboration
* Cost-aware model routing
* Auditable AI execution
* Enterprise-grade governance

SingerOS is built to meet these needs.

Compared to traditional workflow engines such as DeerFlow:

* SingerOS embeds cognitive agents into workflows
* SingerOS includes model routing and cost governance
* SingerOS enforces Skill isolation via Skill Proxy
* SingerOS supports multi-tenant enterprise deployment
* SingerOS is designed as an AI OS, not just a flow engine

---

## 🎯 Design Principles

SingerOS enforces strict architectural invariants to ensure governance and reliability:

1. **Agent never directly calls external systems** - All external interactions go through Tools
2. **Skill never performs orchestration logic** - Skills compose Tools, not workflows
3. **Control plane never executes runtime logic** - Clear separation of concerns
4. **All workflow execution must be persisted** - Replayable and auditable
5. **All model usage must be measurable** - Cost-aware and governable

For detailed design philosophy, see [Design Philosophy](docs/DESIGN_PHILOSOPHY.md).

---

## 🏢 Target Scenarios

SingerOS is designed for:

### Enterprise Internal Digital Transformation

* Digital employees for operations
* Intelligent approval systems
* Automated reporting
* Cross-system workflow automation
* AI-assisted decision engines

### SaaS Sandbox Mode

* Demonstration environments
* Trial accounts
* Limited skill library
* Token quota enforcement
* No sensitive system integration

---

## 🔐 Enterprise-First Capabilities

* Multi-tenant isolation
* RBAC access control
* Audit logs
* Skill-level permission control
* Cost tracking
* SLA-aware execution
* Private deployment support

---

## 🔄 Execution Flow

SingerOS follows a unified event-driven execution model:

```
User → Event Gateway → EventBus → Control Plane → Orchestrator 
→ Runtime Manager → Agent/Edge Runtime → Skill → Tool → EventBus → Client
```

All execution is:

* **Replayable** - Complete execution history recorded
* **Observable** - Full链路 tracing and monitoring
* **Auditable** - Comprehensive audit logs

For detailed architecture, see [Architecture Documentation](docs/ARCHITECTURE.md).

---

## 🧩 Extensibility

SingerOS supports plugin-based architecture:

* Skill plugins
* Agent templates
* Model providers
* Memory backends
* Workflow templates

All plugins must be:

* Versioned
* Isolated
* Auditable

---

## 🛣 Roadmap

### Phase 1 – Core Execution Layer

* DAG execution engine
* Agent runtime
* Skill proxy
* Model router
* Multi-tenant basics

### Phase 2 – Enterprise Intelligence

* Cross-agent collaboration
* Cost optimization engine
* Distributed scheduler
* Observability suite

### Phase 3 – AI OS Evolution

* Agent federation
* Autonomous optimization
* Workflow marketplace
* Digital workforce marketplace

---

## ⚠ Non-Goals

SingerOS is NOT:

* A prompt playground
* A simple chatbot UI
* A research-only autonomous agent simulator
* A decentralized AI experiment

---

## 🧬 Philosophy

SingerOS treats AI agents as:

> First-class digital employees with governance, accountability, and operational boundaries.

We believe the future enterprise stack will include:

* Human employees
* Software systems
* Digital employees (AI Agents)

SingerOS is designed to operate the third category.

---

## 📜 License

(To be determined — Apache 2.0 / Commercial Hybrid / Enterprise License)

---

---

## 📚 Documentation

Complete documentation is available in the `docs/` directory:

| Document | Description |
|----------|-------------|
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | AI OS architecture design (v2 - Three-Plane Model) |
| [DESIGN_PHILOSOPHY.md](docs/DESIGN_PHILOSOPHY.md) | Core design philosophy and principles |
| [PRD.md](docs/PRD.md) | Product requirements (Employee View/AI Workbench) |
| [GITHUB_AUTH_SETUP.md](docs/GITHUB_AUTH_SETUP.md) | GitHub OAuth integration guide |
| [GITHUB_WEBHOOK_TROUBLESHOOTING.md](docs/GITHUB_WEBHOOK_TROUBLESHOOTING.md) | GitHub webhook troubleshooting |
| [PR_EVENT_FLOW.md](docs/PR_EVENT_FLOW.md) | GitHub PR event processing verification |
| [TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) | Common issues and solutions |

---

## 🤝 Contributing

We welcome:

* Skill plugins
* Model adapters
* Workflow templates
* Observability integrations
* Security enhancements

Enterprise partners are welcome to collaborate.

---

## 🐶 Why “Singer”?

Singers are:

* Expressive
* Adaptive
* Highly disciplined
* Excellent collaborators

SingerOS aims to embody the same traits in enterprise AI systems.

