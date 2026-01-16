package main

import (
	"fmt"
	"strings"
)

func main() {
	line := "  author    User     @relation(fields: [authorId], references: [id], onDelete: Cascade)"
	
	fmt.Printf("Line: %q\n", line)
	fmt.Printf("Contains '  ': %v\n", strings.Contains(line, "  "))
	fmt.Printf("Contains '@relation': %v\n", strings.Contains(line, "@relation"))
	fmt.Printf("Starts with '@@': %v\n", strings.HasPrefix(line, "@@"))
	fmt.Printf("Starts with '}': %v\n", strings.HasPrefix(line, "}"))
	
	// Test the condition
	condition := strings.Contains(line, "  ") && !strings.HasPrefix(line, "@@") && !strings.HasPrefix(line, "}")
	fmt.Printf("Full condition: %v\n", condition)
	
	if condition {
		if strings.Contains(line, "@relation") {
			fmt.Println("Would process relationship")
			parts := strings.Fields(strings.TrimSpace(line))
			if len(parts) > 0 {
				fieldName := parts[0]
				fmt.Printf("Field name: %s\n", fieldName)
			}
		} else {
			fmt.Println("Would process column")
		}
	} else {
		fmt.Println("Condition not met")
	}
}
EOF && go run debug_prisma.go && rm debug_prisma.go