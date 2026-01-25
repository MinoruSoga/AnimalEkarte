# History Directory

This directory is used by Gemini CLI to store session history, checkpoints, and logs.
Depending on your OS and configuration, these files might be located in `~/.gemini/history` or within the project directory.

Common contents:
- `shell_history`: Persistent history of shell commands executed by the agent.
- `checkpoints/`: Database of conversation states for `/chat resume`.
