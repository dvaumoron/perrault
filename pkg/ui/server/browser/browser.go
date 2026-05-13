/*
 *
 * Copyright 2026 perrault authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package browser

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"

	ui "github.com/dvaumoron/perrault/pkg/ui"
)

const defaultHttpPort = 8080

var _ ui.UI = &UI{}

type UI struct {
	Port           int
	WaitScreen     []byte
	ChoiceTemplate template.Template

	choiceChan     chan int
	inDisplayChan  chan<- []byte
	outDisplayChan <-chan []byte
}

func NewUI() *UI {
	return &UI{
		Port: defaultHttpPort,
	}
}

func (ui *UI) Start() {
	choiceChan := make(chan int)
	inDisplayChan := make(chan []byte)
	outDisplayChan := make(chan []byte)

	ui.choiceChan = choiceChan
	ui.inDisplayChan = inDisplayChan
	ui.outDisplayChan = outDisplayChan

	go updateBuffer(ui.WaitScreen, inDisplayChan, outDisplayChan)

	mux := http.NewServeMux()
	mux.HandleFunc("/", ui.handleDisplay)
	mux.HandleFunc("/{n}", ui.handleChoice)

	s := &http.Server{
		Addr:    ":" + strconv.Itoa(ui.Port),
		Handler: mux,
	}

	go runAndExit(s)
}

func (ui *UI) AskUserChoice(title string, choices []string) int {
	var buffer bytes.Buffer
	err := ui.ChoiceTemplate.Execute(&buffer, choiceData{
		Title:   title,
		Choices: choices,
	})
	if err != nil {
		return -1
	}

	ui.inDisplayChan <- buffer.Bytes()
	n := <-ui.choiceChan
	if n < 0 || n >= len(choices) {
		return -1
	}
	return n
}

func (ui *UI) handleDisplay(w http.ResponseWriter, r *http.Request) {
	w.Write(<-ui.outDisplayChan)
}

func (ui *UI) handleChoice(w http.ResponseWriter, r *http.Request) {
	nStr := r.PathValue("n")
	if nStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	n, err := strconv.Atoi(nStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ui.inDisplayChan <- ui.WaitScreen
	ui.choiceChan <- n

	http.Redirect(w, r, "/", http.StatusFound)
}

type choiceData struct {
	Title   string
	Choices []string
}

func updateBuffer(buffer []byte, inDisplayChan <-chan []byte, outDisplayChan chan<- []byte) {
	for {
		select {
		case updatedBuffer := <-inDisplayChan:
			buffer = updatedBuffer
		case outDisplayChan <- buffer:
		}
	}
}

func runAndExit(s *http.Server) {
	if err := s.ListenAndServe(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
