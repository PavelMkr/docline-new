package main

import (
	"fmt"
	"encoding/json"

	ui "github.com/webui-dev/go-webui/v2"
)

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

func main() {
	// Create a window.
	w := ui.NewWindow()  
	// Bind a Go function.
	ui.Bind(w, "SendSettings", SendSettings)
	// Show frontend.
	w.Show("index.html")
	// Wait until all windows get closed.
	ui.Wait()
}