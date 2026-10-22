package config

import "fmt"

// Dependency describes an infrastructure component and its optional local endpoint.
type Dependency struct {
	Name       string
	Deployment string
	Port       PortBinding
	Database   string
}

func Dependencies() []Dependency {
	return []Dependency{
		{
			Name:       "rental-postgres",
			Deployment: "rental-postgres",
			Port: PortBinding{
				Host:      25440,
				Node:      30432,
				LocalOnly: true,
			},
			Database: "rental",
		},
		{
			Name:       "pricing-postgres",
			Deployment: "pricing-postgres",
			Port: PortBinding{
				Host:      25441,
				Node:      30433,
				LocalOnly: true,
			},
			Database: "pricing",
		},
		{
			Name:       "profile-postgres",
			Deployment: "profile-postgres",
			Port: PortBinding{
				Host:      25442,
				Node:      30434,
				LocalOnly: true,
			},
			Database: "profile",
		},
		{
			Name:       "scooter-postgres",
			Deployment: "scooter-postgres",
			Port: PortBinding{
				Host:      25443,
				Node:      30435,
				LocalOnly: true,
			},
			Database: "scooter",
		},
		{
			Name:       "kafka",
			Deployment: "rental-kafka",
		},
		{
			Name:       "config-etcd",
			Deployment: "config-etcd",
			Port: PortBinding{
				Host:      12379,
				Node:      32379,
				LocalOnly: true,
			},
		},
		{
			Name:       "akhq",
			Deployment: "akhq",
			Port: PortBinding{
				Host:      18090,
				Node:      30089,
				LocalOnly: true,
			},
		},
	}
}

func (c Config) Components() []Service {
	result := make([]Service, 0, len(c.Services)+len(Dependencies()))
	for _, dependency := range Dependencies() {
		result = append(result, Service{
			Name:       dependency.Name,
			Deployment: dependency.Deployment,
			Namespace:  "scooterflow",
		})
	}
	return append(result, c.Services...)
}

func (c Config) SelectComponents(names []string) ([]Service, error) {
	available := c.Components()
	var result []Service
	seen := map[string]bool{}
	for _, name := range names {
		if seen[name] {
			return nil, fmt.Errorf("компонент %q указан дважды", name)
		}
		seen[name] = true
		found := false
		for _, component := range available {
			if component.Name == name {
				result = append(result, component)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("неизвестный компонент %q; список: scooterctl components", name)
		}
	}
	return result, nil
}
