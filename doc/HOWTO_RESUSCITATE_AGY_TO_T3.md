# ⚰️💀 How-To: Resuscitate Antigravity CLI (`agy`) Sessions into T3 Code

This guide explains how to take an existing **Antigravity CLI** (`agy`) or **Herdr** conversation, transfer its full context and history, and **resuscitate** it directly into an active **T3 Code** chat thread via MCP tools.

---

## 🧭 Overview

When working across multiple AI coding harnesses (such as the headless terminal CLI `agy` orchestrated via Herdr and the full IDE environment in `T3 Code`), conversations often need to be handed off seamlessly without losing project context, previous learnings, or worktree state.

Using `agy-herdr-resuscitate` and the T3 Code MCP server, you can:
1. List all active/sleeping conversations across local repositories.
2. Read the latest transcript turns directly from the agent's SQLite brain and JSONL logs.
3. Launch a new, project-bound thread in T3 Code.
4. Cleanly archive the source session in Antigravity annotations and Herdr.

---

## 🛠️ Step 1: List Sessions with `agy-herdr-resuscitate`

Run the session manager script:
```bash
agy-herdr-resuscitate list
```
Example output:
```
📂 ~/git/pupurabbu-3d
  🟢⬆️ [w1:p1] 8ea41a43-5760   1d  Autonomous WIP Task Execution🏷️ prendi un task WIP e...

📂 ~/git/rails8-app-on-gcp
  🟢⬆️ [w5:p1] 97d19a0b-e2f1   1d  Git Status And Commit Automation🏷️ check git, and check st...
```
Identify the session short ID or folder you wish to migrate.

---

## 🔍 Step 2: Read Transcript & Memory

Locate the full conversation UUID in `~/.gemini/antigravity-cli/conversation_summaries.db`:
```python
import sqlite3
conn = sqlite3.connect('/Users/ricc/.gemini/antigravity-cli/conversation_summaries.db')
c = conn.cursor()
c.execute("SELECT conversation_id, title, workspace_uris FROM conversation_summaries WHERE conversation_id LIKE '8ea41a43%'")
row = c.fetchone()
print(row)
```

The conversation transcript resides at:
`~/.gemini/antigravity-cli/brain/<uuid>/.system_generated/logs/transcript.jsonl`

Read the last few user instructions and the final model response to summarize:
- What the user originally requested.
- What changes or commits were made.
- Where the agent stopped (e.g., waiting for input, `/chumbia`, or completed PR).

---

## 🚀 Step 3: Launch in T3 Code via MCP

1. **Verify or create the Project in T3 Code:**
   Call MCP tool `t3_project_list`. If the folder isn't listed, call `t3_project_create`:
   ```json
   {
     "title": "pupurabbu-3d",
     "workspaceRoot": "/Users/ricc/git/pupurabbu-3d"
   }
   ```

2. **Launch the Resumed Thread:**
   Call MCP tool `t3_thread_launch`:
   ```json
   {
     "title": "⚰️💀 [RESUME] Autonomous WIP Task Execution",
     "projectId": "<project_id>",
     "workspaceStrategy": { "type": "root" },
     "message": "⚰️💀 **[RIESUMAZIONE CHAT / RESUME DA AGY]**\n\nGuarda che ti ho spostato dal folder `~/git/pupurabbu-3d` (sessione CLI Antigravity `8ea41a43-5760...`) a qui dentro T3 Code, e Riccardo vuole riesumarla per proseguire il lavoro!\n\n### 📜 Storico della sessione precedente:\n- ..."
   }
   ```
   T3 Code will instantiate a fresh agent runtime pre-loaded with the project workspace and the handover prompt.

---

## 📦 Step 4: Archive the Source Session in `agy` & Herdr

To avoid duplicate active panes and keep CLI listings clean:

1. **Mark as archived in Antigravity annotations:**
   In `~/.gemini/antigravity-cli/annotations/<uuid>.pbtxt`:
   ```
   title:"<Title>" archived:true archival_status_timestamp:{seconds:<timestamp> nanos:0}
   ```

2. **Close the pane/workspace in Herdr:**
   ```bash
   herdr workspace close <workspace_id>
   ```

Once closed, `agy-herdr-resuscitate list` will only display remaining unarchived sessions.
