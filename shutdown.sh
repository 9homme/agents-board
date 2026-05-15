#!/bin/bash
# shutdown.sh — Stop all Agents Board services.

echo "🛑 Shutting down Agents Board services..."

# 1. Kill via PID files if they exist
for pid_file in .mcp.pid .api.pid .web.pid; do
  if [ -f "$pid_file" ]; then
    PID=$(cat "$pid_file")
    if ps -p $PID > /dev/null; then
       # Kill the process group to catch 'go run' children
       pkill -P $PID 2>/dev/null
       kill $PID 2>/dev/null
       echo "  [OK] Terminated process $PID (from $pid_file)"
    fi
    rm "$pid_file"
  fi
done

# 2. Force cleanup by ports (ensure nothing is left hanging)
echo "  -> Cleaning up ports..."
for port in 8080 8081 3000 3001; do
  PIDS=$(lsof -ti :$port)
  if [ -n "$PIDS" ]; then
    for pid in $PIDS; do
      kill -9 $pid 2>/dev/null
      echo "  [OK] Force killed PID $pid on port $port"
    done
  fi
done

echo "✅ All services stopped."
