package tree

import (
	"fmt"
	"os"
	"strings"

	kdl "github.com/sblinch/kdl-go"
)

// Config holds the configuration for a tree.
type Config struct {
	ID            string
	Name          string
	Description   string
	SurnameGroups *SurnameGroups
	Annotations   *Annotations
}

// ReadConfig reads a KDL config file and returns a *Config.
func ReadConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open tree config: %w", err)
	}
	defer f.Close()

	doc, err := kdl.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parse tree config: %w", err)
	}

	cfg := &Config{}

	for _, node := range doc.Nodes {
		switch node.Name.ValueString() {
		case "tree":
			if v, ok := node.Properties.Get("id"); ok {
				cfg.ID, _ = v.Value.(string)
			}
			for _, child := range node.Children {
				switch child.Name.ValueString() {
				case "id":
					if len(child.Arguments) > 0 {
						cfg.ID, _ = child.Arguments[0].Value.(string)
					}
				case "name":
					if len(child.Arguments) > 0 {
						cfg.Name, _ = child.Arguments[0].Value.(string)
					}
				case "description":
					if len(child.Arguments) > 0 {
						s, _ := child.Arguments[0].Value.(string)
						cfg.Description = strings.TrimSpace(s)
					}
				}
			}
		case "annotations":
			a := &Annotations{}
			for _, section := range node.Children {
				kind := section.Name.ValueString()
				var kindName string
				switch kind {
				case "people":
					kindName = "person"
				case "places":
					kindName = "place"
				case "sources":
					kindName = "source"
				default:
					return nil, fmt.Errorf("unknown annotations section %q", kind)
				}
				for _, obj := range section.Children {
					idVal, ok := obj.Properties.Get("id")
					if !ok {
						return nil, fmt.Errorf("%s annotation missing id property", kindName)
					}
					id, _ := idVal.Value.(string)
					for _, field := range obj.Children {
						fieldName := field.Name.ValueString()
						var value any
						switch len(field.Arguments) {
						case 0:
							value = true
						case 1:
							value = field.Arguments[0].Value
						default:
							args := make([]any, len(field.Arguments))
							for i, arg := range field.Arguments {
								args[i] = arg.Value
							}
							value = args
						}
						if err := a.Set(kindName, id, fieldName, value); err != nil {
							return nil, fmt.Errorf("%s %q field %q: %w", kindName, id, fieldName, err)
						}
					}
				}
			}
			cfg.Annotations = a
		case "surname-groups":
			sg := &SurnameGroups{}
			for _, child := range node.Children {
				canonical := child.Name.ValueString()
				variants := make([]string, 0, len(child.Arguments))
				for _, arg := range child.Arguments {
					if s, ok := arg.Value.(string); ok {
						variants = append(variants, s)
					}
				}
				sg.AddGroup(canonical, variants)
			}
			cfg.SurnameGroups = sg
		}
	}

	return cfg, nil
}
