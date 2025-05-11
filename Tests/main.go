package main

import (
	"net"
	"net/http"
	"os"

	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ----------------------------------------------------------------
func main() {
	xip := fmt.Sprintf("%s", GetOutboundIP())
	port := "8080"
	a := app.New()
	w := a.NewWindow("Listening on " + xip + ":" + port)

	memo := widget.NewEntry()
	memo.SetPlaceHolder("Enter your memo here...")
	memo.MultiLine = true               // Enable multiline for larger text fields
	memo.Resize(fyne.NewSize(400, 100)) // Adjust the height (4x the default)

	helloButton := widget.NewButton("Say Hello", func() {
		// Display the value from the memo field in the dialog box
		dialog.ShowInformation("Hello", "Hello, "+memo.Text, w)
	})
	exitButton := widget.NewButton("Exit", func() {
		os.Exit(0)
	})

	w.SetContent(container.NewVBox(
		memo,        // Add the memo field
		helloButton, // Add the "Say Hello" button
		exitButton,  // Add the "Exit" button
	))
	w.Resize(fyne.NewSize(400, 300))
	w.ShowAndRun()
	if err := http.ListenAndServe(xip+":"+port, nil); err != nil {
		panic(err)
	}
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
