# 天河应用中心底层API开发（kubernetes部分）

## kubernetes子系统功能需求

kubernetes将作为独立的子系统（或微服务）为上层提供功能调用，具体需要实现的功能列表如下：

1. 实现基于LDAP的用户账号认证，供上层绑定kubernetes子系统用户账号时使用；
2. 实现基于JWT的认证功能，保证kubernetes子系统的安全；
3. 实现用户账号的添加、查询、更新和删除，供上层管理kubernees子系统用户账号时使用；
4. 实现用户组的添加、查询、更新和删除，供上层管理kubernees子系统用户组时使用；
5. 为上层提供创建、查询、更新和删除应用模板的接口；
6. 为上层提供创建、查询、更新和删除应用的接口；
7. 为上层提供设置、查询和修改用户资源配额的接口；
8. 为上层提供管理kubernetes子系统计算节点的接口；
9. 其他kubernetes子系统内部需要实现的功能。

## 目前已实现功能

1. 实现基于LDAP的用户账号认证；
2. 实现基于JWT的认证功能；
3. 设计完成基于Mysql的内部数据库；
4. 实现用户账号的添加、查询、更新和删除，具体包括LDAP和Mysql数据库相关信息的管理；
5. 实现用户组的添加、查询、更新和删除，具体包括LDAP和Mysql数据库相关信息的管理；
6. 实现创建、查询、更新和删除应用模板的接口；
7. 实现创建、查询、更新和删除应用的接口；
8. 实现带账号验证的二层反向代理功能；


## 测试代码

1. 启动代理服务: go run proxy/proxy.go
2. 启动kubernetes子系统服务: go run server/server.go
3. 测试ldap账号认证: go run test/test_ldapAuth.go
4. 测试JWT认证: go run test/test_jwt.go
5. 测试SSHA功能: go run test/test_ssha.go
6. 测试用户账号管理功能：go run test/test_user.go
7. 测试用户组管理功能：go run test/test_group.go
8. 测试kubernetes接口功能：go run test/test_k8s.go
9. 测试应用模板管理功能：go run test/test_template.go
10. 测试应用管理功能：go run test/test_app.go
