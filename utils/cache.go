package utils

import (
	"sync"
	"time"

	cache "github.com/patrickmn/go-cache"
)

const (
	ClusterCachePrefix          = "cluster-cache-"
	TokenCachePrefix            = "token-cache-"
	UserCachePrefix             = "user-cache-"
	AlertIndicatorCache         = "alert-indicator"
	PromAlertsCache             = "prom-alerts"
	AlertSilenceCache           = "alert-silence"
	AlertTenantClusterCache     = "alert-tenant-clusters"
	AlertClustersCache          = "alert-clusters"
	AlertMessagesCache          = "alert-message"
	AlertClusterCache           = "alert-cluster"
	AlertNodeCache              = "alert-node"
	AlertAppCache               = "alert-app"
	LicAllClusterVcpusCache     = "lic-all-cluster-vcpus"
	LicProductCache             = "lic-product-cache"
	ClusterPrometheusCacheKey   = "ClusterPrometheus__"
	ClusterAlertmanagerCacheKey = "ClusterAlertmanager__"
	ClusterSkywalkingCacheKey   = "ClusterSkywalking__"
)

var (
	// 统计缓存个数,分别是cache和redis
	// 有过期时间的cache
	localCacheWithTTL *cache.Cache
	once              sync.Once
)

func NewCacheCliWithTTL() *cache.Cache {
	once.Do(func() {
		localCacheWithTTL = cache.New(10*time.Minute, 10*time.Minute)
	})
	return localCacheWithTTL
}

type ValueWrapper struct {
	Value     interface{}
	CacheTime time.Time
	Duration  time.Duration
}
