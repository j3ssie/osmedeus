package distributed

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/redis/rueidis"
)

// Redis key prefixes
const (
	KeyPrefix           = "osm:"
	KeyTasksPending     = KeyPrefix + "tasks:pending"
	KeyTasksRunning     = KeyPrefix + "tasks:running"
	KeyTasksCompleted   = KeyPrefix + "tasks:completed"
	KeyWorkers          = KeyPrefix + "workers"
	KeyWorkersHeartbeat = KeyPrefix + "workers:heartbeat"
	KeyMasterLock       = KeyPrefix + "master:lock"
)

// Timeouts and intervals
const (
	HeartbeatInterval     = 30 * time.Second
	HeartbeatTimeout      = 90 * time.Second // 3 missed heartbeats
	TaskPollTimeout       = 5 * time.Second
	DefaultConnectTimeout = 60 * time.Second
)

// Client wraps a rueidis client with helper methods
type Client struct {
	client rueidis.Client
	cfg    *config.RedisConfig
}

// NewClient creates a new Redis client from configuration
func NewClient(cfg *config.RedisConfig) (*Client, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("redis host not configured")
	}

	port := cfg.Port
	if port == 0 {
		port = 6379
	}

	opts := rueidis.ClientOption{
		InitAddress:  []string{fmt.Sprintf("%s:%d", cfg.Host, port)},
		Username:     cfg.Username,
		Password:     cfg.Password,
		SelectDB:     cfg.DB,
		DisableCache: true, // Disable client-side caching for simpler behavior
	}

	client, err := rueidis.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	return &Client{
		client: client,
		cfg:    cfg,
	}, nil
}

// NewClientFromConfig creates a client from the global config
func NewClientFromConfig(cfg *config.Config) (*Client, error) {
	return NewClient(&cfg.Redis)
}

// ParseRedisURL parses a Redis connection URL into RedisConfig
// Format: redis://[username:password@]host:port[/db]
func ParseRedisURL(redisURL string) (*config.RedisConfig, error) {
	if !strings.HasPrefix(redisURL, "redis://") {
		redisURL = "redis://" + redisURL
	}

	u, err := url.Parse(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}

	cfg := &config.RedisConfig{
		Host:              u.Hostname(),
		Port:              6379,
		ConnectionTimeout: 60,
	}

	if u.Port() != "" {
		port, err := strconv.Atoi(u.Port())
		if err != nil {
			return nil, fmt.Errorf("invalid redis port: %w", err)
		}
		cfg.Port = port
	}

	if u.User != nil {
		cfg.Username = u.User.Username()
		cfg.Password, _ = u.User.Password()
	}

	if u.Path != "" && u.Path != "/" {
		db, err := strconv.Atoi(strings.TrimPrefix(u.Path, "/"))
		if err == nil {
			cfg.DB = db
		}
	}

	return cfg, nil
}

// Close closes the Redis client
func (c *Client) Close() {
	c.client.Close()
}

// Ping tests the Redis connection
func (c *Client) Ping(ctx context.Context) error {
	cmd := c.client.B().Ping().Build()
	return c.client.Do(ctx, cmd).Error()
}

// Raw returns the underlying rueidis client
func (c *Client) Raw() rueidis.Client {
	return c.client
}

// PushTask pushes a task to the pending queue
func (c *Client) PushTask(ctx context.Context, task *Task) error {
	data, err := task.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	cmd := c.client.B().Lpush().Key(KeyTasksPending).Element(string(data)).Build()
	return c.client.Do(ctx, cmd).Error()
}

// PopTask pops a task from the pending queue (blocking)
func (c *Client) PopTask(ctx context.Context, timeout time.Duration) (*Task, error) {
	cmd := c.client.B().Brpop().Key(KeyTasksPending).Timeout(timeout.Seconds()).Build()
	result, err := c.client.Do(ctx, cmd).AsStrSlice()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return nil, nil // Timeout, no task available
		}
		return nil, fmt.Errorf("failed to pop task: %w", err)
	}

	if len(result) < 2 {
		return nil, nil // No task
	}

	return UnmarshalTask([]byte(result[1]))
}

// SetTaskRunning moves a task to the running hash
func (c *Client) SetTaskRunning(ctx context.Context, task *Task) error {
	data, err := task.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	cmd := c.client.B().Hset().Key(KeyTasksRunning).FieldValue().FieldValue(task.ID, string(data)).Build()
	return c.client.Do(ctx, cmd).Error()
}

// RemoveTaskRunning removes a task from the running hash
func (c *Client) RemoveTaskRunning(ctx context.Context, taskID string) error {
	cmd := c.client.B().Hdel().Key(KeyTasksRunning).Field(taskID).Build()
	return c.client.Do(ctx, cmd).Error()
}

// SetTaskResult stores a task result in the completed hash
func (c *Client) SetTaskResult(ctx context.Context, result *TaskResult) error {
	data, err := result.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	cmd := c.client.B().Hset().Key(KeyTasksCompleted).FieldValue().FieldValue(result.TaskID, string(data)).Build()
	return c.client.Do(ctx, cmd).Error()
}

// GetTaskResult retrieves a task result from the completed hash
func (c *Client) GetTaskResult(ctx context.Context, taskID string) (*TaskResult, error) {
	cmd := c.client.B().Hget().Key(KeyTasksCompleted).Field(taskID).Build()
	data, err := c.client.Do(ctx, cmd).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get task result: %w", err)
	}

	return UnmarshalTaskResult([]byte(data))
}

// GetRunningTask retrieves a running task by ID
func (c *Client) GetRunningTask(ctx context.Context, taskID string) (*Task, error) {
	cmd := c.client.B().Hget().Key(KeyTasksRunning).Field(taskID).Build()
	data, err := c.client.Do(ctx, cmd).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get running task: %w", err)
	}

	return UnmarshalTask([]byte(data))
}

// GetAllRunningTasks retrieves all running tasks
func (c *Client) GetAllRunningTasks(ctx context.Context) ([]*Task, error) {
	cmd := c.client.B().Hgetall().Key(KeyTasksRunning).Build()
	result, err := c.client.Do(ctx, cmd).AsStrMap()
	if err != nil {
		return nil, fmt.Errorf("failed to get running tasks: %w", err)
	}

	var tasks []*Task
	for _, data := range result {
		task, err := UnmarshalTask([]byte(data))
		if err != nil {
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// RegisterWorker registers a worker in the workers hash
func (c *Client) RegisterWorker(ctx context.Context, worker *WorkerInfo) error {
	data, err := worker.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal worker: %w", err)
	}

	cmd := c.client.B().Hset().Key(KeyWorkers).FieldValue().FieldValue(worker.ID, string(data)).Build()
	return c.client.Do(ctx, cmd).Error()
}

// UpdateWorkerHeartbeat updates a worker's heartbeat timestamp
func (c *Client) UpdateWorkerHeartbeat(ctx context.Context, workerID string) error {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	cmd := c.client.B().Hset().Key(KeyWorkersHeartbeat).FieldValue().FieldValue(workerID, timestamp).Build()
	return c.client.Do(ctx, cmd).Error()
}

// GetWorkerHeartbeat gets a worker's last heartbeat timestamp
func (c *Client) GetWorkerHeartbeat(ctx context.Context, workerID string) (time.Time, error) {
	cmd := c.client.B().Hget().Key(KeyWorkersHeartbeat).Field(workerID).Build()
	data, err := c.client.Do(ctx, cmd).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}

	ts, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(ts, 0), nil
}

// GetAllWorkers retrieves all registered workers
func (c *Client) GetAllWorkers(ctx context.Context) ([]*WorkerInfo, error) {
	cmd := c.client.B().Hgetall().Key(KeyWorkers).Build()
	result, err := c.client.Do(ctx, cmd).AsStrMap()
	if err != nil {
		return nil, fmt.Errorf("failed to get workers: %w", err)
	}

	var workers []*WorkerInfo
	for _, data := range result {
		worker, err := UnmarshalWorkerInfo([]byte(data))
		if err != nil {
			continue
		}
		workers = append(workers, worker)
	}

	return workers, nil
}

// RemoveWorker removes a worker from the registry
func (c *Client) RemoveWorker(ctx context.Context, workerID string) error {
	// Remove from both workers and heartbeat hashes
	cmd1 := c.client.B().Hdel().Key(KeyWorkers).Field(workerID).Build()
	cmd2 := c.client.B().Hdel().Key(KeyWorkersHeartbeat).Field(workerID).Build()

	if err := c.client.Do(ctx, cmd1).Error(); err != nil {
		return err
	}
	return c.client.Do(ctx, cmd2).Error()
}

// AcquireMasterLock tries to acquire the master lock
func (c *Client) AcquireMasterLock(ctx context.Context, masterID string, ttl time.Duration) (bool, error) {
	cmd := c.client.B().Set().Key(KeyMasterLock).Value(masterID).Nx().Ex(ttl).Build()
	result, err := c.client.Do(ctx, cmd).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return false, nil // Lock not acquired
		}
		return false, err
	}
	return result == "OK", nil
}

// RefreshMasterLock refreshes the master lock TTL
func (c *Client) RefreshMasterLock(ctx context.Context, masterID string, ttl time.Duration) error {
	// Only refresh if we still own the lock
	cmd := c.client.B().Get().Key(KeyMasterLock).Build()
	current, err := c.client.Do(ctx, cmd).ToString()
	if err != nil {
		return err
	}
	if current != masterID {
		return fmt.Errorf("master lock lost")
	}

	expireCmd := c.client.B().Expire().Key(KeyMasterLock).Seconds(int64(ttl.Seconds())).Build()
	return c.client.Do(ctx, expireCmd).Error()
}

// ReleaseMasterLock releases the master lock
func (c *Client) ReleaseMasterLock(ctx context.Context, masterID string) error {
	// Only release if we own the lock
	cmd := c.client.B().Get().Key(KeyMasterLock).Build()
	current, err := c.client.Do(ctx, cmd).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return nil // Already released
		}
		return err
	}
	if current != masterID {
		return nil // Not our lock
	}

	delCmd := c.client.B().Del().Key(KeyMasterLock).Build()
	return c.client.Do(ctx, delCmd).Error()
}
