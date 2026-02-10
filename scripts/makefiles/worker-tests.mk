# Image Worker Test Commands
.PHONY: run-image-worker stop-image-worker worker-logs

run-image-worker:
	@echo "Starting image processing worker..."
	@go run cmd/image-worker/main.go

stop-image-worker:
	@echo "Stopping image worker..."
	@pkill -f "image-worker/main.go" || true

worker-logs:
	@echo "Tailing worker logs..."
	@tail -f $(shell ls -t logs/worker-*.log | head -1) 2>/dev/null || echo "No log file found"

.PHONY: help-worker
help-worker:
	@echo "Image Worker Commands:"
	@echo "  make run-image-worker  - Start image processing worker"
	@echo "  make stop-image-worker - Stop image processing worker"
	@echo "  make worker-logs       - Tail worker logs"
