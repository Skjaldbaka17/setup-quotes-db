# The \ is needed for sourcing the env-file
setup:
	source .env; \
	go run setup-quotes.go