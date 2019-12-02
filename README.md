## mainchain

POC+Ethereum mainchain.

A project fork from [go-ethereum](https://github.com/ethereum/go-ethereum).

---------


#### Compile #####
```bash
XXXXXXX/$ git clone https://github.com/XXXXXXX.git
XXXXXXX/$ cd XXXXXXX
XXXXXXX/$ #make poc
XXXXXXX/$ #make poc-linux-amd64
XXXXXXX/$ make poc
build/env.sh go run build/ci.go install ./cmd/poc
>>> /usr/local/go/bin/go install -gcflags="all=-N -l" -ldflags="-linkmode=internal -s" -v ./cmd/poc
XXXXXXX/mainchain/cmd/poc
Done building.
Run "XXXXXXX/mainchain/build/bin/poc" to launch poc.
```

#### Deploy #####
######1.初始网关节点部署

修改./load.sh中的相关变量为实际值，执行`./load.sh start`

```
#CMD='./load.sh'
#STOPEDFILE=.stoped.flag
#BIN=./bin/poc
#GATEWAY=gateway.inner.com
#IP=192.168.0.1
XXXXXXX/$ ./load.sh start
```
注：
>1. 可通过`./load.sh install`等命令安装自检定时器，进程异常退出后可以自动拉起。
>2. 初始节点如果作为其他节点的启动节点，需要公网IP且关闭相应端口的防火墙拦截。 

######2.网络节点部署
方法一：

编译时修改：`params/bootnodes.go`文件：

```
var MainnetBootnodes = []string{
}
var TestnetBootnodes = []string{
}
```

方法二：

设置启动参数:`--bootnodes`

方法三：

启动节点后，通过控制台添加初始节点：

```
./bin/poc attach poce.ipc
>web3.admin.addPeer("enode://...@[::]:30303")
```

#### 重要参数调整 #####
##### 1. 启动节点
编译前修改文件：`params/bootnodes.go`

#### 2. 挖矿总数
编译前修改文件：`consensus/poc/mortgage/mortgage.go`

```
MortgageSystemMaxReward     // 挖矿总数
MortgageOneBlockFullReward  // 最大出块奖励
```

#### 3. 预挖数据
编译前修改文件：`core/genesis_alloc.go`
```
pocNetAllocData             //预挖矿数据
注：
1. go run mkalloc.go xxx.json来生成上述预挖数据
2. keystore.poc.genesis/create_genesis_account.sh 可以生成随机账户
3. xxx.json格式参考keystore.poc.genesis.json，根据生成的账户进行配置
```



