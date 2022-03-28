# Spec Makefile

install:
	@ go install ./cmd/spec

generate:
	@ go generate ./...

test:
	@ go test ./...

clean:
	@ find . -name '*pb.go' -delete
	@ find . -name '*_generated.go' -delete
