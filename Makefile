.PHONY: lint
lint:
	@echo "Ejecutando linters..."
	@golangci-lint run ./...

.PHONY: format
format:
	@echo "Formateando código..."
	@go fmt ./...
	@goimports -w .

.PHONY: check
check: lint format
	@echo "Verificaciones completadas"

.PHONY: install-tools
install-tools:
	@echo "Instalando herramientas de desarrollo..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest