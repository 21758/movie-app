.PHONY: docker-up docker-down test-e2e docker-logs

docker-up:
	docker compose up -d --build
	@echo "Waiting for services to be healthy..."
	@while ! curl -s http://localhost:8080/healthz > /dev/null; do \
		echo "Waiting for app to be ready..."; \
		sleep 5; \
	done
	@echo "All services are up and healthy!"

docker-down:
	docker compose down -v

test-e2e:
	./e2e-test.sh

docker-logs:
	docker compose logs -f