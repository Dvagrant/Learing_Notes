由于2PC的关系，redo log 中事务会有两种状态，`prepare`和`commit`

## 2PC不同时刻异常重启
写入 redo log 处于`prepare`阶段之后、写 binlog 之前，发生了崩溃（crash），由于此时 binlog 还没写，redo log 也还没提交，所以崩溃恢复的时候，这个事务会回滚。这时候，binlog 还没写，所以也不会传到备库。

binlog 写完，redo log 还没 commit 前发生 crash，那崩溃恢复的时候 MySQL 会怎么处理？
- 如果 redo log 里面的事务是完整的，也就是已经有了 commit 标识，则直接提交
- 如果 redo log 里面的事务只有完整的 prepare，则判断对应的事务 binlog 是否存在并完整：
		1. 如果是，则提交事务；
		2. 否则，回滚事务。

>一个事务的 binlog 是有完整格式的：
>- statement 格式的 binlog，最后会有 COMMIT；
>- row 格式的 binlog，最后会有一个 XID event。
>
>另外，在 MySQL 5.6.2 版本以后，还引入了 binlog-checksum 参数，用来验证 binlog 内容的正确性。

#### redo log 和 binlog 的关联
它们有一个共同的数据字段，叫 XID。崩溃恢复的时候，会按顺序扫描 redo log：
- 如果碰到既有 prepare、又有 commit 的 redo log，就直接提交
- 如果碰到只有 parepare、而没有 commit 的 redo log，就拿着 XID 去 binlog 找对应的事务

---
#### 为什么不只用 binlog 来支持崩溃恢复
历史：binlog 是 mysql server 层的，redo log 是 InnoDB 引擎层的。
现实：只用 binlog 不能实现数据页级别的恢复。
binlog 并没有记录数据页中的改动，只记录了 sql 语句。

由于采用 WAL 技术，所以崩溃后必须有日志文件（记录了详细改动）来恢复脏页。即必须有 redo log