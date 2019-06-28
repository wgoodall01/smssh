package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/kevinburke/twilio-go"

	_ "github.com/joho/godotenv/autoload"
)

const inFileName string = "in.fifo"
const outFileName string = "out.fifo"

func fatal(msg string, err error) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

var TwilioHostUrl string
var TwilioAccountSid string
var TwilioAuthToken string
var TwilioPhoneNumber string
var UserPhoneNumber string
var ListenAddr string

func main() {
	// Note: TWILIO_HOST_URL should NOT include "/twilio" at the end of the URL.
	// Example: "https://some.domain.here.com", "https://something.ngrok.io"
	TwilioHostUrl = os.Getenv("TWILIO_HOST_URL")
	TwilioAccountSid = os.Getenv("TWILIO_ACCOUNT_SID")
	TwilioAuthToken = os.Getenv("TWILIO_AUTH_TOKEN")
	TwilioPhoneNumber = os.Getenv("TWILIO_PHONE_NUMBER")
	UserPhoneNumber = os.Getenv("USER_PHONE_NUMBER")
	ListenAddr = fmt.Sprintf(":%s", os.Getenv("PORT"))

	// Initialize the twilio client
	client := twilio.NewClient(TwilioAccountSid, TwilioAuthToken, nil)

	// Update Twilio incoming number with correct host URL
	webhookUrl := fmt.Sprintf("%s/twilio", TwilioHostUrl)
	log.Printf("updating twilio webhook URL to %s", webhookUrl)
	numberPage, err := client.IncomingNumbers.GetPage(
		context.Background(),
		url.Values{"PhoneNumber": []string{TwilioPhoneNumber}},
	)
	fatal("could not get application", err)
	if numCount := len(numberPage.IncomingPhoneNumbers); numCount != 1 {
		log.Fatalf("looking for 1 phone number in twilio api, got %d", numCount)
	}
	incomingNumberSid := numberPage.IncomingPhoneNumbers[0].Sid
	incomingNumber, err := client.IncomingNumbers.Update(
		context.Background(),
		incomingNumberSid,
		url.Values{"SmsMethod": []string{"POST"}, "SmsUrl": []string{webhookUrl}},
	)
	fatal("could not update incoming number", err)
	log.Printf("%+v", incomingNumber)

	// Process stdio pipes
	readStdin, writeStdin := io.Pipe()
	readStdout, writeStdout := io.Pipe()

	// Put lines from SMS callbacks onto stdin
	http.HandleFunc("/twilio", func(w http.ResponseWriter, r *http.Request) {
		err := twilio.ValidateIncomingRequest(TwilioHostUrl, TwilioAuthToken, r)
		if err != nil {
			log.Println("Got invalid Twilio webhook request:", err)
			w.WriteHeader(500)
			return
		}

		if from := r.Form.Get("From"); from != UserPhoneNumber {
			log.Println("ignoring message from", from)
			return
		}

		// Write the body to stdin
		body := r.Form.Get("Body")
		log.Println("<<", body)
		if !strings.HasSuffix(body, "\n") {
			body = body + "\n"
		}
		io.WriteString(writeStdin, body)

		w.WriteHeader(200)
	})

	// Read lines from stdout, sending them as 140-character SMS messages.
	go func() {
		scanner := bufio.NewScanner(readStdout)
		for scanner.Scan() {
			line := scanner.Text()
			log.Println(">>", line)
			client.Messages.SendMessage(TwilioPhoneNumber, UserPhoneNumber, line, nil)
		}
	}()

	// Run the shell, restarting it if it closes.
	go func() {
		for {
			// Open a shell
			log.Println("starting shell process...")
			shell := exec.Command("/usr/bin/env", "bash")
			shell.Stdout = writeStdout
			shell.Stderr = writeStdout
			shell.Stdin = readStdin
			err := shell.Start()
			fatal("shell run failed", err)
			log.Println("shell started")

			err = shell.Wait()
			log.Println("shell stopped")
			if exitErr, ok := err.(*exec.ExitError); ok {
				// Just an exit code, continue.
				log.Printf("shell exited: %v\n", exitErr)
			} else {
				// Fail if it's something else
				fatal("shell exit failed", err)
			}
		}
	}()

	log.Println("listening for incoming messages")
	http.ListenAndServe(":8080", nil)
}
