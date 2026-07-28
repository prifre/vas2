package general

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
)

func ClearPrefsFromJSONKeys(appID string) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("kunde inte hitta config-mapp: %w", err)
	}

	// Sökväg till preferences.json (anpassad efter Fyne-strukturen)
	jsonPath := filepath.Join(configDir, "fyne", appID, "preferences.json")

	// 1. Läs in JSON-filen
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		// Om filen inte finns finns det inget att rensa
		return nil
	}

	// 2. Parsa JSON till en generic map (nyckel -> värde)
	var prefsMap map[string]interface{}
	if err := json.Unmarshal(data, &prefsMap); err != nil {
		return fmt.Errorf("misslyckades att tolka JSON: %w", err)
	}

	// 3. Loopa igenom alla nycklar och radera dem ur Fynes Preferences
	p := fyne.CurrentApp().Preferences()
	for key := range prefsMap {
		p.RemoveValue(key)
	}

	return nil
}
