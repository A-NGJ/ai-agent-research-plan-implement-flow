#!/bin/sh
# PostCompact hook for the rpi Claude Code plugin.
#
# Fires after Claude Code compacts the conversation. Restores a narrow slice
# of the RPI framing (just enough to keep the pipeline coherent) and points
# at the currently active artifact if one exists.

cat <<'EOF'
# RPI workflow (post-compact restore)

Pipeline: Research → Propose → Plan → Implement → Verify
Artifacts live under `.rpi/`. Each skill suggests the next.

Use `mcp__rpi__rpi_session_resume` to recover active artifacts and the next
suggested pipeline step.
EOF
