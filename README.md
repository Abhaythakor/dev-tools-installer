# 🛠️ Automated Development Tools Installer

A flexible, configuration-driven system for automatically managing and installing development tools across different environments. This tool helps you maintain consistent development environments by automatically installing and verifying tool versions.

## ✨ Features

- 🔄 **Dynamic Tool Configuration**
  - YAML-based configuration
  - Multiple installation methods per tool
  - Version management and verification
  - Dependency resolution

- 📦 **Multiple Installation Methods**
  - Official binary downloads
  - Package managers (apt, snap)
  - Go install command
  - Custom installation commands

- 🎨 **Beautiful Console Output**
  - Progress indicators
  - Color-coded status messages
  - Clear error reporting
  - Installation progress tracking

- 🔌 **Extensible Design**
  - Easy to add new tools
  - Configurable installation methods
  - Version detection patterns
  - Custom version flags

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/Abhaythakor/dev-tools-installer.git
cd dev-tools-installer

# Build the project
go build ./cmd/installer

# Run the installer
./installer
```

## 📋 Requirements

- Linux-based operating system
- Go 1.21 or higher
- Sudo access (for some installation methods)
- Internet connection for downloading tools

## 🔧 Configuration

Tools are configured in `installer.yaml`. Here's a comprehensive example:

```yaml
tool_list:
  - go
  - amass
  - subfinder
  - assetfinder

tools:
  go:
    version: "1.23.3"
    methods:
      - name: binary
        commands:
          - wget https://go.dev/dl/go${version}.linux-amd64.tar.gz
          - sudo rm -rf /usr/local/go
          - sudo tar -C /usr/local -xzf go${version}.linux-amd64.tar.gz
          - rm go${version}.linux-amd64.tar.gz
      - name: snap
        commands:
          - sudo snap install go --classic

  amass:
    dependencies: ["go"]
    methods:
      - name: go
        commands:
          - go install -v github.com/owasp-amass/amass/v4/...@master
      - name: apt
        commands:
          - sudo apt update
          - sudo apt install -y amass
```

### Configuration Options

#### Tool Configuration
- `version`: Specify the required version (optional)
- `dependencies`: List of tools that must be installed first
- `version_flag`: Custom flag to check version (optional)
- `methods`: List of installation methods to try

#### Installation Methods
- `name`: Identifier for the installation method
- `commands`: List of commands to execute for installation
- Variables available in commands:
  - `${version}`: Replaced with the tool's version
  - Environment variables (e.g., `$HOME`, `$PATH`)

## 🏗️ Project Structure

```
.
├── cmd/
│   └── installer/
│       └── main.go           # Entry point
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration handling
│   └── installer/
│       └── installer.go      # Core installation logic
└── installer.yaml            # Tool configuration
```

## 🔍 Version Detection

The installer supports multiple version detection strategies:
1. Use version from YAML configuration
2. Try common version flags:
   - `--version`
   - `-version`
   - `version`
   - `-v`
   - `-V`
3. Parse version using regex patterns
4. Fallback to first line of output

## 🛟 Error Handling

The installer provides detailed error handling:
- Installation method failures
- Version detection issues
- Configuration errors
- Command execution problems

Each error includes:
- Clear error message
- Context about the failure
- Suggested next steps

## 🔒 Security

- Uses official package managers and repositories
- Downloads from trusted sources
- Allows review of installation commands
- Supports checksums for binary downloads

## 🤝 Contributing

Contributions are welcome! Here's how you can help:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

Areas for contribution:
- Add new tool configurations
- Improve version detection
- Add installation methods
- Enhance error handling
- Improve documentation

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Go community for excellent tooling
- Open source tool maintainers
- Contributors and users

## 📚 Documentation

For more detailed documentation:
- [Configuration Guide](docs/configuration.md)
- [Installation Methods](docs/installation.md)
- [Version Detection](docs/versions.md)
- [Contributing Guide](CONTRIBUTING.md)

## 🐛 Troubleshooting

Common issues and solutions:
1. **Tool installation fails**
   - Check internet connection
   - Verify sudo access
   - Check disk space
   - Review error messages

2. **Version detection fails**
   - Add custom version flag in config
   - Check tool documentation
   - Update version patterns

3. **Configuration issues**
   - Validate YAML syntax
   - Check file permissions
   - Verify tool names

## 📞 Support

- Open an issue for bugs
- Start a discussion for questions
- Submit a PR for improvements

## 🗺️ Roadmap

Future improvements:
- [ ] Windows support
- [ ] macOS support
- [ ] Configuration validation
- [ ] Installation logs
- [ ] Progress persistence
- [ ] Tool updates
- [ ] Version constraints
- [ ] Plugin system
