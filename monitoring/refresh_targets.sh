#!/usr/bin/env bash
# Usage: CLUSTER=CS6650L2-cluster SERVICE=CS6650L2 ./monitoring/refresh_targets.sh
set -u
DIR="$(cd "$(dirname "$0")" && pwd)/sd"
mkdir -p "$DIR"

while true; do
  IPS=""
  TASKS=$(aws ecs list-tasks --cluster "$CLUSTER" --service-name "$SERVICE" \
          --desired-status RUNNING --query 'taskArns[]' --output text)
  if [ -n "$TASKS" ] && [ "$TASKS" != "None" ]; then
    ENIS=$(aws ecs describe-tasks --cluster "$CLUSTER" --tasks $TASKS \
           --query "tasks[].attachments[].details[?name=='networkInterfaceId'][].value" --output text)
    if [ -n "$ENIS" ]; then
      IPS=$(aws ec2 describe-network-interfaces --network-interface-ids $ENIS \
            --query 'NetworkInterfaces[].Association.PublicIp' --output text)
    fi
  fi

  TARGETS=$(for ip in $IPS; do printf '"%s:8080",' "$ip"; done | sed 's/,$//')
  printf '[{"targets":[%s]}]\n' "$TARGETS" > "$DIR/targets.tmp" && mv "$DIR/targets.tmp" "$DIR/targets.json"
  echo "$(date +%T) targets: ${IPS:-none}"
  sleep 15
done
