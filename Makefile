run:
	docker-compose up  --remove-orphans --build

run_swagat:
	@echo " > Start Swagat service"
	go build -o build/swagat_build github.com/s2ar/swagat/cmd/swagat &&  ./build/uswagat_build -c config/config.yml server

lint:
	@echo " > Start lint"
	@golangci-lint run