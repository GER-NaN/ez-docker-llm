package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sync"
	"time"
)

// DTOs
type healthReply struct {
	Status string   `json:"status"`
	Llama  string   `json:"llama"`
	Models int      `json:"models"`
	Loaded []string `json:"loaded"`
}

type modelReply struct {
	Name   string `json:"name"`
	Loaded bool   `json:"loaded"`
}

type promptRequest struct {
	Model  string          `json:"model"`
	Prompt string          `json:"prompt"`
	Text   string          `json:"text"`
	Format json.RawMessage `json:"format"`
}

type promptReply struct {
	Model   string  `json:"model"`
	Answer  any     `json:"answer"`
	Seconds float64 `json:"seconds"`
}

type errorReply struct {
	Error  string `json:"error"`
	Answer string `json:"answer,omitempty"`
}

// llama.cpp DTOS

type llamaHealth struct {
	Status string `json:"status"`
}

type llamaModels struct {
	Data []llamaModel `json:"data"`
}

type llamaModel struct {
	ID     string      `json:"id"`
	Status llamaStatus `json:"status"`
}

type llamaStatus struct {
	Value string `json:"value"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type       string     `json:"type"`
	JSONSchema jsonSchema `json:"json_schema"`
}

type jsonSchema struct {
	Name   string          `json:"name"`
	Schema json.RawMessage `json:"schema"`
}

type chatReply struct {
	Choices []chatChoice `json:"choices"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

// Our llama.cpp instance we talk with
var llamaURL = "http://llama:8080"
var port = "8085"
var client = &http.Client{Timeout: 15 * time.Minute}

// One prompt runs at a time.
var promptLock sync.Mutex

func getJSON(path string, out any) error {
	resp, err := client.Get(llamaURL + path)
	if err != nil {
		return err
	}
	return readJSON(resp, out)
}

func postJSON(path string, body any, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := client.Post(llamaURL+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	return readJSON(resp, out)
}

func readJSON(resp *http.Response, out any) error {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("llama.cpp answered %d: %s", resp.StatusCode, data)
	}
	return json.Unmarshal(data, out)
}

func listModels() ([]modelReply, error) {
	var found llamaModels
	err := getJSON("/v1/models", &found)
	if err != nil {
		return nil, err
	}
	models := []modelReply{}
	for _, m := range found.Data {
		models = append(models, modelReply{Name: m.ID, Loaded: m.Status.Value == "loaded"})
	}
	return models, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorReply{Error: message})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	down := healthReply{Status: "down", Loaded: []string{}}

	var llama llamaHealth
	err := getJSON("/health", &llama)
	if err != nil {
		down.Llama = err.Error()
		writeJSON(w, http.StatusServiceUnavailable, down)
		return
	}
	models, err := listModels()
	if err != nil {
		down.Llama = err.Error()
		writeJSON(w, http.StatusServiceUnavailable, down)
		return
	}

	loaded := []string{}
	for _, m := range models {
		if m.Loaded {
			loaded = append(loaded, m.Name)
		}
	}
	writeJSON(w, http.StatusOK, healthReply{
		Status: "ok",
		Llama:  llama.Status,
		Models: len(models),
		Loaded: loaded,
	})
}

func handleModels(w http.ResponseWriter, r *http.Request) {
	models, err := listModels()
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models)
}

func handlePrompt(w http.ResponseWriter, r *http.Request) {
	var in promptRequest
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		writeError(w, http.StatusBadRequest, "body must be JSON: "+err.Error())
		return
	}
	if in.Model == "" || in.Prompt == "" {
		writeError(w, http.StatusBadRequest, "model and prompt are required")
		return
	}

	models, err := listModels()
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	known := false
	for _, m := range models {
		if m.Name == in.Model {
			known = true
		}
	}
	if !known {
		writeError(w, http.StatusNotFound, "no model named "+in.Model)
		return
	}

	chat := chatRequest{Model: in.Model}
	if in.Text == "" {
		chat.Messages = []chatMessage{
			{Role: "user", Content: in.Prompt},
		}
	} else {
		chat.Messages = []chatMessage{
			{Role: "system", Content: in.Prompt},
			{Role: "user", Content: in.Text},
		}
	}

	hasFormat := len(in.Format) > 0 && string(in.Format) != "null"
	if hasFormat {
		chat.ResponseFormat = &responseFormat{
			Type:       "json_schema",
			JSONSchema: jsonSchema{Name: "answer", Schema: in.Format},
		}
	}

	promptLock.Lock()
	started := time.Now()
	var out chatReply
	err = postJSON("/v1/chat/completions", chat, &out)
	seconds := time.Since(started).Seconds()
	promptLock.Unlock()

	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if len(out.Choices) == 0 {
		writeError(w, http.StatusBadGateway, "llama.cpp gave no answer")
		return
	}

	answer := out.Choices[0].Message.Content
	result := promptReply{
		Model:   in.Model,
		Answer:  answer,
		Seconds: math.Round(seconds*100) / 100,
	}
	if hasFormat {
		if !json.Valid([]byte(answer)) {
			writeJSON(w, http.StatusBadGateway, errorReply{Error: "answer is not valid JSON", Answer: answer})
			return
		}
		result.Answer = json.RawMessage(answer)
	}
	writeJSON(w, http.StatusOK, result)
}

func main() {
	if os.Getenv("LLAMA_URL") != "" {
		llamaURL = os.Getenv("LLAMA_URL")
	}
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	http.HandleFunc("GET /health", handleHealth)
	http.HandleFunc("GET /models", handleModels)
	http.HandleFunc("POST /prompt", handlePrompt)

	log.Printf("listening on :%s, llama.cpp at %s", port, llamaURL)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
