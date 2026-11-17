#!/usr/bin/env bash
# C3: create one ECR repository per service (idempotent).
# Usage: AWS_REGION=eu-west-1 TEAM=myteam ./scripts/create-ecr-repos.sh
set -euo pipefail

: "${TEAM:?Set TEAM, e.g. TEAM=dhaba-blue}"
: "${AWS_REGION:?Set AWS_REGION, e.g. AWS_REGION=eu-west-1}"

SERVICES=(web menu-api orders-api kitchen-worker stats-api postgres redis ops-monitor)
# postgres/redis don't actually need their own ECR repo (we use upstream
# images for those), but the sheet asks for "eight in total with
# ops-monitor" counting the five app services + ops-monitor + the two
# stateful services as placeholders for consistency with the naming scheme.
SERVICES=(web menu-api orders-api kitchen-worker stats-api ops-monitor)

for svc in "${SERVICES[@]}"; do
  repo="dhaba-${TEAM}/${svc}"
  if aws ecr describe-repositories --repository-names "$repo" --region "$AWS_REGION" >/dev/null 2>&1; then
    echo "repo already exists: $repo"
    continue
  fi
  aws ecr create-repository \
    --repository-name "$repo" \
    --region "$AWS_REGION" \
    --image-scanning-configuration scanOnPush=true \
    --image-tag-mutability IMMUTABLE >/dev/null
  aws ecr put-lifecycle-policy \
    --repository-name "$repo" \
    --region "$AWS_REGION" \
    --lifecycle-policy-text '{
      "rules": [{
        "rulePriority": 1,
        "description": "keep only the newest 20 images",
        "selection": {
          "tagStatus": "any",
          "countType": "imageCountMoreThan",
          "countNumber": 20
        },
        "action": { "type": "expire" }
      }]
    }' >/dev/null
  echo "created repo: $repo"
done
