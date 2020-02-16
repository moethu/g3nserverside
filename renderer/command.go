package renderer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/g3n/engine/window"
)

// Command received from client
type Command struct {
	X     float32
	Y     float32
	Cmd   string
	Val   string
	Moved bool
	Ctrl  bool
}

// mapMouseButton maps js mouse buttons to window mouse buttons
func mapMouseButton(value string) window.MouseButton {
	switch value {
	case "0":
		return window.MouseButtonLeft
	case "1":
		return window.MouseButtonMiddle
	case "2":
		return window.MouseButtonRight
	default:
		return window.MouseButtonLeft
	}
}

// mapKey maps js keys to window keys
func mapKey(value string) window.Key {
	switch value {
	case "38":
		return window.KeyUp
	case "37":
		return window.KeyLeft
	case "39":
		return window.KeyRight
	case "40":
		return window.KeyDown
	default:
		return window.KeyEnter
	}
}

// commandLoop listens for incoming commands and forwards them to the rendering app
func (app *RenderingApp) commandLoop() {
	for {
		message := <-app.cCommands

		cmd := Command{}
		err := json.Unmarshal(message, &cmd)
		if err != nil {
			app.Log().Error(err.Error())
		}

		if cmd.Cmd != "" {
			app.Log().Info("received command: %v", cmd)
		}

		switch cmd.Cmd {
		case "":
			cev := window.CursorEvent{Xpos: cmd.X, Ypos: cmd.Y}
			app.Orbit().OnCursorPos(&cev)
		case "mousedown":
			mev := window.MouseEvent{Xpos: cmd.X, Ypos: cmd.Y,
				Action: window.Press,
				Button: mapMouseButton(cmd.Val)}

			if cmd.Moved {
				app.imageSettings.isNavigating = true
			}

			app.Orbit().OnMouse(&mev)
		case "zoom":
			scrollFactor := float32(10.0)
			mev := window.ScrollEvent{Xoffset: cmd.X, Yoffset: -cmd.Y / scrollFactor}
			app.Orbit().OnScroll(&mev)
		case "mouseup":
			mev := window.MouseEvent{Xpos: cmd.X, Ypos: cmd.Y,
				Action: window.Release,
				Button: mapMouseButton(cmd.Val)}

			app.imageSettings.isNavigating = false
			app.Orbit().OnMouse(&mev)

			// mouse left click
			if cmd.Val == "0" && !cmd.Moved {
				app.selectNode(cmd.X, cmd.Y, cmd.Ctrl)
			}
		case "hide":
			for inode := range app.selectionBuffer {
				inode.GetNode().SetVisible(false)
			}
			app.resetSelection()
		case "unhide":
			for _, node := range app.nodeBuffer {
				node.SetVisible(true)
			}
		case "userdata":
			if node, ok := app.nodeBuffer[cmd.Val]; ok {
				app.sendMessageToClient("userdata", fmt.Sprintf("%v", node.UserData()))
			}
		case "keydown":
			kev := window.KeyEvent{Action: window.Press, Mods: 0, Keycode: mapKey(cmd.Val)}
			app.Orbit().OnKey(&kev)
		case "keyup":
			kev := window.KeyEvent{Action: window.Release, Mods: 0, Keycode: mapKey(cmd.Val)}
			app.Orbit().OnKey(&kev)
		case "view":
			app.setCamera(cmd.Val)
		case "zoomextent":
			app.zoomToExtent()
		case "focus":
			app.focusOnSelection()
		case "invert":
			if app.imageSettings.invert {
				app.imageSettings.invert = false
			} else {
				app.imageSettings.invert = true
			}
		case "imagesettings":
			s := strings.Split(cmd.Val, ":")
			if len(s) == 5 {
				br, err := strconv.Atoi(s[0])
				if err == nil {
					app.imageSettings.brightness = float64(getValueInRange(br, -100, 100))
				}
				ct, err := strconv.Atoi(s[1])
				if err == nil {
					app.imageSettings.contrast = float64(getValueInRange(ct, -100, 100))
				}
				sa, err := strconv.Atoi(s[2])
				if err == nil {
					app.imageSettings.saturation = float64(getValueInRange(sa, -100, 100))
				}
				bl, err := strconv.Atoi(s[3])
				if err == nil {
					app.imageSettings.blur = float64(getValueInRange(bl, 0, 20))
				}
				pix, err := strconv.ParseFloat(s[4], 64)
				if err == nil {
					app.imageSettings.pixelation = getFloatValueInRange(pix, 1.0, 10.0)
				}
			}
		case "quality":
			quality, err := strconv.Atoi(cmd.Val)
			if err == nil {
				switch quality {
				case 0:
					app.imageSettings.quality = highQ
				case 2:
					app.imageSettings.quality = lowQ
				default:
					app.imageSettings.quality = mediumQ
				}
			}
		case "fov":
			fov, err := strconv.Atoi(cmd.Val)
			if err == nil {
				app.CameraPersp().SetFov(float32(getValueInRange(fov, 5, 120)))
			}
		case "debug":
			if app.Debug {
				app.Debug = false
			} else {
				app.Debug = true
			}
		case "close":
			app.Log().Info("close")
			app.Window().SetShouldClose(true)
		default:
			app.Log().Info("Unknown Command: " + cmd.Cmd)
		}
	}
}

// getValueInRange returns a value within bounds
func getValueInRange(value int, lower int, upper int) int {
	if value > upper {
		return upper
	} else if value < lower {
		return lower
	} else {
		return value
	}
}

// getFloatValueInRange returns a value within bounds
func getFloatValueInRange(value float64, lower float64, upper float64) float64 {
	if value > upper {
		return upper
	} else if value < lower {
		return lower
	} else {
		return value
	}
}
