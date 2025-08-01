#!/bin/bash

# Script to fix duplicate import aliases in templates
TEMPLATES_DIR="/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates"

echo "Fixing duplicate import aliases in templates..."

# Fix duplicate aliases by removing the duplicate part
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/authinterface authinterface/authinterface/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/authmiddleware authmiddleware/authmiddleware/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/authmodel authmodel/authmodel/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/authconstants authconstants/authconstants/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/authservice authservice/authservice/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/authgorm authgorm/authgorm/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/authredis authredis/authredis/g' {} \;

find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/roleinterface roleinterface/roleinterface/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/roleconstants roleconstants/roleconstants/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/rolemodel rolemodel/rolemodel/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/rolemiddleware rolemiddleware/rolemiddleware/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/roleservice roleservice/roleservice/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/rolegorm rolegorm/rolegorm/g' {} \;

find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/healthinterface healthinterface/healthinterface/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/healthconstants healthconstants/healthconstants/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/healthmodel healthmodel/healthmodel/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/healthservice healthservice/healthservice/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/healthcheckers healthcheckers/healthcheckers/g' {} \;

find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/teaminterface teaminterface/teaminterface/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/teamconstants teamconstants/teamconstants/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/teammodel teammodel/teammodel/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/teammiddleware teammiddleware/teammiddleware/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/teamservice teamservice/teamservice/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/teamgorm teamgorm/teamgorm/g' {} \;

find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/subscriptioninterface subscriptioninterface/subscriptioninterface/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/subscriptionconstants subscriptionconstants/subscriptionconstants/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/subscriptionmodel subscriptionmodel/subscriptionmodel/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/subscriptionmiddleware subscriptionmiddleware/subscriptionmiddleware/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/subscriptionservice subscriptionservice/subscriptionservice/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/subscriptiongorm subscriptiongorm/subscriptiongorm/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/subscriptionprovider subscriptionprovider/subscriptionprovider/g' {} \;

find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/jobinterface jobinterface/jobinterface/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/jobconstants jobconstants/jobconstants/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/jobmodel jobmodel/jobmodel/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/jobservice jobservice/jobservice/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/jobgorm jobgorm/jobgorm/g' {} \;

find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/notificationinterface notificationinterface/notificationinterface/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/notificationconstants notificationconstants/notificationconstants/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/notificationmodel notificationmodel/notificationmodel/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/notificationservice notificationservice/notificationservice/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/notificationsmtp notificationsmtp/notificationsmtp/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/notificationdev notificationdev/notificationdev/g' {} \;

find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/sseinterface sseinterface/sseinterface/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/sseconstants sseconstants/sseconstants/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/ssemodel ssemodel/ssemodel/g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's/sseservice sseservice/sseservice/g' {} \;

echo "Duplicate aliases fixed!"