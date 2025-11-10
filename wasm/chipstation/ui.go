//go:build js && wasm

package main

import (
	"bytes"
	"html/template"
	"io/fs"
	"log"
	"strconv"
	"syscall/js"

	chip8web "github.com/mrchip53/chip-station/cores/chip8/webgl"
)

// UI manages the HTML interface and emulator controls
type UI struct {
	document    js.Value
	container   js.Value
	containerID string
	emulator    *chip8web.Chip8WebEmulator // Your emulator instance
	elements    map[string]js.Value
	handlers    map[string]js.Func
	template    *template.Template
}

// UIData holds data for template rendering
type UIData struct {
	DisplayWidth  int
	DisplayHeight int
	Speeds        []SpeedOption
	ROMs          []ROMOption
}

type SpeedOption struct {
	Value int
	Label string
}

type ROMOption struct {
	Value string
	Label string
}

const styledTemplate = `
<div style="display: flex; justify-content: center;">
    <div style="position: relative; display: inline-block;">
		<canvas id="cs-screen" width="{{.DisplayWidth}}" height="{{.DisplayHeight}}"></canvas>
		<div style="position:absolute; top:0; left:0; width:100%; padding:4px; background:rgba(0,0,0,0.4); color:white; font:12px monospace; box-sizing:border-box;">
			<button id="startBtn" class="btn btn-primary">Start</button>
        	<button id="stopBtn" class="btn btn-secondary">Stop</button>
        	<button id="resetBtn" class="btn btn-warning">Reset</button>
			<button onclick="downloadRom()">Download ROM</button>
			<select id="speedDropdown" style="width: 100px;">
				{{range .Speeds}}
				<option value="{{.Value}}">{{.Label}}</option>
				{{end}}
        	</select>
			<select id="romSelector" style="width: 100px;">
				{{range .ROMs}}
				<option value="{{.Value}}">{{.Label}}</option>
				{{end}}
            </select>
		</div>
		<div style="position:absolute; bottom:0; left:0; width:100%; padding:4px; background:rgba(0,0,0,0.4); color:white; font:12px monospace; box-sizing:border-box;">
			<a href="https://github.com/mrchip53/chip-station" target="_blank" rel="noreferrer noopener">Chip Station Source</a> | <a href="https://www.shadertoy.com/view/XlVczc" target="_blank" rel="noreferrer noopener">CRT Shader Source</a>
		</div>
	</div>
</div>
`

// NewUI creates a new UI manager with a default container ID
func NewUI(emu *chip8web.Chip8WebEmulator) *UI {
	return NewUIWithContainer(emu, "chip8-ui")
}

// NewUIWithContainer creates a new UI manager with a custom container ID
func NewUIWithContainer(emu *chip8web.Chip8WebEmulator, containerID string) *UI {
	tmpl := template.Must(template.New("ui").Parse(styledTemplate))

	ui := &UI{
		document:    js.Global().Get("document"),
		containerID: containerID,
		emulator:    emu,
		elements:    make(map[string]js.Value),
		handlers:    make(map[string]js.Func),
		template:    tmpl,
	}
	return ui
}

func (ui *UI) SetEmulator(emu *chip8web.Chip8WebEmulator) {
	ui.emulator = emu
}

// SetContainer changes the container div ID (must call before Build)
func (ui *UI) SetContainer(containerID string) {
	ui.containerID = containerID
}

// SetTemplate sets a custom HTML template
func (ui *UI) SetTemplate(tmplString string) error {
	tmpl, err := template.New("ui").Parse(tmplString)
	if err != nil {
		return err
	}
	ui.template = tmpl
	return nil
}

// Build constructs the UI HTML and attaches event handlers
func (ui *UI) Build() error {
	// log to console
	log.Printf("Building UI in container: %s", ui.containerID)
	// Get or validate container
	ui.container = ui.document.Call("getElementById", ui.containerID)
	if ui.container.IsNull() || ui.container.IsUndefined() {
		return js.Error{Value: js.ValueOf("Container element not found: " + ui.containerID)}
	}

	// Build HTML structure from template
	if err := ui.buildHTML(); err != nil {
		return err
	}

	// Get UI elements (now that they exist)
	ui.elements["startBtn"] = ui.document.Call("getElementById", "startBtn")
	ui.elements["stopBtn"] = ui.document.Call("getElementById", "stopBtn")
	ui.elements["resetBtn"] = ui.document.Call("getElementById", "resetBtn")
	ui.elements["speedDropdown"] = ui.document.Call("getElementById", "speedDropdown")
	ui.elements["romSelector"] = ui.document.Call("getElementById", "romSelector")
	ui.elements["cs-screen"] = ui.document.Call("getElementById", "cs-screen")

	// Attach event handlers
	ui.attachHandler("startBtn", "click", ui.handleStart)
	ui.attachHandler("stopBtn", "click", ui.handleStop)
	ui.attachHandler("resetBtn", "click", ui.handleReset)
	ui.attachHandler("speedDropdown", "change", ui.handleSpeedChange)
	ui.attachHandler("romSelector", "change", ui.handleRomLoad)

	return nil
}

// buildHTML renders the template and injects it into the container
func (ui *UI) buildHTML() error {
	// Get all files in romAssets/roms
	romFiles, err := fs.ReadDir(romAssets, "roms")
	if err != nil {
		return err
	}

	roms := make([]ROMOption, 0, len(romFiles))
	for _, file := range romFiles {
		if !file.IsDir() {
			roms = append(roms, ROMOption{
				Value: file.Name(),
				Label: file.Name(),
			})
		}
	}

	data := UIData{
		DisplayWidth:  640,
		DisplayHeight: 320,
		Speeds: []SpeedOption{
			{Value: 7, Label: "7 cycles/frame"},
			{Value: 15, Label: "15 cycles/frame"},
			{Value: 20, Label: "20 cycles/frame"},
			{Value: 30, Label: "30 cycles/frame"},
			{Value: 100, Label: "100 cycles/frame"},
			{Value: 200, Label: "200 cycles/frame"},
			{Value: 500, Label: "500 cycles/frame"},
			{Value: 1000, Label: "1000 cycles/frame"},
		},
		ROMs: roms,
	}

	var buf bytes.Buffer
	if err := ui.template.Execute(&buf, data); err != nil {
		return err
	}

	ui.container.Set("innerHTML", buf.String())
	return nil
}

// attachHandler registers an event handler for an element
func (ui *UI) attachHandler(elementKey, event string, handler func(js.Value, []js.Value) interface{}) {
	elem := ui.elements[elementKey]
	if elem.IsUndefined() || elem.IsNull() {
		return
	}

	jsFunc := js.FuncOf(handler)
	ui.handlers[elementKey+":"+event] = jsFunc
	elem.Call("addEventListener", event, jsFunc)
}

// Event handlers with emulator side effects
func (ui *UI) handleStart(this js.Value, args []js.Value) interface{} {
	ui.emulator.Resume()
	return nil
}

func (ui *UI) handleStop(this js.Value, args []js.Value) interface{} {
	ui.emulator.Pause()
	return nil
}

func (ui *UI) handleReset(this js.Value, args []js.Value) interface{} {
	ui.emulator.Start()
	return nil
}

func (ui *UI) handleSpeedChange(this js.Value, args []js.Value) interface{} {
	event := args[0]
	target := event.Get("target")
	val := target.Get("value").String()

	speed, err := strconv.Atoi(val)
	if err != nil {
		speed = 20
		log.Print("Error setting speed, defaulting to 20")
	}

	ui.emulator.SetIPF(speed)
	return nil
}

func (ui *UI) handleRomLoad(this js.Value, args []js.Value) interface{} {
	event := args[0]
	target := event.Get("target")
	romName := target.Get("value").String()
	if romName != "" {
		content, err := fs.ReadFile(romAssets, "roms/"+romName)
		if err != nil {
			log.Printf("Error loading ROM: %v", err)
			return nil
		}
		ui.emulator.SwapROM(content)
		ui.emulator.Start()
	}
	return nil
}

// Cleanup releases all event handlers
func (ui *UI) Cleanup() {
	for _, handler := range ui.handlers {
		handler.Release()
	}
}
