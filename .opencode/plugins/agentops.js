/**
 * AgentOps plugin for OpenCode.ai
 *
 * Injects AgentOps bootstrap context via system prompt transform.
 * Skills are discovered via OpenCode's native skill tool from symlinked directory.
 */

import path from 'path';
import fs from 'fs';
import os from 'os';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

// Simple frontmatter extraction (avoid dependency on skills-core for bootstrap)
const extractAndStripFrontmatter = (content) => {
  const match = content.match(/^---\n([\s\S]*?)\n---\n([\s\S]*)$/);
  if (!match) return { frontmatter: {}, content };

  const frontmatterStr = match[1];
  const body = match[2];
  const frontmatter = {};

  for (const line of frontmatterStr.split('\n')) {
    const colonIdx = line.indexOf(':');
    if (colonIdx > 0) {
      const key = line.slice(0, colonIdx).trim();
      const value = line.slice(colonIdx + 1).trim().replace(/^["']|["']$/g, '');
      frontmatter[key] = value;
    }
  }

  return { frontmatter, content: body };
};

// Normalize a path: trim whitespace, expand ~, resolve to absolute
const normalizePath = (p, homeDir) => {
  if (!p || typeof p !== 'string') return null;
  let normalized = p.trim();
  if (!normalized) return null;
  if (normalized.startsWith('~/')) {
    normalized = path.join(homeDir, normalized.slice(2));
  } else if (normalized === '~') {
    normalized = homeDir;
  }
  return path.resolve(normalized);
};

export const AgentOpsPlugin = async ({ client, directory }) => {
  const homeDir = os.homedir();
  const agentopsSkillsDir = path.resolve(__dirname, '../../skills');
  const envConfigDir = normalizePath(process.env.OPENCODE_CONFIG_DIR, homeDir);
  const configDir = envConfigDir || path.join(homeDir, '.config/opencode');

  // Helper to generate bootstrap content
  const getBootstrapContent = () => {
    // Try to load using-agentops skill
    const skillPath = path.join(agentopsSkillsDir, 'using-agentops', 'SKILL.md');
    if (!fs.existsSync(skillPath)) return null;

    const fullContent = fs.readFileSync(skillPath, 'utf8');
    const { content } = extractAndStripFrontmatter(fullContent);

    const toolMapping = `**Tool Mapping for OpenCode:**
When skills reference tools you don't have, substitute OpenCode equivalents:
- \`TodoWrite\` → \`update_plan\`
- \`Task\` tool with subagents → Use OpenCode's subagent system (@mention)
- \`Skill\` tool → OpenCode's native \`skill\` tool (READ-ONLY — see below)
- \`Read\`, \`Write\`, \`Edit\`, \`Bash\` → Your native tools
- \`AskUserQuestion\` → Skip in headless mode, auto-proceed with defaults
- \`TeamCreate\`, \`SendMessage\`, \`TaskCreate\`, \`TaskList\` → Not available; work inline

**CRITICAL — Skill Chaining Rules:**
OpenCode's \`skill\` tool is READ-ONLY. It loads skill content into your context. It does NOT execute skills.

When a loaded skill tells you to invoke another skill (e.g., \`Skill(skill="council")\` or \`/council validate\`):
1. Use the \`skill\` tool to LOAD that skill's content
2. Then FOLLOW the loaded instructions INLINE in your current turn
3. NEVER use the \`slashcommand\` tool to invoke a skill — this will crash

Example — skill says \`Skill(skill="pre-mortem", args="--quick")\`:
  ✅ CORRECT: Use skill tool to load "pre-mortem", then follow its --quick instructions inline
  ❌ WRONG: Use slashcommand {"command":"pre-mortem"} — this crashes OpenCode

When a skill references \`spawn_agent()\` or \`TeamCreate\` for parallel agents:
  → These are not available. Execute the work serially in your current turn instead.

**Skills location:**
AgentOps skills are in \`${configDir}/skills/agentops/\`
Use OpenCode's native \`skill\` tool to list and load skills.`;

    return `<EXTREMELY_IMPORTANT>
You have AgentOps superpowers.

**IMPORTANT: The using-agentops skill content is included below. It is ALREADY LOADED - you are currently following it. Do NOT use the skill tool to load "using-agentops" again - that would be redundant.**

${content}

${toolMapping}
</EXTREMELY_IMPORTANT>`;
  };

  return {
    // Use system prompt transform to inject bootstrap (fixes agent reset bug)
    'experimental.chat.system.transform': async (_input, output) => {
      const bootstrap = getBootstrapContent();
      if (bootstrap) {
        (output.system ||= []).push(bootstrap);
      }
    },

    // Guard against OpenCode titlecase crash (sst/opencode#13933)
    // When model calls `task` tool with undefined subagent_type, OpenCode's
    // Locale.titlecase(undefined) crashes in the UI rendering layer.
    // This hook fills in a default for the execution layer (rendering crash is upstream).
    'tool.execute.before': async (input, output) => {
      if (input.tool === 'task' && output.args) {
        if (!output.args.subagent_type) {
          output.args.subagent_type = output.args.subagent_type || 'general';
        }
      }
    },

    // Patch task tool description to prevent model from sending undefined subagent_type.
    // The rendering crash in run.ts happens BEFORE tool.execute.before, so we must
    // prevent the model from ever omitting subagent_type in the first place.
    'tool.definition': async (input, output) => {
      if (input.toolID === 'task') {
        output.description = `CRITICAL: The "subagent_type" parameter is REQUIRED and MUST always be a non-empty string. ` +
          `If you don't know what agent type to use, set subagent_type to "general". ` +
          `NEVER omit subagent_type or set it to null/undefined — this will crash the application.\n\n` +
          output.description;
      }
    },

    // Intercept slashcommand calls to skills — redirect to skill tool loading
    'command.execute.before': async (input, output) => {
      // Known skill names that should be loaded via skill tool, not executed via slashcommand
      const skillNames = [
        'council', 'vibe', 'pre-mortem', 'post-mortem', 'retro', 'crank',
        'swarm', 'research', 'plan', 'implement', 'rpi', 'status',
        'complexity', 'knowledge', 'bug-hunt', 'doc', 'handoff', 'learn',
        'release', 'product', 'quickstart', 'trace', 'inbox', 'recover',
        'evolve', 'codex-team', 'beads', 'standards', 'inject', 'extract',
        'forge', 'provenance', 'ratchet', 'flywheel', 'update', 'using-agentops'
      ];

      if (skillNames.includes(input.command)) {
        // Rewrite to a system message telling the model to use skill tool instead
        (output.parts ||= []).push({
          type: 'text',
          text: `⚠️ "${input.command}" is an AgentOps skill. Do NOT use slashcommand to invoke it — use the \`skill\` tool to load "${input.command}" content, then follow its instructions inline.`
        });
      }
    }
  };
};
