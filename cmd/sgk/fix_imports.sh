#!/bin/bash

# Script to fix import aliases in templates
TEMPLATES_DIR="/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates"

echo "Fixing import aliases in templates..."

# Fix auth module imports
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/auth/interface"|authinterface "{{.Project.GoModule}}/internal/auth/interface"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/auth/constants"|authconstants "{{.Project.GoModule}}/internal/auth/constants"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/auth/model"|authmodel "{{.Project.GoModule}}/internal/auth/model"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/auth/middleware"|authmiddleware "{{.Project.GoModule}}/internal/auth/middleware"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/auth/service"|authservice "{{.Project.GoModule}}/internal/auth/service"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/auth/repository/gorm"|authgorm "{{.Project.GoModule}}/internal/auth/repository/gorm"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/auth/repository/redis"|authredis "{{.Project.GoModule}}/internal/auth/repository/redis"|g' {} \;

# Fix role module imports
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/role/interface"|roleinterface "{{.Project.GoModule}}/internal/role/interface"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/role/constants"|roleconstants "{{.Project.GoModule}}/internal/role/constants"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/role/model"|rolemodel "{{.Project.GoModule}}/internal/role/model"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/role/middleware"|rolemiddleware "{{.Project.GoModule}}/internal/role/middleware"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/role/service"|roleservice "{{.Project.GoModule}}/internal/role/service"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/role/repository/gorm"|rolegorm "{{.Project.GoModule}}/internal/role/repository/gorm"|g' {} \;

# Fix health module imports
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/health/interface"|healthinterface "{{.Project.GoModule}}/internal/health/interface"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/health/constants"|healthconstants "{{.Project.GoModule}}/internal/health/constants"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/health/model"|healthmodel "{{.Project.GoModule}}/internal/health/model"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/health/service"|healthservice "{{.Project.GoModule}}/internal/health/service"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/health/checkers"|healthcheckers "{{.Project.GoModule}}/internal/health/checkers"|g' {} \;

# Fix team module imports
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/team/interface"|teaminterface "{{.Project.GoModule}}/internal/team/interface"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/team/constants"|teamconstants "{{.Project.GoModule}}/internal/team/constants"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/team/model"|teammodel "{{.Project.GoModule}}/internal/team/model"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/team/middleware"|teammiddleware "{{.Project.GoModule}}/internal/team/middleware"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/team/service"|teamservice "{{.Project.GoModule}}/internal/team/service"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/team/repository/gorm"|teamgorm "{{.Project.GoModule}}/internal/team/repository/gorm"|g' {} \;

# Fix subscription module imports
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/subscription/interface"|subscriptioninterface "{{.Project.GoModule}}/internal/subscription/interface"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/subscription/constants"|subscriptionconstants "{{.Project.GoModule}}/internal/subscription/constants"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/subscription/model"|subscriptionmodel "{{.Project.GoModule}}/internal/subscription/model"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/subscription/middleware"|subscriptionmiddleware "{{.Project.GoModule}}/internal/subscription/middleware"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/subscription/service"|subscriptionservice "{{.Project.GoModule}}/internal/subscription/service"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/subscription/repository/gorm"|subscriptiongorm "{{.Project.GoModule}}/internal/subscription/repository/gorm"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/subscription/provider"|subscriptionprovider "{{.Project.GoModule}}/internal/subscription/provider"|g' {} \;

# Fix job module imports
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/job/interface"|jobinterface "{{.Project.GoModule}}/internal/job/interface"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/job/constants"|jobconstants "{{.Project.GoModule}}/internal/job/constants"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/job/model"|jobmodel "{{.Project.GoModule}}/internal/job/model"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/job/service"|jobservice "{{.Project.GoModule}}/internal/job/service"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/job/repository/gorm"|jobgorm "{{.Project.GoModule}}/internal/job/repository/gorm"|g' {} \;

# Fix notification module imports
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/notification/interface"|notificationinterface "{{.Project.GoModule}}/internal/notification/interface"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/notification/constants"|notificationconstants "{{.Project.GoModule}}/internal/notification/constants"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/notification/model"|notificationmodel "{{.Project.GoModule}}/internal/notification/model"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/notification/service"|notificationservice "{{.Project.GoModule}}/internal/notification/service"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/notification/provider/smtp"|notificationsmtp "{{.Project.GoModule}}/internal/notification/provider/smtp"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/notification/provider/dev"|notificationdev "{{.Project.GoModule}}/internal/notification/provider/dev"|g' {} \;

# Fix sse module imports
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/sse/interface"|sseinterface "{{.Project.GoModule}}/internal/sse/interface"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/sse/constants"|sseconstants "{{.Project.GoModule}}/internal/sse/constants"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/sse/model"|ssemodel "{{.Project.GoModule}}/internal/sse/model"|g' {} \;
find "$TEMPLATES_DIR" -name "*.go" -exec sed -i 's|"{{\.Project\.GoModule}}/internal/sse/service"|sseservice "{{.Project.GoModule}}/internal/sse/service"|g' {} \;

echo "Import aliases fixed!"