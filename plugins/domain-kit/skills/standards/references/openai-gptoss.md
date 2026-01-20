# OpenAI GPT-OSS Standards

<!-- Last synced: 2026-01-19 -->

> **Purpose:** GPT-OSS-specific patterns. For tool calling, reasoning effort, and agent loops, see [Reasoning Models](./openai-reasoning.md).

## Quick Reference

| Feature | GPT-OSS-120B | GPT-OSS-20B |
|---------|--------------|-------------|
| **Total Parameters** | 117B | 21B |
| **Active Parameters** | 5.1B (MoE) | 3.6B (MoE) |
| **Context Window** | 128K | 128K |
| **Minimum GPU** | 80GB (H100) | 16GB (RTX 4080) |
| **License** | Apache 2.0 | Apache 2.0 |
| **Response Format** | **Harmony REQUIRED** | **Harmony REQUIRED** |

---

## CRITICAL: Harmony Response Format

**GPT-OSS models WILL NOT WORK without Harmony format.** This is non-negotiable.

```python
# REQUIRED on every request
response = client.responses.create(
    model="gpt-oss-120b",
    input="Your prompt here",
    response_format={"type": "harmony"},  # MANDATORY
)
```

Without Harmony, models produce garbled/incomplete responses.

---

## Deployment Options

### Self-Hosted (vLLM)

```bash
# Single H100 (120B)
python -m vllm.entrypoints.openai.api_server \
    --model openai/gpt-oss-120b \
    --dtype bfloat16 \
    --response-format harmony

# Consumer GPU (20B)
python -m vllm.entrypoints.openai.api_server \
    --model openai/gpt-oss-20b \
    --dtype float16 \
    --response-format harmony
```

### Cloud Providers

| Provider | Endpoint |
|----------|----------|
| Groq | `https://api.groq.com/openai/v1` |
| Together | `https://api.together.xyz/v1` |
| Fireworks | `https://api.fireworks.ai/inference/v1` |

---

## Hardware Requirements

| Model | VRAM | Examples |
|-------|------|----------|
| GPT-OSS-120B | 80GB | H100, A100-80GB |
| GPT-OSS-120B (TP=4) | 4x 24GB | 4x RTX 4090 |
| GPT-OSS-20B | 16GB | RTX 4080, A10, L4 |

---

## Common Errors

| Symptom | Cause | Fix |
|---------|-------|-----|
| Garbled output | Missing Harmony | Add `response_format={"type": "harmony"}` |
| OOM on GPU | Model too large | Use 20B or tensor parallelism |
| "Unknown model" | Wrong model name | Use `gpt-oss-120b` or `gpt-oss-20b` |

---

## Anti-Patterns

| Name | Pattern | Instead |
|------|---------|---------|
| No Harmony | Standard chat completions | Always use `response_format: harmony` |
| Single GPU 120B | Running on <80GB VRAM | Use 20B or multi-GPU |
| Float32 | Full precision inference | Use bfloat16/float16 |

---

## AI Agent Guidelines

| Guideline | Rationale |
|-----------|-----------|
| **ALWAYS use Harmony format** | Models malfunction without it |
| **NEVER use chat.completions** | Use responses API with Harmony |
| **PREFER 20B for development** | Faster iteration on consumer hardware |
| For tool calling, reasoning effort, agent loops | See [openai-reasoning.md](./openai-reasoning.md) |
