#!/usr/bin/env bash
# ==============================================================================
# Script: restart-service.sh
# Purpose: Gracefully stop or restart the DigitalPersona fingerprint session
# ==============================================================================

SESSION_NAME="${1:-finger_session}"

# Send Ctrl+C interrupt to the tmux session
if tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
    echo "[INFO] Sending SIGINT to tmux session '$SESSION_NAME'..."
    tmux send-keys -t "$SESSION_NAME" C-c
else
    # Also support legacy session name
    if tmux has-session -t "finger_sesion" 2>/dev/null; then
        echo "[INFO] Sending SIGINT to legacy tmux session 'finger_sesion'..."
        tmux send-keys -t "finger_sesion" C-c
    else
        echo "[WARN] No active tmux session found for '$SESSION_NAME'."
    fi
fi
