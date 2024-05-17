package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"regexp"
)

type Body struct {
	Prepend bool  `json:"prepend"`
	Batch   Batch `json:"batch"`
}

type Batch struct {
	Graph Graph    `json:"graph"`
	Runs  int      `json:"runs"`
	Data  [][]Item `json:"data"`
}

type Graph struct {
	Id    string `json:"id"`
	Nodes Nodes  `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type CoreMetadata struct {
	Node
	ControlLayers        ControlLayers `json:"control_layers,omitempty"`
	CfgRescaleMultiplier int           `json:"cfg_rescale_multiplier"`
	NegativePrompt       string        `json:"negative_prompt"`
	ClipSkipInt          uint          `json:"clip_skip"`
}

type ClipSkip struct {
	Node
	SkippedLayers  uint `json:"skipped_layers"`
	IsIntermediate bool `json:"is_intermediate"`
}

type Conditioning struct {
	Node
	Type           string `json:"type"`
	Id             string `json:"id"`
	Prompt         string `json:"prompt"`
	IsIntermediate bool   `json:"is_intermediate"`
}

type Noise struct {
	Node
	Seed           uint8 `json:"seed"`
	IsIntermediate bool  `json:"is_intermediate"`
}

type Latents struct {
	Node
	IsIntermediate       bool    `json:"is_intermediate"`
	CfgRescaleMultiplier int     `json:"cfg_rescale_multiplier"`
	DenoisingStart       float64 `json:"denoising_start"`
}

type ToImage struct {
	Node
	IsIntermediate bool `json:"is_intermediate"`
	UseCache       bool `json:"use_cache"`
}

type MainModelLoader struct {
	Node
	IsIntermediate bool `json:"is_intermediate"`
}

type Node struct {
	Type           string  `json:"type"`
	Id             string  `json:"id"`
	Model          *Model  `json:"model,omitempty"`
	Width          int     `json:"width,omitempty"`
	Height         int     `json:"height,omitempty"`
	UseCPU         bool    `json:"use_cpu,omitempty"`
	CfgScale       float64 `json:"cfg_scale,omitempty"`
	Scheduler      string  `json:"scheduler,omitempty"`
	Steps          int     `json:"steps,omitempty"`
	DenoisingEnd   float64 `json:"denoising_end,omitempty"`
	Fp32           bool    `json:"fp32,omitempty"`
	GenerationMode string  `json:"generation_mode,omitempty"`
	RandDevice     string  `json:"rand_device,omitempty"`
}

type Nodes struct {
	MainModelLoader             MainModelLoader `json:"main_model_loader,omitempty"`
	ClipSkip                    ClipSkip        `json:"clip_skip,omitempty"`
	PositiveConditioning        Conditioning    `json:"positive_conditioning,omitempty"`
	NegativeConditioning        Conditioning    `json:"negative_conditioning,omitempty"`
	Noise                       Noise           `json:"noise,omitempty"`
	DenoiseLatents              Latents         `json:"denoise_latents,omitempty"`
	LatentsToImage              ToImage         `json:"latents_to_image,omitempty"`
	CoreMetadata                CoreMetadata    `json:"core_metadata,omitempty"`
	PositiveConditioningCollect Node            `json:"positive_conditioning_collect,omitempty"`
	NegativeConditioningCollect Node            `json:"negative_conditioning_collect,omitempty"`
}

type Model struct {
	Key  string `json:"key"`
	Hash string `json:"hash"`
	Name string `json:"name"`
	Base string `json:"base"`
	Type string `json:"type"`
}

type ControlLayers struct {
	Layers  []interface{} `json:"layers"`
	Version int           `json:"version"`
}

type Edge struct {
	Source      Field `json:"source"`
	Destination Field `json:"destination"`
}

type Field struct {
	NodeID string `json:"node_id"`
	Field  string `json:"field"`
}

type Item struct {
	NodePath  string        `json:"node_path"`
	FieldName string        `json:"field_name"`
	Items     []interface{} `json:"items"`
}

type Data struct {
	Data [][]Item `json:"data"`
}

func getPrompt() string {
	base_filepath := "/Users/adellehousker/fun/website/cs/blog/content"
	source_filepath := fmt.Sprintf("%s/%s", base_filepath, "en/posts")
	source_filename := os.Args[1]

	content, err := os.ReadFile(fmt.Sprintf("%s/%s", source_filepath, source_filename))
	if err != nil {
		fmt.Printf("err reading file %s\n", err)
	}

	re_title := regexp.MustCompile(`title = '(\w+)'`)
	source_title := re_title.FindStringSubmatch(string(content))[1]

	re_body := regexp.MustCompile(`\+{3}(?s).*\+{3}((?s).*)`)
	source_body := re_body.FindStringSubmatch(string(content))[1]

	return fmt.Sprintf("%s %s", source_title, source_body)
}

func getData(prompt string) [][]Item {
	min := int64(math.Pow(10, 10))
	max := int64(math.Pow(10, 11))
	seed := rand.Int63n(int64(max-min) + min)

	item1 := Item{
		NodePath:  "noise",
		FieldName: "seed",
		Items:     []interface{}{seed},
	}

	item2 := Item{
		NodePath:  "core_metadata",
		FieldName: "seed",
		Items:     []interface{}{seed},
	}

	item3 := Item{
		NodePath:  "positive_conditioning",
		FieldName: "prompt",
		Items:     []interface{}{prompt},
	}

	item4 := Item{
		NodePath:  "core_metadata",
		FieldName: "positive_prompt",
		Items:     []interface{}{prompt},
	}

	return [][]Item{
		{item1, item2},
		{item3, item4},
	}
}

func getBody(prompt string, data [][]Item) Body {
	file, err := os.Open("/Users/adellehousker/fun/ai/Columbia/project3/columbia-project3/Adelle/blog-main/imaging/graph.json")
	if err != nil {
		fmt.Println("Error opening JSON file:", err)
		return Body{}
	}
	defer file.Close()

	var graph Graph
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&graph)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return Body{}
	}

	graph.Nodes.PositiveConditioning = Conditioning{
		Type:           "compel",
		Id:             "positive_conditioning",
		Prompt:         prompt,
		IsIntermediate: true,
	}

	return Body{
		Prepend: false,
		Batch: Batch{
			Graph: graph,
			Runs:  1,
			Data:  data,
		},
	}
}

type LoggingTransport struct {
	Transport http.RoundTripper
}

// FOR DEBUGGING
func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Print the request method, URL, and headers
	fmt.Println("Request:")
	fmt.Printf("%s %s %s\n", req.Method, req.URL, req.Proto)
	req.Header.Write(os.Stdout)
	fmt.Println()

	// Print the request body if present
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		fmt.Println("Request Body:")
		fmt.Println(string(body))
		// Restore the request body so it can be read again
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	// Proceed with the actual HTTP request
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func main() {
	prompt := getPrompt()
	data := getData(prompt)
	body := getBody(prompt, data)

	payload, err := json.Marshal(body)
	if err != nil {
		fmt.Println("Error marshaling body:", err)
		return
	}

	client := &http.Client{
		// FOR DEBUGGING
		// Transport: &LoggingTransport{
		// 	Transport: http.DefaultTransport,
		// },
	}

	req, err := http.NewRequest("POST", "http://127.0.0.1:9091/api/v1/queue/default/enqueue_batch", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Origin", "http://127.0.0.1:9091")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "http://127.0.0.1:9091/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Sec-Ch-Ua", "\"Chromium\";v=\"124\", \"Google Chrome\";v=\"124\", \"Not-A.Brand\";v=\"99\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "\"macOS\"")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status)
}
