package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/rpi"
)

func init() {
	// init needs to be outside the goroutine
	if _, err := host.Init(); err != nil {
		panic(err)
	}
}

func main() {

	discordWebhookURL, ok := os.LookupEnv("DISCORD_WEBHOOK_URL")
	if !ok {
		fmt.Fprintln(
			os.Stderr,
			"'DISCORD_WEBHOOK_URL' env var is required",
		)
		os.Exit(1)
	}

	// physical pin number my boy!
	if err := rpi.P1_12.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		log.Fatal(err)
	}

	spotted := false
	fmt.Println("Sensing enabled")
	for {
		if rpi.P1_12.WaitForEdge(time.Millisecond * 500) {
			if spotted {
				spotted = false
				continue
			}

			spotted = true
			fmt.Println("Spotted! Taking snapshot")

			// The camera can only take one image at a time.
			// Cannot run in separate goroutines without synchronization
			// So just run the http code in the goroutine.
			capture, err := captureImage()
			if err != nil {
				log.Fatal("Error capturing picture:", err)
			}
			fmt.Println("Picture captured")

			go sendImage(discordWebhookURL, capture)
		}

	}
}

// Takes a picture using raspicam-still, sends it to Discord, and then deletes the file.
// If sending the picture to Discord fails, the file is not deleted.
func sendImage(discordWebhookURL string, imageFile string) {
	multipartReq, err := newMultiPartRequest(imageFile, discordWebhookURL)

	if err != nil {
		fmt.Fprintln(
			os.Stderr,
			"Failed to create file multipart request:",
			err,
		)
		return
	}

	if err := sendRequest(multipartReq); err != nil {
		fmt.Fprintln(
			os.Stderr,
			"Capture failed to post to Discord channel:",
			err,
		)
		return
	}

	fmt.Println("Image posted to Discord channel")
	os.Remove(imageFile)
}

// Use rpicam-still to take a picture and save it to a temporary file
//
// Returns the name of the temporary file or an error
func captureImage() (string, error) {
	output, err := os.CreateTemp("", "capture*.jpg")
	if err != nil {
		return "", err
	}

	output.Close() // Not using the ptr, close it before rpicam-jpeg

	cmd := exec.Command(
		"rpicam-jpeg",
		"--output", output.Name(),
	)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to capture image: %s", err)
	}

	return output.Name(), nil
}

// Create the request to send to Discord
func newMultiPartRequest(imagePath string, discordWebhook string) (*http.Request, error) {
	f, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	body := bytes.Buffer{}
	bodywriter := multipart.NewWriter(&body)

	part, err := bodywriter.CreateFormFile(
		"multipart/form-data",
		filepath.Base(f.Name()),
	)

	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, f)
	if err != nil {
		return nil, err
	}

	err = bodywriter.Close()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, discordWebhook, &body)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", bodywriter.FormDataContentType())
	return request, nil
}

func sendRequest(request *http.Request) error {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"invalid response from discord channel: %s",
			response.Status,
		)
	}

	return nil
}
