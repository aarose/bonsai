# Bonsai
🌳 Bonsai is a CLI tool and for managing LLM conversation trees. AI's in the name.

![Bonsai logo](assets/bonsai.png)

It provides a structured way to explore multiple directions of a conversation, experiment freely, and return to earlier points to continue down different paths without friction. Think of it as Git for your LLM sessions: a system that lets you branch, reseed, and graft ideas without losing track of the bigger picture.

With Bonsai, you can:
- Maintain clean histories
- Compare alternate paths
- Keep your conversations organized as they grow
- Use the visualization tool to better understand the tree

The result is a workflow where creativity and control coexist, making it easy to cultivate ideas, revisit roots, and guide your conversations toward meaningful
outcomes.

## Installation

```bash
go install github.com/aarose/bonsai@latest
```

Or build from source:
```bash
git clone https://github.com/aarose/bonsai.git
cd bonsai
go build -o bai
```

## LLM API Setup

To chat with LLMs, you'll need to set up API keys for your preferred providers:

### OpenAI (GPT models)
```bash
export OPENAI_API_KEY="sk-your-openai-api-key-here"
```

### Anthropic (Claude models)
```bash
export ANTHROPIC_API_KEY="sk-ant-your-anthropic-api-key-here"
```

### Supported Models
- **OpenAI**: `gpt-3.5-turbo`, `gpt-4`, `gpt-4-turbo`, `gpt-4o`
- **Anthropic**: `claude-3-haiku`, `claude-3-sonnet`, `claude-3-opus`, `claude-3-5-sonnet`

Don't have an API key? See instructions below to generate a fake conversation tree to test bai out.

## Usage

### Basic Commands
```bash
bai              # Run the tool
bai version      # Check version
bai --help       # See help
```

### Conversation Management
```bash
# Plant a seed and start a new conversation
bai seed "I want to plan a trip to Japan"

# Add to current conversation (automatically gets LLM response if model is set)
bai "What's the best time of year to visit?" --llm gpt-4

# View all conversation branches
bai seeds

# Show conversation history on this branch
bai log

# Switch to different conversation branch
bai checkout <node-id>

# View branching options
bai offshoots

# Copy some context from one conversation to another
bai cherry-pick <node-id>
```

Example output:
```
$bai seeds
🌱 Found 2 seed(s) in the Bonsai garden:

ID: f47ac10b-58cc-4372-a567-0e02b2c3d479
🧠 Model: claude-3-haiku
💬 Message: I want to plan a trip by train from Ottawa to a nearby city. What are my options?

ID: fbb2e9dd-4b0b-4462-93eb-1554db2e9fe1 (contains current working node)
🧠 Model: claude-3-haiku
💬 Message: Hi there, please tell me a knock knock joke
```

### LLM Integration
When you use the `--llm` flag or set a model on a seed conversation, bai will:
1. Create your user message as a node
2. Automatically generate an LLM response
3. Add the response as a child node
4. Continue the conversation tree

```bash
# Start conversation with specific model
bai seed "Help me learn Go programming" --llm claude-3-5-sonnet

# Add message and get response
bai "What are goroutines?"

# Switch models mid-conversation
bai "Explain with a code example" --llm gpt-4
```

## Dev Notes

### Key Features
- Uses Cobra framework for CLI structure
- Base command "bai" with help and version subcommands
- Modular command structure in the `cmd` package
- Makefile for easy building and installation
- Ready for `go install` to make it globally available

### Development Commands
```bash
go mod download    # Install dependencies
go build -o bai    # Build binary
make build         # Build using Makefile
make install       # Install globally
make test          # Run tests
```

### Test fixture -- generate fake convo

Run with default database (bonsai.db)
Default behavior: Uses ~/.bonsai/bonsai.db (same as CLI tool)
```bash
./scripts/generate_fake_data.sh
```

Or specify custom database path
```bash
./scripts/generate_fake_data.sh path/to/your/database.db
```

Or run the Go script directly
```bash
go run scripts/generate_fake_data.go [database_path]
```

### Visualization
![Visualization Screenshot](assets/visualization.png)

Launch visualization (opens on http://localhost:8080)
```bash
./bai visualize
```

Custom port
```bash
./bai visualize --port 3000
```

Custom database
```
./bai visualize --database ./custom.db
```

## How is `bai` pronounced?
It is pronounced "bye", as in *NSYNC's "Bye Bye Bye".