package sqlx

import (
	"database/sql"
	"io"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/syncx"
)

const (
	maxIdleConns = 64
	maxOpenConns = 64
	maxLifetime  = time.Minute
)

var connManager = syncx.NewResourceManager()

type pingedDB struct {
	*sql.DB
	once sync.Once
}

func getCachedSqlConn(driverName, server string) (*pingedDB, error) {
	// driverName 是 mysqlDriverName -> "mysql"
	// server 是 资源链接 data source -> "root:123456@tcp(mysql:3306)/mall?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
	// GetResource 方法 实现一次创建，共享连接的功能，防止多次创建
	val, err := connManager.GetResource(server, func() (io.Closer, error) {
		conn, err := newDBConnection(driverName, server) // 创建 DB 连接，并进行相关配置
		if err != nil {
			return nil, err
		}

		return &pingedDB{
			DB: conn,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	return val.(*pingedDB), nil
}

func getSqlConn(driverName, server string) (*sql.DB, error) {
	// 获取可能已经创建的 db 连接
	pdb, err := getCachedSqlConn(driverName, server)
	if err != nil {
		return nil, err
	}

	pdb.once.Do(func() { // 测试能否连通
		err = pdb.Ping()
	})
	if err != nil {
		return nil, err
	}

	return pdb.DB, nil
}

func newDBConnection(driverName, datasource string) (*sql.DB, error) {
	conn, err := sql.Open(driverName, datasource) // 创建 DB 连接
	if err != nil {
		return nil, err
	}

	// we need to do this until the issue https://github.com/golang/go/issues/9851 get fixed
	// discussed here https://github.com/go-sql-driver/mysql/issues/257
	// if the discussed SetMaxIdleTimeout methods added, we'll change this behavior
	// 8 means we can't have more than 8 goroutines to concurrently access the same database.
	conn.SetMaxIdleConns(maxIdleConns)
	conn.SetMaxOpenConns(maxOpenConns)
	conn.SetConnMaxLifetime(maxLifetime)

	return conn, nil
}
