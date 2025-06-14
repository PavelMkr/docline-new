//go:build !ui_test
// +build !ui_test

package main

import (
	"encoding/json"
	"fmt"

	ui "github.com/webui-dev/go-webui/v2"
)

func startUI() {
	// UI
	w := ui.NewWindow()
	w.SetSize(800, 800)
	w.SetMinimumSize(400, 800)
	ui.Bind(w, "SendSettings", SendSettings)
	w.Show("index.html")
	ui.Wait()
}

func SendSettings(e ui.Event) string {
	// json string
	jsonStr, err := ui.GetArg[string](e)
	if err != nil {
		return fmt.Sprintf("Error getting argument: %v", err)
	}

	// json to map
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &settings); err != nil {
		return fmt.Sprintf("Error parsing JSON: %v", err)
	}

	// print data
	fmt.Printf("Received settings: %+v\n", settings)

	// FIXME unnessesarry return
	return "Settings received and parsed successfully"
}
