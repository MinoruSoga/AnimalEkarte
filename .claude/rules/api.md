---
paths: "backend/**/*.go"
---

# API Development Rules (Go/Gin)

## Endpoint Design

- Use RESTful conventions
- Include input validation on all endpoints
- Return consistent error response format
- Use proper HTTP status codes

## Handler Pattern

```go
func (h *Handler) GetResource(c *gin.Context) {
    id := c.Param("id")

    result, err := h.service.FindByID(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, result)
}
```

## Error Response Format

```json
{
  "error": "Human readable message"
}
```

## Security

- Validate all user input
- Use parameterized queries (prevent SQL injection)
- Implement rate limiting
- Sanitize output (prevent XSS)
- Use context for request-scoped values

## Documentation

- Add Swagger comments for API documentation
- Document all request/response schemas
- Use meaningful error messages
