package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"waylight/internal/modules/docker/domain"
)

const maxResponseBytes = 4 << 20

// API lists containers via the Docker Engine HTTP API.
type API struct {
	client  *http.Client
	baseURL string
}

// NewAPI returns a client for baseURL, e.g. http://docker-socket-proxy:2375.
func NewAPI(client *http.Client, baseURL string) *API {
	return &API{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

// containerPayload is the subset of GET /containers/json we use.
type containerPayload struct {
	ID      string            `json:"Id"`
	Names   []string          `json:"Names"`
	Image   string            `json:"Image"`
	State   string            `json:"State"`
	Status  string            `json:"Status"`
	Created int64             `json:"Created"`
	Labels  map[string]string `json:"Labels"`
	Ports   []struct {
		PrivatePort int    `json:"PrivatePort"`
		PublicPort  int    `json:"PublicPort"`
		Type        string `json:"Type"`
	} `json:"Ports"`
	NetworkSettings struct {
		Networks map[string]struct {
			IPAddress string `json:"IPAddress"`
		} `json:"Networks"`
	} `json:"NetworkSettings"`
}

// Containers implements usecase.ContainerSource.
func (a *API) Containers(ctx context.Context) ([]domain.Container, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/containers/json?all=1", nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	res, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("docker api: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docker api returned %s", res.Status)
	}
	var payload []containerPayload
	if err := json.NewDecoder(io.LimitReader(res.Body, maxResponseBytes)).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode docker response: %w", err)
	}
	out := make([]domain.Container, 0, len(payload))
	for _, p := range payload {
		name := p.ID
		if len(p.Names) > 0 {
			name = strings.TrimPrefix(p.Names[0], "/")
		}
		id := p.ID
		if len(id) > 12 {
			id = id[:12]
		}
		out = append(out, domain.Container{
			ID: id, Name: name, Image: p.Image, State: p.State, Status: p.Status,
			Project: p.Labels["com.docker.compose.project"], Created: time.Unix(p.Created, 0).UTC(), Labels: p.Labels,
			Ports: ports(p), Addresses: addresses(p),
		})
	}
	return out, nil
}

// ports keeps the TCP ports once each; Docker lists a port per host
// address (IPv4 and IPv6) and per protocol.
func ports(p containerPayload) []domain.Port {
	var out []domain.Port
	seen := map[domain.Port]bool{}
	for _, x := range p.Ports {
		port := domain.Port{Private: x.PrivatePort, Public: x.PublicPort}
		if x.Type != "tcp" || x.PrivatePort == 0 || seen[port] {
			continue
		}
		seen[port] = true
		out = append(out, port)
	}
	return out
}

// addresses lists the container's IPs in its networks, sorted.
func addresses(p containerPayload) []string {
	var out []string
	for _, n := range p.NetworkSettings.Networks {
		if n.IPAddress != "" {
			out = append(out, n.IPAddress)
		}
	}
	slices.Sort(out)
	return out
}
