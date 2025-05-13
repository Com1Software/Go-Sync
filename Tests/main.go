package main

import (
	"bufio"
	"encoding/xml"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ----------------------------------------------------------------
func main() {
	agent := SSE()
	xip := fmt.Sprintf("%s", GetOutboundIP())
	port := "8080"
	a := app.New()
	w := a.NewWindow("Listening on " + xip + ":" + port)
	tctl := 0
	tc := 0
	memo := widget.NewEntry()
	memo.SetPlaceHolder("Enter an IP address to sync with...")
	memo.MultiLine = true               // Enable multiline for larger text fields
	memo.Resize(fyne.NewSize(400, 100)) // Adjust the height (4x the default)

	memo1 := widget.NewEntry()
	memo1.SetPlaceHolder("...")
	memo1.MultiLine = true               // Enable multiline for larger text fields
	memo1.Resize(fyne.NewSize(400, 100)) // Adjust the height (4x the default)

	input := widget.NewEntry()
	input.SetPlaceHolder("Enter Command...")

	inputvalue := widget.NewEntry()
	inputvalue.SetPlaceHolder("Enter Value...")

	button := widget.NewButton("Cmd", func() {
	})

	helloButton := widget.NewButton("Connect", func() {
		url := memo.Text
		go func() {
			for {
				data := ReadURL(url)

				// Use fyne.Do to safely update the UI from a goroutine
				fyne.Do(func() {
					memo1.SetText(data)
				})

				time.Sleep(1 * time.Second) // Adjust the refresh rate as needed
			}
		}()

	})
	exitButton := widget.NewButton("Exit", func() {
		os.Exit(0)
	})

	inputContainer := container.NewGridWrap(fyne.NewSize(165, 40), input)
	inputContainerValue := container.NewGridWrap(fyne.NewSize(165, 40), inputvalue)
	buttonContainer := container.NewGridWrap(fyne.NewSize(50, 40), button)
	layoutContainer := container.NewHBox(buttonContainer, inputContainer, inputContainerValue)

	w.SetContent(container.NewVBox(
		memo,            // Add the memo field
		memo1,           // Add the second memo field
		layoutContainer, // Add the input field and button
		helloButton,     // Add the "Say Hello" button
		exitButton,      // Add the "Exit" button
	))
	w.Resize(fyne.NewSize(400, 300))

	go func() {
		for {
			switch {
			case tctl == 0:
				time.Sleep(time.Second * 1)
			case tctl == 1:
				time.Sleep(time.Second * -1)
				tc++
				fmt.Printf("loop count = %d\n", tc)
			}
			dtime := fmt.Sprintf("%s", time.Now())
			msg := "<message>"
			msg = msg + "<controller>" + fmt.Sprint(GetOutboundIP()) + "</controller>"
			msg = msg + "<date_time>" + dtime[0:24] + "</date_time>"
			msg = msg + "<rand_num>" + fmt.Sprintf("%s", "Test") + "</rand_num>"
			msg = msg + "/<message>\n"
			event := msg
			//		event := fmt.Sprintf("Controller=%s Time=%v\n", GetOutboundIP(), dtime[0:24])
			agent.Notifier <- []byte(event)
		}
	}()
	go fmt.Printf("Listening at  : %s Port : %s\n", xip, port)
	go http.ListenAndServe(":"+port, agent)

	w.ShowAndRun()

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

type message struct {
	Controller string `xml:"controller"`
	DateTime   string `xml:"date_time"`
	Command    string `xml:"command"`
	Value      string `xml:"value"`
}

func ReadURL(url string) string {
	msg := &message{}
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body)

	for {
		line, erra := reader.ReadBytes('\n')
		if erra != nil {
			log.Fatal(erra) // Ensure correct error logging
		}
		xml.Unmarshal(line, &msg)
		break // Exit after the first read
	}

	// Construct the string with extracted values
	return msg.Controller + " " + msg.DateTime + " " + msg.Command + " " + msg.Value
}

type Agent struct {
	Notifier    chan []byte
	newuser     chan chan []byte
	closinguser chan chan []byte
	user        map[chan []byte]bool
}

func SSE() (agent *Agent) {
	agent = &Agent{
		Notifier:    make(chan []byte, 1),
		newuser:     make(chan chan []byte),
		closinguser: make(chan chan []byte),
		user:        make(map[chan []byte]bool),
	}
	go agent.listen()
	return
}

func (agent *Agent) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Error ", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	mChan := make(chan []byte)
	agent.newuser <- mChan
	defer func() {
		agent.closinguser <- mChan
	}()
	notify := req.Context().Done()
	go func() {
		<-notify
		agent.closinguser <- mChan
	}()
	for {
		fmt.Fprintf(rw, "%s", <-mChan)
		flusher.Flush()
	}

}

func (agent *Agent) listen() {
	for {
		select {
		case s := <-agent.newuser:
			agent.user[s] = true
		case s := <-agent.closinguser:
			delete(agent.user, s)
		case event := <-agent.Notifier:
			for userMChan, _ := range agent.user {
				userMChan <- event
			}
		}
	}

}
