.PHONY: build build-provider build-cli clean install

# Build both provider and CLI
build: build-provider build-cli

# Build Terraform provider
build-provider:
	go build -o terraform-provider-ssmready

# Build CLI for macOS ARM64
build-cli:
	GOOS=darwin GOARCH=arm64 go build -o ansible-ssm cmd/ansible-ssm/main.go

# Clean build artifacts
clean:
	rm -f terraform-provider-ssmready ansible-ssm

# Install CLI to ~/bin (no sudo required)
install: build-cli
	./install.sh
