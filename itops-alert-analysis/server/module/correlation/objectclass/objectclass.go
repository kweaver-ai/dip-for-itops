package objectclass

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/config"
	"github.com/pkg/errors"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/cache"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/dip"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/infra/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/utils"
	"github.com/spf13/cast"
)

const (
	// 缓存键前缀
	cacheKeyPrefix = "objectclass:hostname:" // hostname -> EntityObjectInfo JSON
	// 分页大小
	pageLimit = 1000
	// 缓存过期时间：1小时
	cacheTTL = 1 * time.Hour
	// 刷新间隔：10分钟
	//TODO 可配置
	refreshInterval = 30 * time.Second
)

// EntityObjectInfo 实体对象信息
type EntityObjectInfo struct {
	ObjectTypeID string `json:"object_type_id"` // 对象类 ID
	ObjectID     string `json:"object_id"`      // 对象实例 ID
	Name         string `json:"name"`           // 对象名称
}

// ObjectClass 对象类缓存，负责维护 hostname -> object_type_id 的映射。
type ObjectClass struct {
	dipClient *dip.Client //dip客户端用于拉取 对象缓存
	cache     cache.Cache //具体的cache 实现
	mu        sync.RWMutex
}

// New 创建对象类缓存实例。
func New(cfg *config.Config, dipClient *dip.Client) (*ObjectClass, error) {
	// 初始化 Redis 缓存
	redisCache, err := cache.NewRedisCache(cache.RedisConfig{
		MasterName: cfg.DepServices.Redis.ConnectInfo.MasterGroupName,
		SentinelAddrs: []string{
			fmt.Sprintf("%s:%d", cfg.DepServices.Redis.ConnectInfo.SentinelHost, cfg.DepServices.Redis.ConnectInfo.SentinelPort),
		},
		SentinelUsername: cfg.DepServices.Redis.ConnectInfo.SentinelUsername,
		SentinelPassword: cfg.DepServices.Redis.ConnectInfo.SentinelPassword,

		Username: cfg.DepServices.Redis.ConnectInfo.Username,
		Password: cfg.DepServices.Redis.ConnectInfo.Password,
	})
	if err != nil {
		return nil, errors.Wrap(err, "初始化 Redis 缓存失败")
	}
	return &ObjectClass{
		dipClient: dipClient,
		cache:     redisCache,
	}, nil
}

// Run 在 errgroup 中运行缓存刷新器，初始预热后定期刷新。
func (c *ObjectClass) Run(ctx context.Context) error {
	// 初始预热
	if err := c.Warmup(ctx); err != nil {
		log.Warnf("对象类缓存初始预热失败: %v，服务将继续启动并在后续定时刷新中重试", err)
	} else {
		log.Info("对象类缓存预热完成")
	}

	log.Infof("启动对象类缓存定时刷新（间隔: %v）", refreshInterval)

	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("对象类缓存刷新器收到停止信号")
			return ctx.Err()
		case <-ticker.C:
			log.Info("开始定时刷新对象类缓存...")
			if err := c.Warmup(ctx); err != nil {
				log.Warnf("定时刷新失败: %v", err)
			} else {
				log.Info("定时刷新完成")
			}
		}
	}
}

// Warmup 预热缓存，拉取所有对象类及其对象数据，构建 hostname -> ot_id 映射。
func (c *ObjectClass) Warmup(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 第一步：获取所有对象类列表
	objectTypes, err := c.dipClient.GetObjectTypes(ctx)
	if err != nil {
		return errors.Wrap(err, "获取对象类列表失败")
	}

	log.Infof("获取到 %d 个对象类", len(objectTypes))

	// 第二步：遍历每个对象类，查询对象数据
	var failedCount int
	for _, ot := range objectTypes {
		if err := c.warmupObjectType(ctx, ot.ID); err != nil {
			failedCount++
			log.Warnf("预热对象类 %s 失败: %v", ot.ID, err)
			// 继续处理其他对象类
			continue
		}
	}

	if failedCount > 0 {
		log.Warnf("对象类缓存预热完成，共 %d 个对象类，其中 %d 个预热失败", len(objectTypes), failedCount)
	}

	return nil
}

// warmupObjectType 预热单个对象类的数据。
func (c *ObjectClass) warmupObjectType(ctx context.Context, otID string) error {
	// 第三步：查询对象类的所有对象数据（自动处理分页）
	objectData, err := c.dipClient.QueryAllObjectData(ctx, otID, pageLimit)
	if err != nil {
		return errors.Wrap(err, "查询对象数据失败")
	}

	log.Debugf("对象类 %s 有 %d 个对象实例", otID, len(objectData))

	// 第四步：遍历对象数据，提取字段并缓存
	for _, obj := range objectData {
		// 统一格式：k8s_cluster:xxx,namespace:xxx,name:xxx（不存在的字段使用空字符串）
		k8sCluster := cast.ToString(obj["k8s_cluster"])
		namespace := cast.ToString(obj["namespace"])
		name := cast.ToString(obj["name"])
		cacheKeySuffix := "k8s_cluster:" + k8sCluster + ",namespace:" + namespace + ",name:" + name

		// 提取 id 字段（对象实例 ID）
		objIDStr := cast.ToString(obj["s_id"])
		if objIDStr == "" {
			continue
		}

		// 构建对象信息
		objInfo := EntityObjectInfo{
			ObjectTypeID: otID,
			ObjectID:     objIDStr,
			Name:         name,
		}

		// 序列化为 JSON
		jsonData := utils.JsonEncode(objInfo)

		// 缓存 hostname -> EntityObjectInfo JSON（1小时过期）
		cacheKey := cacheKeyPrefix + cacheKeySuffix
		if err := c.cache.Set(ctx, cacheKey, jsonData, cacheTTL); err != nil {
			log.Warnf("缓存对象信息失败, key=%s, objectTypeID=%s, objectID=%s, 原因: %v", cacheKey, otID, objIDStr, err)
			// 继续处理其他数据
		}
	}

	return nil
}

// GetEntityObjectInfo 根据 hostname 查询对应的对象信息（包含 object_type_id、object_id、name）。
func (c *ObjectClass) GetEntityObjectInfo(ctx context.Context, hostname string) (*EntityObjectInfo, error) {
	if hostname == "" {
		return nil, errors.New("hostname 不能为空")
	}

	cacheKey := cacheKeyPrefix + hostname
	jsonData, err := c.cache.Get(ctx, cacheKey)
	if err != nil {
		// 缓存未命中
		return nil, errors.Errorf("未找到 hostname 对应的对象信息: %s", hostname)
	}

	var objInfo EntityObjectInfo
	if err := json.Unmarshal([]byte(jsonData), &objInfo); err != nil {
		return nil, errors.Wrap(err, "反序列化对象信息失败")
	}

	return &objInfo, nil
}

// Close 关闭缓存资源。
func (c *ObjectClass) Close() error {
	if c.cache != nil {
		return c.cache.Close()
	}
	return nil
}
