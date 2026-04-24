# Handler Layer — P5 / P6 / P7 / P12 / P14 / P15 / P18

## P7: toXxxResponse() wrapping (MANDATORY)

```go
// ✅
c.JSON(http.StatusOK, toVaccineResponse(vaccine))
c.JSON(http.StatusOK, toVaccineListResponse(vaccines))

// ❌ Direct model or gin.H
c.JSON(http.StatusOK, vaccine)
c.JSON(http.StatusOK, gin.H{"data": vaccine})
```

## P12: ShouldBindJSON error handling (MANDATORY)

```go
// ✅
var req CreateVaccineRequest
if err := c.ShouldBindJSON(&req); err != nil {
    RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
    return
}

// ❌
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
```

## P14: No direct Repository injection (MANDATORY)

```go
// ✅
type VaccineHandler struct {
    svc service.VaccineService
}

// ❌
type VaccineHandler struct {
    svc  service.VaccineService
    repo repository.VaccineRepository  // 直接注入禁止
}
```

## P15: POST returns 201 + Location header (MANDATORY)

```go
// ✅ master routes
c.Header("Location", fmt.Sprintf("/v1/masters/vaccines/%d", vaccine.ID))
c.JSON(http.StatusCreated, toVaccineResponse(vaccine))

// ✅ business routes
c.Header("Location", fmt.Sprintf("/api/v1/owners/%d", owner.ID))
c.JSON(http.StatusCreated, toOwnerResponse(owner))

// ❌
c.JSON(http.StatusOK, toVaccineResponse(vaccine))    // 200 is wrong
c.JSON(http.StatusCreated, toVaccineResponse(vaccine)) // no Location
```

## P18: toXxxResponse function naming (MANDATORY)

```go
// ✅
func toVaccineResponse(v *model.Vaccine) VaccineResponse { ... }
func toVaccineListResponse(vs []*model.Vaccine) []VaccineResponse { ... }

// ❌ Wrong prefixes
func convertToVaccine(...)     // convert
func buildVaccineResponse(...) // build
func mapVaccine(...)           // map
func newVaccineResponse(...)   // new
```

## P5: RequirePermission on write routes (MANDATORY)

```go
// ✅ in RegisterXxxRoutes
masters.POST("/vaccines", RequirePermission("edit"), h.Create)
masters.PUT("/vaccines/:id", RequirePermission("edit"), h.Update)
masters.DELETE("/vaccines/:id", RequirePermission("delete"), h.Delete)

// Exemptions (no RequirePermission):
// /login, /logout, /auth/*, /me, LIFF public APIs

// ❌
masters.POST("/vaccines", h.Create)  // missing RequirePermission
```

## P6: DELETE uses "delete" permission, not "edit" (MANDATORY)

```go
// ✅
masters.DELETE("/vaccines/:id", RequirePermission("delete"), h.Delete)

// ❌
masters.DELETE("/vaccines/:id", RequirePermission("edit"), h.Delete)
```
