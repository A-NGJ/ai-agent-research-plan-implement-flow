#!/bin/sh
# Stop hook for the rpi Claude Code plugin.
#
# Fires when Claude Code finishes a response. Asks the rpi binary (if
# available) for the next pipeline suggestion based on .rpi/ state and
# emits it to context. Stays silent when there is no clear next step or
# when the binary is unavailable — silence is the right default.

if [ ! -x "$HOME/.rpi/bin/rpi" ]; then
  exit 0
fi

# `rpi suggest` prints a single-line next-step hint or exits non-zero / empty
# when there is nothing to suggest. Discard stderr to stay quiet on errors.
suggestion=$("$HOME/.rpi/bin/rpi" suggest 2>/dev/null) || exit 0

if [ -z "$suggestion" ]; then
  exit 0
fi

printf '# RPI next step\n\n%s\n' "$suggestion"
