# DarwinFlow

**A learning workflow framework that transforms personal AI assistants from static prompt executors into adaptive systems that improve through use.**

## The Problem

When you use AI assistants for repetitive work:
- ❌ You provide the same context repeatedly
- ❌ You make the same corrections over and over
- ❌ You pay for the same reasoning each time
- ❌ The AI never learns from your corrections

**Current AI assistants have amnesia.** They don't get better with use.

## The Solution

DarwinFlow learns from your corrections:

1. **You work normally** - Use AI assistant as you do today
2. **System observes** - Captures what happens during execution
3. **You reflect** - Manually trigger reflection on recent work
4. **LLM analyzes** - Finds patterns in logs and corrections
5. **New version created** - System generates improved workflow automatically
6. **You choose** - Use v1, v2, or v3 - whichever works best

**Workflow versions emerge from your actual usage patterns.**

## Example

### Using v1 (Initial Workflow)
```
You: "Add Telegram bot support"
AI: Starts implementing
You: "Wait, read vision.md first to classify this"
AI: Reads vision.md, classifies, continues
Task completes in 20 minutes

[All logged to logs/runs/run_abc123.json]
```

### After Using v1 for a While
```
You: darwinflow reflect --workflow workflows/assistant_v1.yaml --recent-runs 20

[LLM analyzes logs]
Pattern found: "Read vision.md first" corrected 3 times

[Creates workflows/assistant_v2.yaml automatically]
[Creates workflows/assistant_v2_reasoning.md]

New version available: v2 now reads vision.md before starting
```

### Using v2 (Evolved Workflow)
```
You: "Add Slack integration"
    (using workflows/assistant_v2.yaml)

AI: [Automatically reads vision.md first]
AI: "Based on vision.md, this is an Integration..."
Task completes in 12 minutes, no correction needed!
```

## Key Benefits

- **Fewer corrections**: Measurable reduction over time
- **Faster completion**: Significant improvement for repeated tasks
- **Lower costs**: Learned patterns replace expensive reasoning
- **Better quality**: Consistent approach from learned behaviors
- **Personalized**: Workflow adapts to *your* specific needs

## How It's Different

### vs Workflow Automation (Zapier, n8n)
- **Them**: Configure workflows upfront
- **Us**: Workflows emerge from your actual corrections

### vs Prompt Engineering
- **Them**: Write better prompts manually
- **Us**: System learns what context to provide

### vs Fine-tuning
- **Them**: Train models (expensive, complex)
- **Us**: Adapt workflows (fast, transparent)

### vs Current AI Assistants
- **Them**: Static, no learning
- **Us**: Improves every time you use it

## Architecture

### Three Layers

1. **Core Framework** (domain-agnostic)
   - Workflow execution and learning engine
   - Pattern recognition from corrections
   - Workflow evolution mechanisms

2. **Integrations** (reusable tools)
   - Communication platforms (Telegram, Slack)
   - Development tools (GitHub, file systems)
   - LLM providers and AI services

3. **Reference Workflows** (domain demonstrations)
   - Software engineering workflows
   - Product management workflows
   - Architecture review workflows

**Philosophy**: Never hardcode domain logic. Provide primitives others build upon.

## Current Status

**Status**: Requirements phase - MVP specification complete

### Next Steps
1. **Step 1**: Build workflow execution + logging (file-based)
2. **Step 2**: Add metrics tracking (separate from logs)
3. **Step 3**: Build reflection command (LLM analyzes → creates new versions)
4. **Dogfood**: Use DarwinFlow to build DarwinFlow itself

## Quick Start

### For Users (Coming Soon)
```bash
# Install
pip install darwinflow

# Run workflow
darwinflow run --workflow workflows/assistant_v1.yaml "Your task here"

# After some usage, reflect to create improved versions
darwinflow reflect --workflow workflows/assistant_v1.yaml --recent-runs 20
```

### For Developers
```bash
# Clone
git clone https://github.com/yourusername/darwinflow-pub.git
cd darwinflow-pub

# Read MVP specification
cat docs/mvp_simple.md

# Read product vision
cat docs/product/vision.md
```

## Documentation

- **[MVP Specification](docs/mvp_simple.md)** - Complete MVP description (start here)
- **[Product Vision](docs/product/vision.md)** - Overall strategy and principles
- **[Initial Workflow](workflows/assistant_v1.yaml)** - Simple starting workflow

## Use Cases

### Software Engineering
Learn to:
- Always check architecture docs before features
- Check recent commits before debugging
- Classify changes as Core/Integration/Workflow
- Run specific test suites after certain changes

### Product Management
Learn to:
- Check analytics before feature requests
- Validate against target persona
- Follow prioritization framework
- Maintain consistent documentation

### Any Repetitive Knowledge Work
If you:
- Do similar tasks repeatedly
- Provide same context multiple times
- Make same corrections frequently
- Want AI to learn your style

Then DarwinFlow helps.

## Success Metrics

### MVP Success
- 3+ workflow evolutions from real corrections
- 30%+ reduction in correction frequency
- 30%+ reduction in task completion time
- User confirms system "getting better"

### Long-term Vision
- 70%+ reduction in corrections
- 10x cost reduction (learned vs reasoning)
- Community of shared workflows
- Multiple domain workflows

## Principles

1. **Learn, Don't Hardcode** - Solutions emerge from patterns
2. **General Over Specific** - Framework capabilities over domain solutions
3. **Efficiency Through Learning** - Every interaction reduces future cost
4. **Human-Guided Evolution** - Learn from usage, user chooses versions
5. **Transparent Learning** - All workflow changes visible and understandable

## Contributing

DarwinFlow is in early development. We're focused on:
1. Building the MVP (workflow execution + logging + reflection)
2. Validating the learning loop works
3. Proving measurable improvement through dogfooding

Interested in contributing? Start by:
- Reading `docs/mvp_simple.md` - the complete MVP spec
- Understanding `docs/product/vision.md` - the product vision
- Trying the system when ready
- Sharing feedback on the learning approach

## License

[To be determined - likely MIT or Apache 2.0]

## Project Status

**Phase**: MVP specification complete, implementation starting
**Approach**: Dogfooding (using DarwinFlow to build DarwinFlow)

## Contact

- **Issues**: [GitHub Issues](https://github.com/yourusername/darwinflow-pub/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/darwinflow-pub/discussions)

---

**The future of AI assistants isn't better prompts. It's systems that learn from your corrections.**

**That's DarwinFlow.**
