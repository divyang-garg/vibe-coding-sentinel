# Phase 6: AST Analysis Integration

## Overview

This document describes the planned enhancement to use Abstract Syntax Tree (AST) analysis for more accurate function extraction in test requirement generation and related code analysis tasks.

## Current Implementation

### Pattern Matching Approach

The current implementation uses pattern matching (regex-based) to extract function names from code:

- **Location**: `hub/api/services/test_requirement_helpers.go:42` and `hub/api/test_requirement_generator.go:325`
- **Method**: `extractFunctionNameFromCode(code, keyword string)`
- **Approach**: Uses regex patterns to find function definitions containing keywords

### Limitations

1. **Accuracy**: Pattern matching can produce false positives and miss complex function structures
2. **Nested Functions**: Cannot accurately handle nested functions or closures
3. **Language Support**: Limited to simple patterns, doesn't leverage language-specific parsing
4. **Context Awareness**: Lacks understanding of code structure and context
5. **Edge Cases**: Struggles with:
   - Arrow functions
   - Method definitions in classes
   - Function expressions
   - Async/await patterns
   - Decorators and annotations

## Planned AST-Based Implementation

### Benefits

1. **Accuracy**: AST parsing provides precise understanding of code structure
2. **Language Support**: Can support multiple languages with appropriate parsers
3. **Context Awareness**: Understands relationships between functions, classes, and modules
4. **Robustness**: Handles edge cases and complex code structures
5. **Extensibility**: Easy to extend for additional analysis capabilities

### Implementation Plan

#### Phase 6.1: Parser Selection

- **JavaScript/TypeScript**: Use `@babel/parser` or `typescript` compiler API
- **Python**: Use `ast` module from Python standard library
- **Go**: Use `go/parser` and `go/ast` packages
- **Java**: Use `javaparser` or similar

#### Phase 6.2: AST Extraction

Create a unified AST extraction interface:

```go
type ASTExtractor interface {
    ExtractFunctions(code string, keyword string) ([]FunctionInfo, error)
    ExtractClasses(code string) ([]ClassInfo, error)
    ExtractImports(code string) ([]ImportInfo, error)
}
```

#### Phase 6.3: Function Analysis

Enhance function extraction to include:
- Function signatures (parameters, return types)
- Function visibility (public/private)
- Function decorators/annotations
- Function documentation
- Function complexity metrics

#### Phase 6.4: Integration

Replace pattern matching with AST-based extraction:
- Update `extractFunctionNameFromCode` to use AST extractors
- Maintain backward compatibility during transition
- Add fallback to pattern matching if AST parsing fails

### Estimated Effort

- **Phase 6.1**: 2-3 days (parser selection and setup)
- **Phase 6.2**: 3-4 days (AST extraction interface and implementations)
- **Phase 6.3**: 2-3 days (function analysis enhancements)
- **Phase 6.4**: 1-2 days (integration and testing)

**Total**: 8-12 days

### Dependencies

- Language-specific parser libraries
- AST manipulation utilities
- Testing framework for multi-language support
- Documentation updates

### Success Criteria

1. AST-based extraction achieves >95% accuracy vs. current ~70% accuracy
2. Support for at least 3 languages (JavaScript/TypeScript, Python, Go)
3. Backward compatibility maintained
4. Performance impact <10% compared to pattern matching
5. Comprehensive test coverage (>90%)

## Migration Strategy

1. **Parallel Implementation**: Implement AST extraction alongside pattern matching
2. **Feature Flag**: Use feature flag to toggle between implementations
3. **Gradual Rollout**: Enable AST extraction for specific languages first
4. **Monitoring**: Track accuracy and performance metrics
5. **Full Migration**: Once validated, remove pattern matching fallback

## Related Files

- `hub/api/services/test_requirement_helpers.go:42`
- `hub/api/test_requirement_generator.go:325`
- Future: `hub/api/services/ast_extractor.go` (to be created)
- Future: `hub/api/services/ast_parsers/` (language-specific parsers)

## References

- [Babel Parser Documentation](https://babeljs.io/docs/en/babel-parser)
- [Go AST Package](https://pkg.go.dev/go/ast)
- [Python AST Module](https://docs.python.org/3/library/ast.html)
