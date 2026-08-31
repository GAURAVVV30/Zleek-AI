package aiengine

// VoiceEngine is a faithful Go port of app/core/voice_engine.py.
//
// ASR: NVIDIA Riva (NIM, OpenAI-compatible /audio/transcriptions) → Groq
//      Whisper fallback. Both use multipart uploads exactly like Python.
// TTS: NVIDIA Riva (NIM) only. The MMS-TTS / SpeechT5 fallback is a local
//      PyTorch pipeline, which cannot run inside the Go runtime; without
//      NVIDIA_API_KEY synthesis fails with an explicit error (matching the
//      Python failure path when the local model cannot load).

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	nvidiaASRModel  = "nvidia/parakeet-ctc-1.1b-asr"
	nvidiaTTSVoice  = "English-US.Female-1"
	nvidiaRivaTTS   = "nvidia/riva-tts"
	groqWhisper     = "whisper-large-v3-turbo"
	voiceSampleRate = 16000
)

var voiceAudioTypes = map[string]string{
	"audio/wav":   "audio.wav",
	"audio/x-wav": "audio.wav",
	"audio/mpeg":  "audio.mp3",
	"audio/mp3":   "audio.mp3",
	"audio/webm":  "audio.webm",
	"audio/ogg":   "audio.ogg",
	"audio/flac":  "audio.flac",
	"audio/mp4":   "audio.mp4",
}

type VoiceEngine struct {
	nvidiaKey string
	groqKey   string
	client    *http.Client
}

func NewVoiceEngine() *VoiceEngine {
	return &VoiceEngine{
		nvidiaKey: strings.TrimSpace(os.Getenv("NVIDIA_API_KEY")),
		groqKey:   strings.TrimSpace(os.Getenv("GROQ_API_KEY")),
		client:    &http.Client{Timeout: 90 * time.Second},
	}
}

// Transcribe mirrors VoiceEngine.transcribe(): Riva → Groq cascade.
func (v *VoiceEngine) Transcribe(audioBytes []byte, mimeType string) (string, map[string]any, error) {
	if v.nvidiaKey != "" {
		text, meta, err := v.asrMultipart(
			"https://integrate.api.nvidia.com/v1", v.nvidiaKey,
			nvidiaASRModel, false, audioBytes, mimeType,
		)
		if err == nil {
			return text, meta, nil
		}
	}
	if v.groqKey != "" {
		text, meta, err := v.asrMultipart(
			"https://api.groq.com/openai/v1", v.groqKey,
			groqWhisper, true, audioBytes, mimeType,
		)
		if err != nil {
			return "", nil, fmt.Errorf("all ASR providers failed. Last error: %v", err)
		}
		return text, meta, nil
	}
	return "", nil, fmt.Errorf("no ASR provider available. Set NVIDIA_API_KEY (Riva) or GROQ_API_KEY (Whisper)")
}

// asrMultipart POSTs audio to /audio/transcriptions (OpenAI-compatible).
func (v *VoiceEngine) asrMultipart(baseURL, apiKey, model string, verbose bool, audioBytes []byte, mimeType string) (string, map[string]any, error) {
	filename := voiceAudioTypes[strings.ToLower(mimeType)]
	if filename == "" {
		filename = "audio.webm"
	}
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if err := w.WriteField("model", model); err != nil {
		return "", nil, err
	}
	if verbose {
		if err := w.WriteField("response_format", "verbose_json"); err != nil {
			return "", nil, err
		}
		if err := w.WriteField("language", "en"); err != nil {
			return "", nil, err
		}
	}
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return "", nil, err
	}
	if _, err := fw.Write(audioBytes); err != nil {
		return "", nil, err
	}
	if err := w.Close(); err != nil {
		return "", nil, err
	}

	req, err := http.NewRequest(http.MethodPost, strings.TrimSuffix(baseURL, "/")+"/audio/transcriptions", &buf)
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := v.client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", nil, fmt.Errorf("ASR HTTP %d: %s", resp.StatusCode, truncate(string(raw), 300))
	}

	var payload struct {
		Text     string   `json:"text"`
		Duration *float64 `json:"duration"`
		Language string   `json:"language"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", nil, err
	}

	meta := map[string]any{"provider": "nvidia_riva", "model": nvidiaASRModel}
	if verbose {
		meta = map[string]any{"provider": "groq_whisper", "model": groqWhisper}
		if payload.Language != "" {
			meta["language"] = payload.Language
		}
		if payload.Duration != nil {
			meta["duration_seconds"] = floatRound(*payload.Duration, 2)
		}
	}
	return strings.TrimSpace(payload.Text), meta, nil
}

// Synthesize mirrors VoiceEngine.synthesize(): Riva TTS via NIM, else error.
func (v *VoiceEngine) Synthesize(text string) ([]byte, error) {
	if v.nvidiaKey != "" {
		return v.synthesizeRiva(text)
	}
	return nil, fmt.Errorf("TTS unavailable: no NVIDIA_API_KEY. The MMS-TTS fallback (facebook/mms-tts-eng) is a local PyTorch pipeline not available in the Go runtime")
}

func (v *VoiceEngine) synthesizeRiva(text string) ([]byte, error) {
	if len(text) > 4096 {
		text = text[:4096]
	}
	payload := map[string]any{"model": nvidiaRivaTTS, "voice": nvidiaTTSVoice, "input": text}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, "https://integrate.api.nvidia.com/v1/audio/speech", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+v.nvidiaKey)

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TTS HTTP %d: %s", resp.StatusCode, truncate(string(raw), 300))
	}
	return io.ReadAll(resp.Body)
}

// Status mirrors VoiceEngine.status().
func (v *VoiceEngine) Status() map[string]any {
	asrProvider := "none"
	if v.nvidiaKey != "" {
		asrProvider = "nvidia_riva"
	} else if v.groqKey != "" {
		asrProvider = "groq_whisper"
	}
	asrModel := "whisper-large-v3-turbo"
	if v.nvidiaKey != "" {
		asrModel = nvidiaASRModel
	}
	ttsProvider := "mms_tts_local"
	ttsModel := "facebook/mms-tts-eng"
	ttsAvailable := false
	if v.nvidiaKey != "" {
		ttsProvider = "nvidia_riva"
		ttsModel = nvidiaRivaTTS
		ttsAvailable = true
	}
	return map[string]any{
		"asr_available":       v.nvidiaKey != "" || v.groqKey != "",
		"asr_provider":        asrProvider,
		"asr_model":           asrModel,
		"tts_available":       ttsAvailable,
		"tts_provider":        ttsProvider,
		"tts_model":           ttsModel,
		"nvidia_key_present":  v.nvidiaKey != "",
		"groq_key_present":    v.groqKey != "",
		"openai_key_required": false,
	}
}

func floatRound(v float64, places int) float64 {
	pow := math.Pow(10, float64(places))
	return math.Round(v*pow) / pow
}

// Embedding of the WAV writer used by the Python _numpy_to_wav_bytes fallback.
// Kept exported for the handler's silent-audio graceful path.

// BuildSilentWav returns a valid 16kHz mono WAV of `seconds` of silence.
func BuildSilentWav(seconds int) []byte {
	const sampleRate = 16000
	numSamples := sampleRate * seconds
	var hdr bytes.Buffer
	appendStr := func(s string) { hdr.WriteString(s) }
	append32 := func(n uint32) {
		for i := 0; i < 4; i++ {
			hdr.WriteByte(byte(n >> (8 * i)))
		}
	}
	append16 := func(n uint16) {
		for i := 0; i < 2; i++ {
			hdr.WriteByte(byte(n >> (8 * i)))
		}
	}

	dataSize := numSamples * 2
	appendStr("RIFF")
	append32(36 + uint32(dataSize))
	appendStr("WAVE")
	appendStr("fmt ")
	append32(16)
	append16(1) // PCM
	append16(1) // mono
	append32(sampleRate)
	append32(sampleRate * 2)
	append16(2)
	append16(16)
	appendStr("data")
	append32(uint32(dataSize))
	hdr.Write(make([]byte, dataSize))
	return hdr.Bytes()
}
