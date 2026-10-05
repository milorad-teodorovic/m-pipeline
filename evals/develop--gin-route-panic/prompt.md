---
max_turns: 160
timeout_seconds: 3600
allowed_tools: [Read, Glob, Grep, Skill, Agent, Bash, Write, Edit]
tags: [execution, develop, pipeline, real-repo]
---
/seamark:develop Resolve this issue filed against the gin repository in this directory. Add tests. Keep the existing tests. All requirements are settled. Work autonomously without questions. Do not commit.

Issue: Panic: slice bounds out of range in getValue with HandleMethodNotAllowed enabled

### Description

With `HandleMethodNotAllowed = true`, gin can panic during route matching:                                                                                                                                                                                                           
```
runtime error: slice bounds out of range [:x+1] with capacity x
```                                            
                                                                                                                            
`c.skippedNodes` is allocated once per pooled Context with capacity `engine.maxSections` (`gin.go:254`). `getValue` pushes onto it with a raw reslice i.e. `(*skippedNodes)[:index+1]` (`tree.go:435`), so it cannot grow. `Context.reset()` clears it once per request (`context.go:117`), but `handleHTTPRequest` calls `getValue` once per method tree in the `HandleMethodNotAllowed` loop (`gin.go:746`) reusing the same stack, and `getValue` neither resets it on entry nor drains it before returning. Residue accumulates per tree until the reslice exceeds capacity:
```
getValue call #1: len=0 cap=6   push -> 1, push -> 2
getValue call #2: len=2 cap=6   <-- residue
getValue call #3: len=4 cap=6   <-- residue
getValue call #4: len=6 cap=6   push -> 7  => panic
```                                                                                                                                         

Because the panic happens during route matching, *before* the handler chain, `gin.Recovery()` cannot catch it, `net/http` recovers per connection, logs `http: panic serving ...`, and closes the connection with no response, so clients see an empty reply/reset rather than a 5xx.

### Gin Version

v1.12.0

### Can you reproduce the bug?

Yes

### Source Code

gin v1.12.0, go 1.26.4

```go                                                                                                                                       
package main                                                                                                                                
                                                                                                                                            
import (                                                                                                                                    
      "net/http"                                                                                                                            
      "net/http/httptest"                                                                                                                   
      "testing"                                                                                                                             
                                                                                                                                            
      "github.com/gin-gonic/gin"                                                                                                            
)                                                                                                                                           
                                                                                                                                            
func TestPanic(t *testing.T) {                                                                                                              
      gin.SetMode(gin.ReleaseMode)                                                                                                          
      router := gin.New()                                                                                                                   
      router.HandleMethodNotAllowed = true                                                                                                  
                                                                                                                                            
      h := func(c *gin.Context) {}                                                                                                          
      router.OPTIONS("/:p0/:p1/a/:p2", h)                                                                                                   
      router.GET("/:p0/:p1/a/:p2", h)                                                                                                       
      router.PATCH("/b/:p0/:p1/c", h)                                                                                                       
      router.DELETE("/b/:p0/:p1/d/:p3", h)                                                                                                  
      router.GET("/b/:p0/:p1/e/f", h)                                                                                                       
      router.POST("/b/:p0/:p1/g/:p4/h", h)                                                                                                  
      router.OPTIONS("/b/:p0/:p1/g/:p4/h", h)                                                                                               
      router.DELETE("/b/cache", h)                                                                                                          
      router.GET("/b/clients/:p1/g", h)                                                                                                     
      router.POST("/b/clients/:p1/g", h)                                                                                                    
      router.PATCH("/b/clients/:p1/g/:p4", h)                                                                                               
      router.OPTIONS("/b/clients/:p1/g/:p4", h)
                                                                                                                                            
      req := httptest.NewRequest(http.MethodPost, "/b/clients/42", nil)                                                                     
      router.ServeHTTP(httptest.NewRecorder(), req)                                                                                         
}                                                                                                                                           
```

## Expectations                                                                                                                                                                                                                                                                           
`405 Method Not Allowed`.

## Actual result                                                                                                                  
```
panic: runtime error: slice bounds out of range [:7] with capacity 6                                                                        
github.com/gin-gonic/gin.(*node).getValue(...)                                                                                              
      gin@v1.12.0/tree.go:435                                                                                                               
github.com/gin-gonic/gin.(*Engine).handleHTTPRequest(...)                                                                                   
      gin@v1.12.0/gin.go:746                                                                                                                
github.com/gin-gonic/gin.(*Engine).ServeHTTP(...)                                                                                           
      gin@v1.12.0/gin.go:672                                                                                                                
```

### Go Version

go 1.26.4

### Operating System

_No response_
