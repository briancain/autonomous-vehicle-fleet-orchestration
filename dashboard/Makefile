.PHONY: help install build watch clean open

help: ## Show this help message
	@echo "Dashboard Build Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: ## Install dependencies
	npm install

build: ## Build TypeScript to JavaScript
	npm run build

watch: ## Watch TypeScript files and rebuild on changes
	npm run watch

clean: ## Clean built files
	rm -rf dist/ node_modules/

open: ## Open dashboard in browser (requires services to be running)
	@echo "Starting dashboard server on http://localhost:3000"
	@node server.js &
	@sleep 1
	@echo "Opening dashboard..."
	open http://localhost:3000 || xdg-open http://localhost:3000 || start http://localhost:3000

dev: install build ## Install dependencies and build
	@echo "Dashboard ready! Run 'make open' to view in browser"

stop-server: ## Stop dashboard server
	@pkill -f "node server.js" || true
	@echo "Dashboard server stopped"
