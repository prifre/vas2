# Variabler
VERSION=2.0.7
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
	@echo "Väntar på att bygg-workflowet ska bli klart på GitHub..."
	@sleep 3  # Ger GitHub 3 sekunder att registrera din push
	@gh run watch $$(gh run list --workflow="build.yml" --limit 1 --json databaseId -q '.[0].databaseId')
	@echo "Hämtar artifacts..."
	@gh run download --name vas-macOS --dir $(BUILD_DIR_MAC) || true
	@gh run download --name vas-android --dir $(BUILD_DIR_ANDROID) || true
# 3. Kopiera readme och packa alla plattformar till public_html
.PHONY: pack
pack: fetch-artifacts
	@echo "Kopierar readme.txt till alla byggmappar..."
	@cp readme.txt $(BUILD_DIR_LINUX)/
	@cp readme.txt $(BUILD_DIR_WIN)/
	@cp readme.txt $(BUILD_DIR_MAC)/
	@cp readme.txt $(BUILD_DIR_ANDROID)/ 2>/dev/null || true
	
	@echo "Packar filer till $(OUTPUT_DIR)..."
	@mkdir -p $(OUTPUT_DIR)
	
	# Packa Windows (.zip)
	@zip -j $(OUTPUT_DIR)/vas_$(VERSION).zip $(BUILD_DIR_WIN)/*
	
	# Packa Linux (.tar.gz)
	@tar -czvf $(OUTPUT_DIR)/vas_$(VERSION).tar.gz -C $(BUILD_DIR_LINUX) .
	
	# Packa macOS (.zip)
	@zip -j $(OUTPUT_DIR)/vas_$(VERSION)darwin.zip $(BUILD_DIR_MAC)/*
	
	# Kopiera Android APK direkt till public_html (eller zippa om du föredrar det)
	@cp $(BUILD_DIR_ANDROID)/*.apk $(OUTPUT_DIR)/vas_$(VERSION).apk 2>/dev/null || true

# Rensa bygg- och paketmappar
.PHONY: clean
clean:
	@echo "Rensar gamla byggfiler och paket..."
	@rm -rf $(BUILD_DIR_LINUX) $(BUILD_DIR_WIN) $(BUILD_DIR_MAC) $(BUILD_DIR_ANDROID)
	@rm -rf $(OUTPUT_DIR)/vas_*.zip $(OUTPUT_DIR)/vas_*.tar.gz $(OUTPUT_DIR)/vas_*.apk