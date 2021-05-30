# BlockChain-Learning
## SHA256

### 预处理

假设要计算的字符串的**二进制**长度为l，首先在串末尾加上1，再加上k位0，需要满足再补上64位就是512的倍数，即：
```
l + 1 + k === -64 mod 512
```
再补上64位的原字符串长度l的二进制表示

例如对于字符串`abc`，其二进制表示为`01100001 01100010 01100011`，长度为`l = 24`

计算得`k = 512 - 24 - 1 - 64 = 423`

`24`的64为二进制表示为`11000`

故对于`abc`而言，得到的结果为：
```go
01100001011000100110001110000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000011000
```
然后还需要对得到的字符串按照每512位进行分块，`abc`很短，就只有一块

预处理的步骤使用go写成如下
```go
func prepare(s []byte) [][]byte {
    // append a bytes started with bit '1'
    tmp := append(s, 0x80)
    l := len(s)
    // zero-padding
    if l % 64 < 56 {
        tmp = append(tmp, make([]byte, 55 - l % 64)...)
    } else {
        tmp = append(tmp, make([]byte, 121 - l % 64)...)
    }
    // raw data length
    tmp2 := make([]byte, 8)
    binary.BigEndian.PutUint64(tmp2, uint64(l<<3))
    tmp = append(tmp, tmp2...)
    // chunk
    var prepared [][]byte
    for i := 0; i < len(tmp) / 64; i++ {
        prepared = append(prepared, tmp[i * 64 : i * 64 + 64])
    }
    return prepared
}
```

### 计算
对每一块的计算过程如下(512位 64字节):

分割成若干小块，每小块是32位(4字节)。所以现在我们有16个小块。在后面补上48个为0的小块。一共有64个小块，用`w[0...63]`表示

后面就是计算，定义了若干常量，涉及大量位运算

这篇文章把计算步骤写得很清楚[https://qvault.io/cryptography/how-sha-2-works-step-by-step-sha-256/](https://qvault.io/cryptography/how-sha-2-works-step-by-step-sha-256/)

完整脚本在sha256.go

## Merkle Tree

原理挺简单的。将n条数据的哈希值(例如使用上面的SHA256)作为叶子节点，两两组合并再次计算哈希值并将该值作为这两个节点的父节点，如此往复，直到计算根节点。

在p2p文件传输完整性校验的应用中比较好理解。校验时，同样计算出获取到的n条数据的Merkle Tree。首先比较根节点，根节点包含所有数据的缩略信息，因此可以判断出整体是否发生错误。若根节点值不一致，再比较左子树和右子树，如此往复。最终能定位到是哪一段数据的问题，做到在log(n)的复杂度下对n条数据进行完整性校验。

[https://github.com/cbergoon/merkletree](https://github.com/cbergoon/merkletree) 用go实现了Merkle Tree

## bitcoin

An electronic coin is a chain of digital signatures. 每次交易时，拥有者用自己的私钥对上次交易的信息和下一位拥有者的公钥进行签名。因此可以用拥有者的公钥对此次交易进行验证。

每个区块包含若干个交易，使用Merkle Tree将区块中的交易进行哈希，获取最终的Merkle root值，这样做既能节省空间，又能快速索引每一笔交易。

区块头包含Merkle root，时间戳，前一个区块的哈希值，当前难度(target)等等。区块头中还有个nonce字段，定义为uint32，随意取值，只要它使得对区块头计算出hash值小于难度target，就是一个能被承认的区块。

通过对难度进行调整使得尽量每10分钟出一个区块，每2016个区块就会进行一次难度调整，如果过去这些区块平均时间小于十分钟，就会增加难度，反之减小难度。

对成功计算出区块的用户，会奖励若干BTC，当前的奖励是6.25BTC。每210000个区块会减半，如上所说会动态调整近似每10分钟一个区块，那么210000个区块大概需要4年时间，也就是每四年奖励减半。矿工另一个收益来自于交易手续费，即交易输入的BTC数量减去需要输出的BTC数量。

一方面矿工为了利润最大化，一定会优先选择手续费高的交易，另一方面，每个区块有最大size限制，低手续费或无手续费的交易可能无法被矿工打包计算。

## Ethereum

### 地址生成

参考[https://ethbook.abyteahead.com/index.html](https://ethbook.abyteahead.com/index.html) 和《智能合约安全分析和审计指南》

先生成随机私钥，在计算出对应公钥及地址

keccak库似乎有内存泄漏的问题

```javascript
'use strict'

const randomBytes = require('randombytes')
const secp256k1 = require('secp256k1')
const keccak = require('keccak')

const createRandomPrivateKey = function () {
    return randomBytes(32)
}

const privateKeyToAddress = function (privateKey) {
    return keccak('keccak256').update(Buffer.from(secp256k1.publicKeyCreate(privateKey, false).slice(1))).digest().slice(-20)
}

let privateKey = createRandomPrivateKey()
let address = privateKeyToAddress(privateKey)
console.log('0x'+address.toString('hex'))
console.log(privateKey.toString('hex'))
```

### geth

具有账户管理、网络、合约等功能。

#### create private network

创建目录结构

```
ether-test/
├── db
└── gensis.json
```

配置gensis.json

`chainId`是私链的网络地址，不同的链无法互联通讯。`homesteadBlock`表示是否使用homesteadBlock版本的eth协议。
`eip150Block`,`eip155Block`和`eip158Block`是两个以太坊分叉提议，表示是否需要支持其相应标准。

```json
{
  "config": {
    "chainId": 987,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0
  },
  "difficulty": "0x400",
  "gasLimit": "0xffffffff",
  "alloc": {}
}
```

配置gensis.json(上面这个配置文件不兼容较新的硬分叉，比如19年的更新新增了一些EVM指令，后面编译合约的时候要选择旧些的版本)是为了创世块的产生：

```shell
geth --datadir "./db" init gensis.json
```

一些启动参数：
- `--rpc` `--rpcaddr` `rpcport`  是否开启JSON-RPC调用调试功能及地址端口
- `--nodiscover`  避免别人加入
- `--port`  监听节点之间P2P消息的端口，默认是30303
- `--mine -–etherbase`  是否挖矿，相当于console执行`miner.start()`。后面一个参数是挖矿接收奖励地址

```shell
geth --datadir "./db" --nodiscover console
```

在console中有一些内置对象，可以进行如下操作：

```shell
# personal 账号管理相关
personal.newAccount('123')  # 生成账号
personal.unlockAccount(eth.accounts[0], '123', 300)  # 解锁账号，第三个参数是解锁时间


# eth 区块链操作相关
eth.accounts()  # 返回账号数组
eth.getBalance(eth.coinbase)  # 余额
eth.getBlock(1)  # 查询区块信息

eth.sendTransaction({from:eth.accounts[0], to:eth.accounts[1], value:web3.toWei(10, 'ether')})  # 进行交易


# miner
miner.start()  # 开始挖矿
miner.stop()  # 停止挖矿

# txtool  查看交易池相关状态
txtool.status()  # 

# web3 包含上面的对象，还有一些单位换算的方法
web3.fromWei(1, "ether")  # 10 ^18 wei = 1 Ether
```

如果没有把输出流放到文件，执行`miner.start()`之后，会有源源不断的输出，无法继续调试输入命令。
或者运行节点时，没有使用console，
可以用命名管道attach上去再继续调试。在windows中的命令为：

```shell
geth attach ipc://./pipe/geth.ipc
```

### contract

一个简单的例子

```solidity
pragma solidity ^0.4.24;
contract test1 {
    uint vaultData;
    function set(uint data) public{
        vaultData = data;
    }

    function get() public view returns (uint) {
        return vaultData;
    }
}
```

在[https://remix.ethereum.org](https://remix.ethereum.org)中连接到本地私链并部署合约