#!/bin/bash
# startup.sh — Start all Agents Board services in the background.

# Default environment variables
DB_URL=${DB_URL:-"postgres://localhost:5432/agent_board?sslmode=disable"}
DATABASE_URL=${DATABASE_URL:-$DB_URL}
FRONTEND_URL=${FRONTEND_URL:-"http://localhost:3000"}
API_PORT=${API_PORT:-8080}
MCP_PORT=${MCP_PORT:-8081}

echo "🚀 Starting Agents Board services..."

# 1. Start MCP Server (Port 8081)
echo "  -> Starting MCP Server on port $MCP_PORT..."
cd services/agent-board
PORT=$MCP_PORT DB_URL=$DB_URL go run cmd/mcp-server/main.go > ../../mcp-server.log 2>&1 &
echo $! > ../../.mcp.pid

# 2. Start API Server (Port 8080)
echo "  -> Starting API Server on port $API_PORT..."
DATABASE_URL=$DATABASE_URL PORT=$API_PORT FRONTEND_URL=$FRONTEND_URL go run cmd/api-server/main.go > ../../api-server.log 2>&1 &
echo $! > ../../.api.pid

# 3. Start Frontend
echo "  -> Starting Frontend on port 3000..."
cd ../../web
PORT=3000 NEXT_PUBLIC_API_BASE_URL="http://localhost:$API_PORT" npm run dev > ../web.log 2>&1 &
echo $! > ../.web.pid

echo ""
echo "✅ All services initiated."
echo "   - MCP Server: http://localhost:$MCP_PORT/sse"
echo "   - API Server: http://localhost:$API_PORT/api/v1/projects"
echo "   - Frontend:   http://localhost:3000 (or 3001)"
echo ""
echo "Logs are available in: mcp-server.log, api-server.log, web.log"
echo "Use ./shutdown.sh to stop all services."
