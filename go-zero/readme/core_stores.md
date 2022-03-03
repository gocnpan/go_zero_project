文件夹：[core/store](../core/stores)
# 简介
数据缓存、持久化相关

## sqlx
`mysql`持久化

### 关于 `mysql transaction` 事务
`database/sql`提供了事务处理的功能。通过`Tx`对象实现。`db.Begin`会创建tx对象，后者的`Exec`和`Query`执行事务的数据库操作，最后在`tx`的`Commit`和`Rollback`中完成数据库事务的提交和回滚，同时释放连接。

`tx`事务环境中，只有一个数据库连接，事务内的`Eexc`都是**依次执行**的，事务中也可以使用`db`进行查询，但是`db`查询的过程会新建连接，这个连接的操作**不属于该事务**。

### `mysql.go`
初始化一个`mysql`连接

主方法`NewMysql`：
1. 增加 accept 判断
2. 创建 `sql connect`
    ```go
    type commonSqlConn struct {
        // db连接，在 connManager.GetResource 中，通过 manager.singleFlight 确保同一连接创建一次
        connProv connProvider  
        onError  func(error)
        beginTx  beginnable // 创建事务对象 Tx
        brk      breaker.Breaker //
        accept   func(error) bool // 表明 sql 语句是否执行正确
    }
    ```


## sqlc
`redis`缓存