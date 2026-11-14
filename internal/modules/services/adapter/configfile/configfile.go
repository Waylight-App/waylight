// Package configfile maps the manual service tiles to their section of
// trisse-home.yaml (#88) and back.
//
// Ring: Interface Adapters.
package configfile

import "trisse-home/internal/modules/services/usecase"

// Service is one tile. checkUrl is only written when it differs from url.
type Service struct {
	Name          string `yaml:"name"`
	URL           string `yaml:"url"`
	CheckURL      string `yaml:"checkUrl,omitempty"`
	Icon          string `yaml:"icon,omitempty"`
	Description   string `yaml:"description,omitempty"`
	SkipTLSVerify bool   `yaml:"skipTlsVerify,omitempty"`
	MAC           string `yaml:"mac,omitempty"`
}

// Section is the services section: a list of tiles in display order.
type Section []Service

// FromList maps the use case list.
func FromList(list []usecase.ConfigService) *Section {
	out := Section{}
	for _, s := range list {
		check := s.CheckURL
		if check == s.URL {
			check = ""
		}
		out = append(out, Service{Name: s.Name, URL: s.URL, CheckURL: check, Icon: s.Icon, Description: s.Description,
			SkipTLSVerify: s.SkipTLSVerify, MAC: s.MAC})
	}
	return &out
}

// List maps back to the use case list.
func (s Section) List() []usecase.ConfigService {
	out := make([]usecase.ConfigService, 0, len(s))
	for _, v := range s {
		out = append(out, usecase.ConfigService(v))
	}
	return out
}
