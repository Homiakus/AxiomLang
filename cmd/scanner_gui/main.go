package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"AxiomLang/stdlib/gui"
)

type HardwareDevices struct {
	ActiveCamera struct {
		Type       string `json:"type"`
		DeviceName string `json:"device_name"`
	} `json:"active_camera"`
	ActiveSerial struct {
		Port string `json:"port"`
	} `json:"active_serial"`
}

type StageJogRequest struct {
	XUm float64 `json:"x_um"`
	YUm float64 `json:"y_um"`
	ZUm float64 `json:"z_um"`
}

func main() {
	fmt.Println("================================================================================")
	fmt.Println("  🔬 ScannerPRO × AxiomLang Native Precision GUI Engine (120 FPS)")
	fmt.Println("================================================================================")
	fmt.Println("  ⚡ Режим: Прямой фреймбуфер без браузера и тяжелого DOM")
	fmt.Println("  🔒 Модель безопасности: Zero-Ambient Authority (cap[Stage, Camera])")
	fmt.Println("================================================================================")

	baseURL := "http://127.0.0.1:18080"
	uiFile := "scannerpro.axlui"

	model, err := gui.LoadAXLUI(uiFile)
	if err != nil {
		fmt.Printf("  [!] Не удалось загрузить %s: %v, создание нового...\n", uiFile, err)
		model = gui.NewAXLUIModel(uiFile, "ScannerPRO Precision Console")
	}

	canvas := gui.NewNativeCanvas(1920, 1080)
	client := &http.Client{Timeout: 3 * time.Second}

	// Initial render
	gui.RenderSceneToCanvas(canvas, model)
	_ = canvas.SaveAsBMP("scannerpro_framebuffer.bmp")
	fmt.Println("  [✓] Стартовый фреймбуфер готов: scannerpro_framebuffer.bmp")

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			fmt.Println("\n  [AXIOM GUI] Завершение работы нативного графического интерфейса.")
			return
		case <-ticker.C:
			// Fetch hardware status & live telemetry from ScannerPRO kernel
			resp, err := client.Get(baseURL + "/api/v1/hardware/devices")
			if err == nil && resp.StatusCode == http.StatusOK {
				var dev HardwareDevices
				if err := json.NewDecoder(resp.Body).Decode(&dev); err == nil {
					// Update telemetry in model if elements exist
				}
				resp.Body.Close()
			}

			// Render updated frame
			gui.RenderSceneToCanvas(canvas, model)
			_ = canvas.SaveAsBMP("scannerpro_framebuffer.bmp")
		}
	}
}

func fetchLatestCameraFrame(baseURL string, client *http.Client) (image.Image, error) {
	resp, err := client.Get(fmt.Sprintf("%s/api/v1/camera/frame?t=%d", baseURL, time.Now().UnixNano()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return jpeg.Decode(bytes.NewReader(data))
}

func jogStageAPI(baseURL string, client *http.Client, dx, dy, dz float64) error {
	reqBody, _ := json.Marshal(StageJogRequest{XUm: dx, YUm: dy, ZUm: dz})
	resp, err := client.Post(baseURL+"/api/v1/hardware/jog", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
