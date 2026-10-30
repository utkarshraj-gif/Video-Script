package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Video struct {
	ID        string    `json:"video_id"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Uploader  string    `json:"uploader"`
	Duration  int       `json:"duration_seconds"`
	Platform  string    `json:"platform"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type YouTubeMetadata struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Uploader   string  `json:"uploader"`
	Duration   float64 `json:"duration"`
	WebpageURL string  `json:"webpage_url"`
}

type YouTubeOEmbed struct {
	Title        string `json:"title"`
	AuthorName   string `json:"author_name"`
	AuthorURL    string `json:"author_url"`
	ProviderName string `json:"provider_name"`
}

type DownloadResult struct {
	VideoID          string  `json:"video_id"`
	Title            string  `json:"title"`
	DownloadType     string  `json:"download_type"`
	FileName         string  `json:"file_name"`
	FilePath         string  `json:"file_path"`
	FileSizeBytes    int64   `json:"file_size_bytes"`
	FileSizeReadable string  `json:"file_size_readable"`
	DurationSeconds  float64 `json:"download_duration_seconds"`
	Message          string  `json:"message"`
}

type TranscriptionSegment struct {
	ID    int     `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

type TranscriptionResponse struct {
	Status                   string                 `json:"status"`
	VideoID                  string                 `json:"video_id"`
	Model                    string                 `json:"model"`
	DetectedLanguage         string                 `json:"detected_language"`
	LanguageProbability      float64                `json:"language_probability"`
	DurationSeconds          float64                `json:"duration_seconds"`
	TranscriptionTimeSeconds float64                `json:"transcription_time_seconds"`
	FullTranscript           string                 `json:"full_transcript"`
	SegmentsCount            int                    `json:"segments_count"`
	Segments                 []TranscriptionSegment `json:"segments"`
}

type ChunkItem struct {
	ChunkID          string  `json:"chunk_id"`
	VideoID          string  `json:"video_id"`
	ChunkIndex       int     `json:"chunk_index"`
	StartTime        float64 `json:"start_time"`
	EndTime          float64 `json:"end_time"`
	TimestampDisplay string  `json:"timestamp_display"`
	Text             string  `json:"text"`
	WordCount        int     `json:"word_count"`
	SegmentIDs       []int   `json:"segment_ids"`
}

type ChunkWithEmbedding struct {
	ChunkID             string    `json:"chunk_id"`
	VideoID             string    `json:"video_id"`
	ChunkIndex          int       `json:"chunk_index"`
	StartTime           float64   `json:"start_time"`
	EndTime             float64   `json:"end_time"`
	TimestampDisplay    string    `json:"timestamp_display"`
	Text                string    `json:"text"`
	WordCount           int       `json:"word_count"`
	SegmentIDs          []int     `json:"segment_ids"`
	EmbeddingDimensions int       `json:"embedding_dimensions"`
	EmbeddingPreview    []float64 `json:"embedding_preview"`
	EmbeddingVector     []float64 `json:"embedding_vector,omitempty"`
	StoredInQdrant      bool      `json:"stored_in_qdrant"`
}

type EmbedChunksResponse struct {
	Status              string               `json:"status"`
	VideoID             string               `json:"video_id"`
	Title               string               `json:"title,omitempty"`
	EmbeddingModel      string               `json:"embedding_model"`
	EmbeddingDimensions int                  `json:"embedding_dimensions"`
	QdrantMode          string               `json:"qdrant_mode"`
	CollectionName      string               `json:"collection_name"`
	QdrantStored        bool                 `json:"qdrant_stored"`
	TotalChunks         int                  `json:"total_chunks"`
	Chunks              []ChunkWithEmbedding `json:"chunks"`
	Message             string               `json:"message"`
}

type ChunkAndEmbedResponse struct {
	Status              string      `json:"status"`
	VideoID             string      `json:"video_id"`
	CollectionName      string      `json:"collection_name"`
	EmbeddingModel      string      `json:"embedding_model"`
	EmbeddingDimensions int         `json:"embedding_dimensions"`
	ChunksCount         int         `json:"chunks_count"`
	Chunks              []ChunkItem `json:"chunks"`
	Message             string      `json:"message"`
}

type SearchMatch struct {
	Score               float64 `json:"score"`
	ChunkID             string  `json:"chunk_id"`
	VideoID             string  `json:"video_id"`
	StartTime           float64 `json:"start_time"`
	EndTime             float64 `json:"end_time"`
	TimestampDisplay    string  `json:"timestamp_display"`
	Text                string  `json:"text"`
	YouTubeTimestampURL string  `json:"youtube_timestamp_url,omitempty"`
}

type SearchResponse struct {
	Status       string        `json:"status"`
	Query        string        `json:"query"`
	MatchesCount int           `json:"matches_count"`
	Matches      []SearchMatch `json:"matches"`
}

type CypherStatement struct {
	Statement  string                 `json:"statement"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

type Neo4jTxRequest struct {
	Statements []CypherStatement `json:"statements"`
}

type Neo4jTxResponse struct {
	Results []struct {
		Columns []string `json:"columns"`
		Data    []struct {
			Row  []interface{} `json:"row"`
			Meta []interface{} `json:"meta"`
		} `json:"data"`
	} `json:"results"`
	Errors []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

type GraphSyncResponse struct {
	Status             string      `json:"status"`
	VideoID            string      `json:"video_id"`
	Title              string      `json:"title"`
	Neo4jHost          string      `json:"neo4j_host"`
	BrowserURL         string      `json:"neo4j_browser_url"`
	VideoNodes         int         `json:"video_nodes"`
	ChunkNodes         int         `json:"chunk_nodes"`
	ConceptNodes       int         `json:"concept_nodes"`
	TotalNodes         int         `json:"total_nodes"`
	RelationshipsCount int         `json:"relationships_count"`
	RelationshipTypes  []string    `json:"relationship_types"`
	Chunks             []ChunkItem `json:"chunks"`
	Message            string      `json:"message"`
}

type GraphNeighbor struct {
	ChunkID          string  `json:"chunk_id"`
	ChunkIndex       int     `json:"chunk_index"`
	StartTime        float64 `json:"start_time"`
	EndTime          float64 `json:"end_time"`
	TimestampDisplay string  `json:"timestamp_display"`
	Text             string  `json:"text"`
}

type ChunkContextResponse struct {
	Status        string         `json:"status"`
	VideoID       string         `json:"video_id"`
	TargetChunkID string         `json:"target_chunk_id"`
	TargetChunk   ChunkItem      `json:"target_chunk"`
	PrevChunk     *GraphNeighbor `json:"prev_chunk,omitempty"`
	NextChunk     *GraphNeighbor `json:"next_chunk,omitempty"`
	Concepts      []string       `json:"concepts"`
	FullContext   string         `json:"full_context"`
}

type ChatRequest struct {
	Query   string `json:"query"`
	VideoID string `json:"video_id"`
}

type ChatResponse struct {
	Status        string                `json:"status"`
	Query         string                `json:"query"`
	VideoID       string                `json:"video_id"`
	Answer        string                `json:"answer"`
	Highlights    []string              `json:"highlights"`
	Concepts      []string              `json:"concepts"`
	Score         float64               `json:"score"`
	Timestamp     string                `json:"timestamp"`
	YouTubeURL    string                `json:"youtube_url"`
	SourceChunkID string                `json:"source_chunk_id"`
	GraphContext  *ChunkContextResponse `json:"graph_context,omitempty"`
	OllamaModel   string                `json:"ollama_model,omitempty"`
	RAGSource     string                `json:"rag_source,omitempty"`
}

var (
	videos        = make(map[string]Video)
	videosMu      sync.RWMutex
	transcripts   = make(map[string]TranscriptionResponse)
	transcriptsMu sync.RWMutex
	videoChunks   = make(map[string][]ChunkItem)
	videoChunksMu sync.RWMutex
)

// formatBytes returns a human-readable file size string.
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// corsMiddleware allows frontend clients and Postman to communicate without restrictions.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// extractVideoInput flexibly reads URL and Title from JSON body, form data, or query params.
func extractVideoInput(c *gin.Context) (rawURL string, title string) {
	var body struct {
		URL      string `json:"url"`
		URLAlt   string `json:"URL"`
		VideoURL string `json:"video_url"`
		VideoUrl string `json:"videoUrl"`
		Link     string `json:"link"`
		Title    string `json:"title"`
	}

	if c.Request.Body != nil {
		if err := c.ShouldBindJSON(&body); err == nil {
			rawURL = body.URL
			if rawURL == "" {
				rawURL = body.URLAlt
			}
			if rawURL == "" {
				rawURL = body.VideoURL
			}
			if rawURL == "" {
				rawURL = body.VideoUrl
			}
			if rawURL == "" {
				rawURL = body.Link
			}
			title = body.Title
		}
	}

	if rawURL == "" {
		rawURL = c.PostForm("url")
		if rawURL == "" {
			rawURL = c.PostForm("video_url")
		}
		if rawURL == "" {
			rawURL = c.PostForm("videoUrl")
		}
		if rawURL == "" {
			rawURL = c.PostForm("link")
		}
	}
	if title == "" {
		title = c.PostForm("title")
	}

	if rawURL == "" {
		rawURL = c.Query("url")
		if rawURL == "" {
			rawURL = c.Query("video_url")
		}
		if rawURL == "" {
			rawURL = c.Query("videoUrl")
		}
		if rawURL == "" {
			rawURL = c.Query("link")
		}
	}
	if title == "" {
		title = c.Query("title")
	}

	return strings.TrimSpace(rawURL), strings.TrimSpace(title)
}

// validateYouTubeURL parses and validates standard, short, embed, shorts, and live YouTube URLs.
func validateYouTubeURL(rawURL string) (string, string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", "", fmt.Errorf("URL cannot be empty; please provide a YouTube video URL")
	}

	if len(rawURL) == 11 && !strings.Contains(rawURL, ".") && !strings.Contains(rawURL, "/") {
		return rawURL, "YouTube", nil
	}

	if !strings.HasPrefix(strings.ToLower(rawURL), "http://") && !strings.HasPrefix(strings.ToLower(rawURL), "https://") {
		rawURL = "https://" + rawURL
	}

	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil || parsedURL == nil || parsedURL.Hostname() == "" {
		return "", "", fmt.Errorf("invalid URL; provide a valid HTTP or HTTPS YouTube URL")
	}

	host := strings.ToLower(parsedURL.Hostname())

	isYouTube := false
	domains := []string{"youtube.com", "www.youtube.com", "m.youtube.com", "youtu.be", "youtube-nocookie.com", "www.youtube-nocookie.com"}
	for _, d := range domains {
		if host == d || strings.HasSuffix(host, "."+d) {
			isYouTube = true
			break
		}
	}

	if !isYouTube {
		return "", "", fmt.Errorf("unsupported platform '%s'; only YouTube URLs are supported", host)
	}

	videoID := ""

	if host == "youtu.be" {
		pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
		if len(pathParts) > 0 {
			videoID = pathParts[0]
		}
	} else {
		videoID = parsedURL.Query().Get("v")
		if videoID == "" {
			parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
			if len(parts) >= 2 {
				action := strings.ToLower(parts[0])
				if action == "shorts" || action == "embed" || action == "live" || action == "v" {
					videoID = parts[1]
				}
			}
		}
	}

	videoID = strings.Trim(videoID, "/")
	if idx := strings.IndexAny(videoID, "?&#"); idx != -1 {
		videoID = videoID[:idx]
	}

	if len(videoID) != 11 {
		return "", "", fmt.Errorf("invalid or missing YouTube video ID; expected an 11-character ID (got '%s')", videoID)
	}

	return videoID, "YouTube", nil
}

// fetchYouTubeMetadata retrieves metadata using yt-dlp if available, or falls back to YouTube's public oEmbed API.
func fetchYouTubeMetadata(rawURL, videoID string, fallbackTitle string) (YouTubeMetadata, error) {
	if _, err := exec.LookPath("yt-dlp"); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()

		cmd := exec.CommandContext(
			ctx,
			"yt-dlp",
			"--no-warnings",
			"--no-update",
			"--no-playlist",
			"--dump-single-json",
			"--skip-download",
			rawURL,
		)

		output, err := cmd.Output()
		if err == nil {
			var metadata YouTubeMetadata
			if jsonErr := json.Unmarshal(output, &metadata); jsonErr == nil && metadata.ID != "" && metadata.Title != "" {
				return metadata, nil
			}
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	oEmbedURL := fmt.Sprintf("https://www.youtube.com/oembed?url=https://www.youtube.com/watch?v=%s&format=json", videoID)
	resp, err := client.Get(oEmbedURL)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		var oEmbed YouTubeOEmbed
		if err := json.NewDecoder(resp.Body).Decode(&oEmbed); err == nil && oEmbed.Title != "" {
			return YouTubeMetadata{
				ID:         videoID,
				Title:      oEmbed.Title,
				Uploader:   oEmbed.AuthorName,
				Duration:   0,
				WebpageURL: fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
			}, nil
		}
	}

	title := fallbackTitle
	if title == "" {
		title = fmt.Sprintf("YouTube Video (%s)", videoID)
	}

	return YouTubeMetadata{
		ID:         videoID,
		Title:      title,
		Uploader:   "YouTube Creator",
		Duration:   0,
		WebpageURL: fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
	}, nil
}

// getOrCreateVideo looks up a video in memory or automatically creates/registers it on the fly.
func getOrCreateVideo(input string) (Video, error) {
	input = strings.TrimSpace(input)
	videoID, _, err := validateYouTubeURL(input)
	if err != nil {
		if len(input) == 11 {
			videoID = input
		} else {
			return Video{}, fmt.Errorf("invalid video ID or URL: %s", input)
		}
	}

	videosMu.RLock()
	video, exists := videos[videoID]
	videosMu.RUnlock()

	if exists {
		return video, nil
	}

	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	metadata, _ := fetchYouTubeMetadata(videoURL, videoID, "")

	title := metadata.Title
	if title == "" {
		title = fmt.Sprintf("YouTube Video (%s)", videoID)
	}

	video = Video{
		ID:        videoID,
		URL:       videoURL,
		Title:     title,
		Uploader:  metadata.Uploader,
		Duration:  int(metadata.Duration),
		Platform:  "YouTube",
		Status:    "ready",
		CreatedAt: time.Now().UTC(),
	}

	videosMu.Lock()
	videos[video.ID] = video
	videosMu.Unlock()

	return video, nil
}

// fetchYouTubeCaptions extracts captions using yt-dlp.
func fetchYouTubeCaptions(rawURL string) (string, error) {
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return "", fmt.Errorf("yt-dlp is not installed in the current environment. To extract captions, run the backend via Docker or install yt-dlp on your system")
	}

	tempDir, err := os.MkdirTemp("", "video-rag-captions-*")
	if err != nil {
		return "", fmt.Errorf("could not create temporary directory")
	}
	defer os.RemoveAll(tempDir)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	outputTemplate := filepath.Join(tempDir, "captions.%(ext)s")

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"--no-warnings",
		"--no-update",
		"--no-playlist",
		"--skip-download",
		"--write-subs",
		"--write-auto-subs",
		"--sub-langs", "en",
		"--sub-format", "vtt",
		"--output", outputTemplate,
		"--ignore-errors",
		rawURL,
	)

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("caption retrieval timed out")
	}
	if err != nil {
		return "", fmt.Errorf("could not retrieve captions: %s", strings.TrimSpace(string(output)))
	}

	files, err := os.ReadDir(tempDir)
	if err != nil {
		return "", fmt.Errorf("could not read caption directory")
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".vtt") {
			continue
		}

		captionBytes, err := os.ReadFile(filepath.Join(tempDir, file.Name()))
		if err != nil {
			return "", fmt.Errorf("could not read caption content")
		}

		captions := strings.TrimSpace(string(captionBytes))
		if captions != "" {
			return captions, nil
		}
	}

	return "", fmt.Errorf("no English captions were found for this video")
}

// downloadMedia downloads audio or video using yt-dlp.
func downloadMedia(rawURL, videoID, title, mediaType string) (*DownloadResult, error) {
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return nil, fmt.Errorf("yt-dlp is not installed in the current environment. Please ensure the backend is running inside Docker with yt-dlp and ffmpeg installed")
	}

	downloadsDir := "downloads"
	if err := os.MkdirAll(downloadsDir, 0755); err != nil {
		return nil, fmt.Errorf("could not create downloads directory: %w", err)
	}

	outputTemplate := filepath.Join(downloadsDir, fmt.Sprintf("%s.%%(ext)s", videoID))

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	var args []string
	if mediaType == "audio" {
		args = []string{
			"--no-warnings",
			"--no-playlist",
			"-x",
			"--audio-format", "mp3",
			"--output", outputTemplate,
			rawURL,
		}
	} else {
		args = []string{
			"--no-warnings",
			"--no-playlist",
			"-f", "worst[ext=mp4]/best[height<=480]/best",
			"--output", outputTemplate,
			rawURL,
		}
	}

	startTime := time.Now()
	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("media download timed out after 3 minutes")
	}
	if err != nil {
		return nil, fmt.Errorf("yt-dlp download failed: %s", strings.TrimSpace(string(output)))
	}

	duration := time.Since(startTime).Seconds()

	files, err := os.ReadDir(downloadsDir)
	if err != nil {
		return nil, fmt.Errorf("could not read downloads directory: %w", err)
	}

	var foundName string
	var foundSize int64
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if strings.HasPrefix(f.Name(), videoID+".") {
			ext := strings.ToLower(filepath.Ext(f.Name()))
			if mediaType == "audio" && (ext == ".mp3" || ext == ".m4a" || ext == ".opus" || ext == ".wav") {
				foundName = f.Name()
				if info, err := f.Info(); err == nil {
					foundSize = info.Size()
				}
				break
			}
			if mediaType == "video" && (ext == ".mp4" || ext == ".webm" || ext == ".mkv") {
				foundName = f.Name()
				if info, err := f.Info(); err == nil {
					foundSize = info.Size()
				}
				break
			}
		}
	}

	if foundName == "" {
		return nil, fmt.Errorf("download completed, but the %s file could not be found on disk", mediaType)
	}

	fullPath, _ := filepath.Abs(filepath.Join(downloadsDir, foundName))

	return &DownloadResult{
		VideoID:          videoID,
		Title:            title,
		DownloadType:     mediaType,
		FileName:         foundName,
		FilePath:         fullPath,
		FileSizeBytes:    foundSize,
		FileSizeReadable: formatBytes(foundSize),
		DurationSeconds:  duration,
		Message:          fmt.Sprintf("%s successfully downloaded via yt-dlp in %.2f seconds (%s).", strings.ToUpper(mediaType[:1])+mediaType[1:], duration, formatBytes(foundSize)),
	}, nil
}

// ensureAudioFile checks if audio is already downloaded, or downloads it automatically.
func ensureAudioFile(video Video) (string, error) {
	downloadsDir := "downloads"
	files, err := os.ReadDir(downloadsDir)
	if err == nil {
		for _, f := range files {
			if !f.IsDir() && strings.HasPrefix(f.Name(), video.ID+".") {
				ext := strings.ToLower(filepath.Ext(f.Name()))
				if ext == ".mp3" || ext == ".m4a" || ext == ".opus" || ext == ".wav" {
					return f.Name(), nil
				}
			}
		}
	}

	dlResult, err := downloadMedia(video.URL, video.ID, video.Title, "audio")
	if err != nil {
		return "", fmt.Errorf("failed to download audio for transcription: %w", err)
	}
	return dlResult.FileName, nil
}

// requestWhisperTranscription calls the Python Faster-Whisper microservice over HTTP.
func requestWhisperTranscription(videoID, audioFileName string) (*TranscriptionResponse, error) {
	whisperURL := os.Getenv("WHISPER_SERVICE_URL")
	if whisperURL == "" {
		whisperURL = "http://whisper-service:5000"
	}

	payload := map[string]string{
		"video_id":   videoID,
		"audio_path": filepath.Join("/app/downloads", audioFileName),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("could not serialize transcription request: %w", err)
	}

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Post(whisperURL+"/transcribe", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("could not connect to faster-whisper service at %s: %w. Ensure whisper-service container is running", whisperURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Detail string `json:"detail"`
		}
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil && errResp.Detail != "" {
			return nil, fmt.Errorf("faster-whisper service error: %s", errResp.Detail)
		}
		return nil, fmt.Errorf("faster-whisper service returned HTTP %d", resp.StatusCode)
	}

	var result TranscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("could not parse transcription response: %w", err)
	}

	return &result, nil
}

// requestQdrantPoints queries the Python microservice for points stored in Qdrant.
func requestQdrantPoints(videoID string, limit int) (map[string]interface{}, error) {
	whisperURL := os.Getenv("WHISPER_SERVICE_URL")
	if whisperURL == "" {
		whisperURL = "http://whisper-service:5000"
	}

	qURL := fmt.Sprintf("%s/qdrant-points?limit=%d", whisperURL, limit)
	if videoID != "" {
		qURL += fmt.Sprintf("&video_id=%s", url.QueryEscape(videoID))
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(qURL)
	if err != nil {
		return nil, fmt.Errorf("could not connect to qdrant proxy at %s: %w", qURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("qdrant points request returned HTTP %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("could not decode qdrant points: %w", err)
	}
	return result, nil
}

// Ollama API integration structures
type OllamaGenerateRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type OllamaGenerateResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}

type OllamaTagResponse struct {
	Models []struct {
		Name  string `json:"name"`
		Model string `json:"model"`
	} `json:"models"`
}

func getOllamaURL() string {
	u := os.Getenv("OLLAMA_URL")
	if u == "" {
		u = "http://host.docker.internal:11434"
	}
	return strings.TrimRight(u, "/")
}

func getOllamaModel() string {
	m := os.Getenv("OLLAMA_MODEL")
	if m != "" {
		return m
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(getOllamaURL() + "/api/tags")
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		var tagResp OllamaTagResponse
		if json.NewDecoder(resp.Body).Decode(&tagResp) == nil && len(tagResp.Models) > 0 {
			for _, mItem := range tagResp.Models {
				if strings.Contains(mItem.Name, "llama3.2") {
					return mItem.Name
				}
			}
			return tagResp.Models[0].Name
		}
	}
	return "llama3.2"
}

func cleanExcerpt(t string, maxLen int) string {
	t = strings.TrimSpace(strings.ReplaceAll(t, "\n", " "))
	if len(t) > maxLen {
		return t[:maxLen] + "..."
	}
	return t
}

func queryOllama(prompt string, model string, timeout time.Duration) (string, error) {
	if model == "" {
		model = getOllamaModel()
	}
	ollamaReq := OllamaGenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": 0.2,
		},
	}
	b, err := json.Marshal(ollamaReq)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Post(getOllamaURL()+"/api/generate", "application/json", bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("ollama connection failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error HTTP %d: %s", resp.StatusCode, string(body))
	}
	var genResp OllamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return "", fmt.Errorf("failed to decode ollama response: %w", err)
	}
	return strings.TrimSpace(genResp.Response), nil
}

// fetchRepresentativeChunks retrieves chronological excerpts spread across the video for generating comprehensive summaries.
func fetchRepresentativeChunks(videoID string) ([]string, error) {
	// 1. In-memory check
	videoChunksMu.RLock()
	cachedChunks, has := videoChunks[videoID]
	videoChunksMu.RUnlock()

	if has && len(cachedChunks) > 0 {
		var selected []string
		n := len(cachedChunks)
		if n <= 5 {
			for _, c := range cachedChunks {
				selected = append(selected, fmt.Sprintf("[%s]: %s", c.TimestampDisplay, cleanExcerpt(c.Text, 350)))
			}
		} else {
			indices := []int{0, n / 4, n / 2, (3 * n) / 4, n - 1}
			seen := make(map[int]bool)
			for _, idx := range indices {
				if idx >= 0 && idx < n && !seen[idx] {
					seen[idx] = true
					c := cachedChunks[idx]
					selected = append(selected, fmt.Sprintf("[%s]: %s", c.TimestampDisplay, cleanExcerpt(c.Text, 350)))
				}
			}
		}
		return selected, nil
	}

	// 2. Fetch from Qdrant
	pointsResp, err := requestQdrantPoints(videoID, 150)
	if err != nil {
		return nil, err
	}
	rawPoints, ok := pointsResp["points"].([]interface{})
	if !ok || len(rawPoints) == 0 {
		return nil, fmt.Errorf("no chunks found in Qdrant for video %s", videoID)
	}

	type chunkData struct {
		index int
		time  string
		text  string
	}
	var list []chunkData
	for _, rp := range rawPoints {
		pMap, ok := rp.(map[string]interface{})
		if !ok {
			continue
		}
		payload, ok := pMap["payload"].(map[string]interface{})
		if !ok {
			continue
		}
		idx := 0
		if f, ok := payload["chunk_index"].(float64); ok {
			idx = int(f)
		}
		timeStr, _ := payload["timestamp_display"].(string)
		txt, _ := payload["text"].(string)
		if txt != "" {
			list = append(list, chunkData{index: idx, time: timeStr, text: txt})
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].index < list[j].index
	})

	var selected []string
	n := len(list)
	if n <= 5 {
		for _, c := range list {
			selected = append(selected, fmt.Sprintf("[%s]: %s", c.time, cleanExcerpt(c.text, 350)))
		}
	} else {
		indices := []int{0, n / 4, n / 2, (3 * n) / 4, n - 1}
		seen := make(map[int]bool)
		for _, idx := range indices {
			if idx >= 0 && idx < n && !seen[idx] {
				seen[idx] = true
				selected = append(selected, fmt.Sprintf("[%s]: %s", list[idx].time, cleanExcerpt(list[idx].text, 350)))
			}
		}
	}
	return selected, nil
}

// executeCypher runs Cypher queries against Neo4j transactional HTTP API.
func executeCypher(statements []CypherStatement) (*Neo4jTxResponse, error) {
	neo4jURL := os.Getenv("NEO4J_HTTP_URL")
	if neo4jURL == "" {
		neo4jURL = "http://neo4j:7474"
	}
	user := os.Getenv("NEO4J_USER")
	if user == "" {
		user = "neo4j"
	}
	pass := os.Getenv("NEO4J_PASSWORD")
	if pass == "" {
		pass = "videorag123"
	}

	endpoint := strings.TrimRight(neo4jURL, "/") + "/db/neo4j/tx/commit"
	reqBytes, err := json.Marshal(Neo4jTxRequest{Statements: statements})
	if err != nil {
		return nil, fmt.Errorf("failed to encode cypher statements: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(user, pass)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not connect to neo4j at %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	var txResp Neo4jTxResponse
	if err := json.NewDecoder(resp.Body).Decode(&txResp); err != nil {
		return nil, fmt.Errorf("could not parse neo4j response: %w", err)
	}

	if len(txResp.Errors) > 0 {
		return nil, fmt.Errorf("neo4j cypher error [%s]: %s", txResp.Errors[0].Code, txResp.Errors[0].Message)
	}

	return &txResp, nil
}

var domainKeywords = []string{
	"celery", "worker", "broker", "redis", "rabbitmq", "task", "queue",
	"concurrency", "async", "gevent", "thread", "process", "io", "cpu",
	"report", "background", "scale", "latency", "message", "event",
	"app", "server", "cluster", "performance", "database", "load",
}

func extractKeyConcepts(text string) []string {
	lower := strings.ToLower(text)
	var found []string
	seen := make(map[string]bool)
	for _, kw := range domainKeywords {
		if strings.Contains(lower, kw) && !seen[kw] {
			seen[kw] = true
			found = append(found, strings.ToUpper(kw[:1])+kw[1:])
		}
	}
	return found
}

// syncVideoToNeo4j syncs video, chunks, temporal NEXT/PREV links, and concept nodes into Neo4j.
func syncVideoToNeo4j(video Video, chunks []ChunkItem) (*GraphSyncResponse, error) {
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks to sync to neo4j")
	}

	var statements []CypherStatement

	// 1. Merge Video node
	statements = append(statements, CypherStatement{
		Statement: `
			MERGE (v:Video {id: $id})
			SET v.title = $title,
				v.url = $url,
				v.uploader = $uploader,
				v.duration_seconds = $duration,
				v.platform = $platform
		`,
		Parameters: map[string]interface{}{
			"id":       video.ID,
			"title":    video.Title,
			"url":      video.URL,
			"uploader": video.Uploader,
			"duration": video.Duration,
			"platform": video.Platform,
		},
	})

	// 2. Merge Chunks and connect to Video via HAS_CHUNK
	conceptSet := make(map[string]bool)
	for _, c := range chunks {
		statements = append(statements, CypherStatement{
			Statement: `
				MERGE (ch:Chunk {chunk_id: $chunk_id})
				SET ch.video_id = $video_id,
					ch.chunk_index = $chunk_index,
					ch.start_time = $start_time,
					ch.end_time = $end_time,
					ch.timestamp_display = $timestamp_display,
					ch.text = $text,
					ch.word_count = $word_count
				WITH ch
				MATCH (v:Video {id: $video_id})
				MERGE (v)-[r:HAS_CHUNK {index: $chunk_index}]->(ch)
			`,
			Parameters: map[string]interface{}{
				"chunk_id":          c.ChunkID,
				"video_id":          c.VideoID,
				"chunk_index":       c.ChunkIndex,
				"start_time":        c.StartTime,
				"end_time":          c.EndTime,
				"timestamp_display": c.TimestampDisplay,
				"text":              c.Text,
				"word_count":        c.WordCount,
			},
		})

		// Extract concepts and link via DISCUSSES
		concepts := extractKeyConcepts(c.Text)
		for _, cp := range concepts {
			conceptSet[cp] = true
			statements = append(statements, CypherStatement{
				Statement: `
					MERGE (concept:Concept {name: $concept_name})
					WITH concept
					MATCH (ch:Chunk {chunk_id: $chunk_id})
					MERGE (ch)-[:DISCUSSES]->(concept)
				`,
				Parameters: map[string]interface{}{
					"concept_name": cp,
					"chunk_id":     c.ChunkID,
				},
			})
		}
	}

	// 3. Connect temporal sequential chain: NEXT_CHUNK & PREV_CHUNK
	for i := 0; i < len(chunks)-1; i++ {
		c1 := chunks[i]
		c2 := chunks[i+1]
		statements = append(statements, CypherStatement{
			Statement: `
				MATCH (c1:Chunk {chunk_id: $c1_id}), (c2:Chunk {chunk_id: $c2_id})
				MERGE (c1)-[:NEXT_CHUNK {video_id: $video_id}]->(c2)
				MERGE (c2)-[:PREV_CHUNK {video_id: $video_id}]->(c1)
			`,
			Parameters: map[string]interface{}{
				"c1_id":    c1.ChunkID,
				"c2_id":    c2.ChunkID,
				"video_id": video.ID,
			},
		})
	}

	// Execute transaction
	_, err := executeCypher(statements)
	if err != nil {
		return nil, fmt.Errorf("failed to sync graph to neo4j: %w", err)
	}

	// 4. Query relationships count
	countStmt := []CypherStatement{
		{
			Statement: `
				MATCH (v:Video {id: $video_id})-[:HAS_CHUNK]->(c:Chunk)
				OPTIONAL MATCH (c)-[r]-()
				RETURN count(DISTINCT c) AS chunk_count, count(DISTINCT r) AS rel_count
			`,
			Parameters: map[string]interface{}{
				"video_id": video.ID,
			},
		},
	}
	countResp, err := executeCypher(countStmt)
	relCount := 0
	if err == nil && len(countResp.Results) > 0 && len(countResp.Results[0].Data) > 0 {
		row := countResp.Results[0].Data[0].Row
		if len(row) > 1 {
			if rc, ok := row[1].(float64); ok {
				relCount = int(rc)
			}
		}
	}

	return &GraphSyncResponse{
		Status:             "success",
		VideoID:            video.ID,
		Title:              video.Title,
		Neo4jHost:          "http://localhost:7474",
		BrowserURL:         "http://localhost:7474/browser/",
		VideoNodes:         1,
		ChunkNodes:         len(chunks),
		ConceptNodes:       len(conceptSet),
		TotalNodes:         1 + len(chunks) + len(conceptSet),
		RelationshipsCount: relCount,
		RelationshipTypes:  []string{"HAS_CHUNK", "NEXT_CHUNK", "PREV_CHUNK", "DISCUSSES"},
		Chunks:             chunks,
		Message:            fmt.Sprintf("Synchronized %d chunks, %d concepts, and %d relationships into Neo4j graph.", len(chunks), len(conceptSet), relCount),
	}, nil
}

// fetchVideoGraph retrieves graph nodes and relations for a video from Neo4j.
func fetchVideoGraph(videoID string) (map[string]interface{}, error) {
	stmt := []CypherStatement{
		{
			Statement: `
				MATCH (v:Video {id: $video_id})
				OPTIONAL MATCH (v)-[r1:HAS_CHUNK]->(c:Chunk)
				OPTIONAL MATCH (c)-[r2:NEXT_CHUNK]->(next:Chunk)
				OPTIONAL MATCH (c)-[r3:DISCUSSES]->(cp:Concept)
				RETURN v.title AS title, 
				       collect(DISTINCT {id: c.chunk_id, index: c.chunk_index, time: c.timestamp_display, text: c.text}) AS chunks,
				       collect(DISTINCT {from: c.chunk_id, to: next.chunk_id, type: 'NEXT_CHUNK'}) AS sequences,
				       collect(DISTINCT {chunk: c.chunk_id, concept: cp.name}) AS concept_links,
				       collect(DISTINCT cp.name) AS concepts
			`,
			Parameters: map[string]interface{}{
				"video_id": videoID,
			},
		},
	}

	resp, err := executeCypher(stmt)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"status":            "success",
		"video_id":          videoID,
		"neo4j_browser_url": "http://localhost:7474/browser/",
	}

	if len(resp.Results) > 0 && len(resp.Results[0].Data) > 0 {
		row := resp.Results[0].Data[0].Row
		if len(row) >= 5 {
			result["title"] = row[0]
			result["chunks"] = row[1]
			result["temporal_sequences"] = row[2]
			result["concept_links"] = row[3]
			result["concepts"] = row[4]
		}
	}

	return result, nil
}

// fetchChunkContext retrieves target chunk with PREV_CHUNK and NEXT_CHUNK context from Neo4j.
func fetchChunkContext(videoID, chunkID string) (*ChunkContextResponse, error) {
	stmt := []CypherStatement{
		{
			Statement: `
				MATCH (target:Chunk {chunk_id: $chunk_id})
				OPTIONAL MATCH (prev:Chunk)-[:NEXT_CHUNK]->(target)
				OPTIONAL MATCH (target)-[:NEXT_CHUNK]->(next:Chunk)
				OPTIONAL MATCH (target)-[:DISCUSSES]->(cp:Concept)
				RETURN target, prev, next, collect(DISTINCT cp.name) AS concepts
			`,
			Parameters: map[string]interface{}{
				"chunk_id": chunkID,
			},
		},
	}

	resp, err := executeCypher(stmt)
	if err != nil {
		return nil, err
	}

	if len(resp.Results) == 0 || len(resp.Results[0].Data) == 0 {
		return nil, fmt.Errorf("chunk '%s' not found in Neo4j graph", chunkID)
	}

	row := resp.Results[0].Data[0].Row
	var target ChunkItem
	var prev *GraphNeighbor
	var next *GraphNeighbor
	var concepts []string

	if len(row) > 0 && row[0] != nil {
		if m, ok := row[0].(map[string]interface{}); ok {
			target.ChunkID, _ = m["chunk_id"].(string)
			if idx, ok := m["chunk_index"].(float64); ok {
				target.ChunkIndex = int(idx)
			}
			target.StartTime, _ = m["start_time"].(float64)
			target.EndTime, _ = m["end_time"].(float64)
			target.TimestampDisplay, _ = m["timestamp_display"].(string)
			target.Text, _ = m["text"].(string)
			if wc, ok := m["word_count"].(float64); ok {
				target.WordCount = int(wc)
			}
			target.VideoID, _ = m["video_id"].(string)
		}
	}

	if len(row) > 1 && row[1] != nil {
		if m, ok := row[1].(map[string]interface{}); ok {
			prev = &GraphNeighbor{
				ChunkID:          m["chunk_id"].(string),
				TimestampDisplay: m["timestamp_display"].(string),
				Text:             m["text"].(string),
			}
			if idx, ok := m["chunk_index"].(float64); ok {
				prev.ChunkIndex = int(idx)
			}
		}
	}

	if len(row) > 2 && row[2] != nil {
		if m, ok := row[2].(map[string]interface{}); ok {
			next = &GraphNeighbor{
				ChunkID:          m["chunk_id"].(string),
				TimestampDisplay: m["timestamp_display"].(string),
				Text:             m["text"].(string),
			}
			if idx, ok := m["chunk_index"].(float64); ok {
				next.ChunkIndex = int(idx)
			}
		}
	}

	if len(row) > 3 && row[3] != nil {
		if arr, ok := row[3].([]interface{}); ok {
			for _, item := range arr {
				if s, ok := item.(string); ok {
					concepts = append(concepts, s)
				}
			}
		}
	}

	var sb strings.Builder
	if prev != nil {
		sb.WriteString(fmt.Sprintf("[PREVIOUS CONTEXT (%s)]: %s\n\n", prev.TimestampDisplay, prev.Text))
	}
	sb.WriteString(fmt.Sprintf("[CURRENT CHUNK (%s)]: %s\n\n", target.TimestampDisplay, target.Text))
	if next != nil {
		sb.WriteString(fmt.Sprintf("[NEXT CONTEXT (%s)]: %s\n", next.TimestampDisplay, next.Text))
	}

	return &ChunkContextResponse{
		Status:        "success",
		VideoID:       videoID,
		TargetChunkID: chunkID,
		TargetChunk:   target,
		PrevChunk:     prev,
		NextChunk:     next,
		Concepts:      concepts,
		FullContext:   sb.String(),
	}, nil
}

// renderEmbeddingsHTML produces an interactive, browser-friendly dashboard for stored Qdrant embeddings.
func renderEmbeddingsHTML(res *EmbedChunksResponse) string {
	return renderProductionUI(res)
}

// extractHighlights splits transcription text into concise, high-value bullet points.
func extractHighlights(text string) []string {
	var highlights []string
	clean := strings.ReplaceAll(text, "\n", " ")
	sentences := strings.Split(clean, ". ")
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if len(s) > 20 {
			if !strings.HasSuffix(s, ".") {
				s += "."
			}
			highlights = append(highlights, s)
		}
		if len(highlights) >= 3 {
			break
		}
	}
	if len(highlights) == 0 && text != "" {
		highlights = append(highlights, text)
	}
	return highlights
}
func requestChunkAndEmbed(videoID string, segments []TranscriptionSegment, targetWords int, overlapSegs int) (*ChunkAndEmbedResponse, error) {
	whisperURL := os.Getenv("WHISPER_SERVICE_URL")
	if whisperURL == "" {
		whisperURL = "http://whisper-service:5000"
	}

	payload := map[string]interface{}{
		"video_id":         videoID,
		"segments":         segments,
		"target_words":     targetWords,
		"overlap_segments": overlapSegs,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("could not serialize chunking request: %w", err)
	}

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Post(whisperURL+"/chunk-and-embed", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("could not connect to embedding service at %s: %w", whisperURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Detail string `json:"detail"`
		}
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil && errResp.Detail != "" {
			return nil, fmt.Errorf("embedding service error: %s", errResp.Detail)
		}
		return nil, fmt.Errorf("embedding service returned HTTP %d", resp.StatusCode)
	}

	var result ChunkAndEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("could not parse chunking response: %w", err)
	}

	return &result, nil
}

// requestEmbedChunks calls the Python service to create chunks, 384-dim embeddings, and store in Qdrant.
func requestEmbedChunks(videoID string, segments []TranscriptionSegment, targetWords int, overlapSegs int, storeInQdrant bool, includeFull bool) (*EmbedChunksResponse, error) {
	whisperURL := os.Getenv("WHISPER_SERVICE_URL")
	if whisperURL == "" {
		whisperURL = "http://whisper-service:5000"
	}

	payload := map[string]interface{}{
		"video_id":            videoID,
		"segments":            segments,
		"target_words":        targetWords,
		"overlap_segments":    overlapSegs,
		"store_in_qdrant":     storeInQdrant,
		"include_full_vector": includeFull,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("could not serialize request: %w", err)
	}

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Post(whisperURL+"/embed-chunks", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("could not connect to embedding service at %s: %w", whisperURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Detail string `json:"detail"`
		}
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil && errResp.Detail != "" {
			return nil, fmt.Errorf("embedding service error: %s", errResp.Detail)
		}
		return nil, fmt.Errorf("embedding service returned HTTP %d", resp.StatusCode)
	}

	var result EmbedChunksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("could not parse embedding response: %w", err)
	}

	return &result, nil
}

// requestQdrantSearch calls the Python service to embed query and search Qdrant.
func requestQdrantSearch(query string, videoID string, limit int) (*SearchResponse, error) {
	whisperURL := os.Getenv("WHISPER_SERVICE_URL")
	if whisperURL == "" {
		whisperURL = "http://whisper-service:5000"
	}

	payload := map[string]interface{}{
		"query":    query,
		"video_id": videoID,
		"limit":    limit,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("could not serialize search request: %w", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Post(whisperURL+"/search", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("could not connect to search service at %s: %w", whisperURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Detail string `json:"detail"`
		}
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil && errResp.Detail != "" {
			return nil, fmt.Errorf("search service error: %s", errResp.Detail)
		}
		return nil, fmt.Errorf("search service returned HTTP %d", resp.StatusCode)
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("could not parse search response: %w", err)
	}

	return &result, nil
}

func main() {
	router := gin.Default()
	router.Use(corsMiddleware())

	// 1. Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Video RAG API is running!",
		})
	})

	// 2. Validate URL format (Supports BOTH POST and GET)
	handleValidateURL := func(c *gin.Context) {
		rawURL, _ := extractVideoInput(c)

		if rawURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"valid":   false,
				"message": "Please provide a video URL via JSON body ({\"url\": \"...\"}), form data, or query param (?url=...).",
			})
			return
		}

		videoID, platform, err := validateYouTubeURL(rawURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"valid":   false,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"valid":    true,
			"platform": platform,
			"video_id": videoID,
			"message":  "URL format is valid.",
		})
	}

	router.POST("/api/v1/videos/validate-url", handleValidateURL)
	router.GET("/api/v1/videos/validate-url", handleValidateURL)
	router.POST("/videos/validate-url", handleValidateURL)
	router.GET("/videos/validate-url", handleValidateURL)

	// 3. Register video and retrieve metadata
	handleCreateVideo := func(c *gin.Context) {
		rawURL, userTitle := extractVideoInput(c)

		if rawURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Please provide a video URL via JSON body, form data, or query parameter.",
			})
			return
		}

		videoID, platform, err := validateYouTubeURL(rawURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		metadata, _ := fetchYouTubeMetadata(rawURL, videoID, userTitle)

		canonicalURL := metadata.WebpageURL
		if canonicalURL == "" {
			canonicalURL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
		}

		title := metadata.Title
		if title == "" {
			if userTitle != "" {
				title = userTitle
			} else {
				title = fmt.Sprintf("YouTube Video (%s)", videoID)
			}
		}

		video := Video{
			ID:        videoID,
			URL:       canonicalURL,
			Title:     title,
			Uploader:  metadata.Uploader,
			Duration:  int(metadata.Duration),
			Platform:  platform,
			Status:    "ready",
			CreatedAt: time.Now().UTC(),
		}

		videosMu.Lock()
		videos[video.ID] = video
		videosMu.Unlock()

		c.JSON(http.StatusCreated, gin.H{
			"status":  "success",
			"message": "Video registered successfully.",
			"video":   video,
		})
	}

	router.POST("/api/v1/videos", handleCreateVideo)
	router.POST("/api/v1/video", handleCreateVideo)
	router.POST("/videos", handleCreateVideo)
	router.POST("/video", handleCreateVideo)

	// 4. List registered videos
	handleListVideos := func(c *gin.Context) {
		videosMu.RLock()
		result := make([]Video, 0, len(videos))
		for _, video := range videos {
			result = append(result, video)
		}
		videosMu.RUnlock()

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"count":  len(result),
			"videos": result,
		})
	}

	router.GET("/api/v1/videos", handleListVideos)
	router.GET("/api/v1/video", handleListVideos)
	router.GET("/videos", handleListVideos)
	router.GET("/video", handleListVideos)

	// 5. Get a single registered video by ID (Auto-registers on the fly if not yet registered)
	handleGetVideo := func(c *gin.Context) {
		videoID := c.Param("id")

		video, err := getOrCreateVideo(videoID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"video":  video,
		})
	}

	router.GET("/api/v1/videos/:id", handleGetVideo)
	router.GET("/api/v1/video/:id", handleGetVideo)
	router.GET("/videos/:id", handleGetVideo)
	router.GET("/video/:id", handleGetVideo)

	// 6. Retrieve available captions (Supports BOTH POST and GET; auto-registers if needed)
	handleCaptions := func(c *gin.Context) {
		videoID := c.Param("id")

		video, err := getOrCreateVideo(videoID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		captions, err := fetchYouTubeCaptions(video.URL)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "success",
			"video_id": video.ID,
			"title":    video.Title,
			"captions": captions,
			"message":  "Captions retrieved successfully.",
		})
	}

	router.POST("/api/v1/videos/:id/captions", handleCaptions)
	router.GET("/api/v1/videos/:id/captions", handleCaptions)
	router.POST("/videos/:id/captions", handleCaptions)
	router.GET("/videos/:id/captions", handleCaptions)

	// 7. Download Audio via yt-dlp (Supports direct browser attachment download and API JSON)
	handleDownloadAudio := func(c *gin.Context) {
		target := c.Param("id")
		if target == "" {
			target, _ = extractVideoInput(c)
		}
		if target == "" {
			target = c.Query("video_id")
		}

		if target == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Please provide a video ID in the URL path (e.g. /api/v1/videos/Rr4CW2quazA/download-audio) or in the request body.",
			})
			return
		}

		video, err := getOrCreateVideo(target)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		// Fast lookup: if audio file already exists on disk, use it immediately
		audioFileName, err := ensureAudioFile(video)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		audioPath := filepath.Join("downloads", audioFileName)

		// Check if requested as a direct browser attachment download
		isDirectDownload := c.Query("download") == "true" || c.Query("download") == "1" || c.Query("format") == "file" || c.Query("action") == "download" || c.Query("file") == "1"
		acceptHeader := c.GetHeader("Accept")
		if strings.Contains(acceptHeader, "text/html") && c.Request.Method == "GET" && !strings.Contains(c.Query("format"), "json") {
			isDirectDownload = true
		}

		if isDirectDownload {
			downloadName := fmt.Sprintf("%s.mp3", video.ID)
			if video.Title != "" {
				safeTitle := strings.Map(func(r rune) rune {
					if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' || r == '-' || r == '_' {
						return r
					}
					return -1
				}, video.Title)
				safeTitle = strings.TrimSpace(safeTitle)
				if safeTitle != "" {
					downloadName = fmt.Sprintf("%s.mp3", safeTitle)
				}
			}
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", downloadName))
			c.Header("Content-Type", "audio/mpeg")
			c.FileAttachment(audioPath, downloadName)
			return
		}

		info, _ := os.Stat(audioPath)
		var sizeBytes int64
		if info != nil {
			sizeBytes = info.Size()
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": DownloadResult{
				VideoID:          video.ID,
				Title:            video.Title,
				DownloadType:     "audio",
				FileName:         audioFileName,
				FilePath:         audioPath,
				FileSizeBytes:    sizeBytes,
				FileSizeReadable: formatBytes(sizeBytes),
				DurationSeconds:  float64(video.Duration),
				Message:          fmt.Sprintf("Audio file '%s' is ready (%s).", audioFileName, formatBytes(sizeBytes)),
			},
		})
	}

	router.POST("/api/v1/videos/:id/download-audio", handleDownloadAudio)
	router.GET("/api/v1/videos/:id/download-audio", handleDownloadAudio)
	router.POST("/videos/:id/download-audio", handleDownloadAudio)
	router.GET("/videos/:id/download-audio", handleDownloadAudio)
	router.POST("/api/v1/videos/download-audio", handleDownloadAudio)
	router.GET("/api/v1/videos/download-audio", handleDownloadAudio)

	// 8. Download Video via yt-dlp (Supports BOTH POST and GET; auto-registers if needed)
	handleDownloadVideo := func(c *gin.Context) {
		target := c.Param("id")
		if target == "" {
			target, _ = extractVideoInput(c)
		}

		if target == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Please provide a video ID in the URL path (e.g. /api/v1/videos/Rr4CW2quazA/download-video) or in the request body.",
			})
			return
		}

		video, err := getOrCreateVideo(target)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		result, err := downloadMedia(video.URL, video.ID, video.Title, "video")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   result,
		})
	}

	router.POST("/api/v1/videos/:id/download-video", handleDownloadVideo)
	router.GET("/api/v1/videos/:id/download-video", handleDownloadVideo)
	router.POST("/videos/:id/download-video", handleDownloadVideo)
	router.GET("/videos/:id/download-video", handleDownloadVideo)
	router.POST("/api/v1/videos/download-video", handleDownloadVideo)
	router.GET("/api/v1/videos/download-video", handleDownloadVideo)

	// 9. Check downloaded media files on disk for a video
	handleGetMediaFiles := func(c *gin.Context) {
		videoID := c.Param("id")

		downloadsDir := "downloads"
		files, err := os.ReadDir(downloadsDir)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status":   "success",
				"video_id": videoID,
				"count":    0,
				"files":    []gin.H{},
				"message":  "No downloads folder or files exist yet.",
			})
			return
		}

		var matchedFiles []gin.H
		for _, f := range files {
			if !f.IsDir() && strings.HasPrefix(f.Name(), videoID+".") {
				info, _ := f.Info()
				matchedFiles = append(matchedFiles, gin.H{
					"file_name":     f.Name(),
					"size_bytes":    info.Size(),
					"size_readable": formatBytes(info.Size()),
					"modified_at":   info.ModTime(),
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "success",
			"video_id": videoID,
			"count":    len(matchedFiles),
			"files":    matchedFiles,
		})
	}

	router.GET("/api/v1/videos/:id/files", handleGetMediaFiles)
	router.GET("/videos/:id/files", handleGetMediaFiles)

	// 10. Stream / Play audio directly in browser or Postman
	handleStreamAudio := func(c *gin.Context) {
		videoID := c.Param("id")
		if videoID == "" {
			videoID = c.Query("video_id")
		}

		downloadsDir := "downloads"
		var audioPath string
		files, err := os.ReadDir(downloadsDir)
		if err == nil {
			for _, f := range files {
				if !f.IsDir() && strings.HasPrefix(f.Name(), videoID+".") {
					ext := strings.ToLower(filepath.Ext(f.Name()))
					if ext == ".mp3" || ext == ".m4a" || ext == ".opus" || ext == ".wav" {
						audioPath = filepath.Join(downloadsDir, f.Name())
						break
					}
				}
			}
		}

		if audioPath == "" {
			video, vErr := getOrCreateVideo(videoID)
			if vErr == nil {
				if fName, aErr := ensureAudioFile(video); aErr == nil {
					audioPath = filepath.Join(downloadsDir, fName)
				}
			}
		}

		if audioPath == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("No downloaded audio file found for video '%s'.", videoID),
			})
			return
		}

		c.Header("Content-Type", "audio/mpeg")
		c.Header("Accept-Ranges", "bytes")
		c.File(audioPath)
	}

	router.GET("/api/v1/videos/:id/stream-audio", handleStreamAudio)
	router.GET("/videos/:id/stream-audio", handleStreamAudio)

	// 11. Transcribe audio to text using the Python Faster-Whisper microservice
	handleTranscribe := func(c *gin.Context) {
		target := c.Param("id")
		if target == "" {
			target, _ = extractVideoInput(c)
		}

		if target == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Please provide a video ID in the URL path (e.g. /api/v1/videos/Rr4CW2quazA/transcribe) or in the request body.",
			})
			return
		}

		video, err := getOrCreateVideo(target)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		audioFileName, err := ensureAudioFile(video)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		transcription, err := requestWhisperTranscription(video.ID, audioFileName)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		transcriptsMu.Lock()
		transcripts[video.ID] = *transcription
		transcriptsMu.Unlock()

		c.JSON(http.StatusOK, gin.H{
			"status":        "success",
			"video_id":      video.ID,
			"title":         video.Title,
			"transcription": transcription,
		})
	}

	router.POST("/api/v1/videos/:id/transcribe", handleTranscribe)
	router.GET("/api/v1/videos/:id/transcribe", handleTranscribe)
	router.POST("/videos/:id/transcribe", handleTranscribe)
	router.GET("/videos/:id/transcribe", handleTranscribe)
	router.POST("/api/v1/videos/transcribe", handleTranscribe)
	router.GET("/api/v1/videos/transcribe", handleTranscribe)

	// 12. Retrieve existing stored transcript for a video
	handleGetTranscript := func(c *gin.Context) {
		videoID := c.Param("id")

		transcriptsMu.RLock()
		transcript, exists := transcripts[videoID]
		transcriptsMu.RUnlock()

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("No transcript found for video '%s'. Please call POST /api/v1/videos/%s/transcribe first.", videoID, videoID),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":        "success",
			"video_id":      videoID,
			"transcription": transcript,
		})
	}

	router.GET("/api/v1/videos/:id/transcript", handleGetTranscript)
	router.GET("/videos/:id/transcript", handleGetTranscript)

	// 13. Convert Transcript into Chunks, Generate 384-dim Embeddings, and Store in Qdrant
	handleChunkAndEmbed := func(c *gin.Context) {
		target := c.Param("id")
		if target == "" {
			target, _ = extractVideoInput(c)
		}

		if target == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Please provide a video ID in the URL path (e.g. /api/v1/videos/Rr4CW2quazA/index) or in the request body.",
			})
			return
		}

		video, err := getOrCreateVideo(target)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		// Check if transcription exists; if not, automatically transcribe now
		transcriptsMu.RLock()
		transcript, hasTranscript := transcripts[video.ID]
		transcriptsMu.RUnlock()

		if !hasTranscript || len(transcript.Segments) == 0 {
			audioFileName, err := ensureAudioFile(video)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "error",
					"message": fmt.Sprintf("Failed to prepare audio for transcription: %s", err.Error()),
				})
				return
			}

			newTranscript, err := requestWhisperTranscription(video.ID, audioFileName)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{
					"status":  "error",
					"message": fmt.Sprintf("Failed to auto-transcribe: %s", err.Error()),
				})
				return
			}

			transcript = *newTranscript
			transcriptsMu.Lock()
			transcripts[video.ID] = transcript
			transcriptsMu.Unlock()
		}

		// Optional parameters from request body
		var bodyParams struct {
			TargetWords     int `json:"target_words"`
			OverlapSegments int `json:"overlap_segments"`
		}
		_ = c.ShouldBindJSON(&bodyParams)
		targetWords := bodyParams.TargetWords
		if targetWords <= 0 {
			targetWords = 120
		}
		overlapSegs := bodyParams.OverlapSegments
		if overlapSegs <= 0 {
			overlapSegs = 1
		}

		// Call Python microservice to chunk, embed (384 dim), and upsert to Qdrant
		result, err := requestChunkAndEmbed(video.ID, transcript.Segments, targetWords, overlapSegs)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		videoChunksMu.Lock()
		videoChunks[video.ID] = result.Chunks
		videoChunksMu.Unlock()

		c.JSON(http.StatusOK, gin.H{
			"status":               "success",
			"video_id":             video.ID,
			"title":                video.Title,
			"collection_name":      result.CollectionName,
			"embedding_model":      result.EmbeddingModel,
			"embedding_dimensions": result.EmbeddingDimensions,
			"chunks_count":         result.ChunksCount,
			"chunks":               result.Chunks,
			"message":              result.Message,
		})
	}

	router.POST("/api/v1/videos/:id/index", handleChunkAndEmbed)
	router.GET("/api/v1/videos/:id/index", handleChunkAndEmbed)
	router.POST("/api/v1/videos/:id/process-chunks", handleChunkAndEmbed)
	router.GET("/api/v1/videos/:id/process-chunks", handleChunkAndEmbed)
	router.POST("/api/v1/videos/index", handleChunkAndEmbed)
	router.POST("/videos/:id/index", handleChunkAndEmbed)

	// 14. List generated chunks for a video
	handleGetChunks := func(c *gin.Context) {
		videoID := c.Param("id")

		videoChunksMu.RLock()
		chunks, exists := videoChunks[videoID]
		videoChunksMu.RUnlock()

		if !exists || len(chunks) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("No chunks indexed yet for video '%s'. Please call POST /api/v1/videos/%s/index first.", videoID, videoID),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":       "success",
			"video_id":     videoID,
			"chunks_count": len(chunks),
			"chunks":       chunks,
		})
	}

	router.GET("/api/v1/videos/:id/chunks", handleGetChunks)
	router.GET("/videos/:id/chunks", handleGetChunks)

	// 15. Semantic vector search across chunks in Qdrant
	handleSearch := func(c *gin.Context) {
		videoID := c.Param("id")

		var searchReq struct {
			Query   string `json:"query" binding:"required"`
			VideoID string `json:"video_id"`
			Limit   int    `json:"limit"`
		}

		if err := c.ShouldBindJSON(&searchReq); err != nil {
			// Also check query param ?q=... or ?query=...
			q := c.Query("q")
			if q == "" {
				q = c.Query("query")
			}
			if q == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Please provide a query in the JSON body ({\"query\": \"...\"}) or via query param (?query=...).",
				})
				return
			}
			searchReq.Query = q
		}

		if videoID != "" {
			searchReq.VideoID = videoID
		}
		if searchReq.Limit <= 0 {
			searchReq.Limit = 3
		}

		result, err := requestQdrantSearch(searchReq.Query, searchReq.VideoID, searchReq.Limit)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}

	// 16. View Embeddings with respect to chunks and store in Qdrant
	handleEmbeddings := func(c *gin.Context) {
		target := c.Param("id")
		if target == "" {
			target, _ = extractVideoInput(c)
		}

		if target == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Please provide a video ID in the URL path (e.g. /api/v1/videos/Rr4CW2quazA/embeddings) or in the request body.",
			})
			return
		}

		video, err := getOrCreateVideo(target)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		// Ensure transcription exists
		transcriptsMu.RLock()
		transcript, hasTranscript := transcripts[video.ID]
		transcriptsMu.RUnlock()

		if !hasTranscript || len(transcript.Segments) == 0 {
			audioFileName, err := ensureAudioFile(video)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "error",
					"message": fmt.Sprintf("Failed to prepare audio: %s", err.Error()),
				})
				return
			}

			newTranscript, err := requestWhisperTranscription(video.ID, audioFileName)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{
					"status":  "error",
					"message": fmt.Sprintf("Failed to auto-transcribe: %s", err.Error()),
				})
				return
			}

			transcript = *newTranscript
			transcriptsMu.Lock()
			transcripts[video.ID] = transcript
			transcriptsMu.Unlock()
		}

		// Check query param full_vector
		includeFull := true
		if fv := c.Query("full_vector"); fv == "false" || fv == "0" {
			includeFull = false
		}

		// Call python service
		result, err := requestEmbedChunks(video.ID, transcript.Segments, 120, 1, true, includeFull)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		result.Title = video.Title

		// Cache regular chunks
		var simpleChunks []ChunkItem
		for _, cItem := range result.Chunks {
			simpleChunks = append(simpleChunks, ChunkItem{
				ChunkID:          cItem.ChunkID,
				VideoID:          cItem.VideoID,
				ChunkIndex:       cItem.ChunkIndex,
				StartTime:        cItem.StartTime,
				EndTime:          cItem.EndTime,
				TimestampDisplay: cItem.TimestampDisplay,
				Text:             cItem.Text,
				WordCount:        cItem.WordCount,
				SegmentIDs:       cItem.SegmentIDs,
			})
		}
		videoChunksMu.Lock()
		videoChunks[video.ID] = simpleChunks
		videoChunksMu.Unlock()

		// Automatically sync graph and relationships into Neo4j
		go syncVideoToNeo4j(video, simpleChunks)

		// If visited in a browser (Accept: text/html) or format=html, render interactive HTML dashboard
		acceptHeader := c.GetHeader("Accept")
		formatQuery := c.Query("format")
		viewQuery := c.Query("view")
		if strings.Contains(acceptHeader, "text/html") || formatQuery == "html" || viewQuery == "html" {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, renderEmbeddingsHTML(result))
			return
		}

		c.JSON(http.StatusOK, result)
	}

	handleQdrantPoints := func(c *gin.Context) {
		videoID := c.Param("id")
		if videoID == "" {
			videoID = c.Query("video_id")
		}
		points, err := requestQdrantPoints(videoID, 100)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, points)
	}

	handleDashboard := func(c *gin.Context) {
		target := c.Query("video_id")
		if target == "" {
			target = "Rr4CW2quazA"
		}
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/api/v1/videos/%s/embeddings?format=html", target))
	}

	router.GET("/dashboard", handleDashboard)
	router.GET("/dashboard/embeddings", handleDashboard)
	router.GET("/api/v1/videos/:id/qdrant-points", handleQdrantPoints)
	router.GET("/api/v1/qdrant/points", handleQdrantPoints)

	router.POST("/api/v1/videos/:id/embeddings", handleEmbeddings)
	router.GET("/api/v1/videos/:id/embeddings", handleEmbeddings)
	router.POST("/api/v1/videos/:id/embed", handleEmbeddings)
	router.GET("/api/v1/videos/:id/embed", handleEmbeddings)
	router.POST("/api/v1/videos/embeddings", handleEmbeddings)
	router.GET("/api/v1/videos/embeddings", handleEmbeddings)
	router.POST("/videos/:id/embeddings", handleEmbeddings)
	router.GET("/videos/:id/embeddings", handleEmbeddings)

	// 17. Sync video chunks to Neo4j graph & relationships
	handleSyncNeo4j := func(c *gin.Context) {
		videoID := c.Param("id")
		if videoID == "" {
			videoID, _ = extractVideoInput(c)
		}
		if videoID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "video ID is required"})
			return
		}

		video, err := getOrCreateVideo(videoID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
			return
		}

		videoChunksMu.RLock()
		chunks, exists := videoChunks[video.ID]
		videoChunksMu.RUnlock()

		if !exists || len(chunks) == 0 {
			transcriptsMu.RLock()
			t, hasT := transcripts[video.ID]
			transcriptsMu.RUnlock()
			if !hasT || len(t.Segments) == 0 {
				audioFile, aErr := ensureAudioFile(video)
				if aErr != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": aErr.Error()})
					return
				}
				nt, tErr := requestWhisperTranscription(video.ID, audioFile)
				if tErr != nil {
					c.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": tErr.Error()})
					return
				}
				t = *nt
				transcriptsMu.Lock()
				transcripts[video.ID] = t
				transcriptsMu.Unlock()
			}
			res, err := requestEmbedChunks(video.ID, t.Segments, 120, 1, true, false)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
				return
			}
			for _, ci := range res.Chunks {
				chunks = append(chunks, ChunkItem{
					ChunkID:          ci.ChunkID,
					VideoID:          ci.VideoID,
					ChunkIndex:       ci.ChunkIndex,
					StartTime:        ci.StartTime,
					EndTime:          ci.EndTime,
					TimestampDisplay: ci.TimestampDisplay,
					Text:             ci.Text,
					WordCount:        ci.WordCount,
					SegmentIDs:       ci.SegmentIDs,
				})
			}
			videoChunksMu.Lock()
			videoChunks[video.ID] = chunks
			videoChunksMu.Unlock()
		}

		graphResp, err := syncVideoToNeo4j(video, chunks)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, graphResp)
	}

	// 18. Retrieve Neo4j Graph for a video
	handleGetGraph := func(c *gin.Context) {
		videoID := c.Param("id")
		if videoID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "video ID is required"})
			return
		}

		graph, err := fetchVideoGraph(videoID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, graph)
	}

	// 19. Retrieve Expanded Graph Context for a chunk (Prev, Target, Next, Concepts)
	handleChunkContext := func(c *gin.Context) {
		videoID := c.Param("id")
		chunkID := c.Param("chunk_id")
		if videoID == "" || chunkID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "video ID and chunk_id are required"})
			return
		}

		ctxResp, err := fetchChunkContext(videoID, chunkID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, ctxResp)
	}

	// 20. Neo4j Status check
	handleNeo4jStatus := func(c *gin.Context) {
		res, err := executeCypher([]CypherStatement{{Statement: "RETURN 'Neo4j Connected!' AS message"}})
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"status": "error", "connected": false, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":            "success",
			"connected":         true,
			"neo4j_browser_url": "http://localhost:7474/browser/",
			"credentials": gin.H{
				"username": "neo4j",
				"password": "videorag123",
			},
			"result": res,
		})
	}

	// 21. Chatbot endpoint: Qdrant semantic vector search + Neo4j Graph RAG context expansion + Ollama Local LLM Synthesis
	handleChat := func(c *gin.Context) {
		videoID := c.Param("id")
		var chatReq ChatRequest
		if err := c.ShouldBindJSON(&chatReq); err != nil {
			q := c.Query("q")
			if q == "" {
				q = c.Query("query")
			}
			chatReq.Query = q
		}

		if videoID == "" {
			videoID = chatReq.VideoID
		}
		if videoID == "" {
			videoID = c.Query("video_id")
		}
		if videoID == "" {
			videoID = "Rr4CW2quazA"
		}

		if strings.TrimSpace(chatReq.Query) == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Query string is required via JSON body ({\"query\": \"...\"}) or query parameter (?query=...).",
			})
			return
		}

		videosMu.RLock()
		vInfo, vExists := videos[videoID]
		videosMu.RUnlock()
		vTitle := videoID
		if vExists && vInfo.Title != "" {
			vTitle = vInfo.Title
		}

		// 1. Check if request is a summary / overview request
		qLower := strings.ToLower(chatReq.Query)
		isSummary := strings.Contains(qLower, "summar") || strings.Contains(qLower, "overview") || strings.Contains(qLower, "what is this video about") || strings.Contains(qLower, "explain this video")
		activeModel := getOllamaModel()

		if isSummary {
			repChunks, _ := fetchRepresentativeChunks(videoID)
			if len(repChunks) == 0 {
				searchRes, _ := requestQdrantSearch(chatReq.Query, videoID, 6)
				if searchRes != nil {
					for _, m := range searchRes.Matches {
						repChunks = append(repChunks, fmt.Sprintf("[%s]: %s", m.TimestampDisplay, m.Text))
					}
				}
			}

			if len(repChunks) > 0 {
				contextBlock := strings.Join(repChunks, "\n\n")
				prompt := fmt.Sprintf(`You are an expert AI video assistant analyzing a YouTube video. Below are chronological transcript excerpts from the video:

Video Title: "%s"

Chronological Transcript Excerpts:
%s

Instructions:
Provide a clear, high-quality, structured summary of this video strictly based on the excerpts above.
Format your answer using clean Markdown with:
- 🎯 **Overview**: A concise 2-sentence explanation of what this video teaches.
- 🔑 **Key Topics & Steps**: 4-6 bullet points highlighting the specific tools, code, libraries, and concepts covered.
- 💡 **Core Takeaway**: The primary conclusion from the video.

Rules:
1. Ground your answer ONLY in the provided transcript excerpts.
2. Do NOT hallucinate or assume facts not present in the excerpts.`, vTitle, contextBlock)

				aiAnswer, err := queryOllama(prompt, activeModel, 120*time.Second)
				if err == nil && len(aiAnswer) > 20 {
					var allConcepts []string
					if g, _ := fetchChunkContext(videoID, ""); g != nil && len(g.Concepts) > 0 {
						allConcepts = g.Concepts
					}
					c.JSON(http.StatusOK, ChatResponse{
						Status:        "success",
						Query:         chatReq.Query,
						VideoID:       videoID,
						Answer:        aiAnswer,
						Highlights:    extractHighlights(aiAnswer),
						Concepts:      allConcepts,
						Score:         1.0,
						Timestamp:     "Full Video",
						YouTubeURL:    fmt.Sprintf("https://youtu.be/%s", videoID),
						SourceChunkID: "summary_multi_chunks",
						OllamaModel:   activeModel,
						RAGSource:     "Ollama Grounded Summary (" + activeModel + ")",
					})
					return
				}
			}
		}

		// 2. Standard Question RAG: search top 4 chunks in Qdrant
		searchRes, err := requestQdrantSearch(chatReq.Query, videoID, 4)
		if err != nil || len(searchRes.Matches) == 0 {
			c.JSON(http.StatusOK, ChatResponse{
				Status:     "success",
				Query:      chatReq.Query,
				VideoID:    videoID,
				Answer:     "I could not find a relevant section in the video matching your question. Try asking about concepts covered in this video.",
				Highlights: []string{"No direct matching video segment found."},
				Concepts:   []string{},
			})
			return
		}

		top := searchRes.Matches[0]

		// 3. Query Neo4j for ChunkContext (PREV_CHUNK, NEXT_CHUNK, and DISCUSSES Concepts)
		ctxResp, _ := fetchChunkContext(videoID, top.ChunkID)

		// 4. Extract or use Neo4j concepts
		var concepts []string
		if ctxResp != nil && len(ctxResp.Concepts) > 0 {
			concepts = ctxResp.Concepts
		} else {
			concepts = extractKeyConcepts(top.Text)
		}

		// 5. Build multi-chunk grounded RAG context for Ollama
		var excerpts []string
		for idx, m := range searchRes.Matches {
			if idx >= 3 {
				break
			}
			excerpts = append(excerpts, fmt.Sprintf("Excerpt %d [%s]: %s", idx+1, m.TimestampDisplay, cleanExcerpt(m.Text, 350)))
		}
		if ctxResp != nil && ctxResp.FullContext != "" {
			excerpts = append(excerpts, fmt.Sprintf("Graph Context: %s", cleanExcerpt(ctxResp.FullContext, 300)))
		}

		prompt := fmt.Sprintf(`You are an AI video assistant answering questions about a YouTube video strictly using the transcript excerpts and knowledge graph concepts provided below.

Video Title: "%s"

Video Transcript Excerpts:
%s

Concepts Identified: %s

User Question: "%s"

Instructions:
1. Answer the user's question accurately, directly, and concisely using ONLY facts explicitly stated in the transcript excerpts above.
2. If the excerpts do not contain enough information to answer the question, state: "The video does not provide enough detail to answer this question."
3. Do NOT make outside assumptions or add information from the internet.`, vTitle, strings.Join(excerpts, "\n\n"), strings.Join(concepts, ", "), chatReq.Query)

		aiAnswer, err := queryOllama(prompt, activeModel, 90*time.Second)
		finalAnswer := top.Text
		ragSource := "Direct Transcript Chunk"
		highlights := extractHighlights(top.Text)
		if err == nil && len(aiAnswer) > 10 {
			finalAnswer = aiAnswer
			ragSource = "Ollama Grounded RAG (" + activeModel + ")"
			highlights = extractHighlights(aiAnswer)
		} else {
			finalAnswer = fmt.Sprintf("%s\n\n*(Note: Ollama offline/busy (%v), displaying direct matching transcript excerpt)*", top.Text, err)
		}

		c.JSON(http.StatusOK, ChatResponse{
			Status:        "success",
			Query:         chatReq.Query,
			VideoID:       videoID,
			Answer:        finalAnswer,
			Highlights:    highlights,
			Concepts:      concepts,
			Score:         top.Score,
			Timestamp:     top.TimestampDisplay,
			YouTubeURL:    top.YouTubeTimestampURL,
			SourceChunkID: top.ChunkID,
			GraphContext:  ctxResp,
			OllamaModel:   activeModel,
			RAGSource:     ragSource,
		})
	}

	// 22. End-to-end Process Video URL endpoint (Download -> Transcribe -> Embed 384-dim -> Qdrant -> Neo4j)
	handleProcessVideo := func(c *gin.Context) {
		rawURL, userTitle := extractVideoInput(c)
		if rawURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Please provide a valid YouTube video URL or ID via JSON body ({\"url\": \"...\"}), form data, or query param (?url=...).",
			})
			return
		}

		video, err := getOrCreateVideo(rawURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		if userTitle != "" {
			video.Title = userTitle
			videosMu.Lock()
			videos[video.ID] = video
			videosMu.Unlock()
		}

		// 1. Ensure audio file is downloaded
		audioFileName, err := ensureAudioFile(video)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("Failed to download audio for '%s': %s", video.Title, err.Error()),
			})
			return
		}

		// 2. Transcribe with Faster-Whisper
		transcriptsMu.RLock()
		transcript, hasT := transcripts[video.ID]
		transcriptsMu.RUnlock()

		if !hasT || len(transcript.Segments) == 0 {
			newT, err := requestWhisperTranscription(video.ID, audioFileName)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{
					"status":  "error",
					"message": fmt.Sprintf("Faster-Whisper transcription failed: %s", err.Error()),
				})
				return
			}
			transcript = *newT
			transcriptsMu.Lock()
			transcripts[video.ID] = transcript
			transcriptsMu.Unlock()
		}

		// 3. Chunk and generate 384-dim FastEmbed embeddings & store in Qdrant
		embedRes, err := requestEmbedChunks(video.ID, transcript.Segments, 120, 1, true, true)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("FastEmbed / Qdrant indexing failed: %s", err.Error()),
			})
			return
		}
		embedRes.Title = video.Title

		// 4. Cache simple chunks
		var simpleChunks []ChunkItem
		for _, cItem := range embedRes.Chunks {
			simpleChunks = append(simpleChunks, ChunkItem{
				ChunkID:          cItem.ChunkID,
				VideoID:          cItem.VideoID,
				ChunkIndex:       cItem.ChunkIndex,
				StartTime:        cItem.StartTime,
				EndTime:          cItem.EndTime,
				TimestampDisplay: cItem.TimestampDisplay,
				Text:             cItem.Text,
				WordCount:        cItem.WordCount,
				SegmentIDs:       cItem.SegmentIDs,
			})
		}
		videoChunksMu.Lock()
		videoChunks[video.ID] = simpleChunks
		videoChunksMu.Unlock()

		// 5. Synchronize knowledge graph into Neo4j
		graphRes, _ := syncVideoToNeo4j(video, simpleChunks)

		redirectURL := fmt.Sprintf("/api/v1/videos/%s/embeddings?format=html", video.ID)

		// If user requested via browser redirect
		acceptHeader := c.GetHeader("Accept")
		if strings.Contains(acceptHeader, "text/html") && c.Request.Method == "GET" {
			c.Redirect(http.StatusTemporaryRedirect, redirectURL)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":       "success",
			"video_id":     video.ID,
			"title":        video.Title,
			"chunks_count": len(simpleChunks),
			"neo4j_synced": graphRes != nil,
			"redirect_url": redirectURL,
			"message":      fmt.Sprintf("Video '%s' successfully indexed in Qdrant and mapped in Neo4j!", video.Title),
		})
	}

	// 23. Run custom Cypher query from UI
	handleRunCypher := func(c *gin.Context) {
		var body struct {
			Query string `json:"query"`
		}
		if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Query) == "" {
			q := c.Query("query")
			if q == "" {
				c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "query string is required"})
				return
			}
			body.Query = q
		}

		res, err := executeCypher([]CypherStatement{{Statement: body.Query}})
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success", "result": res})
	}

	// Root redirects directly to dashboard / UI
	router.GET("/", handleDashboard)

	router.POST("/api/v1/neo4j/query", handleRunCypher)
	router.GET("/api/v1/neo4j/query", handleRunCypher)
	router.POST("/neo4j/query", handleRunCypher)
	router.GET("/neo4j/query", handleRunCypher)

	// URL Processing & Extraction routes
	router.POST("/api/v1/videos/process", handleProcessVideo)
	router.POST("/api/v1/videos/extract", handleProcessVideo)
	router.GET("/api/v1/videos/extract", handleProcessVideo)
	router.POST("/api/v1/extract", handleProcessVideo)
	router.GET("/api/v1/extract", handleProcessVideo)
	router.POST("/process", handleProcessVideo)
	router.GET("/process", handleProcessVideo)

	// Chatbot routes
	router.POST("/api/v1/videos/:id/chat", handleChat)
	router.GET("/api/v1/videos/:id/chat", handleChat)
	router.POST("/api/v1/chat", handleChat)
	router.GET("/api/v1/chat", handleChat)

	router.POST("/api/v1/videos/:id/graph", handleSyncNeo4j)
	router.GET("/api/v1/videos/:id/graph", handleGetGraph)
	router.POST("/api/v1/videos/:id/neo4j", handleSyncNeo4j)
	router.GET("/api/v1/videos/:id/neo4j", handleGetGraph)
	router.GET("/api/v1/videos/:id/chunks/:chunk_id/context", handleChunkContext)
	router.GET("/api/v1/neo4j/status", handleNeo4jStatus)
	router.GET("/neo4j/status", handleNeo4jStatus)

	router.POST("/api/v1/videos/:id/search", handleSearch)
	router.GET("/api/v1/videos/:id/search", handleSearch)
	router.POST("/api/v1/videos/search", handleSearch)
	router.GET("/api/v1/videos/search", handleSearch)
	router.POST("/videos/:id/search", handleSearch)

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
