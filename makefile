# Variabler
VERSION=2.0.5
BINARY_NAME=vas2
BUILD_DIR_LINUX=vas2_linux
BUILD_DIR_WIN=vas2_windows
BUILD_DIR_MAC=vas2_darwin
OUTPUT_DIR=public_html

# Standardmål
.PHONY: all
all: clean pack

# 1. Skapa mappar och kompilera lokalt för Linux & Windows
.PHONY: build-local
build-local:
	@echo "Kompilerar lokalt för Linux och Windows..."
	@mkdir -p $(BUILD_DIR_LINUX) $(BUILD_DIR_WIN) $(BUILD_DIR_MAC)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR_LINUX)/$(BINARY_NAME) .
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o $(BUILD_DIR_WIN)/$(BINARY_NAME).exe .

# 2. Skicka till GitHub, vänta på Actions-bygget och hämta Mac-binären
# 2. Skicka till GitHub, vänta på Actions-bygget och hämta Mac-binären
.PHONY: fetch-mac
fetch-mac: build-local
	@echo "Pushar till GitHub och väntar på Mac-bygget..."
	@git add .
	@git commit -m "Auto-build release v$(VERSION)" || true
	@git push
	@echo "Väntar på att GitHub Actions ska bli klar..."
	@gh run watch $$(gh run list --limit 1 --json databaseId -q '.[0].databaseId')
	@echo "Hämtar Mac-artifact från GitHub..."
	@gh run download --name vas-macOS --dir $(BUILD_DIR_MAC)
	
# 3. Kopiera readme och packa alla tre plattformar till public_html
.PHONY: pack
pack: fetch-mac
	@echo "Kopierar readme.txt till alla byggmappar..."
	@cp readme.txt $(BUILD_DIR_LINUX)/
	@cp readme.txt $(BUILD_DIR_WIN)/
	@cp readme.txt $(BUILD_DIR_MAC)/
	
	@echo "Packar filer till $(OUTPUT_DIR)..."
	@mkdir -p $(OUTPUT_DIR)
	
	# Packa Windows (.zip)
	@zip -j $(OUTPUT_DIR)/vas_$(VERSION).zip $(BUILD_DIR_WIN)/*
	
	# Packa Linux (.tar.gz)
	@tar -czvf $(OUTPUT_DIR)/vas_$(VERSION).tar.gz -C $(BUILD_DIR_LINUX) .
	
	# Packa macOS (.zip)
	@zip -j $(OUTPUT_DIR)/vas_$(VERSION)darwin.zip $(BUILD_DIR_MAC)/*

# Rensa bygg- och paketmappar
.PHONY: clean
clean:
	@echo "Rensar gamla byggfiler och paket..."
	@rm -rf $(BUILD_DIR_LINUX) $(BUILD_DIR_WIN) $(BUILD_DIR_MAC)
	@rm -rf $(OUTPUT_DIR)/vas_*.zip $(OUTPUT_DIR)/vas_*.tar.gz