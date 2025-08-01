#!/bin/bash

# Script to add fmt imports to files that use fmt.Errorf but don't import fmt
TEMPLATES_DIR="/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates"

echo "Adding missing fmt imports..."

# Files that use fmt.Errorf but might not import fmt
FILES_NEEDING_FMT=(
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/subscription/module.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/subscription/middleware/subscription.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/subscription/controller/subscription_controller.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/team/module.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/team/middleware/team.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/team/controller/team_controller.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/job/module.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/job/service/job_service.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/job/repositories/gorm/job_result_repository.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/job/repositories/gorm/job_repository.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/job/controller/job_controller.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/job/worker/handlers/email_handler.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/job/worker/job_worker.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/job/service.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/role/module.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/role/model/role.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/role/service/role_service.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/role/middleware/rbac.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/role/controller/role_controller.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/health/module.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/health/controller/health_controller.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/auth/module.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/auth/service/auth_service.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/auth/controller/auth_controller.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/notification/provider/smtp/smtp_provider.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/notification/module.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/notification/controller/notification_controller.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/container/container.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/sse/module.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/sse/service/sse_service.go"
"/home/carlos/Documents/repositories/saas-go-kit/cmd/sgk/internal/embed/templates/sse/controller/sse_controller.go"
)

for file in "${FILES_NEEDING_FMT[@]}"; do
    if [[ -f "$file" ]]; then
        # Check if fmt is already imported
        if ! grep -q '"fmt"' "$file"; then
            echo "Adding fmt import to $file"
            # Add fmt import after the first import line or create import block
            if grep -q 'import (' "$file"; then
                # Add to existing import block
                sed -i '/import (/a\	"fmt"' "$file"
            elif grep -q 'import "' "$file"; then
                # Add after single import
                sed -i '/import "/a\import "fmt"' "$file"
            else
                # Add new import block after package line
                sed -i '/^package /a\\nimport (\n\t"fmt"\n)' "$file"
            fi
        fi
    fi
done

echo "Fmt imports added!"