package main

import (
	"net"
	"net/http"
	"os"
	"time"

	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

	helloButton := widget.NewButton("Connect", func() {
		//		url := memo.Text
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
			msg = msg + "<rand_num>" + fmt.Sprintf("%d", 1) + "</rand_num>"
			msg = msg + "/<message>\n"
			event := msg
			//		event := fmt.Sprintf("Controller=%s Time=%v\n", GetOutboundIP(), dtime[0:24])
			agent.Notifier <- []byte(event)
		}
	}()
	go fmt.Printf("Listening at  : %s Port : %s\n", xip, port)
	go http.ListenAndServe(":"+port, agent)

	w.ShowAndRun()

	//	if err := http.ListenAndServe(xip+":"+port, nil); err != nil {
	//		panic(err)
	//	}
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

func ReadURL(url string) {

	return
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
