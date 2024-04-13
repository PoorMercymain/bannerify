up:
	docker-compose up

test-service:
	docker-compose -f docker-compose.testing.yml up --build -d
	docker wait bannerify-e2e-1
	docker logs bannerify-e2e-1
	docker-compose -f docker-compose.testing.yml down -v