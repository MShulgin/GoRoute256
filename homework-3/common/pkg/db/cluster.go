package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/logger"
	"go.etcd.io/etcd/client/v3"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	dialTimeout = 2 * time.Second
)

type PgCluster struct {
	etcd   *clientv3.Client
	name   string
	stopCh chan bool
	Shards *Shards
}

func NewPgCluster(ctx context.Context, etcdServers []string, clusterName string) (*PgCluster, error) {
	etcd, err := clientv3.New(clientv3.Config{Endpoints: etcdServers, DialTimeout: dialTimeout})
	if err != nil {
		panic(err)
	}
	kv := clientv3.NewKV(etcd)
	initKey, err := kv.Get(ctx, clusterName)
	if err != nil {
		panic(err)
	}
	if len(initKey.Kvs) <= 0 {
		panic(errors.New("etcd cluster config not found"))
	}
	value := initKey.Kvs[0]
	var clusterInfo ClusterInfo
	err = json.Unmarshal(value.Value, &clusterInfo)
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("Using PG cluster config:\n%s", clusterInfo))
	shards, err := initShards(clusterInfo.Servers, clusterInfo.Buckets)
	if err != nil {
		panic(err)
	}

	cluster := PgCluster{
		etcd:   etcd,
		name:   clusterName,
		stopCh: make(chan bool),
		Shards: shards,
	}

	go watchConfigUpdates(&cluster)

	return &cluster, nil
}

func watchConfigUpdates(cluster *PgCluster) {
	watcher := cluster.etcd.Watch(context.TODO(), cluster.name)
	for {
		select {
		case watchResp := <-watcher:
			fmt.Println(watchResp)
			for _, event := range watchResp.Events {
				if event.Type != clientv3.EventTypePut {
					continue
				}
				value := event.Kv.Value
				var clusterInfo ClusterInfo
				err := json.Unmarshal(value, &clusterInfo)
				if err != nil {
					panic(err)
				}
				fmt.Println(clusterInfo)
				err = cluster.Shards.ReconfigShards(clusterInfo.Servers, clusterInfo.Buckets)
				if err != nil {
					fmt.Printf("failed to update cluster config, %e\n", err)
				}
			}
		case <-cluster.stopCh:
			fmt.Println("Stop watching cluster configuration")
			return
		}
	}
}

func (s *Shards) GetShard(key string) (*sqlx.DB, error) {
	s.RLock()
	defer s.RUnlock()

	bucketNum := hash(key) % s.NumServers
	server := s.Buckets[bucketNum]
	if conn, ok := s.Connections[server]; ok {
		return conn, nil
	}

	return nil, errors.New("no connection for bucket: " + strconv.FormatUint(uint64(bucketNum), 10))
}

func (s *Shards) ReconfigShards(connMap map[string]string, buckets map[uint32]string) error {
	s.Lock()
	defer s.Unlock()
	logger.Info("Update PG cluster configuration")

	newConn := make(map[string]*sqlx.DB)
	for server, connInfo := range connMap {
		conn, err := sqlx.Connect("pgx", connInfo)
		if err != nil {
			for _, c := range newConn {
				c.Close()
			}
			return err
		}
		newConn[server] = conn
	}
	for _, c := range s.Connections {
		c.Close()
	}
	s.Connections = newConn
	s.Buckets = buckets
	s.NumServers = uint32(len(newConn))

	logger.Info("Finish PG cluster reconfiguration")

	return nil
}

func (cluster *PgCluster) Close() error {
	cluster.stopCh <- true
	_ = cluster.etcd.Close()
	for _, conn := range cluster.Shards.Connections {
		_ = conn.Close()
	}
	return nil
}

func (cluster *PgCluster) GetShard(key string) (*sqlx.DB, error) {
	return cluster.Shards.GetShard(key)
}

func (cluster *PgCluster) QueryRow(dest any, query string, args ...any) error {
	var wg sync.WaitGroup
	wg.Add(int(cluster.Shards.NumServers))
	rows := make([]*sqlx.Row, cluster.Shards.NumServers)
	idx := 0
	for i := range cluster.Shards.Connections {
		db := cluster.Shards.Connections[i]
		go func(x int) {
			defer wg.Done()

			row := db.QueryRowx(query, args...)
			rows[x] = row
		}(idx)
		idx += 1
	}
	wg.Wait()

	rowsCount := 0
	for _, row := range rows {
		err := row.StructScan(dest)
		switch err {
		case nil:
			rowsCount += 1
		case sql.ErrNoRows:
			continue
		default:
			return err
		}
	}
	if rowsCount == 0 {
		return sql.ErrNoRows
	}
	if rowsCount > 1 {
		return errors.New(fmt.Sprintf("more than one rows found: query=%s", query))
	}

	return nil
}

func hash(t string) uint32 {
	hasher := fnv.New32()
	_, _ = hasher.Write([]byte(t))
	return hasher.Sum32()
}

type Shards struct {
	sync.RWMutex
	Connections map[string]*sqlx.DB
	Buckets     map[uint32]string
	NumServers  uint32
}

func initShards(connMap map[string]string, buckets map[uint32]string) (*Shards, error) {
	shards := Shards{
		RWMutex:     sync.RWMutex{},
		Connections: make(map[string]*sqlx.DB),
		Buckets:     make(map[uint32]string),
		NumServers:  0,
	}
	if err := shards.ReconfigShards(connMap, buckets); err != nil {
		return nil, err
	}

	return &shards, nil
}

type ClusterInfo struct {
	Buckets map[uint32]string `json:"buckets"`
	Servers map[string]string `json:"servers"`
}

func (info ClusterInfo) String() string {
	bucks := make(map[string][]uint32)
	for b, s := range info.Buckets {
		if bucks[s] == nil {
			bucks[s] = make([]uint32, 0)
		}
		bucks[s] = append(bucks[s], b)
	}
	keys := make([]string, 0)
	for k := range bucks {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, server := range keys {
		b := bucks[server]
		sort.Slice(b, func(i, j int) bool {
			return b[i] < b[j]
		})
		sb.WriteString(fmt.Sprintf("%s ---> buckets:%d\n", server, b))
	}
	return sb.String()
}
