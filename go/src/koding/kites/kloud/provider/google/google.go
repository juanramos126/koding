package google

import (
	"errors"
	"fmt"
	"strconv"

	"koding/kites/kloud/stack"
	"koding/kites/kloud/stack/provider"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

var p = &provider.Provider{
	Name:         "google",
	ResourceName: "compute_instance",
	Machine:      newMachine,
	Stack:        newStack,
	Schema: &provider.Schema{
		NewCredential: newCredential,
		NewBootstrap:  newBootstrap,
		NewMetadata:   newMetadata,
	},
}

func init() {
	provider.Register(p)
}

// Region represents google's geographical region code.
type Region string

var regions = []Region{
	"asia-east1",
	"europe-west1",
	"us-central1",
	"us-east1",
	"us-west1",
}

// Enum returns all available regions for "google" provider.
func (Region) Enum() (rs []interface{}) {
	for _, region := range regions {
		rs = append(rs, region)
	}
	return rs
}

// Valid checks if stored region code is available in GCP.
func (r Region) Valid() error {
	if r == "" {
		return fmt.Errorf("region name is not set")
	}

	for _, region := range regions {
		if r == region {
			return nil
		}
	}

	return fmt.Errorf("unknown region name: %v", r)
}

// Cred represents jCredentialDatas.meta for "google" provider.
type Cred struct {
	Credentials string `json:"credentials" bson:"credentials" hcl:"credentials" kloud:",secret"`
	Project     string `json:"project" bson:"project" hcl:"project"`
	Region      Region `json:"region" bson:"region" hcl:"region"`
}

var _ stack.Validator = (*Cred)(nil)

func newCredential() interface{} {
	return &Cred{}
}

func (c *Cred) Valid() error {
	if c.Credentials == "" {
		return errors.New(`cred value for "credentials" is empty`)
	}
	if c.Project == "" {
		return errors.New(`cred value for "project" is empty`)
	}

	return c.Region.Valid()
}

func (c *Cred) ComputeService() (*compute.Service, error) {
	cfg, err := google.JWTConfigFromJSON([]byte(c.Credentials), compute.ComputeScope)
	if err != nil {
		return nil, err
	}

	return compute.New(cfg.Client(context.Background()))
}

type Bootstrap struct {
	KodingFirewall string `json:"koding_firewall" bson:"koding_firewall" hcl:"koding_firewall"`
}

var _ stack.Validator = (*Bootstrap)(nil)

func newBootstrap() interface{} {
	return &Bootstrap{}
}

func (b *Bootstrap) Valid() error {
	if b.KodingFirewall == "" {
		return errors.New(`bootstrap value for "koding_firewall" is empty`)
	}
	return nil
}

type Meta struct {
	Name        string `json:"name" bson:"name" hcl:"name"`
	Region      Region `json:"region" bson:"region" hcl:"region"`
	Zone        string `json:"zone" bson:"zone" hcl:"zone"`
	Image       string `json:"image" bson:"image" hcl:"image"`
	StorageSize int    `json:"storage_size" bson:"storage_size" hcl:"storage_size"`
	MachineType string `json:"machine_type" bson:"machine_type" hcl:"machine_type"`
}

var _ stack.Validator = (*Meta)(nil)

func newMetadata(m *stack.Machine) interface{} {
	if m == nil {
		return &Meta{}
	}

	meta := &Meta{
		Name:        m.Attributes["name"],
		Zone:        m.Attributes["zone"],
		Image:       m.Attributes["disk.0.image"],
		MachineType: m.Attributes["machine_type"],
	}

	if n, err := strconv.Atoi(m.Attributes["disk.0.size"]); err == nil {
		meta.StorageSize = n
	}

	if cred, ok := m.Credential.Credential.(*Cred); ok {
		meta.Region = cred.Region
	}

	return meta
}

func (m *Meta) Valid() error {
	if m.Name == "" {
		return errors.New(`metadata value for "name" is empty`)
	}
	if err := m.Region.Valid(); err != nil {
		return fmt.Errorf(`metadata region is unknown: %v`, err)
	}
	if m.Zone == "" {
		return errors.New(`metadata value for "zone" is empty`)
	}
	if m.Image == "" {
		return errors.New(`metadata value for "image" is empty`)
	}
	if m.StorageSize == 0 {
		return errors.New(`metadata value for "storage_size" is 0`)
	}
	if m.MachineType == "" {
		return errors.New(`metadata value for "machie_type" is empty`)
	}

	return nil
}
