# E. Morricone Ag

> **E. Morricone Ag** (`emorr-agy`) is a Go-based Telegram bot and CLI orchestrator designed to manage, monitor, and interact with multiple `tmux` sessions running on remote machines.

---

<p align="center">
  <img src="doc/img/conductor_swiss.png" alt="E. Morricone A.g. Conductor in Switzerland" width="600px" style="border-radius: 8px;" />
</p>

---

## Project Architecture

```mermaid
graph TD
    A[Riccardo] -- Telegram --> B{Telegram API};
    
    subgraph Host 1
        B -- Command --> C1[emorr-agy service];
        C1 --> D1_1[tmux s1 - active];
        C1 --> D1_2[tmux s2 - active];
        C1 --> D1_3[tmux s3 - idle];
        C1 --> D1_4[tmux s4 - active];
        C1 --> D1_5[tmux s5 - idle];
    end

    subgraph Host 2
        B -- Command --> C2[emorr-agy service];
        C2 --> D2_1[tmux s1 - active];
        C2 --> D2_2[tmux s2 - idle];
        C2 --> D2_3[tmux s3 - active];
        C2 --> D2_4[tmux s4 - active];
        C2 --> D2_5[tmux s5 - idle];
    end

    subgraph Host 3
        B -- Command --> C3[emorr-agy service];
        C3 --> D3_1[tmux s1 - idle];
        C3 --> D3_2[tmux s2 - active];
        C3 --> D3_3[tmux s3 - active];
        C3 --> D3_4[tmux s4 - idle];
        C3 --> D3_5[tmux s5 - active];
    end

    style D1_1 fill:#d4edda,stroke:#c3e6cb
    style D1_2 fill:#d4edda,stroke:#c3e6cb
    style D1_3 fill:#f8f9fa,stroke:#dee2e6
    style D1_4 fill:#d4edda,stroke:#c3e6cb
    style D1_5 fill:#f8f9fa,stroke:#dee2e6

    style D2_1 fill:#d4edda,stroke:#c3e6cb
    style D2_2 fill:#f8f9fa,stroke:#dee2e6
    style D2_3 fill:#d4edda,stroke:#c3e6cb
    style D2_4 fill:#d4edda,stroke:#c3e6cb
    style D2_5 fill:#f8f9fa,stroke:#dee2e6

    style D3_1 fill:#f8f9fa,stroke:#dee2e6
    style D3_2 fill:#d4edda,stroke:#c3e6cb
    style D3_3 fill:#d4edda,stroke:#c3e6cb
    style D3_4 fill:#f8f9fa,stroke:#dee2e6
    style D3_5 fill:#d4edda,stroke:#c3e6cb
```

## Features (BDD Specs)

* **Multi-Project Management**: Switch context via `/projects` to view only active `tmux` sessions relating to a specific project.
* **Standardized Session Names**: Dynamically spawn `tmux` sessions with standardized, structured names (e.g. `emorragi-data-analysis-1654172400`).
* **Real-time Status Monitoring**: Instantly query machine status via `/status` to see which sessions are **[BUSY]** or **[IDLE]**.
* **Persistence on Reboot**: Restores sessions and executes startup workflows automatically if the remote host machine reboots.
* **Granular Navigation & Interaction**: Select specific **[IDLE]** sessions and focus on them to send interactive inputs, view recent console outputs, or terminate them.

---

## Technical Stack

* **Language**: Go
* **Telegram Integration**: `go-telegram-bot-api`
* **Session Management**: Native `tmux` CLI integration with process status inference.
* **State Persistence**: Simple JSON state tracking (e.g., `~/.emorragi_state.json`) for session restoration.
