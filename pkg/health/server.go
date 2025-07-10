package health

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// Server provides health monitoring endpoints
type Server struct {
	port         int
	agentInfo    *AgentInfo
	statusGetter StatusGetter
	server       *http.Server
}

// AgentInfo contains basic agent information
type AgentInfo struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Wallet       string   `json:"wallet"`
	Capabilities []string `json:"capabilities"`
	Description  string   `json:"description"`
}

// StatusGetter interface for getting agent status
type StatusGetter interface {
	IsConnected() bool
	IsAuthenticated() bool
	GetActiveTaskCount() int
	GetUptime() time.Duration
}

// HealthStatus represents the agent's health status
type HealthStatus struct {
	Status        string    `json:"status"`
	Connected     bool      `json:"connected"`
	Authenticated bool      `json:"authenticated"`
	ActiveTasks   int       `json:"active_tasks"`
	Uptime        string    `json:"uptime"`
	Timestamp     time.Time `json:"timestamp"`
	Agent         AgentInfo `json:"agent"`
}

// NewServer creates a new health monitoring server
func NewServer(port int, agentInfo *AgentInfo, statusGetter StatusGetter) *Server {
	return &Server{
		port:         port,
		agentInfo:    agentInfo,
		statusGetter: statusGetter,
	}
}

// Start starts the health monitoring server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Health endpoints
	mux.HandleFunc("/", s.rootHandler)
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/status", s.statusHandler)
	mux.HandleFunc("/info", s.infoHandler)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	log.Printf("🌐 Starting health server on port %d...", s.port)
	return s.server.ListenAndServe()
}

// Stop stops the health monitoring server
func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// rootHandler handles the root endpoint
func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "Hello World from %s!\n", s.agentInfo.Name)
	fmt.Fprintf(w, "Agent: %s v%s\n", s.agentInfo.Name, s.agentInfo.Version)
	fmt.Fprintf(w, "Wallet: %s\n", s.agentInfo.Wallet)
	fmt.Fprintf(w, "Connected: %v\n", s.statusGetter.IsConnected())
	fmt.Fprintf(w, "Authenticated: %v\n", s.statusGetter.IsAuthenticated())
	fmt.Fprintf(w, "Active Tasks: %d\n", s.statusGetter.GetActiveTaskCount())
	fmt.Fprintf(w, "Capabilities: %s\n", strings.Join(s.agentInfo.Capabilities, ", "))
	fmt.Fprintf(w, "Uptime: %v\n", s.statusGetter.GetUptime())
	fmt.Fprintf(w, "\nEndpoints:\n")
	fmt.Fprintf(w, "  /health - Health check\n")
	fmt.Fprintf(w, "  /status - Detailed status (JSON)\n")
	fmt.Fprintf(w, "  /info   - Agent information (JSON)\n")
}

// healthHandler provides a simple health check
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	connected := s.statusGetter.IsConnected()
	authenticated := s.statusGetter.IsAuthenticated()

	var status string
	var statusCode int

	if connected && authenticated {
		status = "healthy"
		statusCode = http.StatusOK
	} else if connected {
		status = "connected_not_authenticated"
		statusCode = http.StatusOK
	} else {
		status = "disconnected"
		statusCode = http.StatusServiceUnavailable
	}

	w.WriteHeader(statusCode)

	health := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now(),
		"agent":     s.agentInfo.Name,
	}

	json.NewEncoder(w).Encode(health)
}

// statusHandler provides detailed status information
func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	connected := s.statusGetter.IsConnected()
	authenticated := s.statusGetter.IsAuthenticated()

	var status string
	if connected && authenticated {
		status = "operational"
	} else if connected {
		status = "connected"
	} else {
		status = "disconnected"
	}

	healthStatus := HealthStatus{
		Status:        status,
		Connected:     connected,
		Authenticated: authenticated,
		ActiveTasks:   s.statusGetter.GetActiveTaskCount(),
		Uptime:        s.statusGetter.GetUptime().String(),
		Timestamp:     time.Now(),
		Agent:         *s.agentInfo,
	}

	json.NewEncoder(w).Encode(healthStatus)
}

// infoHandler provides agent information
func (s *Server) infoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(s.agentInfo)
}

// UpdateAgentInfo updates the agent information
func (s *Server) UpdateAgentInfo(info *AgentInfo) {
	s.agentInfo = info
}
