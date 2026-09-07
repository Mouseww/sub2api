package middleware

import (
	"bufio"
	_ "embed"
	"log"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	iputil "github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

//go:embed data/china_cidr.txt
var chinaCIDRData string

// chinaNets holds the compiled China IP networks, loaded once at package init.
var chinaNets = compileChinaNets()

func compileChinaNets() []*net.IPNet {
	var nets []*net.IPNet
	scanner := bufio.NewScanner(strings.NewReader(chinaCIDRData))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		_, network, err := net.ParseCIDR(line)
		if err != nil {
			log.Printf("geo_block: invalid CIDR %q: %v", line, err)
			continue
		}
		nets = append(nets, network)
	}
	return nets
}

// geoBlockCache caches the runtime-configurable settings (enabled flag + whitelist).
type geoBlockCache struct {
	enabled   bool
	whitelist []*net.IPNet
	loadedAt  time.Time
}

const geoBlockCacheTTL = 30 * time.Second

// GeoBlock returns a Gin middleware that blocks mainland China IP requests when
// the GeoBlockEnabled setting is true.
//
// Settings are fetched from settingService and cached for geoBlockCacheTTL.
// Whitelisted IPs/CIDRs bypass the geo-block check.
func GeoBlock(settingService *service.SettingService) gin.HandlerFunc {
	var cache atomic.Pointer[geoBlockCache]

	loadCache := func(c *gin.Context) *geoBlockCache {
		if p := cache.Load(); p != nil && time.Since(p.loadedAt) < geoBlockCacheTTL {
			return p
		}
		settings, err := settingService.GetAllSettings(c.Request.Context())
		if err != nil {
			// 加载失败时沿用旧缓存，若无缓存则默认关闭
			if p := cache.Load(); p != nil {
				return p
			}
			return &geoBlockCache{}
		}
		p := &geoBlockCache{
			enabled:   settings.GeoBlockEnabled,
			whitelist: parseNetworks(settings.GeoBlockWhitelist),
			loadedAt:  time.Now(),
		}
		cache.Store(p)
		return p
	}

	return func(c *gin.Context) {
		cfg := loadCache(c)
		if !cfg.enabled {
			c.Next()
			return
		}

		clientIP := iputil.GetTrustedClientIP(c)

		// 白名单直通
		if matchesNetworks(clientIP, cfg.whitelist) {
			c.Next()
			return
		}

		// 命中中国大陆 CIDR → 拒绝
		if matchesNetworks(clientIP, chinaNets) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied: this service is not available in your region.",
				"code":  "GEO_BLOCK",
			})
			return
		}

		c.Next()
	}
}

// parseNetworks 将字符串切片解析为 *net.IPNet 列表，支持单IP和CIDR格式。
func parseNetworks(entries []string) []*net.IPNet {
	var nets []*net.IPNet
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		// 尝试解析为CIDR
		if strings.Contains(entry, "/") {
			_, network, err := net.ParseCIDR(entry)
			if err == nil {
				nets = append(nets, network)
			}
			continue
		}
		// 单IP：转为 /32 或 /128
		ip := net.ParseIP(entry)
		if ip == nil {
			continue
		}
		bits := 32
		if ip.To4() == nil {
			bits = 128
		}
		nets = append(nets, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
	}
	return nets
}

// matchesNetworks 检查 IP 字符串是否命中任一网络段。
func matchesNetworks(ipStr string, nets []*net.IPNet) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	for _, network := range nets {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
