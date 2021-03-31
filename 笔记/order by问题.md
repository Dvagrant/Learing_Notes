MySQL 会给每个线程分配一块内存用于排序，称为`sort_buffer`

`max_length_for_sort_data`用于控制排序中单行数据的最长长度，当单行数据长度小于此值时，使用全字段排序，当单行数据大于此值时，使用 rowID 排序

## 全字段排序
`sort_buffer_size`为排序内存`sort_buffer`的大小，当需要排序的数据量小于此值时，排序在内存中完成，如果内存放不下，只能利用磁盘临时文件。

> 打开optimizer_trace，只对本线程有效 
>`SET optimizer_trace='enabled=on'`
>查看内容 
>`SELECT * FROM 'information_schema'.'OPTIMIZER_TRACE'\G`


![[全排序的optimizer_trace结果.png]]
全排序的optimizer_trace结果，其中
`number_of_tmp_files`代表排序中使用到的临时文件数目
`examined_rows`代表用来排序的行数
`sort_mode`中的`packed_additional_fields`代表对VARCHAR 类型的数据做了紧凑处理

***查询 `OPTIMIZER_TRACE`这个表时，需要用到临时表，而 `internal_tmp_disk_storage_engine`的默认值是 InnoDB。如果使用的是 InnoDB 引擎的话，把数据从临时表取出来的时候，会让 InnoDB引擎的读取值加 1。所以应当将`internal_tmp_disk_storage_engine`设置成 MyISAM***

---
## rowID 排序
当使用 rowID 排序时，在内存中排序好结果后，直接使用主键索引回表查询记录并返回给客户端，无需继续在服务端打包成一整个结果返回

>`select VARIABLE_VALUE into @a from performance_schema.session_status where variable_name = 'Innodb_rows_read'`
>可以使用此语句，将引擎读出的行数做记录

此时的读取行会加上回表的行数。
并且 `OPTIMEZER_TRACE`表中 `sort_mode`字段会变成`<sort_key,rowID>`

---
## 排序的优化
- 可以通过索引的天然有序性避免排序。
- 可以使用覆盖索引的效果，将读取行数小的查询直接返回数据。