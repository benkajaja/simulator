.PHONY: agent controller
all: agent controller
agent:
	@go build -o agent Agent/main.go
controller:
	@go build -o controller Controller/main.go
clean:
	rm -f agent controller
