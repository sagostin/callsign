# FreeSWITCH mod_xml_curl with Go Backend

> Architecture guide for using mod_xml_curl with a Go backend to serve dynamic dialplan, directory, and configuration to FreeSWITCH.

---

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [FreeSWITCH Configuration](#freeswitch-configuration)
3. [Go Backend Structure](#go-backend-structure)
4. [HTTP Endpoints](#http-endpoints)
5. [Request/Response Format](#requestresponse-format)
6. [Handler Implementations](#handler-implementations)
7. [Caching Strategy](#caching-strategy)
8. [Performance Considerations](#performance-considerations)

---

## Architecture Overview

```
┌─────────────────┐         HTTP POST          ┌─────────────────┐
│   FreeSWITCH    │ ◄─────────────────────────►│   Go Backend    │
│  mod_xml_curl   │     XML Response           │   (callsign)    │
└─────────────────┘                            └────────┬────────┘
                                                        │
                                                        ▼
                                               ┌─────────────────┐
                                               │   PostgreSQL    │
                                               └─────────────────┘
```

### Benefits of This Approach

1. **Single Language** - All logic in Go, no Lua maintenance
2. **Full Tooling** - Use Go's testing, debugging, profiling
3. **Flexible Deployment** - Backend can run on separate server
4. **Easy Integration** - Same backend serves UI API and FreeSWITCH
5. **Better Observability** - Middleware for logging, metrics, tracing

---

## FreeSWITCH Configuration

### Install mod_xml_curl

```bash
# Already included in most FreeSWITCH installations
# Verify it's loaded
fs_cli -x "module_exists mod_xml_curl"
```

### xml_curl.conf.xml

```xml
<!-- /etc/freeswitch/autoload_configs/xml_curl.conf.xml -->
<configuration name="xml_curl.conf" description="cURL XML Gateway">
  <bindings>
    
    <!-- Configuration binding (sofia.conf, ivr.conf, etc.) -->
    <binding name="configuration">
      <param name="gateway-url" value="http://127.0.0.1:8080/freeswitch/configuration"/>
      <param name="method" value="POST"/>
      <param name="timeout" value="10"/>
      <param name="enable-cacert-check" value="false"/>
      <!-- Optional: Add auth header -->
      <param name="gateway-credentials" value="freeswitch:your-secret-key"/>
    </binding>
    
    <!-- Directory binding (user authentication) -->
    <binding name="directory">
      <param name="gateway-url" value="http://127.0.0.1:8080/freeswitch/directory"/>
      <param name="method" value="POST"/>
      <param name="timeout" value="5"/>
    </binding>
    
    <!-- Dialplan binding (call routing) -->
    <binding name="dialplan">
      <param name="gateway-url" value="http://127.0.0.1:8080/freeswitch/dialplan"/>
      <param name="method" value="POST"/>
      <param name="timeout" value="5"/>
    </binding>
    
  </bindings>
</configuration>
```

### modules.conf.xml

```xml
<!-- Ensure mod_xml_curl is loaded -->
<load module="mod_xml_curl"/>
```

---

## Go Backend Structure

```
callsign-backend/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── freeswitch/
│   │   ├── handler.go          # HTTP handlers
│   │   ├── configuration.go    # Sofia, IVR config generators
│   │   ├── dialplan.go         # Dialplan XML generator
│   │   ├── directory.go        # User auth/directory
│   │   └── xml.go              # XML helper functions
│   ├── models/
│   │   ├── extension.go
│   │   ├── dialplan.go
│   │   ├── sip_profile.go
│   │   └── ivr_menu.go
│   └── database/
│       └── postgres.go
├── pkg/
│   └── xmlbuilder/
│       └── builder.go          # XML construction utilities
└── go.mod
```

---

## HTTP Endpoints

### POST /freeswitch/configuration

Called when FreeSWITCH needs configuration (sofia.conf, ivr.conf, etc.)

### POST /freeswitch/directory

Called for:
- SIP REGISTER (authentication)
- User lookup when calling an extension
- Voicemail MWI checks

### POST /freeswitch/dialplan

Called when a call needs routing decisions.

---

## Request/Response Format

### Request Parameters (form-encoded POST)

FreeSWITCH sends these as `application/x-www-form-urlencoded`:

#### Configuration Request
```
section=configuration
tag_name=configuration
key_name=name
key_value=sofia.conf      // or ivr.conf, acl.conf, etc.
hostname=fs1.example.com
```

#### Directory Request
```
section=directory
tag_name=domain
key_name=name
key_value=example.com
user=1001
domain=example.com
purpose=network-list      // or blank for auth
action=sip_auth           // or user_call, message-count
sip_auth_username=1001
sip_auth_realm=example.com
```

#### Dialplan Request
```
section=dialplan
tag_name=context
key_name=name
key_value=example.com
Caller-Context=example.com
Caller-Destination-Number=1001
Caller-Caller-ID-Number=5551234567
Hunt-Context=example.com
variable_domain_uuid=abc-123
```

### Response Format

Return `Content-Type: text/xml` with valid FreeSWITCH XML:

```xml
<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<document type="freeswitch/xml">
  <section name="dialplan">
    <!-- content -->
  </section>
</document>
```

### Not Found Response

Return this to tell FreeSWITCH "I don't handle this":

```xml
<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<document type="freeswitch/xml">
  <section name="result">
    <result status="not found"/>
  </section>
</document>
```

---

## Handler Implementations

### Main Handler Router

```go
// internal/freeswitch/handler.go
package freeswitch

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type Handler struct {
    db     *database.DB
    cache  *cache.Cache
}

func NewHandler(db *database.DB, cache *cache.Cache) *Handler {
    return &Handler{db: db, cache: cache}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
    fs := r.Group("/freeswitch")
    {
        fs.POST("/configuration", h.HandleConfiguration)
        fs.POST("/directory", h.HandleDirectory)
        fs.POST("/dialplan", h.HandleDialplan)
    }
}

// HandleConfiguration routes to specific config handlers
func (h *Handler) HandleConfiguration(c *gin.Context) {
    keyValue := c.PostForm("key_value")
    
    var xml string
    var err error
    
    switch keyValue {
    case "sofia.conf":
        xml, err = h.generateSofiaConfig(c)
    case "ivr.conf":
        xml, err = h.generateIVRConfig(c)
    case "acl.conf":
        xml, err = h.generateACLConfig(c)
    case "conference.conf":
        xml, err = h.generateConferenceConfig(c)
    default:
        xml = notFoundXML()
    }
    
    if err != nil {
        c.String(http.StatusInternalServerError, errorXML(err))
        return
    }
    
    c.Header("Content-Type", "text/xml")
    c.String(http.StatusOK, xml)
}
```

### Sofia Configuration Handler

```go
// internal/freeswitch/configuration.go
package freeswitch

import (
    "bytes"
    "text/template"
)

const sofiaTemplate = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<document type="freeswitch/xml">
  <section name="configuration">
    <configuration name="sofia.conf" description="sofia Endpoint">
      <global_settings>
        {{range .GlobalSettings}}
        <param name="{{.Name}}" value="{{.Value}}"/>
        {{end}}
      </global_settings>
      <profiles>
        {{range .Profiles}}
        <profile name="{{.Name}}">
          <gateways>
            {{range .Gateways}}
            <gateway name="{{.UUID}}">
              <param name="username" value="{{.Username}}"/>
              <param name="password" value="{{.Password}}"/>
              <param name="proxy" value="{{.Proxy}}"/>
              <param name="register" value="{{.Register}}"/>
              <param name="expire-seconds" value="{{.ExpireSeconds}}"/>
              <param name="context" value="{{.Context}}"/>
            </gateway>
            {{end}}
          </gateways>
          <domains>
            <domain name="all" alias="true" parse="true"/>
          </domains>
          <settings>
            {{range .Settings}}
            <param name="{{.Name}}" value="{{.Value}}"/>
            {{end}}
          </settings>
        </profile>
        {{end}}
      </profiles>
    </configuration>
  </section>
</document>`

func (h *Handler) generateSofiaConfig(c *gin.Context) (string, error) {
    hostname := c.PostForm("hostname")
    
    // Check cache first
    cacheKey := "sofia.conf:" + hostname
    if cached, ok := h.cache.Get(cacheKey); ok {
        return cached.(string), nil
    }
    
    // Query global settings
    globalSettings, err := h.db.GetSofiaGlobalSettings()
    if err != nil {
        return "", err
    }
    
    // Query profiles with their settings and gateways
    profiles, err := h.db.GetSIPProfiles(hostname)
    if err != nil {
        return "", err
    }
    
    // Render template
    tmpl, _ := template.New("sofia").Parse(sofiaTemplate)
    var buf bytes.Buffer
    err = tmpl.Execute(&buf, map[string]interface{}{
        "GlobalSettings": globalSettings,
        "Profiles":       profiles,
    })
    if err != nil {
        return "", err
    }
    
    xml := buf.String()
    
    // Cache for 1 hour
    h.cache.Set(cacheKey, xml, time.Hour)
    
    return xml, nil
}
```

### Directory Handler (User Authentication)

```go
// internal/freeswitch/directory.go
package freeswitch

const directoryTemplate = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<document type="freeswitch/xml">
  <section name="directory">
    <domain name="{{.Domain}}">
      <params>
        <param name="dial-string" value="{presence_id=${dialed_user}@${dialed_domain}}${sofia_contact(*/${dialed_user}@${dialed_domain})}"/>
      </params>
      <groups>
        <group name="default">
          <users>
            <user id="{{.Extension}}"{{if .NumberAlias}} number-alias="{{.NumberAlias}}"{{end}}>
              <params>
                <param name="password" value="{{.Password}}"/>
                <param name="vm-enabled" value="{{.VMEnabled}}"/>
                {{if .VMMail}}<param name="vm-mailto" value="{{.VMMail}}"/>{{end}}
              </params>
              <variables>
                <variable name="domain_uuid" value="{{.DomainUUID}}"/>
                <variable name="extension_uuid" value="{{.ExtensionUUID}}"/>
                <variable name="user_context" value="{{.UserContext}}"/>
                <variable name="effective_caller_id_name" value="{{.CallerIDName}}"/>
                <variable name="effective_caller_id_number" value="{{.CallerIDNumber}}"/>
                <variable name="outbound_caller_id_name" value="{{.OutboundCIDName}}"/>
                <variable name="outbound_caller_id_number" value="{{.OutboundCIDNumber}}"/>
                <variable name="toll_allow" value="{{.TollAllow}}"/>
                <variable name="call_timeout" value="{{.CallTimeout}}"/>
                <variable name="accountcode" value="{{.Accountcode}}"/>
              </variables>
            </user>
          </users>
        </group>
      </groups>
    </domain>
  </section>
</document>`

func (h *Handler) HandleDirectory(c *gin.Context) {
    action := c.PostForm("action")
    user := c.PostForm("user")
    domain := c.PostForm("domain")
    
    // Handle different directory actions
    switch action {
    case "sip_auth", "user_call", "":
        h.handleUserAuth(c, user, domain)
    case "message-count":
        h.handleMessageCount(c, user, domain)
    case "group_call":
        h.handleGroupCall(c, c.PostForm("group"), domain)
    default:
        c.Header("Content-Type", "text/xml")
        c.String(http.StatusOK, notFoundXML())
    }
}

func (h *Handler) handleUserAuth(c *gin.Context, user, domain string) {
    // Get domain UUID
    domainUUID, err := h.db.GetDomainUUID(domain)
    if err != nil {
        c.String(http.StatusOK, notFoundXML())
        return
    }
    
    // Get extension by username or number_alias
    ext, err := h.db.GetExtension(domainUUID, user)
    if err != nil || ext == nil {
        c.String(http.StatusOK, notFoundXML())
        return
    }
    
    // Get voicemail settings
    vm, _ := h.db.GetVoicemail(domainUUID, ext.Extension)
    
    // Build response
    data := DirectoryData{
        Domain:           domain,
        DomainUUID:       domainUUID,
        Extension:        ext.Extension,
        ExtensionUUID:    ext.ExtensionUUID,
        NumberAlias:      ext.NumberAlias,
        Password:         ext.Password,
        UserContext:      ext.UserContext,
        CallerIDName:     ext.EffectiveCallerIDName,
        CallerIDNumber:   ext.EffectiveCallerIDNumber,
        OutboundCIDName:  ext.OutboundCallerIDName,
        OutboundCIDNumber: ext.OutboundCallerIDNumber,
        TollAllow:        ext.TollAllow,
        CallTimeout:      ext.CallTimeout,
        Accountcode:      ext.Accountcode,
        VMEnabled:        boolToString(vm != nil && vm.Enabled),
        VMMail:           vmMailTo(vm),
    }
    
    xml := renderTemplate(directoryTemplate, data)
    c.Header("Content-Type", "text/xml")
    c.String(http.StatusOK, xml)
}
```

### Dialplan Handler

```go
// internal/freeswitch/dialplan.go
package freeswitch

func (h *Handler) HandleDialplan(c *gin.Context) {
    context := c.PostForm("Caller-Context")
    destination := c.PostForm("Caller-Destination-Number")
    hostname := c.PostForm("hostname")
    
    // Check cache first
    cacheKey := fmt.Sprintf("dialplan:%s", context)
    if cached, ok := h.cache.Get(cacheKey); ok {
        c.Header("Content-Type", "text/xml")
        c.String(http.StatusOK, cached.(string))
        return
    }
    
    // Query dialplans from database
    // FusionPBX stores pre-generated XML in dialplan_xml column
    dialplans, err := h.db.GetDialplans(context, hostname)
    if err != nil {
        c.String(http.StatusOK, notFoundXML())
        return
    }
    
    // Build complete dialplan XML
    var xml bytes.Buffer
    xml.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
    xml.WriteString(`<document type="freeswitch/xml">`)
    xml.WriteString(`<section name="dialplan">`)
    xml.WriteString(fmt.Sprintf(`<context name="%s">`, escapeXML(context)))
    
    for _, dp := range dialplans {
        // Append pre-generated XML
        xml.WriteString(dp.DialplanXML)
    }
    
    xml.WriteString(`</context>`)
    xml.WriteString(`</section>`)
    xml.WriteString(`</document>`)
    
    result := xml.String()
    
    // Cache result
    h.cache.Set(cacheKey, result, 5*time.Minute)
    
    c.Header("Content-Type", "text/xml")
    c.String(http.StatusOK, result)
}
```

### IVR Configuration Handler

```go
// internal/freeswitch/ivr.go
package freeswitch

func (h *Handler) generateIVRConfig(c *gin.Context) (string, error) {
    menuUUID := c.PostForm("Menu-Name")
    
    // Check cache
    cacheKey := "ivr.conf:" + menuUUID
    if cached, ok := h.cache.Get(cacheKey); ok {
        return cached.(string), nil
    }
    
    // Get IVR menu
    menu, err := h.db.GetIVRMenu(menuUUID)
    if err != nil {
        return notFoundXML(), nil
    }
    
    // Get domain for recordings path
    domain, _ := h.db.GetDomain(menu.DomainUUID)
    
    // Get menu options
    options, err := h.db.GetIVRMenuOptions(menuUUID)
    if err != nil {
        return "", err
    }
    
    // Build XML
    var xml bytes.Buffer
    xml.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
    xml.WriteString(`<document type="freeswitch/xml">`)
    xml.WriteString(`<section name="configuration">`)
    xml.WriteString(`<configuration name="ivr.conf" description="IVR Menus">`)
    xml.WriteString(`<menus>`)
    
    xml.WriteString(fmt.Sprintf(`<menu name="%s" `+
        `greet-long="%s" `+
        `greet-short="%s" `+
        `invalid-sound="%s" `+
        `exit-sound="%s" `+
        `timeout="%d" `+
        `inter-digit-timeout="%d" `+
        `max-failures="%d" `+
        `max-timeouts="%d" `+
        `digit-len="%d">`,
        menu.UUID,
        menu.GreetLong,
        menu.GreetShort,
        menu.InvalidSound,
        menu.ExitSound,
        menu.Timeout,
        menu.InterDigitTimeout,
        menu.MaxFailures,
        menu.MaxTimeouts,
        menu.DigitLen,
    ))
    
    for _, opt := range options {
        xml.WriteString(fmt.Sprintf(
            `<entry action="%s" digits="%s" param="%s"/>`,
            opt.Action,
            opt.Digits,
            opt.Param,
        ))
    }
    
    xml.WriteString(`</menu>`)
    xml.WriteString(`</menus>`)
    xml.WriteString(`</configuration>`)
    xml.WriteString(`</section>`)
    xml.WriteString(`</document>`)
    
    result := xml.String()
    h.cache.Set(cacheKey, result, time.Hour)
    
    return result, nil
}
```

---

## Caching Strategy

### Cache Keys

| Key Pattern | Content | TTL |
|-------------|---------|-----|
| `sofia.conf:{hostname}` | Full Sofia config | 1 hour |
| `dialplan:{context}` | Dialplan for context | 5 minutes |
| `ivr.conf:{uuid}` | IVR menu config | 1 hour |
| `acl.conf` | Access control lists | 1 hour |

### Cache Invalidation

When config changes via API, clear relevant cache:

```go
// internal/api/handlers.go

func (h *Handler) UpdateSIPProfile(c *gin.Context) {
    // ... update database ...
    
    // Clear cache
    h.cache.Delete("sofia.conf:" + hostname)
    
    // Tell FreeSWITCH to reload
    h.esl.Send("api sofia profile internal restart")
}

func (h *Handler) UpdateDialplan(c *gin.Context) {
    // ... update database ...
    
    // Clear cache
    h.cache.Delete("dialplan:" + context)
    
    // Tell FreeSWITCH to reload XML
    h.esl.Send("api reloadxml")
}
```

### Using Redis for Caching

```go
// internal/cache/redis.go
package cache

import (
    "context"
    "time"
    "github.com/redis/go-redis/v9"
)

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(addr string) *RedisCache {
    return &RedisCache{
        client: redis.NewClient(&redis.Options{
            Addr: addr,
        }),
    }
}

func (c *RedisCache) Get(key string) (string, bool) {
    val, err := c.client.Get(context.Background(), key).Result()
    if err != nil {
        return "", false
    }
    return val, true
}

func (c *RedisCache) Set(key string, value string, ttl time.Duration) {
    c.client.Set(context.Background(), key, value, ttl)
}

func (c *RedisCache) Delete(key string) {
    c.client.Del(context.Background(), key)
}

func (c *RedisCache) DeletePattern(pattern string) {
    ctx := context.Background()
    keys, _ := c.client.Keys(ctx, pattern).Result()
    if len(keys) > 0 {
        c.client.Del(ctx, keys...)
    }
}
```

---

## Performance Considerations

### 1. Keep Handlers Fast

- Use connection pooling for PostgreSQL
- Cache heavily (dialplan, sofia.conf change rarely)
- Pre-generate dialplan XML when saving (like FusionPBX)

### 2. Connection Pooling

```go
// internal/database/postgres.go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

func NewDB(dsn string) (*sql.DB, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    
    // Connection pool settings
    db.SetMaxOpenConns(50)
    db.SetMaxIdleConns(10)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    return db, nil
}
```

### 3. Pre-generate Dialplan XML

When saving a dialplan in your API, generate the XML immediately:

```go
func (s *Service) SaveDialplan(dp *models.Dialplan) error {
    // Generate XML from details
    xml := s.generateDialplanXML(dp)
    dp.DialplanXML = xml
    
    // Save to database
    err := s.db.SaveDialplan(dp)
    if err != nil {
        return err
    }
    
    // Invalidate cache
    s.cache.Delete("dialplan:" + dp.Context)
    
    return nil
}

func (s *Service) generateDialplanXML(dp *models.Dialplan) string {
    var xml bytes.Buffer
    
    xml.WriteString(fmt.Sprintf(
        `<extension name="%s" continue="%s">`,
        dp.Name,
        boolToString(dp.Continue),
    ))
    
    // Group details by group number
    for _, detail := range dp.Details {
        switch detail.Tag {
        case "condition":
            xml.WriteString(fmt.Sprintf(
                `<condition field="%s" expression="%s">`,
                detail.Type,
                escapeXML(detail.Data),
            ))
        case "action":
            xml.WriteString(fmt.Sprintf(
                `<action application="%s" data="%s"/>`,
                detail.Type,
                escapeXML(detail.Data),
            ))
        case "anti-action":
            xml.WriteString(fmt.Sprintf(
                `<anti-action application="%s" data="%s"/>`,
                detail.Type,
                escapeXML(detail.Data),
            ))
        }
    }
    
    xml.WriteString(`</condition>`)
    xml.WriteString(`</extension>`)
    
    return xml.String()
}
```

### 4. Timeouts

Configure appropriate timeouts in FreeSWITCH:

```xml
<param name="timeout" value="5"/>  <!-- 5 seconds max -->
```

And in your Go server:

```go
server := &http.Server{
    Addr:         ":8080",
    Handler:      router,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 5 * time.Second,
}
```

---

## Summary

| Component | Database Tables | Cache Key |
|-----------|----------------|-----------|
| Sofia Config | `sip_profiles`, `sip_profile_settings`, `gateways` | `sofia.conf:{host}` |
| Directory | `extensions`, `voicemails` | (per-request) |
| Dialplan | `dialplans` (pre-generated XML) | `dialplan:{context}` |
| IVR | `ivr_menus`, `ivr_menu_options` | `ivr.conf:{uuid}` |
| ACL | `access_controls`, `access_control_nodes` | `acl.conf` |
