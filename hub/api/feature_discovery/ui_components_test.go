// Package feature_discovery provides tests for UI component discovery
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package feature_discovery

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestDiscoverUIComponents_React tests React component discovery
func TestDiscoverUIComponents_React(t *testing.T) {
	tempDir := t.TempDir()

	// Create a React component file
	componentContent := `import React, { useState } from 'react';
import { Button } from './Button';

interface UserCardProps {
  name: string;
  email: string;
  avatar?: string;
}

const UserCard: React.FC<UserCardProps> = ({ name, email, avatar }) => {
  const [isExpanded, setIsExpanded] = useState(false);

  const handleToggle = () => {
    setIsExpanded(!isExpanded);
  };

  return (
    <div className="user-card">
      <img src={avatar} alt={name} />
      <h3>{name}</h3>
      <p>{email}</p>
      <Button onClick={handleToggle}>
        {isExpanded ? 'Collapse' : 'Expand'}
      </Button>
    </div>
  );
};

export default UserCard;`

	componentPath := filepath.Join(tempDir, "UserCard.tsx")
	err := os.WriteFile(componentPath, []byte(componentContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create component file: %v", err)
	}

	// Test component discovery
	components, err := discoverUIComponents(context.Background(), tempDir, "user", "react")
	if err != nil {
		t.Fatalf("discoverUIComponents returned error: %v", err)
	}

	if components.Framework != "react" {
		t.Errorf("Expected framework 'react', got '%s'", components.Framework)
	}

	if len(components.Components) == 0 {
		t.Errorf("Expected to find components, but found none")
	}

	if len(components.Components) > 0 {
		component := components.Components[0]
		if component.Name != "UserCard" {
			t.Errorf("Expected component name 'UserCard', got '%s'", component.Name)
		}

		if component.Type != "functional" {
			t.Errorf("Expected component type 'functional', got '%s'", component.Type)
		}

		// Check props extraction
		if len(component.Props) == 0 {
			t.Errorf("Expected to find props, but found none")
		}

		// Check state extraction
		if len(component.State) == 0 {
			t.Errorf("Expected to find state, but found none")
		}

		// Check dependencies
		if len(component.Dependencies) == 0 {
			t.Errorf("Expected to find dependencies, but found none")
		}
	}
}

// TestDiscoverUIComponents_Vue tests Vue component discovery
func TestDiscoverUIComponents_Vue(t *testing.T) {
	tempDir := t.TempDir()

	// Create a Vue component file
	componentContent := `<template>
  <div class="user-profile">
    <img :src="avatar" :alt="name" />
    <h3>{{ name }}</h3>
    <p>{{ email }}</p>
    <button @click="toggleExpanded">
      {{ isExpanded ? 'Collapse' : 'Expand' }}
    </button>
  </div>
</template>

<script>
export default {
  name: 'UserProfile',
  props: {
    name: {
      type: String,
      required: true
    },
    email: {
      type: String,
      required: true
    },
    avatar: {
      type: String,
      default: ''
    }
  },
  data() {
    return {
      isExpanded: false
    }
  },
  methods: {
    toggleExpanded() {
      this.isExpanded = !this.isExpanded;
    }
  }
}
</script>

<style scoped>
.user-profile {
  border: 1px solid #ddd;
  padding: 1rem;
  border-radius: 8px;
}
</style>`

	componentPath := filepath.Join(tempDir, "UserProfile.vue")
	err := os.WriteFile(componentPath, []byte(componentContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create component file: %v", err)
	}

	// Test component discovery
	components, err := discoverUIComponents(context.Background(), tempDir, "user", "vue")
	if err != nil {
		t.Fatalf("discoverUIComponents returned error: %v", err)
	}

	if components.Framework != "vue" {
		t.Errorf("Expected framework 'vue', got '%s'", components.Framework)
	}

	if len(components.Components) == 0 {
		t.Errorf("Expected to find components, but found none")
	}

	if len(components.Components) > 0 {
		component := components.Components[0]
		if component.Name != "UserProfile" {
			t.Errorf("Expected component name 'UserProfile', got '%s'", component.Name)
		}

		// Check props extraction
		if len(component.Props) < 3 {
			t.Errorf("Expected to find at least 3 props, but found %d", len(component.Props))
		}

		// Check state extraction
		if len(component.State) == 0 {
			t.Errorf("Expected to find state, but found none")
		}

		// Check methods extraction
		if len(component.Methods) == 0 {
			t.Errorf("Expected to find methods, but found none")
		}
	}
}

// TestDiscoverUIComponents_Angular tests Angular component discovery
func TestDiscoverUIComponents_Angular(t *testing.T) {
	tempDir := t.TempDir()

	// Create an Angular component file
	componentContent := `import { Component, Input, OnInit } from '@angular/core';

@Component({
  selector: 'app-product-card',
  templateUrl: './product-card.component.html',
  styleUrls: ['./product-card.component.scss']
})
export class ProductCardComponent implements OnInit {
  @Input() product: Product;
  @Input() showDetails = false;

  isExpanded = false;
  currentPrice: number;

  constructor() { }

  ngOnInit(): void {
    this.currentPrice = this.product.price;
  }

  toggleDetails(): void {
    this.isExpanded = !this.isExpanded;
  }

  updatePrice(newPrice: number): void {
    this.currentPrice = newPrice;
  }
}

interface Product {
  id: number;
  name: string;
  price: number;
  description: string;
}`

	componentPath := filepath.Join(tempDir, "product-card.component.ts")
	err := os.WriteFile(componentPath, []byte(componentContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create component file: %v", err)
	}

	// Test component discovery
	components, err := discoverUIComponents(context.Background(), tempDir, "product", "angular")
	if err != nil {
		t.Fatalf("discoverUIComponents returned error: %v", err)
	}

	if components.Framework != "angular" {
		t.Errorf("Expected framework 'angular', got '%s'", components.Framework)
	}

	if len(components.Components) == 0 {
		t.Errorf("Expected to find components, but found none")
	}

	if len(components.Components) > 0 {
		component := components.Components[0]
		if component.Name != "product-card" {
			t.Errorf("Expected component name 'product-card', got '%s'", component.Name)
		}

		// Check metadata extraction
		if component.Metadata["selector"] != "app-product-card" {
			t.Errorf("Expected selector 'app-product-card', got '%s'", component.Metadata["selector"])
		}

		// Check methods extraction
		if len(component.Methods) < 2 {
			t.Errorf("Expected to find at least 2 methods, but found %d", len(component.Methods))
		}
	}
}

// TestDetectStylingFrameworks tests styling framework detection
func TestDetectStylingFrameworks(t *testing.T) {
	tempDir := t.TempDir()

	// Create package.json with styling dependencies
	packageJSON := `{
  "name": "test-app",
  "dependencies": {
    "react": "^18.0.0",
    "tailwindcss": "^3.2.0",
    "styled-components": "^5.3.0",
    "sass": "^1.55.0"
  }
}`

	packagePath := filepath.Join(tempDir, "package.json")
	err := os.WriteFile(packagePath, []byte(packageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create a CSS file with Tailwind classes
	cssContent := `.container {
  @apply flex items-center justify-center;
}

.btn {
  @apply bg-blue-500 text-white px-4 py-2 rounded;
}`

	cssPath := filepath.Join(tempDir, "styles.css")
	err = os.WriteFile(cssPath, []byte(cssContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create CSS file: %v", err)
	}

	// Test styling framework detection
	frameworks := detectStylingFrameworks(tempDir)

	if len(frameworks) == 0 {
		t.Errorf("Expected to find styling frameworks, but found none")
	}

	// Check for expected frameworks
	foundTailwind := false
	foundStyledComponents := false
	foundSass := false

	for _, fw := range frameworks {
		switch fw.Name {
		case "Tailwind CSS":
			foundTailwind = true
			if fw.Type != "utility" {
				t.Errorf("Expected Tailwind type 'utility', got '%s'", fw.Type)
			}
		case "styled-components":
			foundStyledComponents = true
			if fw.Type != "css-in-js" {
				t.Errorf("Expected styled-components type 'css-in-js', got '%s'", fw.Type)
			}
		case "SCSS/Sass":
			foundSass = true
			if fw.Type != "scss" {
				t.Errorf("Expected SCSS type 'scss', got '%s'", fw.Type)
			}
		}
	}

	if !foundTailwind {
		t.Errorf("Expected to find Tailwind CSS framework")
	}
	if !foundStyledComponents {
		t.Errorf("Expected to find styled-components framework")
	}
	if !foundSass {
		t.Errorf("Expected to find SCSS/Sass framework")
	}
}

// TestDiscoverUIComponents_NoComponents tests case with no matching components
func TestDiscoverUIComponents_NoComponents(t *testing.T) {
	tempDir := t.TempDir()

	// Create a non-component file
	randomFile := filepath.Join(tempDir, "utils.js")
	utilsContent := `export const formatDate = (date) => {
  return date.toISOString();
};`

	err := os.WriteFile(randomFile, []byte(utilsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create utils file: %v", err)
	}

	// Test component discovery with non-matching feature name
	components, err := discoverUIComponents(context.Background(), tempDir, "nonexistent", "react")
	if err != nil {
		t.Fatalf("discoverUIComponents returned error: %v", err)
	}

	if len(components.Components) != 0 {
		t.Errorf("Expected to find no components, but found %d", len(components.Components))
	}
}
