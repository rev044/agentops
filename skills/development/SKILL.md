---
name: development
description: >
  Use when: "API", "backend", "frontend", "fullstack", "mobile", "iOS", "React Native",
  "Flutter", "deploy", "CI/CD", "Docker", "Kubernetes", "GitHub Actions", "LLM",
  "RAG", "prompt engineering", "microservices", "database schema", "SwiftUI".
version: 1.0.0
author: "AgentOps Team"
license: "MIT"
---

# Development Skill

Software development patterns for backend, frontend, mobile, deployment, and AI engineering.

## Quick Reference

| Area | Key Patterns | When to Use |
|------|--------------|-------------|
| **Backend** | REST APIs, microservices, database schema | Server-side systems |
| **Frontend** | React, state management, accessibility | Web applications |
| **Fullstack** | End-to-end features, API integration | Complete features |
| **Mobile** | React Native, Flutter, cross-platform | Mobile apps |
| **iOS** | Swift/SwiftUI, iOS 18 features | Native iOS apps |
| **Deployment** | CI/CD, Docker, Kubernetes, GitHub Actions | Infrastructure |
| **AI** | LLM integration, RAG, prompt pipelines | AI-powered apps |
| **Prompts** | System prompts, agent optimization | LLM behavior |

---

## Backend Architecture

### Focus Areas
- RESTful API design with proper versioning and error handling
- Service boundary definition and inter-service communication
- Database schema design (normalization, indexes, sharding)
- Caching strategies and performance optimization
- Security patterns (auth, rate limiting)

### Approach
1. Start with clear service boundaries
2. Design APIs contract-first
3. Consider data consistency requirements
4. Plan for horizontal scaling from day one
5. Keep it simple - avoid premature optimization

### Output
- API endpoint definitions with example requests/responses
- Service architecture diagram (mermaid or ASCII)
- Database schema with key relationships
- Technology recommendations with rationale
- Bottlenecks and scaling considerations

---

## Frontend Development

### Focus Areas
- React and modern component patterns
- State management (Redux, Zustand, Context)
- Performance optimization (memoization, code splitting)
- Accessibility (WCAG compliance)
- Responsive design and CSS-in-JS

### Approach
1. Component-first architecture
2. Lift state only when necessary
3. Optimize renders with React.memo, useMemo
4. Test with React Testing Library
5. Accessibility from the start

### Output
- React components with TypeScript
- State management setup
- Performance optimizations
- Accessibility compliance
- Responsive layouts

---

## Fullstack Development

### Focus Areas
- End-to-end feature implementation
- API integration and data flow
- Database design and ORM usage
- Authentication and authorization
- Error handling across the stack

### Approach
1. Define data models first
2. Build API endpoints with validation
3. Create frontend components that consume APIs
4. Handle loading, error, and success states
5. Test the full flow

### Output
- Complete feature implementation
- API contracts (OpenAPI)
- Database migrations
- Frontend components
- Integration tests

---

## Mobile Development

### Focus Areas
- React Native and Flutter patterns
- Cross-platform code sharing
- Native module integration
- Offline-first architecture
- App store optimization

### Approach
1. Share business logic, customize UI
2. Use platform-specific navigation
3. Implement offline support early
4. Optimize for battery and data usage
5. Test on real devices

### Output
- Cross-platform mobile code
- Native integrations where needed
- Offline sync strategy
- Platform-specific optimizations

---

## iOS Development

### Focus Areas
- Swift and SwiftUI patterns
- iOS 18 features and APIs
- Core Data and persistence
- UIKit integration when needed
- App Store guidelines

### Approach
1. SwiftUI-first, UIKit when necessary
2. Use Combine for reactive programming
3. Leverage async/await for concurrency
4. Follow Apple Human Interface Guidelines
5. Test with XCTest and UI testing

### Output
- SwiftUI views and modifiers
- Swift data models
- Core Data schema
- App lifecycle handling
- TestFlight configuration

---

## Deployment Engineering

### Focus Areas
- CI/CD pipeline configuration
- Docker containerization
- Kubernetes deployment manifests
- GitHub Actions workflows
- Infrastructure as Code

### Approach
1. Containerize everything
2. Use multi-stage Docker builds
3. Implement proper health checks
4. Automate testing in pipelines
5. Use GitOps for deployments

### Output
- Dockerfiles with best practices
- Kubernetes manifests (or Helm charts)
- GitHub Actions workflows
- Environment configuration
- Deployment documentation

### Common Patterns

```yaml
# Multi-stage Dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/dist ./dist
CMD ["node", "dist/index.js"]
```

```yaml
# GitHub Actions workflow
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
      - run: npm ci
      - run: npm test
```

---

## AI Engineering

### Focus Areas
- LLM integration (OpenAI, Anthropic, local models)
- RAG systems with vector databases (Qdrant, Pinecone, Weaviate)
- Prompt engineering and optimization
- Agent frameworks (LangChain, LangGraph, CrewAI patterns)
- Token optimization and cost management

### Approach
1. Start with simple prompts, iterate based on outputs
2. Implement fallbacks for AI service failures
3. Monitor token usage and costs
4. Use structured outputs (JSON mode, function calling)
5. Test with edge cases and adversarial inputs

### Output
- LLM integration code with error handling
- RAG pipeline with chunking strategy
- Prompt templates with variable injection
- Vector database setup and queries
- Evaluation metrics for AI outputs

---

## Prompt Engineering

### Focus Areas
- System prompt optimization
- Few-shot example selection
- Chain-of-thought prompting
- Output format control
- Token efficiency

### Approach
1. Clear role definition in system prompt
2. Use examples to demonstrate desired behavior
3. Request structured output (JSON, markdown)
4. Test with diverse inputs
5. Version control prompts

### Output
- System prompts for agents
- Prompt templates with variables
- Evaluation criteria
- A/B testing setup
- Cost analysis
