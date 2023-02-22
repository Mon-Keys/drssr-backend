#!/bin/sh

kill $(lsof -t -sTCP:LISTEN -i:3001) || true
nohup make -C /home/ubuntu/backend/drssr run-api >/dev/null 2>&1 &