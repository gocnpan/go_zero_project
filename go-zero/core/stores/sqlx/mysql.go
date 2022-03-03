package sqlx

import (
	"github.com/go-sql-driver/mysql"
)

const (
	mysqlDriverName           = "mysql"
	duplicateEntryCode uint16 = 1062
)

// NewMysql returns a mysql connection.
func NewMysql(datasource string, opts ...SqlOption) SqlConn {
	opts = append(opts, withMysqlAcceptable()) // accept 判断
	// 返回
	//	commonSqlConn struct {
	//		connProv connProvider  // db连接，在 connManager.GetResource 中，通过 manager.singleFlight 确保同一连接创建一次
	//		onError  func(error)
	//		beginTx  beginnable // 创建事务对象 Tx
	//		brk      breaker.Breaker //
	//		accept   func(error) bool // 表明 sql 语句是否执行正确
	//	}
	return NewSqlConn(mysqlDriverName, datasource, opts...)
}

// mysqlAcceptable
// mysql 语句执行正确与否
func mysqlAcceptable(err error) bool {
	if err == nil {
		return true
	}

	myerr, ok := err.(*mysql.MySQLError)
	if !ok {
		return false
	}

	switch myerr.Number {
	case duplicateEntryCode: // 向唯一字段插入相同值
		return true
	default:
		return false
	}
}

func withMysqlAcceptable() SqlOption {
	return func(conn *commonSqlConn) {
		conn.accept = mysqlAcceptable
	}
}
