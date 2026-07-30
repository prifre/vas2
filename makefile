# Variabler
VERSION=2.0.5
BINARY_NAME=vas2
BUILD_DIR_LINUX=vas2_linux
BUILD_DIR_WIN=vas2_windows
BUILD_DIR_MAC=vas2_darwin
BUILD_DIR_ANDROID=vas2_android
OUTPUT_DIR=public_html

# Standardmål
.PHONY: all
all: clean pack

# 1. Skapa mappar och kompilera lokalt för Linux & Windows
.PHONY: build-local
build-local:
	@echo "Kompilerar lokalt för Linux och Windows..."
	@mkdir -p $(BUILD_DIR_LINUX) $(BUILD_DIR_WIN) $(BUILD_DIR_MAC) $(BUILD_DIR_ANDROID)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR_LINUX)/$(BINARY_NAME) .
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -ldflags="-s -w" -o $(BUILD_DIR_WIN)/$(BINARY_NAME).exe .

# 2. Skicka till GitHub, vänta på Actions-bygget och hämta Mac & Android
.PHONY: fetch-artifacts
fetch-artifacts: build-local
	@echo "Pushar till GitHub och väntar på molnbyggena..."
	@git add .
	@git commit -m "Auto-build release v$(VERSION)" || true
	@git push
	@echo "Väntar på att GitHub Actions ska bli klar..."
	@gh run watch $$(gh run list --limit 1 --json databaseId -q '.[0].databaseId')
	@echo "Hämtar artifacts från GitHub..."
	@gh run download --name vas-macOS --dir $(BUILD_DIR_MAC)
	@gh run download --name vas-android --dir $(BUILD_DIR_ANDROID)

# 3. Kopiera readme och packa alla plattformar till public_html
# 2. Skicka till GitHub, vänta på Actions-bygget och hämta Mac & Android
.PHONY: fetch-artifacts
fetch-artifacts: build-local
	@echo "Pushar till GitHub och väntar på molnbyggena..."
	@git add .
	@git commit -m "Auto-build release v$(VERSION)" || true
	@git push
	@echo "Väntar på att bygg-workflowet ska bli klart på GitHub..."
	@sleep 3  # Ger GitHub 3 sekunder att registrera din push
	@gh run watch $$(gh run list --workflow="build.yml" --limit 1 --json databaseId -q '.[0].databaseId')
	@echo "Hämtar artifacts..."
	@gh run download --name vas-macOS --dir $(BUILD_DIR_MAC) || true
	@gh run download --name vas-android --dir $(BUILD_DIR_ANDROID) || true