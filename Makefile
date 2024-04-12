up:
	docker-compose up

test-service:
	docker-compose -f docker-compose.testing.yml up -d
	docker wait bannerify-test-1
	docker logs bannerify-test-1
	docker-compose -f docker-compose.testing.yml down -v