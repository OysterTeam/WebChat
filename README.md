# WebChat

基于Golang的聊天后端服务器。

使用到的第三方库：
```
github.com/gorilla/websocket
github.com/mattn/go-sqlite3
github.com/dgrijalva/jwt-go
```

目前已支持的功能：注册、登录、点对点聊天。

## 文档

### 建表
这里采用`sqlite`，文件放在`db`文件夹下，如需更改请跳转到`server.go`的`NewSqliteDB("db/my.db")`语句。

若希望自己建表，可以参照以下语句，否则不需要执行，可以跳到下一步。

```sql
CREATE TABLE `User`
(
    `U_ID`             INTEGER PRIMARY KEY AUTOINCREMENT,
    `U_Nickname`       varchar(45)  DEFAULT NULL,
    `U_Gender`         bit          DEFAULT NULL,
    `U_Email`          varchar(45)  DEFAULT NULL,
    `U_Telephone`      varchar(45)  DEFAULT NULL,
    `U_AvatarUrl`      varchar(500) DEFAULT NULL,
    `U_SignUpTime`     TIMESTAMP    DEFAULT (datetime(CURRENT_TIMESTAMP, 'localtime')),
    `U_LastOnlineTime` datetime     DEFAULT NULL
);
CREATE TABLE `UserPwd`
(
    `U_ID`   INTEGER PRIMARY KEY AUTOINCREMENT,
    `U_Pwd`  VARCHAR(20) DEFAULT NULL,
    `U_Salt` VARCHAR(20) DEFAULT NULL,
    FOREIGN KEY (`U_ID`) REFERENCES User (`U_ID`)
);
```
想要让用户的id从`300000`开始自增，则：
```sql
UPDATE sqlite_sequence set seq=300000 where name='User';
UPDATE sqlite_sequence set seq=300000 where name='UserPwd'
```

### build and run

由于用到了go-sqlite3，需要gcc才能正常build。

```shell
go mod tidy
go run run_server.go -addr 7001
```

可以去掉`-addr 7001`，默认端口是7001。

### 注册

服务器启动后，可以向`localhost:7001/signup`发送POST请求，例如：

```json
{
  "nick_name":"testuser1",
  "email":"test1@email.com",
  "gender":"1",
  "pwd_raw":"123456"
}
```

其中，`nick_name`、`email`和`pwd_raw`字段是必填项。

`nick_name`代表昵称，不是唯一的，可以重复。

`email`可以用来登录，唯一，不可以重复。服务端会校验email的格式是否正确，以及是否存在MX。

`gender`代表性别，"1"代表男性，"2"代表女性。

`pwd_raw`代表未加密的密码，最少6位。由于HTTPS证书问题，此处暂未设计加密，后续再更新。

如果请求成功，会返回：

```json
{
  "http_response_code": 201,
  "bool_status": true,
  "response_msg": "成功: 已创建用户，UserID请见ResponseData字段。",
  "response_data": 300001
}
```

### 登录
先注册两个账号，用于模拟，再向`localhost:7001/signin`发送POST请求，例如：
```json
{
    "email": "test1@email.com",
    "pwd_raw": "123456"
}
```
登录支持使用`email`和`user_id`字段，也可以这样发送：
```json
{
    "user_id": "300002",
    "pwd_raw": "123456"
}
```
但不能同时填写，以避免冲突。

如果密码正确，会返回：

```json
{
    "http_response_code": 201,
    "bool_status": true,
    "response_msg": "密码匹配状态:密码匹配。若匹配，ResponseData字段为token。",
    "response_data": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NjkyNzQ5NTksImlzcyI6InNlcnZlciIsInN1YiI6IjMwMDAwMSJ9.kucmgw2JHmTVBe4khJliYne9H4xylwlTrL3fiqYzgGM"
}
```

拿到token之后，可以凭据此`token`建立`WebSocket`连接。

### 建立WebSocket连接

向`ws://localhost:7001/ws`发送http请求升级`WebSocket`协议，将先前拿到的`token`附在`Authorization`字段中。

注意，token的有效期只有60秒。如需更改时间，请到`token.go`文件更改：

```go
var ExpiresSecond int64 = 60
```

### 发送消息

建立ws连接后，可以发送消息，举例：

```json
{
    "msg_id":1,
    "msg_code":2001,
    "msg_to":300002,
    "msg_to_group":false,
    "msg_content":"aGVsbG8="
}
```

`msg_id`代表消息id，具体使用暂未设计。
`msg_code`代表消息类型，具体设计见`message.go`的`const`
`msg_to`代表消息发送给哪个用户，此处只能填写`user_id`
`msg_to_group`代表是否是群组消息。群组功能暂未设计。
`msg_content`代表消息内容，采用`base64`编码。

现在发送消息还没有落库，所以如果对方不在线，会发送失败。

现在还没有好友设计，后续会添加只有好友才能发送消息的功能。

如果接收到消息，服务段会主动发送json。例如`user_id=300002`的用户登录时，会接收到：
```json
{
    "msg_id": 1,
    "msg_code": 2001,
    "msg_from": 300001,
    "msg_to": 300002,
    "msg_to_group": false,
    "msg_content": "aGVsbG8="
}
```