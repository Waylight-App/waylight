package usecase

import (
	"testing"

	"waylight/internal/modules/proxmox/domain"
	"waylight/internal/shared/kernel"
)

func TestCompactState(t *testing.T) {
	cluster := domain.Cluster{
		Nodes:  []domain.Node{{Name: "pve", Online: true}},
		Guests: []domain.Guest{{ID: 100, Status: "running"}, {ID: 101, Status: "running"}, {ID: 102, Status: "running"}, {ID: 103, Status: "stopped"}},
	}
	s := CompactState(StatusView{Cluster: cluster})
	if s.Title.String() != "3 of 4 guests running" || s.Detail.Template != "" || s.Level != kernel.CompactNone {
		t.Fatalf("healthy: %+v", s)
	}
	cluster.Nodes = append(cluster.Nodes, domain.Node{Name: "pve2"})
	if s := CompactState(StatusView{Cluster: cluster}); s.Level != kernel.CompactAttention || s.Detail.String() != "1 of 2 nodes offline" || *s.Count != 1 {
		t.Fatalf("node offline: %+v", s)
	}
	if s := CompactState(StatusView{Cluster: cluster, Error: "timeout"}); s.Level != kernel.CompactAttention || s.Detail.String() != "Proxmox VE is not reachable" {
		t.Fatalf("unreachable: %+v", s)
	}
}
